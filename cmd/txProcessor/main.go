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
	solSvc := solana.NewSolanaService(ctx)

	tf := solana.NewTokenFinder(c, solSvc, mRepo)
	tf.NewTokenProcessor()
	tf.NewMintBurnProcessor()
	pf := solana.NewPairsService(c, tf, solSvc, mRepo)
	pf.NewPairProcessor()

	sh := solana.NewSwapHandler(tf, pf)

	txHandler := solana.NewTxHandler(sh, solSvc, mRepo, pRepo, websocketServer)

	queueHandler := solana.NewSolanaQueueHandler(txHandler, pRepo)

	log.Println("Listening for new solana txs...")
	defer log.Println("Stopped listening for new solana txs...")
	queueHandler.ListenToSolanaQueue(ctx)

}
