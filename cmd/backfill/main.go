package main

import (
	"blocsy/internal/db"
	"blocsy/internal/solana"
	"blocsy/internal/utils"
	"context"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	utils.LoadEnvironment()

	go monitor(ctx)

	mCli, err := utils.GetMongoConnection(ctx)
	if err != nil {
		log.Fatalf("Error connecting to mongo: %v", err)
	}

	defer func() {
		if err := mCli.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from mongo: %v", err)
		}
	}()

	dbx, err := utils.GetDBConnection(ctx)
	if err != nil {
		log.Fatalf("Error connecting to db: %v", err)
	}

	defer dbx.Close()

	solCli := solana.NewSolanaService(ctx)
	pRepo := db.NewTimescaleRepository(dbx)

	queueHandler := solana.NewSolanaQueueHandler(nil, nil)
	backfillService := solana.NewBackfillService(solCli, pRepo, queueHandler)

	numWorkers := 2

	for {
		blocks, err := pRepo.FindMissingBlocks(ctx)
		if err != nil {
			log.Fatalf("Error getting missing blocks: %v", err)
		}

		if len(blocks) == 0 {
			log.Println("No missing blocks found, sleeping for a while...")
			time.Sleep(5 * time.Minute)
			continue
		}

		blockRanges := make(chan [2]int, len(blocks))
		for _, blockRange := range blocks {
			if len(blockRange) == 2 {
				blockRanges <- [2]int{blockRange[0], blockRange[1]}
			} else {
				log.Printf("Invalid block range: %v", blockRange)
			}
		}
		close(blockRanges)

		wg := &sync.WaitGroup{}

		wg.Add(numWorkers)
		for i := 0; i < numWorkers; i++ {
			go func() {
				defer wg.Done()
				for blockRange := range blockRanges {
					err := backfillService.HandleBackFill(ctx, blockRange[0], blockRange[1], true)
					if err != nil {
						log.Printf("Error handling backfill for block range %v: %v", blockRange, err)
					}
				}
			}()
		}

		wg.Wait()
		log.Println("Finished processing current set of missing blocks")
	}

}

func monitor(ctx context.Context) {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			log.Println("Heartbeat: Application is still running")
		case <-ctx.Done():
			log.Println("Application is shutting down")
			return
		}
	}
}
