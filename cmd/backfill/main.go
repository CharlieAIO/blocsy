package main

import (
	"blocsy/internal/db"
	"blocsy/internal/utils"
	"context"
	"log"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	utils.LoadEnvironment()

	go monitor(ctx)

	dbx, err := utils.GetDBConnection(ctx)
	if err != nil {
		log.Fatalf("Error connecting to db: %v", err)
	}

	defer dbx.Close()

	mCli, err := utils.GetMongoConnection(ctx)
	if err != nil {
		log.Fatalf("Error connecting to mongo: %v", err)
	}

	defer func() {
		if err := mCli.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from mongo: %v", err)
		}
	}()

	//solCli := solana.NewSolanaService(ctx)
	pRepo := db.NewTimescaleRepository(dbx)
	mRepo := db.NewMongoRepository(mCli)

	tokenCh, errCh := mRepo.PullTokens(ctx)
	count := 0

	// Process tokens as they are received.
	for token := range tokenCh {

		if err := pRepo.InsertToken(ctx, token); err != nil {
			log.Printf("Failed to insert token (address %s): %v", token.Address, err)
			continue
		}
		count++
		if count%1000 == 0 {
			log.Printf("Inserted %d tokens so far...", count)
		}
	}

	// Check if any error occurred during token retrieval.
	if err, ok := <-errCh; ok && err != nil {
		log.Fatalf("Error while pulling tokens: %v", err)
	}

	log.Printf("Migration complete. Total tokens inserted: %d", count)

	//queueHandler := solana.NewSolanaQueueHandler(nil, nil)
	//backfillService := solana.NewBackfillService(solCli, pRepo, queueHandler)
	//
	//numWorkers := 1
	//
	//for {
	//	blocks, err := pRepo.FindMissingBlocks(ctx)
	//	if err != nil {
	//		log.Fatalf("Error getting missing blocks: %v", err)
	//	}
	//
	//	if len(blocks) == 0 {
	//		log.Println("No missing blocks found, sleeping for a while...")
	//		time.Sleep(5 * time.Minute)
	//		continue
	//	}
	//
	//	blockRanges := make(chan [2]int, len(blocks))
	//	for _, blockRange := range blocks {
	//		if len(blockRange) == 2 {
	//			blockRanges <- [2]int{blockRange[0], blockRange[1]}
	//		} else {
	//			log.Printf("Invalid block range: %v", blockRange)
	//		}
	//	}
	//	close(blockRanges)
	//
	//	wg := &sync.WaitGroup{}
	//
	//	wg.Add(numWorkers)
	//	for i := 0; i < numWorkers; i++ {
	//		go func() {
	//			defer wg.Done()
	//			for blockRange := range blockRanges {
	//				err := backfillService.HandleBackFill(ctx, blockRange[0], blockRange[1], true)
	//				if err != nil {
	//					log.Printf("Error handling backfill for block range %v: %v", blockRange, err)
	//				}
	//			}
	//		}()
	//	}
	//
	//	wg.Wait()
	//	log.Println("Finished processing current set of missing blocks")
	//}

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
