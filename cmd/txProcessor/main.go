package main

import (
	"context"
	"defi-intel/cmd/api/websocket"
	"defi-intel/internal/cache"
	"defi-intel/internal/db"
	"defi-intel/internal/solana"
	"defi-intel/internal/utils"
	"log"
	"os"
	"os/signal"
)

func main() {

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	utils.LoadEnvironment()

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

	defer func() {
		log.Println("Closing DB connection...")
		dbx.Close()
	}()

	c := cache.NewCache()
	mRepo := db.NewMongoRepository(mCli)
	pRepo := db.NewTimescaleRepository(dbx)

	websocketServer := websocket.NewWebSocketServer()
	go websocketServer.Start()

	go solanaTxHandler(ctx, c, mRepo, pRepo, websocketServer)

	<-ctx.Done()
	log.Println("Shutting down tx processor...")
}

func solanaTxHandler(ctx context.Context, c *cache.Cache, mRepo *db.MongoRepository, pRepo *db.TimescaleRepository, websocketServer *websocket.WebSocketServer) {
	solCli := solana.NewSolanaService(ctx)

	tf := solana.NewTokenFinder(c, solCli, mRepo)
	pf := solana.NewPairsService(c, tf, solCli, mRepo)
	sh := solana.NewSwapHandler(tf, pf)

	txHandler := solana.NewTxHandler(sh, solCli, mRepo, pRepo, websocketServer)

	queueHandler := solana.NewSolanaQueueHandler(txHandler, pRepo)

	log.Println("Listening for new solana txs...")
	defer log.Println("Stopped listening for new solana txs...")
	queueHandler.ListenToSolanaQueue(ctx)

}
