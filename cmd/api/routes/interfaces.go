package routes

import (
	"blocsy/internal/types"
	"context"
	"time"
)

type SolanaTokenFinder interface {
	FindToken(ctx context.Context, address string, miss bool) (*types.Token, *[]types.Pair, error)
}

type SolanaPairFinder interface {
	FindPair(ctx context.Context, address string, token_ *string) (*types.Pair, *types.QuoteToken, error)
}

type PriceTrackers interface {
	GetUSDPrice(symbol string) float64
}

type SwapsRepo interface {
	GetAllWalletSwaps(ctx context.Context, wallet string) ([]types.SwapLog, error)
	GetSwapsOnDate(ctx context.Context, wallet string, startDate time.Time) ([]types.SwapLog, error)
	FindSwap(ctx context.Context, timestamp int64, token string, amount float64) (*types.SwapLog, error)
	FindFirstTokenSwaps(ctx context.Context, token string) ([]types.SwapLog, error)
}

type Node interface {
	GetTx(ctx context.Context, hash string) (*types.SolanaTx, error)
	GetParsedLogs(rawLogs []string) []types.LogDetails
}
