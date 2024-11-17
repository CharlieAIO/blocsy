package solana

import (
	"context"
	"defi-intel/cmd/api/websocket"
	"defi-intel/internal/types"
)

func NewTxHandler(sh *SwapHandler, solCli *SolanaService, repo TokensAndPairsRepo, pRepo SwapsRepo, websocket *websocket.WebSocketServer) *TxHandler {
	return &TxHandler{
		sh:     sh,
		solCli: solCli,
		repo:   repo,
		pRepo:  pRepo,

		Websocket: websocket,
	}

}

func (t *TxHandler) ProcessTransaction(ctx context.Context, tx *types.SolanaTx, timestamp int64, block uint64) ([]types.SwapLog, error) {
	transfers := GetAllTransfers(tx)
	swaps := t.sh.HandleSwaps(ctx, transfers, tx, timestamp, block)
	if t.Websocket != nil {
		go func() {
			t.Websocket.BroadcastSwaps(swaps)
		}()
	}
	return swaps, nil
}
