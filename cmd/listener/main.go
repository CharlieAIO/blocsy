package main

import (
	"blocsy/internal/db"
	"blocsy/internal/solana"
	"blocsy/internal/utils"
	"context"
	"log"
	"os"
	"os/signal"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	utils.LoadEnvironment()

	dbx, err := utils.GetDBConnection(ctx)
	if err != nil {
		log.Fatalf("Error connecting to db: %v", err)
	}

	defer dbx.Close()

	pRepo := db.NewTimescaleRepository(dbx)
	go solanaListener(ctx, pRepo)

	<-ctx.Done()

}

func solanaListener(ctx context.Context, pRepo *db.TimescaleRepository) {
	url_https := os.Getenv("SOL_HTTPS")
	if url_https == "" {
		log.Fatalf("SOL_HTTPS is required")
	}

	url_wss := os.Getenv("SOL_WSS")
	if url_wss == "" {
		log.Fatalf("SOL_WSS is required")
	}

	solSvc := solana.NewSolanaService(ctx)
	queueHandler := solana.NewSolanaQueueHandler(nil, nil)
	sbl := solana.NewBlockListener(url_wss, solSvc, pRepo, queueHandler)

	go func() {
		log.Println("Listening for new blocks (solana)...")
		defer log.Println("Stopped listening for new blocks (solana)...")
		if err := sbl.Listen(ctx); err != nil {
			log.Fatalf("failed to listen, err: %v", err)
		}

	}()
}
