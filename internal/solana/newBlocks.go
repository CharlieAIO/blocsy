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
	"os"
	"sync"
	"time"
)

var (
	authToken  = flag.String("x-token", "", "Token for authenticating")
	jsonInput  = flag.String("json", "", "JSON for subscription request, prefix with @ to read json from file")
	slots      = flag.Bool("slots", false, "Subscribe to slots update")
	blocks     = flag.Bool("blocks", false, "Subscribe to block update")
	block_meta = flag.Bool("blocks-meta", true, "Subscribe to block metadata update")
	signature  = flag.String("signature", "", "Subscribe to a specific transaction signature")

	accounts = flag.Bool("accounts", false, "Subscribe to accounts")

	transactions       = flag.Bool("transactions", true, "Subscribe to transactions, required for tx_account_include/tx_account_exclude and vote/failed.")
	voteTransactions   = flag.Bool("transactions-vote", false, "Include vote transactions")
	failedTransactions = flag.Bool("transactions-failed", false, "Include failed transactions")

	accountsFilter              types.ArrayFlags
	accountOwnersFilter         types.ArrayFlags
	transactionsAccountsInclude types.ArrayFlags
	transactionsAccountsExclude types.ArrayFlags
)

var kacp = keepalive.ClientParameters{
	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	Timeout:             20 * time.Second,
	PermitWithoutStream: true, // send pings even without active streams
}

type SolanaBlockListener struct {
	//solanaSocketURL    string
	grpcAddress        string
	lastProcessedBlock int
	solSvc             *SolanaService
	pRepo              SwapsRepo
	queueHandler       *SolanaQueueHandler
	errorMutex         sync.Mutex
}

func NewBlockListener(grpc string, solSvc *SolanaService, pRepo SwapsRepo, qHandler *SolanaQueueHandler) *SolanaBlockListener {
	return &SolanaBlockListener{
		grpcAddress:  grpc,
		solSvc:       solSvc,
		pRepo:        pRepo,
		queueHandler: qHandler,
	}
}

func (s *SolanaBlockListener) Listen() error {
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
			conn.Close()                // explicitly close before reconnecting
			time.Sleep(5 * time.Second) // Wait before reconnecting
			continue
		}
		defer conn.Close() // Only defer close on a successful subscription
		break
	}

	return nil
}

func (s *SolanaBlockListener) grpcConnect(address string, plaintext bool) *grpc.ClientConn {
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

func (s *SolanaBlockListener) grpcSubscribe(conn *grpc.ClientConn) error {
	var err error
	client := pb.NewGeyserClient(conn)

	var subscription pb.SubscribeRequest

	// Read json input or JSON file prefixed with @
	if *jsonInput != "" {
		var jsonData []byte

		if (*jsonInput)[0] == '@' {
			jsonData, err = os.ReadFile((*jsonInput)[1:])
			if err != nil {
				return fmt.Errorf("error reading provided json file: %v", err)
			}
		} else {
			jsonData = []byte(*jsonInput)
		}
		err := json.Unmarshal(jsonData, &subscription)
		if err != nil {
			return err
		}
	} else {
		subscription = pb.SubscribeRequest{}
	}

	if *slots {
		if subscription.Slots == nil {
			subscription.Slots = make(map[string]*pb.SubscribeRequestFilterSlots)
		}

		subscription.Slots["slots"] = &pb.SubscribeRequestFilterSlots{}

	}

	if *blocks {
		if subscription.Blocks == nil {
			subscription.Blocks = make(map[string]*pb.SubscribeRequestFilterBlocks)
		}
		subscription.Blocks["blocks"] = &pb.SubscribeRequestFilterBlocks{}
	}

	if *block_meta {
		if subscription.BlocksMeta == nil {
			subscription.BlocksMeta = make(map[string]*pb.SubscribeRequestFilterBlocksMeta)
		}
		subscription.BlocksMeta["block_meta"] = &pb.SubscribeRequestFilterBlocksMeta{}
	}

	if (len(accountsFilter)+len(accountOwnersFilter)) > 0 || (*accounts) {
		if subscription.Accounts == nil {
			subscription.Accounts = make(map[string]*pb.SubscribeRequestFilterAccounts)
		}

		subscription.Accounts["account_sub"] = &pb.SubscribeRequestFilterAccounts{}

		if len(accountsFilter) > 0 {
			subscription.Accounts["account_sub"].Account = accountsFilter
		}

		if len(accountOwnersFilter) > 0 {
			subscription.Accounts["account_sub"].Owner = accountOwnersFilter
		}
	}

	// Set up the transactions subscription
	if subscription.Transactions == nil {
		subscription.Transactions = make(map[string]*pb.SubscribeRequestFilterTransactions)
	}

	// Subscribe to a specific signature
	if *signature != "" {
		tr := true
		subscription.Transactions["signature_sub"] = &pb.SubscribeRequestFilterTransactions{
			Failed: &tr,
			Vote:   &tr,
		}

		if *signature != "" {
			subscription.Transactions["signature_sub"].Signature = signature
		}
	}

	// Subscribe to generic transaction stream
	if *transactions {

		subscription.Transactions["transactions_sub"] = &pb.SubscribeRequestFilterTransactions{
			Failed: failedTransactions,
			Vote:   voteTransactions,
		}

		subscription.Transactions["transactions_sub"].AccountInclude = transactionsAccountsInclude
		subscription.Transactions["transactions_sub"].AccountExclude = transactionsAccountsExclude
	}
	confirmed := pb.CommitmentLevel_CONFIRMED
	subscription.Commitment = &confirmed
	subscriptionJson, err := json.Marshal(&subscription)
	if err != nil {
		log.Printf("Failed to marshal subscription request: %v", err)
	}
	log.Printf("Subscription request: %s", string(subscriptionJson))

	ctx := context.Background()
	if *authToken != "" {
		md := metadata.New(map[string]string{"x-token": *authToken})
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	stream, err := client.Subscribe(ctx)
	if err != nil {
		return fmt.Errorf("subscribe: %v", err)
	}
	err = stream.Send(&subscription)
	if err != nil {
		return fmt.Errorf("send: %v", err)
	}

	// Check for unexpected content-type
	header, err := stream.Header()
	if err != nil {
		return fmt.Errorf("failed to get header: %v", err)
	}
	if contentType := header.Get("content-type"); len(contentType) > 0 && contentType[0] != "application/grpc" {
		return fmt.Errorf("unexpected content-type: %s", contentType)
	}

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

		if block := resp.GetBlockMeta(); block != nil {
			blockTime = block.BlockTime.Timestamp
		}

		if tx := resp.GetTransaction(); tx != nil {
			blockNumber = tx.Slot

			decodedSignatures := []string{}
			for _, sig := range tx.Transaction.Transaction.Signatures {
				b58Sig := base58.Encode(sig)
				decodedSignatures = append(decodedSignatures, b58Sig)
			}

			decodedAccountKeys := []string{}
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

func (s *SolanaBlockListener) HandleTransaction(transaction types.SolanaTx, blockTime int64, block uint64) {

	if s.queueHandler != nil {
		s.queueHandler.AddToSolanaQueue(types.BlockData{
			Transactions: []types.SolanaTx{transaction},
			Block:        block,
			Timestamp:    blockTime,
		})
	}
}
