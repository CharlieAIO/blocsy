package solana

import (
	"context"
	"defi-intel/internal/types"
	"fmt"
	"log"
	"os"
	"strings"
)

func NewBackfillService(solSvc *SolanaService, pRepo SwapsRepo, qHandler *SolanaQueueHandler) *BackfillService {

	var nodes []*Node

	if os.Getenv("SOL_HTTPS_BACKFILL_NODES") != "" {
		nodeUrls := strings.Split(os.Getenv("SOL_HTTPS_BACKFILL_NODES"), ",")
		for i, url := range nodeUrls {
			nodes = append(nodes, NewNode(fmt.Sprintf("node %d", i), url))
		}
	} else {
		log.Fatalf("No nodes provided for backfill")
	}

	return &BackfillService{
		solSvc:       solSvc,
		queueHandler: qHandler,
		pRepo:        pRepo,
		nodeUrls:     nodes,
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

	for i := startBlock; i <= toBlock; i++ {
		n := bs.nodeUrls[i%(len(bs.nodeUrls))]
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
