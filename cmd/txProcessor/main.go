package main

import (
	"blocsy/cmd/api/websocket"
	"blocsy/internal/cache"
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

	defer func() {
		log.Println("Closing DB connection...")
		dbx.Close()
	}()

	c := cache.NewCache()
	pRepo := db.NewTimescaleRepository(dbx)

	websocketServer := websocket.NewWebSocketServer()
	go websocketServer.Start()

	go solanaTxHandler(ctx, c, pRepo, websocketServer)

	<-ctx.Done()
	log.Println("Shutting down tx processor...")
}

func solanaTxHandler(ctx context.Context, c *cache.Cache, pRepo *db.TimescaleRepository, websocketServer *websocket.WebSocketServer) {
	solSvc := solana.NewSolanaService(ctx)

	tf := solana.NewTokenFinder(c, solSvc, pRepo)
	tf.NewTokenProcessor()
	pf := solana.NewPairsService(c, tf, solSvc, pRepo)
	pf.NewPairProcessor()

	sh := solana.NewSwapHandler(tf, pf)

	txHandler := solana.NewTxHandler(sh, solSvc, pRepo, pRepo, websocketServer)

	queueHandler := solana.NewSolanaQueueHandler(txHandler, pRepo)

	log.Println("Listening for solana txs in rabbitMQ...")
	defer log.Println("Stopped listening for solana txs in rabbitMQ...")
	queueHandler.ListenToSolanaQueue(ctx)

}
