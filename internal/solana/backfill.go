package solana

import (
	"context"
	"defi-intel/internal/types"
	"fmt"
	"log"
)

func NewBackfillService(solSvc *SolanaService, pRepo SwapsRepo, qHandler *SolanaQueueHandler) *BackfillService {
	return &BackfillService{
		solSvc:       solSvc,
		queueHandler: qHandler,
		pRepo:        pRepo,
	}
}

func (bs *BackfillService) HandleBackFill(ctx context.Context, startBlock int, toBlock int) error {
	log.Printf("Handling Backfill %d --> %d", startBlock, toBlock)
	defer log.Printf("Backfill completed")

	if toBlock == 0 {
		slot, err := bs.solSvc.GetSlot(ctx)
		if err != nil {
			return fmt.Errorf("failed to get slot: %w", err)
		}
		toBlock = int(slot)
	}

	var nodes []*Node

	for i := startBlock; i <= toBlock; i++ {
		n := nodes[i%4]
		block, err := n.GetBlockMessage(context.Background(), i)
		if err != nil {
			log.Fatalf("Error getting transaction: %v", err)
		}
		_ = bs.pRepo.MarkBlockProcessed(ctx, i)
		if block.Error != nil {
			if block.Error.Code == -32009 {
				log.Printf("Block %d not available", i)
				continue
			} else {
				continue
			}
		}

		bs.queueHandler.AddToSolanaQueue(types.BlockData{
			Transactions: block.Result.Transactions,
			Block:        uint64(i),
			Timestamp:    block.Result.BlockTime,
		})

	}

	return nil
}
