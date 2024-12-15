package main

import (
	"blocsy/internal/cache"
	"blocsy/internal/db"
	"blocsy/internal/solana"
	"blocsy/internal/utils"
	"context"
	"log"
	"os"
	"os/signal"
	"strings"
	"testing"
)

type TestCase struct {
	Target    int
	Signature string
}

func TestSwapHandler(t *testing.T) {

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

	solSvc := solana.NewSolanaService(ctx)

	tf := solana.NewTokenFinder(c, solSvc, mRepo)
	pf := solana.NewPairsService(c, tf, solSvc, mRepo)
	sh := solana.NewSwapHandler(tf, pf)

	nodeUrl := strings.Split(os.Getenv("SOL_HTTPS_BACKFILL_NODES"), ",")[0]

	node := solana.NewNode("test", nodeUrl)

	tests := []TestCase{
		//{Target: 4, Signature: "3Ry66VFMKtLUxwG1FFbAdz2RMZbxdwmDGXjMmtkWcA2PKTuVuQhU8RVMZUw1duciR8eeZLYHYD5tGCy5LVz1wa4f"},
		//{Target: 1, Signature: "2jRqSN5Ckx9sgwBoTkgQ7xz2fZb4TTftdz69noGFESx7eJW6Jxt1s4HP1fce6vzF8rBgzehYD3KSKQoKYWVRUotC"},
		//{Target: 3, Signature: "43skCh7Z9paKpE1Hv52QQ3hqDzEwgtzDmQL77DtpeUon7zJUQFuemwHUPcZhiT5cBp84vdz3Azo2cxn8fTyLyHZV"},
		//{Target: 1, Signature: "3Y2D8SYBBDnYYAwx9ZPasbySnB5AbofpUL1WCZzxVgMZULkCDpgBBChNs4mS8MFGjPUjzMZgtk8CnvyA2xhFk6GU"},
		{Target: 3, Signature: "2fbQoAurQTKEM2v3uYGuxJ5LQkucB5diFeMvZqGZ1ShLB5kixPzLS3Wnf7UCrZBFfSBYGRaySvSK1sgxApJWQR2g"},
		//{Target: 1, Signature: "5F8W4cxpYLM6JK9ASorp2YkcQzCFYcgbRDjA2PXqf6T59QN1iBcUHVycEtvcE7xtnKuYzf3HetTqVZf65eWcntQ9"},
		//{Target: 2, Signature: "K9PFgioRJRn6wgNjTwyW9j8Gtss6juwzvzkwsvXPHo9SUPACG6AexmSx2ZqfytELXgsopfsLEFfZ5AneNsvWc2F"},
		//{Target: 1, Signature: "5RaAG259hpsjEF3pETFBhCxqosvfSz4T4H7REs7KhPKXg8ZKhLAEk3M1uhqsXfKeCqHCT7JkrN6cVmiAPDaGbBU"},
		//{Target: 1, Signature: "2Wq5SoLz1LrrP4QYEgUKozFrq5VTTENE7JLZCWC2ZGMfqoPE4xZGygijZoihdFbLhXxQYGoB6unZqTX2H4qp5dtZ"},
		//{Target: 1, Signature: "FEkv16o2Az1JYChY758dP2NnemMTnEXNjQT8risU9stcS3asv8UTHvXPtzLYox3BzkwU49rQoRwWYSp6frkaNH1"},
	}

	for _, tc := range tests {
		t.Run(tc.Signature, func(t *testing.T) {
			test_tx(ctx, node, sh, t, tc.Signature, tc.Target)
		})
	}

}

func test_tx(ctx context.Context, node *solana.Node, sh *solana.SwapHandler, t *testing.T, signature string, target int) {
	tx, err := node.GetTx(ctx, signature)
	if err != nil {
		t.Fatalf("Error getting tx: %v", err)
	}
	transfers := solana.GetAllTransfers(tx)

	for i, transfer := range transfers {
		t.Logf("Transfer %d: %+v", i, transfer)
	}

	swaps := sh.HandleSwaps(ctx, transfers, tx, 0, 0)

	if len(swaps) != target {
		t.Fatalf("Expected %d from HandleSwaps, got %d", target, len(swaps))
	}

	for i, swap := range swaps {
		t.Logf("Swap %d: %+v", i, swap)
	}

}
