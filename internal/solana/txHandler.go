package solana

import (
	"blocsy/cmd/api/websocket"
	"blocsy/internal/solana/dex"
	"blocsy/internal/types"
	"context"
)

func NewTxHandler(sh *SwapHandler, repo TokensAndPairsRepo, pRepo SwapsRepo, websocket *websocket.WebSocketServer) *TxHandler {
	return &TxHandler{
		sh:        sh,
		repo:      repo,
		pRepo:     pRepo,
		Websocket: websocket,
	}

}

func (t *TxHandler) ProcessTransaction(ctx context.Context, tx *types.SolanaTx, timestamp int64, block uint64, ignoreWS bool) ([]types.SwapLog, error) {
	transfers := GetAllTransfers(tx)
	logs := GetLogs(tx.Meta.LogMessages)
	swaps := t.sh.HandleSwaps(ctx, transfers, tx, timestamp, block)
	pumpFunTokens := dex.HandlePumpFunNewToken(logs, PUMPFUN)

	if len(pumpFunTokens) > 0 {
		go func() {
			for _, token := range pumpFunTokens {
				_ = t.repo.StoreToken(ctx, types.Token{
					Name:             token.Name,
					Symbol:           token.Symbol,
					Address:          token.Mint.String(),
					Decimals:         6,
					Supply:           "1000000000",
					Network:          "solana",
					CreatedBlock:     int64(block),
					CreatedTimestamp: uint64(timestamp),
				})

			}
		}()
	}

	if t.Websocket != nil && !ignoreWS {
		go func() {
			t.Websocket.BroadcastSwaps(swaps)
			if len(pumpFunTokens) > 0 {
				t.Websocket.BroadcastPumpFunTokens(pumpFunTokens)
			}
		}()
	}
	return swaps, nil
}
