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
	//return
	//const signature = "3Ry66VFMKtLUxwG1FFbAdz2RMZbxdwmDGXjMmtkWcA2PKTuVuQhU8RVMZUw1duciR8eeZLYHYD5tGCy5LVz1wa4f"
	const signature = "2pxpG7AL94Mqf82YGF2zzD6LPvVKULqwWZSSKUjEZLh2ENo1aJjw2gSzm1Qp5JCpjNkEYat6n2yMsahwMP1ZFJTN"
	tx, err := node.GetTx(ctx, signature)
	if err != nil {
		t.Fatalf("Error getting tx: %v", err)
	}

	transfers := solana.GetAllTransfers(tx)
	for _, tf := range transfers {
		log.Printf("Transfer: %+v", tf)
	}
	swaps := sh.HandleSwaps(ctx, transfers, tx, 0, 0)
	log.Printf("Swaps: %+v", swaps)
}

func blockTest(ctx context.Context, node *solana.Node, txHandler *solana.TxHandler, t *testing.T) {
	return
	const blockNum = 304413401

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

//2024/12/06 17:35:10 Transfer: {InnerIndex:3 IxIndex:0 ToUserAccount:7HeD6sLLqAnKVRuSfc1Ko3BSPMNKWgGTiWLKXJF31vKM ToTokenAccount: FromUserAccount:63x6REE8XMAZxMwVXjqu4uyg7T6imsVTXfsJvTimQshS FromTokenAccount: Mint:So11111111111111111111111111111111111111112 Decimals:0 Amount:9855000 Type:native ProgramId:}
//2024/12/06 17:35:10 Transfer: {InnerIndex:3 IxIndex:2 ToUserAccount:63x6REE8XMAZxMwVXjqu4uyg7T6imsVTXfsJvTimQshS ToTokenAccount:6LHLAPmYA9QMPLFhwQQM5u9F4pKGtRZfV8D1sBK9vqdZ FromUserAccount:Gspmj5gXGSbggZ3HnUUdYuiKDmtgt766oBAAaqGtQCsy FromTokenAccount:6KrVo5MqSawX4SsZyJBUW9e5qJfFudHCRkKdFZpYQR2E Mint:AvmYhCNVLodqorCC26c39jYnYLfwDbfqcknmzYjypump Decimals:6 Amount:19991859.717541 Type:token ProgramId:TokenkegQfeZyiNwAJbNbGKPFXCWuBvf9Ss623VQ5DA}
//2024/12/06 17:35:10 Transfer: {InnerIndex:3 IxIndex:3 ToUserAccount:Gspmj5gXGSbggZ3HnUUdYuiKDmtgt766oBAAaqGtQCsy ToTokenAccount: FromUserAccount:63x6REE8XMAZxMwVXjqu4uyg7T6imsVTXfsJvTimQshS FromTokenAccount: Mint:So11111111111111111111111111111111111111112 Decimals:0 Amount:1085145000 Type:native ProgramId:}
//2024/12/06 17:35:10 Transfer: {InnerIndex:3 IxIndex:4 ToUserAccount:CebN5WGQ4jvEPvsVU4EoHEpgzq1VV7AbicfhtW4xC9iM ToTokenAccount: FromUserAccount:63x6REE8XMAZxMwVXjqu4uyg7T6imsVTXfsJvTimQshS FromTokenAccount: Mint:So11111111111111111111111111111111111111112 Decimals:0 Amount:10851450 Type:native ProgramId:}
