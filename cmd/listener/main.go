package main

import (
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

	go solanaListener(ctx)

	<-ctx.Done()

}

func solanaListener(ctx context.Context) {

	grpcAddress := os.Getenv("SOL_GRPC")
	if grpcAddress == "" {
		log.Fatalf("SOL_GRPC is required")
	}

	queueHandler := solana.NewSolanaQueueHandler(nil, nil)
	sbl := solana.NewBlockListener(grpcAddress, queueHandler)

	go func() {
		log.Println("Listening for new blocks (solana)...")
		defer log.Println("Stopped listening for new blocks (solana)...")
		if err := sbl.Listen(); err != nil {
			log.Fatalf("failed to listen, err: %v", err)
		}

	}()
}
