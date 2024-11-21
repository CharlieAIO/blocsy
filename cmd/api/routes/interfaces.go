package routes

import (
	"blocsy/internal/types"
	"context"
	"time"
)

type SolanaTokenFinder interface {
	FindToken(ctx context.Context, address string) (*types.Token, error)
}

type SolanaPairFinder interface {
	FindPair(ctx context.Context, address string) (*types.Pair, *types.QuoteToken, error)
}

type PriceTrackers interface {
	GetUSDPrice(symbol string) float64
}

type SwapsRepo interface {
	GetAllWalletSwaps(ctx context.Context, wallet string) ([]types.SwapLog, error)
	GetSwapsOnDate(ctx context.Context, wallet string, startDate time.Time) ([]types.SwapLog, error)
}
