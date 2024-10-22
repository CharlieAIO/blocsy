package main

import (
	"context"
	"defi-intel/internal/utils"
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

	//pRepo := db.NewTimescaleRepository(dbx)

}
