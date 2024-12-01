package solana

import (
	"blocsy/internal/types"
	"context"
	"fmt"
	"log"
	"math/big"
	"strconv"
	"sync"
	"time"
)

const workers = 10

func NewTokenFinder(cache TockenCache, solSvc *SolanaService, repo TokensRepo) *TokenFinder {
	return &TokenFinder{
		cache:     cache,
		solSvc:    solSvc,
		repo:      repo,
		processor: nil,
	}
}

func (tf *TokenFinder) FindToken(ctx context.Context, address string, miss bool) (*types.Token, *[]types.Pair, error) {
	cachedToken, ok := tf.cache.GetToken(address)
	if ok && !miss {
		return cachedToken, nil, nil
	}

	token, pairs, err := tf.repo.LookupByToken(ctx, address, "solana")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup repo token: %w", err)
	}

	if token.Address != "" {
		tf.cache.PutToken(token.Address, token)
		return &token, &pairs, nil
	}

	newToken, err := tf.lookupToken(ctx, address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup token: %w", err)
	}

	if err := tf.repo.StoreToken(ctx, *newToken); err != nil {
		return nil, nil, fmt.Errorf("failed to store token: %w", err)
	}

	return newToken, &pairs, nil

}

func (tf *TokenFinder) lookupToken(ctx context.Context, address string) (*types.Token, error) {
	if address == "" {
		return nil, fmt.Errorf("address is empty")
	}
	metadata, err := tf.solSvc.GetMetadata(ctx, address)

	var (
		name, symbol string
		decimals     uint8
		supply       int64
	)
	if metadata == nil || err != nil {
		return nil, fmt.Errorf("failed to get token metadata %w", err)
	} else {
		name = metadata.Data.Name
		symbol = metadata.Data.Symbol
		//isMutable := metadata.IsMutable
		//updateAuthority := metadata.UpdateAuthority

		tokenData, err := tf.solSvc.GetTokenSupplyAndContext(ctx, address)
		if err != nil {
			return nil, fmt.Errorf("failed to get token supply and context: %w", err)
		}

		decimals = tokenData.Value.Decimals
		supply = int64(tokenData.Value.Amount)

	}

	parsedAmount, err := ParseTokenAmount(strconv.FormatUint(uint64(supply), 10), int(decimals))
	if err != nil {
		return nil, fmt.Errorf("failed to parse token amount: %w", err)
	}

	token := types.Token{
		Address:          address,
		Name:             name,
		Symbol:           symbol,
		Decimals:         decimals,
		Supply:           parsedAmount,
		Network:          "solana",
		CreatedBlock:     0,
		CreatedTimestamp: 0,
	}

	return &token, nil
}

func (tf *TokenFinder) NewTokenProcessor() {
	tf.processor = &TokenProcessor{
		queue: make(chan string, 1000),
		seen:  sync.Map{},
		wg:    sync.WaitGroup{},
	}

	for i := 0; i < workers; i++ {
		tf.processor.wg.Add(1)
		go tf.worker(i)
	}

}

func (tf *TokenFinder) worker(id int) {
	defer tf.processor.wg.Done()

	for token := range tf.processor.queue {
		timeOutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, _, _ = tf.FindToken(timeOutCtx, token, false)
		cancel()
		tf.processor.seen.Delete(token)
	}
}

func (tf *TokenFinder) AddToQueue(token string) {
	if tf.processor == nil {
		return
	}

	if _, exists := tf.processor.seen.LoadOrStore(token, struct{}{}); exists {
		return
	}

	select {
	case tf.processor.queue <- token:
	default:
		tf.processor.seen.Delete(token)
		log.Printf("Queue is full! Token %s discarded.", token)
	}
}

func ParseTokenAmount(tokenAmount string, d int) (string, error) {
	tokenAmountInt := new(big.Int)
	_, ok := tokenAmountInt.SetString(tokenAmount, 10)
	if !ok {
		return "", fmt.Errorf("invalid token amount")
	}

	decimals := new(big.Int).SetInt64(int64(d))
	scale := new(big.Float).SetInt(new(big.Int).Exp(big.NewInt(10), decimals, nil))

	tokenAmountFloat := new(big.Float).SetInt(tokenAmountInt)
	f := new(big.Float).Quo(tokenAmountFloat, scale)

	return f.Text('f', -1), nil
}
