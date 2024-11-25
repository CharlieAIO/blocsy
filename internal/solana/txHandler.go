package solana

import (
	"blocsy/cmd/api/websocket"
	"blocsy/internal/types"
	"context"
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
	transfers := GetAllTransfers(tx)
	swaps := t.sh.HandleSwaps(ctx, transfers, tx, timestamp, block)
	if t.Websocket != nil && !ignoreWS {
		go func() {
			t.Websocket.BroadcastSwaps(swaps)
		}()
	}
	return swaps, nil
}
