package main

import (
	"context"
	"defi-intel/cmd/api/routes"
	"defi-intel/internal/cache"
	"defi-intel/internal/db"
	"defi-intel/internal/solana"
	"defi-intel/internal/trackers"
	"defi-intel/internal/utils"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"
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

	defer dbx.Close()

	solCli := solana.NewSolanaService(ctx)

	pt := trackers.NewPriceTracker()
	go pt.Run(ctx)

	c := cache.NewCache()
	mRepo := db.NewMongoRepository(mCli)
	swapsRepo := db.NewTimescaleRepository(dbx)

	tf := solana.NewTokenFinder(c, solCli, mRepo)
	pf := solana.NewPairsService(c, tf, solCli, mRepo)

	handler := routes.NewHandler(pt, tf, pf, swapsRepo).GetHttpHandler()

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
