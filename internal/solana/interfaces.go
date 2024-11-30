package solana

import (
	"blocsy/internal/types"
	"context"
)

type TokensAndPairsRepo interface {
	TokensRepo
	PairsRepo
}

type PairsCache interface {
	TockenCache
	PutPair(pairAddress string, pairData types.Pair)
	GetPair(pairAddress string) (*types.Pair, bool)
}

type PairsRepo interface {
	StorePair(ctx context.Context, pair types.Pair) error
	LookupByPair(ctx context.Context, address string, network string, tf TokenFinder) (types.Pair, types.Token, types.QuoteToken, error)
}

type TockenCache interface {
	PutToken(tokenAddress string, tokenData types.Token)
	GetToken(tokenAddress string) (*types.Token, bool)
}

type TokensRepo interface {
	LookupByToken(ctx context.Context, address string, network string) (types.Token, []types.Pair, error)
	StoreToken(ctx context.Context, token types.Token) error
}

type TxCacher interface {
	GetTx(string) bool
	PutTx(string)
}

type SwapsRepo interface {
	MarkBlockProcessed(ctx context.Context, blockNumber int) error
	InsertSwaps(ctx context.Context, swap []types.SwapLog) error
	DeleteSwapsUsingTx(ctx context.Context, signature string) error
	FindMissingBlocks(ctx context.Context) ([][]int, error)
}

type SolanaTokenFinder interface {
	FindToken(ctx context.Context, address string) (*types.Token, error)
	AddToQueue(address string)
}

type SolanaPairFinder interface {
	FindPair(ctx context.Context, address string, token_ *string) (*types.Pair, *types.QuoteToken, error)
	AddToQueue(pair PairProcessorQueue)
}
