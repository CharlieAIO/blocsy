package main

import (
	"blocsy/internal/cache"
	"blocsy/internal/db"
	"blocsy/internal/solana"
	"blocsy/internal/types"
	"blocsy/internal/utils"
	"context"
	"log"
	"os"
	"os/signal"
	"sync"
	"testing"
	"time"
)

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
	pRepo := db.NewTimescaleRepository(dbx)

	solSvc := solana.NewSolanaService(ctx)

	tf := solana.NewTokenFinder(c, solSvc, mRepo)
	pf := solana.NewPairsService(c, tf, solSvc, mRepo)
	sh := solana.NewSwapHandler(tf, pf)
	txHandler := solana.NewTxHandler(sh, solSvc, mRepo, pRepo, nil)

	node := solana.NewNode("test", "https://magical-warmhearted-diagram.solana-mainnet.quiknode.pro/8fec389fad577ed79b0f04773fea0e00db485712/")

	txTest(ctx, node, sh, t)
	blockTest(ctx, node, txHandler, t)

}

func txTest(ctx context.Context, node *solana.Node, sh *solana.SwapHandler, t *testing.T) {
	return

	const signature = "43QW8kcBpkrxvSauLZkhBMD5ToLEo5zRFmHPfHmLJaLw3ED8LogTevcNrjzm2aWTYqzGvc97dNfgKHYxuqmZ6PCr"

	tx, err := node.GetTx(ctx, signature)
	if err != nil {
		t.Fatalf("Error getting tx: %v", err)
	}

	transfers := solana.GetAllTransfers(tx)
	swaps := sh.HandleSwaps(ctx, transfers, tx, 0, 0)
	log.Printf("Swaps: %+v", swaps)
}

func blockTest(ctx context.Context, node *solana.Node, txHandler *solana.TxHandler, t *testing.T) {
	//return
	const blockNum = 304129216

	blockMsg, err := node.GetBlockMessage(ctx, blockNum)
	if err != nil {
		t.Fatalf("Error getting tx: %v", err)
	}

	startTime := time.Now()

	txChan := make(chan types.SolanaTx, len(blockMsg.Result.Transactions))
	resultChan := make(chan []types.SwapLog)
	var wg sync.WaitGroup

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for tx := range txChan {
				if tx.Meta.Err != nil {
					continue
				}
				_swaps, err := txHandler.ProcessTransaction(ctx, &tx, blockMsg.Result.BlockTime, uint64(blockMsg.Result.ParentSlot), true)
				if err != nil {
					continue
				}
				resultChan <- _swaps
			}
		}()
	}

	go func() {
		for _, tx := range blockMsg.Result.Transactions {
			txChan <- tx
		}
		close(txChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
	}()

	swaps := make([]types.SwapLog, 0)
	for result := range resultChan {
		swaps = append(swaps, result...)
	}

	log.Printf("Successfully processed block. Time taken: %v | Got %d swaps", time.Since(startTime), len(swaps))
}
