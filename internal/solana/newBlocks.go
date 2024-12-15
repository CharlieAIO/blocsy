package solana

import (
	"blocsy/internal/types"
	"context"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/mailru/easyjson"
	"log"
	"sync"
	"time"
)

func subscribeToBlocks(client *websocket.Conn) error {
	const msg = `{
        "jsonrpc": "2.0",
        "id": "1",
        "method": "blockSubscribe",
        "params": [
			"all",
            {
				"commitment": "confirmed",
				"encoding": "json",
				"showRewards": false,
				"transactionDetails": "full",
				"maxSupportedTransactionVersion": 2
            }
        ]
    }`

	return client.WriteMessage(websocket.TextMessage, []byte(msg))
}

type SolanaBlockListener struct {
	solanaSocketURL    string
	lastProcessedBlock int
	solSvc             *SolanaService
	pRepo              SwapsRepo
	queueHandler       *SolanaQueueHandler
	errorMutex         sync.Mutex
}

func NewBlockListener(wssUrl string, solSvc *SolanaService, pRepo SwapsRepo, qHandler *SolanaQueueHandler) *SolanaBlockListener {
	return &SolanaBlockListener{
		solanaSocketURL: wssUrl,
		solSvc:          solSvc,
		pRepo:           pRepo,
		queueHandler:    qHandler,
	}
}

func (s *SolanaBlockListener) Listen(ctx context.Context) error {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var firstNewBlock int = 0
	var errorOccurred bool = false

	for {
		if ctx.Err() != nil {
			return nil
		}

		client, _, err := websocket.DefaultDialer.Dial(s.solanaSocketURL, nil)
		if err != nil {
			log.Printf("setupWSSClient error: %v; retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		connectionOpen := true
		client.SetPingHandler(func(appData string) error {
			log.Println("Received ping, sending pong")
			return client.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(time.Second))
		})

		go s.keepAlive(ctx, client, &connectionOpen)

		if err := subscribeToBlocks(client); err != nil {
			log.Printf("subscribeToBlocks error: %v; retrying in 5 seconds...", err)
			connectionOpen = false
			_ = client.Close()
			time.Sleep(5 * time.Second)
			errorOccurred = true
			continue
		}

		if err := s.readMessages(ctx, client, &firstNewBlock, &errorOccurred); err != nil {
			log.Printf("readMessages error: %v", err)
			errorOccurred = true
		}

		connectionOpen = false
		_ = client.Close()
		time.Sleep(1 * time.Second)
	}
}

func (s *SolanaBlockListener) readMessages(ctx context.Context, client *websocket.Conn, firstNewBlock *int, errorOccurred *bool) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		if err := client.SetReadDeadline(time.Now().Add(25 * time.Second)); err != nil {
			return fmt.Errorf("SetReadDeadline error: %w", err)
		}

		_, message, err := client.ReadMessage()
		if err != nil {
			return fmt.Errorf("error reading message: %w", err)
		}

		if err := s.processMessage(ctx, message, firstNewBlock, errorOccurred); err != nil {
			return fmt.Errorf("error processing message: %w", err)
		}
	}
}

func (s *SolanaBlockListener) processMessage(ctx context.Context, message []byte, firstNewBlock *int, errorOccurred *bool) error {

	var blockMessage types.WSBlockMessage
	if err := easyjson.Unmarshal(message, &blockMessage); err != nil {
		return fmt.Errorf("Error decoding message: %w", err)
	}

	if blockMessage.Result != nil {
		log.Printf("Subscription Connected %d", *blockMessage.Result)
		return nil
	}

	if blockMessage.Params == nil {
		return nil
	}

	block := blockMessage.Params.Result.Value.Block
	slot := 0
	if block.ParentSlot != nil {
		slot = *blockMessage.Params.Result.Value.Slot
	}

	s.errorMutex.Lock()
	if *errorOccurred {
		*errorOccurred = false
		*firstNewBlock = slot

		if s.lastProcessedBlock != 0 {
			bf := NewBackfillService(s.solSvc, s.pRepo, s.queueHandler)

			startBlock := s.lastProcessedBlock + 1
			endBlock := *firstNewBlock - 1

			s.errorMutex.Unlock()

			go func(start, end int) {
				err := bf.HandleBackFill(ctx, start, end, false)
				if err != nil {
					log.Printf("HandleBackFill error: %v", err)
				}
			}(startBlock, endBlock)
		} else {
			s.errorMutex.Unlock()
		}
	} else {
		s.errorMutex.Unlock()
	}

	timestamp := time.Now().Unix()
	if block.BlockTime != nil {
		timestamp = *block.BlockTime
	}

	s.lastProcessedBlock = slot
	go func() {
		s.HandleBlock(block.Transactions, timestamp, uint64(slot))
	}()
	_ = s.pRepo.MarkBlockProcessed(ctx, slot)

	return nil
}

func (s *SolanaBlockListener) HandleBlock(blockTransactions []types.SolanaTx, blockTime int64, block uint64) {

	toProcess := make([]types.SolanaTx, 0)
	for i := range blockTransactions {
		if blockTransactions[i].Meta.Err != nil || !validateTX(&blockTransactions[i]) {
			continue
		}
		toProcess = append(toProcess, blockTransactions[i])

	}
	log.Printf("Block %d has %d transactions to process", block, len(toProcess))

	s.queueHandler.AddToSolanaQueue(types.BlockData{
		Transactions: toProcess,
		Block:        block,
		Timestamp:    blockTime,
	})
}

func (s *SolanaBlockListener) keepAlive(ctx context.Context, client *websocket.Conn, connectionOpen *bool) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !*connectionOpen {
				return
			}
			if err := client.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(5*time.Second)); err != nil {
				log.Printf("Error sending ping: %v", err)
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
