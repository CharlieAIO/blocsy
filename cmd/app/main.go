package main

import (
	"context"
	"defi-intel/internal/db"
	"defi-intel/internal/solana"
	"defi-intel/internal/utils"
	"log"
	"os"
	"os/signal"
	"strings"
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
	url := os.Getenv("SOL_HTTPS")
	if url == "" {
		log.Fatalf("SOL_HTTPS is required")
	}
	// Ensure we have a WSS URL instead of HTTP
	url = strings.Replace(url, "https://", "wss://", 1)

	solCli := solana.NewSolanaService(ctx)
	queueHandler := solana.NewSolanaQueueHandler(nil, nil)
	sbl := solana.NewBlockListener(url, solCli, pRepo, queueHandler)

	go func() {
		log.Println("Listening for new blocks (solana)...")
		defer log.Println("Stopped listening for new blocks (solana)...")
		if err := sbl.Listen(ctx); err != nil {
			log.Fatalf("failed to listen, err: %v", err)
		}

	}()
}
