package solana

import (
	"blocsy/internal/types"
	"context"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/mr-tron/base58"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"log"
	"time"
)

type tokenAuth struct{ token string }

func (t tokenAuth) GetRequestMetadata(context.Context, ...string) (map[string]string, error) {
	return map[string]string{"x-token": t.token}, nil
}
func (t tokenAuth) RequireTransportSecurity() bool { return true }

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             20 * time.Second,
	PermitWithoutStream: true,
}

func NewBlockListener(grpc string, qHandler *QueueHandler, authToken string) *BlockListener {
	return &BlockListener{
		Client:       nil,
		Subscription: nil,
		grpcAddress:  grpc,
		queueHandler: qHandler,
		authToken:    authToken,
		pingId:       0,
	}
}

func (s *BlockListener) Listen() error {
	log.SetFlags(0)
	flag.Parse()

	for {
		conn, err := s.grpcConnect(s.grpcAddress)
		if err != nil {
			log.Printf("Failed to connect to %s. Retrying... | err: %v", s.grpcAddress, err)
			time.Sleep(5 * time.Second)
			continue
		}
		if conn == nil {
			log.Printf("Failed to connect to %s. Retrying...", s.grpcAddress)
			time.Sleep(5 * time.Second)
			continue
		}

		err = s.grpcSubscribe(conn)
		if err != nil {
			log.Printf("Error in grpcSubscribe: %v. Reconnecting...", err)
			conn.Close()
			time.Sleep(5 * time.Second)
			continue
		}
		defer conn.Close()
		break
	}

	return nil
}

func (s *BlockListener) grpcConnect(addr string) (*grpc.ClientConn, error) {
	pool, _ := x509.SystemCertPool()
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(credentials.NewClientTLSFromCert(pool, "")),
		grpc.WithKeepaliveParams(kacp),
		grpc.WithPerRPCCredentials(tokenAuth{token: s.authToken}),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(1024*1024*1024),
			grpc.UseCompressor(gzip.Name),
		),
	}
	return grpc.NewClient(addr, opts...)
}

