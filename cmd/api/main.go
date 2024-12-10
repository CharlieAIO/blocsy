package main

import (
	"blocsy/cmd/api/routes"
	"blocsy/internal/cache"
	"blocsy/internal/db"
	"blocsy/internal/solana"
	"blocsy/internal/trackers"
	"blocsy/internal/utils"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer cancel()

	//key, err := routes.GenerateAPIKey(20)
	//if err != nil {
	//	return
	//}
	//log.Printf("API key: %s", key)
	//return

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

	defer dbx.Close()

	solSvc := solana.NewSolanaService(ctx)

	pt := trackers.NewPriceTracker()
	go pt.Run(ctx)

	c := cache.NewCache()
	mRepo := db.NewMongoRepository(mCli)
	swapsRepo := db.NewTimescaleRepository(dbx)

	tf := solana.NewTokenFinder(c, solSvc, mRepo)
	pf := solana.NewPairsService(c, tf, solSvc, mRepo)

	var nodes []routes.Node

	if os.Getenv("SOL_HTTPS_BACKFILL_NODES") != "" {
		nodeUrls := strings.Split(os.Getenv("SOL_HTTPS_BACKFILL_NODES"), ",")
		for i, url := range nodeUrls {
			//nodes = solana.NewNode(fmt.Sprintf("node %d", i), url)
			nodes = append(nodes, solana.NewNode(fmt.Sprintf("node %d", i), url))
		}
	} else {
		log.Fatalf("No nodes provided")
	}

	handler := routes.NewHandler(pt, tf, pf, swapsRepo, nodes).GetHttpHandler()

	srv := &http.Server{
		Addr:    ":8080",
		Handler: handler,
	}

	go func() {
		log.Println("Starting server on :8080")
		if err := srv.ListenAndServe(); err != nil {
			log.Printf("Error starting server: %v", err)
		}
	}()

	<-ctx.Done()

	log.Println("Shutting down server...")

	ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("Error shutting down server: %v", err)
	}

	log.Println("Server gracefully stopped")
}
