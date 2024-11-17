package solana

import (
	"context"
	"defi-intel/internal/types"
	"fmt"
	"github.com/mailru/easyjson"
	"log"
	"time"

	"github.com/gorilla/websocket"
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
	solanaSockerURL    string
	lastProcessedBlock int
	solCli             *SolanaService
	pRepo              SwapsRepo
	queueHandler       *SolanaQueueHandler
}

func NewBlockListener(solanaSockerURL string, solCli *SolanaService, pRepo SwapsRepo, qHandler *SolanaQueueHandler) *SolanaBlockListener {
	return &SolanaBlockListener{
		solanaSockerURL: solanaSockerURL,
		solCli:          solCli,
		pRepo:           pRepo,
		queueHandler:    qHandler,
	}
}

func (s *SolanaBlockListener) Listen(ctx context.Context) error {
	var firstNewBlock int = 0
	var errorOccurred bool = false

	for {
		// Exit the function if the context is done
		if ctx.Err() != nil {
			return nil
		}

		time.Sleep(1 * time.Second)

		client, _, err := websocket.DefaultDialer.Dial(s.solanaSockerURL, nil)
		if err != nil {
			log.Printf("setupWSSClient error: %v; retrying in 5 seconds...", err)
			time.Sleep(5 * time.Second)
			continue
		}

		if err := subscribeToBlocks(client); err != nil {
			log.Printf("subscribeToBlocks error: %v; retrying in 5 seconds...", err)
			err = client.Close()
			if err != nil {
				continue
			}

			time.Sleep(5 * time.Second)
			errorOccurred = true
			continue
		}

		if err := s.readMessages(ctx, client, &firstNewBlock, &errorOccurred); err != nil {
			log.Printf("readMessages error: %v", err)
			errorOccurred = true
		}

		if err := client.Close(); err != nil {
			continue
		}
	}
}

func (s *SolanaBlockListener) readMessages(ctx context.Context, client *websocket.Conn, firstNewBlock *int, errorOccurred *bool) error {
	for {
		if ctx.Err() != nil {
			return nil
		}

		if err := client.SetReadDeadline(time.Now().Add(30 * time.Second)); err != nil {
			return fmt.Errorf("SetReadDeadline error: %w", err)
		}

		_, message, err := client.ReadMessage()
		if err != nil {
			return fmt.Errorf("Error reading message: %w", err)
		}

		if err := s.processMessage(ctx, message, firstNewBlock, errorOccurred); err != nil {
			return fmt.Errorf("Error processing message: %w", err)
		}
	}
}

func (s *SolanaBlockListener) processMessage(ctx context.Context, message []byte, firstNewBlock *int, errorOccurred *bool) error {

	var blockMessage types.WSBlockMessage
	if err := easyjson.Unmarshal(message, &blockMessage); err != nil {
		//SaveErrorMessageToFile(message)
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

	if *errorOccurred {
		*firstNewBlock = slot
		*errorOccurred = false // Reset the error flag

		if s.lastProcessedBlock != 0 {
			bf := NewBackfillService(s.solCli, s.pRepo, s.queueHandler)
			go func() {
				err := bf.HandleBackFill(ctx, s.lastProcessedBlock+1, *firstNewBlock-1)
				if err != nil {
					log.Printf("HandleBackFill error: %v", err)
				}
			}()
		}
	}

	timestamp := time.Now().Unix()
	if block.BlockTime != nil {
		timestamp = *block.BlockTime
	}

	log.Printf("Block %d has %d transactions", slot, len(block.Transactions))

	s.lastProcessedBlock = slot
	s.HandleBlock(block.Transactions, timestamp, uint64(slot))
	_ = s.pRepo.MarkBlockProcessed(ctx, slot)

	return nil
}

func (s *SolanaBlockListener) HandleBlock(blockTransactions []types.SolanaTx, blockTime int64, block uint64) {
	for i := range blockTransactions {
		allNull := true
		for j := 0; j < len(blockTransactions[i].Meta.Err.InstructionError) && !allNull; j++ {
			allNull = blockTransactions[i].Meta.Err.InstructionError[j] == nil
		}

		if !allNull {
			continue
		}

		s.queueHandler.AddToSolanaQueue(context.Background(), types.SolanaBlockTx{
			Tx:        blockTransactions[i],
			Block:     block,
			Timestamp: blockTime,
		})

	}
}
