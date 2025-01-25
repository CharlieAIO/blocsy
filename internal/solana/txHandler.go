package solana

import (
	"blocsy/cmd/api/websocket"
	"blocsy/internal/solana/dex"
	"blocsy/internal/types"
	"context"
	"log"
)

func NewTxHandler(sh *SwapHandler, solSvc *SolanaService, repo TokensAndPairsRepo, pRepo SwapsRepo, websocket *websocket.WebSocketServer) *TxHandler {
	return &TxHandler{
		sh:     sh,
		solSvc: solSvc,
		repo:   repo,
		pRepo:  pRepo,

		Websocket: websocket,
	}

}

func (t *TxHandler) ProcessTransaction(ctx context.Context, tx *types.SolanaTx, timestamp int64, block uint64, ignoreWS bool) ([]types.SwapLog, error) {
	transfers, burns, mints, tokensCreated := ParseTransaction(tx)
	logs := GetLogs(tx.Meta.LogMessages)
	swaps := t.sh.HandleSwaps(ctx, transfers, tx, timestamp, block)
	pumpFunTokens := dex.HandlePumpFunNewToken(logs, PUMPFUN)

	go func() {
		for _, burn := range burns {
			t.sh.tf.AddToMintBurnQueue(burn.Mint, burn.Amount, "burn")
		}
		for _, mint := range mints {
			t.sh.tf.AddToMintBurnQueue(mint.Mint, mint.Amount, "mint")
		}

		pumpFunTokenMints := make(map[string]bool)

		for _, pfToken := range pumpFunTokens {
			deployer := pfToken.User.String()
			pumpFunTokenMints[pfToken.Mint.String()] = true

			_ = t.repo.StoreToken(ctx, types.Token{
				Name:             pfToken.Name,
				Symbol:           pfToken.Symbol,
				Decimals:         6,
				Network:          "solana",
				CreatedBlock:     int64(block),
				CreatedTimestamp: uint64(timestamp),
				Supply:           "1000000000",
				Deployer:         &deployer,
				Metadata:         &pfToken.Uri,
			})
		}

		for _, token := range tokensCreated {
			if _, exists := pumpFunTokenMints[token.Address]; exists {
				continue
			}
			token.CreatedTimestamp = uint64(timestamp)
			token.CreatedBlock = int64(block)
			_ = t.repo.StoreToken(ctx, token)
		}
	}()

	if t.Websocket != nil && !ignoreWS {
		go func() {
			t.Websocket.BroadcastSwaps(swaps)
			if len(pumpFunTokens) > 0 {
				log.Printf("Broadcasting pumpfun tokens")
				t.Websocket.BroadcastPumpFunTokens(pumpFunTokens)
			}
		}()
	}
	return swaps, nil
}
