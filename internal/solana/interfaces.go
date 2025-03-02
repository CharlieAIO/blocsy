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

//Mongo
//type PairsRepo interface {
//	StorePair(ctx context.Context, pair types.Pair) error
//	LookupByPair(ctx context.Context, address string, network string, tf TokenFinder) (types.Pair, types.Token, types.QuoteToken, error)
//}

type TockenCache interface {
	PutToken(tokenAddress string, tokenData types.Token)
	GetToken(tokenAddress string) (*types.Token, bool)
}

//Mongo
//type TokensRepo interface {
//	LookupByToken(ctx context.Context, address string, network string) (types.Token, []types.Pair, error)
//	StoreToken(ctx context.Context, token types.Token) error
//	UpdateTokenSupply(ctx context.Context, address string, supply string) error
//}

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

type TokensRepo interface {
	InsertToken(ctx context.Context, token types.Token) error
	FindToken(ctx context.Context, address string) (*types.Token, error)
	UpdateTokenSupply(ctx context.Context, address string, supply float64) error
}

type PairsRepo interface {
	InsertPair(ctx context.Context, pair types.Pair) error
	FindPair(ctx context.Context, address string) (*types.Pair, error)
	FindPairsByToken(ctx context.Context, token string) ([]*types.Pair, error)
}

type SolanaTokenFinder interface {
	FindToken(ctx context.Context, address string, miss bool) (*types.Token, *[]types.Pair, error)
	AddToQueue(address string)
	AddToMintBurnQueue(token string, amount string, type_ string)
}

type SolanaPairFinder interface {
	FindPair(ctx context.Context, address string, token_ *string) (*types.Pair, *types.QuoteToken, error)
	AddToQueue(pair PairProcessorQueue)
}
