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

	//log.Printf("Processing %s ~ got %d swaps", tx.Transaction.Signatures[0], len(swaps))

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
