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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/metadata"
	"io"
	"log"
	"time"
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             20 * time.Second,
	PermitWithoutStream: true,
}

func NewBlockListener(grpc string, qHandler *QueueHandler) *BlockListener {
	return &BlockListener{
		grpcAddress:  grpc,
		queueHandler: qHandler,
		authToken:    "",
	}
}

func (s *BlockListener) Listen() error {
	log.SetFlags(0)
	flag.Parse()

	for {
		conn := s.grpcConnect(s.grpcAddress, true)
		if conn == nil {
			log.Printf("Failed to connect to %s. Retrying...", s.grpcAddress)
			time.Sleep(5 * time.Second)
			continue
		}

		err := s.grpcSubscribe(conn)
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

func (s *BlockListener) grpcConnect(address string, plaintext bool) *grpc.ClientConn {
	var opts []grpc.DialOption
	if plaintext {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	} else {
		pool, _ := x509.SystemCertPool()
		creds := credentials.NewClientTLSFromCert(pool, "")
		opts = append(opts, grpc.WithTransportCredentials(creds))
	}

	opts = append(opts, grpc.WithKeepaliveParams(kacp))

	log.Println("Starting grpc client, connecting to", address)
	conn, err := grpc.Dial(address, opts...)
	if err != nil {
		return nil
	}

	return conn
}

func (s *BlockListener) grpcSubscribe(conn *grpc.ClientConn) error {
	var err error
	client := pb.NewGeyserClient(conn)

	subscription, err := s.prepareSubscription()
	if err != nil {
		return fmt.Errorf("failed to prepare subscription: %v", err)
	}

	ctx := context.Background()
	if s.authToken != "" {
		md := metadata.New(map[string]string{"x-token": s.authToken})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	stream, err := client.Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("subscribe: %v", err)
	}
	err = stream.Send(subscription)
	if err != nil {
		return fmt.Errorf("send: %v", err)
	}

	log.Printf("Subscribed to %s", s.grpcAddress)

	var blockNumber uint64
	var blockTime int64

	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			return err
		}
		if err != nil {
			return err
		}
		var capturedTS = time.Now().Unix()
		var solanaTx types.SolanaTx

		log.Printf("resp %v", resp.String())

		if block := resp.GetBlockMeta(); block != nil {
			blockTime = block.BlockTime.Timestamp
		}

		if tx := resp.GetTransaction(); tx != nil {
			log.Printf("tx %+v", tx)
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

		if len(solanaTx.Transaction.Signatures) > 0 {
			if blockTime == 0 {
				blockTime = capturedTS
			}

			s.HandleTransaction(solanaTx, blockTime, blockNumber)
		}
	}

}

func (s *BlockListener) prepareSubscription() (*pb.SubscribeRequest, error) {
	var subscription = pb.SubscribeRequest{}
	var err error

	if subscription.BlocksMeta == nil {
		subscription.BlocksMeta = make(map[string]*pb.SubscribeRequestFilterBlocksMeta)
	}
	subscription.BlocksMeta["block_meta"] = &pb.SubscribeRequestFilterBlocksMeta{}

	if subscription.Transactions == nil {
		subscription.Transactions = make(map[string]*pb.SubscribeRequestFilterTransactions)
	}
	var accounts []string
	subscription.Transactions["transactions_sub"] = &pb.SubscribeRequestFilterTransactions{
		Failed: new(bool),
		Vote:   new(bool),
	}
	subscription.Transactions["transactions_sub"].AccountInclude = accounts
	subscription.Transactions["transactions_sub"].AccountExclude = accounts

	confirmed := pb.CommitmentLevel_CONFIRMED
	subscription.Commitment = &confirmed
	subscriptionJson, err := json.Marshal(&subscription)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal subscription request: %v", err)
	}

	log.Printf("Subscription request: %v", string(subscriptionJson))
	return &subscription, nil

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
