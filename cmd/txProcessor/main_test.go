package main

import (
	"blocsy/internal/cache"
	"blocsy/internal/db"
	"blocsy/internal/solana"
	"blocsy/internal/solana/dex"
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

	solSvc := solana.NewSolanaService(ctx)

	tf := solana.NewTokenFinder(c, solSvc, pRepo)
	pf := solana.NewPairsService(c, tf, solSvc, pRepo)
	sh := solana.NewSwapHandler(tf, pf)

	nodeUrl := strings.Split(os.Getenv("SOL_HTTPS_BACKFILL_NODES"), ",")[0]

	node := solana.NewNode("test", nodeUrl)

	tests := []TestCase{
		{Target: 8, Signature: "3Ry66VFMKtLUxwG1FFbAdz2RMZbxdwmDGXjMmtkWcA2PKTuVuQhU8RVMZUw1duciR8eeZLYHYD5tGCy5LVz1wa4f"},
		{Target: 1, Signature: "2jRqSN5Ckx9sgwBoTkgQ7xz2fZb4TTftdz69noGFESx7eJW6Jxt1s4HP1fce6vzF8rBgzehYD3KSKQoKYWVRUotC"},
		{Target: 3, Signature: "43skCh7Z9paKpE1Hv52QQ3hqDzEwgtzDmQL77DtpeUon7zJUQFuemwHUPcZhiT5cBp84vdz3Azo2cxn8fTyLyHZV"},
		{Target: 1, Signature: "3Y2D8SYBBDnYYAwx9ZPasbySnB5AbofpUL1WCZzxVgMZULkCDpgBBChNs4mS8MFGjPUjzMZgtk8CnvyA2xhFk6GU"},
		{Target: 3, Signature: "2fbQoAurQTKEM2v3uYGuxJ5LQkucB5diFeMvZqGZ1ShLB5kixPzLS3Wnf7UCrZBFfSBYGRaySvSK1sgxApJWQR2g"},
		{Target: 1, Signature: "5F8W4cxpYLM6JK9ASorp2YkcQzCFYcgbRDjA2PXqf6T59QN1iBcUHVycEtvcE7xtnKuYzf3HetTqVZf65eWcntQ9"},
		{Target: 1, Signature: "K9PFgioRJRn6wgNjTwyW9j8Gtss6juwzvzkwsvXPHo9SUPACG6AexmSx2ZqfytELXgsopfsLEFfZ5AneNsvWc2F"},
		{Target: 1, Signature: "5RaAG259hpsjEF3pETFBhCxqosvfSz4T4H7REs7KhPKXg8ZKhLAEk3M1uhqsXfKeCqHCT7JkrN6cVmiAPDaGbBU"},
		{Target: 1, Signature: "2Wq5SoLz1LrrP4QYEgUKozFrq5VTTENE7JLZCWC2ZGMfqoPE4xZGygijZoihdFbLhXxQYGoB6unZqTX2H4qp5dtZ"},
		{Target: 1, Signature: "FEkv16o2Az1JYChY758dP2NnemMTnEXNjQT8risU9stcS3asv8UTHvXPtzLYox3BzkwU49rQoRwWYSp6frkaNH1"},
		{Target: 0, Signature: "5V9ew9x15GwpnSW35BGqVJxvS5bjJx2gK6xDzbniYcC5ac4yb4iQHCKbf7cA7EJtEosj2AqmFQkk3ujwqaFeQVQ1"},
		{Target: 0, Signature: "4MvvtBC8589EWzfWM7oLV9omahn3uqaAU7a3k8Xer4FKoMwfpSL9R2VBAdPtxJVLxwJNqksbp4DJUZZC33nvKou1"},
		{Target: 0, Signature: "2vy8cNnzNxReHeKjz5JxRWJcSEqCcRsYdFtWFoNTorsB2bX17ETXtVci4pDqZktVQNZnFyB7K3dDnv9ewkzsyhAT"},
		{Target: 4, Signature: "aDH8MUnMg34diGmHU5459T8Tpw7XsvAu6ajRaNQuP4DvAAeXcPYK1q9eDoXWRaFkRMMzriXnibAkKroQ461Uixr"},
		{Target: 0, Signature: "4C6rbHs5dZHVNX27Mz1uCZC7TXoq7eFLUUnrtD2SwfgRUQfzcMQsRC2GtbHg1PM6AWptad8S67AicDfnSRnkWQgZ"},
		{Target: 1, Signature: "5kvxhSfoK2FRhVtovvA7QY6rQ6DeiR5qA1XXYL1463EYZGeyvrtoA5v92SLVCv4G86R5EkWBQcz3MzrMTGx6wdh2"},
		{Target: 2, Signature: "65P1vYhqptArJDzetniitQqxzduGVLv8KFVXvGifjo8eWbrDHz3Bs8FBUQBkqVpjEPbmyV85sBqqJTPBDPLy4P8A"},
		{Target: 1, Signature: "3LmwWyTdynRQktorcfDU67wVycrXcxV2FrUJ6zEwJ2va3WVbRCZmfP8BnFyjcStk7y1AGtaCVWcqJktYa2knuBcT"},
		{Target: 1, Signature: "4WYKb2q7DLoVg6direSixARb1CpzXwmGA4To3yd4zsBhMaPmxFXpCTnvtM2UdAjjkvzXQyTuf1NUmh4jCye5xZvA"},
		{Target: 0, Signature: "31b9bX8kaFPJizGx3VkGqWX6juyNWpBSeE2BJxQXALovVRpVAeF1k6p4ZB7jb1QgctcLyoSWpcqD1VXTK4niRmS4"},
		{Target: 1, Signature: "3vr5TJo3d5TUxqECYkKfejg9rNcvh3xeaVcTX2XRcC1sqLXVMtMNeWqAkC8zSnE8j8tiU4TcD21UNwtkXhAUPmp"},
		{Target: 1, Signature: "3WugKiaFfhawNVSQwycW7oYrruFEoyzQoHYupvRWhTSyVcScNAJFzcpcQZL2ahL4y9SKypimjC6d3WqWnm4X9Z9w"},
		{Target: 1, Signature: "2ev694arceQ3HtSCdFKj7UJRZmy8sL2cMGbeme8hvptTCnSJv2p2SoeQRGt3vjbYEHbsu29N6WiMA2ryrW1ADSiF"},
		{Target: 2, Signature: "63Xjzbwc2mczzSUsDRnEYtVFr1Das5GwtUrErTAy492J8ZPS3XjhDxFNVyYYJ4DtRquErTCQZg6iBAY68a4nP6se"},
		{Target: 1, Signature: "3FoXqwcUdSGAbMtZpD9NtcB2dyYkGRoLa2PHJcnxWhY5afbgC5bcZS5hyWbXfQAqbLn3SN78srcTxNK7mKjJEKWS"},
		{Target: 1, Signature: "28rHU4GAqtjrnnZNPCguU21Ak7MdJ1UJhEGMZFGZixgsL5cDGwAQmQenaqK1VkVoW4na3qNvGVQh9vQmDZg3m88V"},
	}

	for _, tc := range tests {
		t.Run(tc.Signature, func(t *testing.T) {
			test_tx(ctx, node, sh, t, tf, tc.Signature, tc.Target)
		})
	}

}

func test_tx(ctx context.Context, node *solana.Node, sh *solana.SwapHandler, t *testing.T, tf *solana.TokenFinder, signature string, target int) {
	tx, err := node.GetTx(ctx, signature)
	if err != nil {
		t.Fatalf("Error getting tx: %v", err)
	}
	transfers, _, _, _ := solana.ParseTransaction(tx)
	logs := solana.GetLogs(tx.Meta.LogMessages)
	pumpFunTokens := dex.HandlePumpFunNewToken(logs, solana.PUMPFUN)
	log.Printf("PumpFun Tokens: %+v", pumpFunTokens)

	swaps := sh.HandleSwaps(ctx, transfers, tx, 0, 0)
	//for i, transfer := range transfers {
	//	t.Logf("Transfer %d: %+v", i, transfer)
	//}

	if len(swaps) != target {
		for i, transfer := range transfers {
			t.Logf("Transfer %d: %+v", i, transfer)
		}

		for i, swap := range swaps {
			t.Logf("Swap %d: %+v", i, swap)
		}

		t.Fatalf("Expected %d from HandleSwaps, got %d", target, len(swaps))
	}

}