func (s *BlockListener) grpcSubscribe(conn *grpc.ClientConn) error {
	var err error
	client := pb.NewGeyserClient(conn)
	s.Client = client

	subscription, err := s.prepareSubscription()
	if err != nil {
		return fmt.Errorf("failed to prepare subscription: %v", err)
	}

	ctx := context.Background()
	if s.authToken != "" {
		md := metadata.New(map[string]string{"x-token": s.authToken})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	stream, err := client.Subscribe(context.Background())
	if err != nil {
		return err
	}
	if err = stream.Send(subscription); err != nil {
		return err
	}

	s.Subscription = stream
	go s.keepAlive(ctx)
	log.Printf("Subscribed to %s", s.grpcAddress)
	var blockNumber uint64
	var blockTime int64

	for {

		upd, err := stream.Recv()
		if err != nil {
			log.Printf("Error in grpcSubscribe: %v. Reconnecting...", err)
			return err // reconnect outside
		}

		var capturedTS = time.Now().Unix()
		var solanaTx types.SolanaTx

		if tx := upd.GetTransaction(); tx != nil {
			blockNumber = tx.Slot

			var decodedSignatures []string
			for _, sig := range tx.Transaction.Transaction.Signatures {
				b58Sig := base58.Encode(sig)
				decodedSignatures = append(decodedSignatures, b58Sig)
			}

			var decodedAccountKeys []string
			for _, key := range tx.Transaction.Transaction.Message.AccountKeys {
				b58key := base58.Encode(key)
				decodedAccountKeys = append(decodedAccountKeys, b58key)
			}

			for i, atl := range tx.Transaction.Transaction.Message.AddressTableLookups {
				if i >= len(solanaTx.Transaction.Message.AddressTableLookups) {
					solanaTx.Transaction.Message.AddressTableLookups = append(solanaTx.Transaction.Message.AddressTableLookups, types.AddressTableLookup{})
				}
				solanaTx.Transaction.Message.AddressTableLookups[i] = types.AddressTableLookup{
					AccountKey:      base58.Encode(atl.AccountKey),
					WritableIndexes: convertToIntSlice(atl.WritableIndexes),
					ReadonlyIndexes: convertToIntSlice(atl.ReadonlyIndexes),
				}
			}

			solanaTx.Transaction.Signatures = decodedSignatures
			solanaTx.Transaction.Message.AccountKeys = decodedAccountKeys
			solanaTx.Transaction.Message.RecentBlockhash = base58.Encode(tx.Transaction.Transaction.Message.RecentBlockhash)
			solanaTx.Transaction.Message.Instructions = make([]types.Instruction, len(tx.Transaction.Transaction.Message.Instructions))
			solanaTx.Transaction.Message.Instructions = convertToInstructions(tx.Transaction.Transaction.Message.Instructions)

			solanaTx.Meta.LogMessages = tx.Transaction.Meta.LogMessages
			solanaTx.Meta.LoadedAddresses = types.LoadedAddresses{
				Readonly: convertToBase58Strings(tx.Transaction.Meta.LoadedReadonlyAddresses),
				Writable: convertToBase58Strings(tx.Transaction.Meta.LoadedWritableAddresses),
			}
			solanaTx.Meta.PreTokenBalances = convertToTokenBalanceSlice(tx.Transaction.Meta.PreTokenBalances)
			solanaTx.Meta.PostTokenBalances = convertToTokenBalanceSlice(tx.Transaction.Meta.PostTokenBalances)
			solanaTx.Meta.PreBalances = tx.Transaction.Meta.PreBalances
			solanaTx.Meta.PostBalances = tx.Transaction.Meta.PostBalances
			solanaTx.Meta.Fee = int64(tx.Transaction.Meta.Fee)
			solanaTx.Meta.Err = &types.TransactionError{}

			solanaTx.Meta.InnerInstructions = make([]types.InnerInstruction, len(tx.Transaction.Meta.InnerInstructions))
			for i, instr := range tx.Transaction.Meta.InnerInstructions {
				solanaTx.Meta.InnerInstructions[i] = types.InnerInstruction{
					Index:        int(instr.Index),
					Instructions: convertToInnerInstructions(instr.Instructions),
				}
			}
		}
		if bm := upd.GetBlockMeta(); bm != nil {
			blockTime = bm.BlockTime.Timestamp
		}

		if len(solanaTx.Transaction.Signatures) > 0 {
			if blockTime == 0 {
				blockTime = capturedTS
			}

			s.HandleTransaction(solanaTx, blockTime, blockNumber)
		}
	}

}

func (s *BlockListener) keepAlive(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if s.Subscription != nil {
				s.pingId++
				ping := &pb.SubscribeRequest{
					Ping: &pb.SubscribeRequestPing{
						Id: s.pingId,
					},
				}
				err := s.Subscription.Send(ping)
				if err != nil {
					return
				}

				_, err = s.Client.Ping(ctx, &pb.PingRequest{
					Count: s.pingId,
				})
				if err != nil {
					return
				}
			}
		}
	}
}

func (s *BlockListener) prepareSubscription() (*pb.SubscribeRequest, error) {
	voteFalse, failedFalse := false, false
	var stringArray []string

	sub := &pb.SubscribeRequest{
		Transactions: map[string]*pb.SubscribeRequestFilterTransactions{
			"tx_sub": {
				Vote:            &voteFalse,
				Failed:          &failedFalse,
				AccountExclude:  stringArray,
				AccountInclude:  stringArray,
				AccountRequired: stringArray,
				Signature:       nil,
			},
		},
		BlocksMeta: map[string]*pb.SubscribeRequestFilterBlocksMeta{
			"block_meta": {},
		},
	}
	confirmed := pb.CommitmentLevel_CONFIRMED
	sub.Commitment = &confirmed

	b, _ := json.Marshal(sub)
	log.Printf("Subscription request: %s", b)
	return sub, nil
}

func (s *BlockListener) HandleTransaction(transaction types.SolanaTx, blockTime int64, block uint64) {
	if s.queueHandler != nil {
		s.queueHandler.AddToSolanaQueue(types.BlockData{
			Transactions: []types.SolanaTx{transaction},
			Block:        block,
			Timestamp:    blockTime,
		})
	}
}
