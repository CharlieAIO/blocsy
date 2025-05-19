package solana

import (
	"blocsy/internal/types"
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"github.com/blocto/solana-go-sdk/rpc"
	"log"
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
		return cachedToken, &[]types.Pair{}, nil
	}

	token, err := tf.repo.FindToken(ctx, address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup repo token: %w", err)
	}

	var pairs []types.Pair

	if token.Address != "" && token.Name != "" && token.Symbol != "" {
		tf.cache.PutToken(token.Address, *token)
		return token, &pairs, nil
	}

	metadata, err := tf.solSvc.GetMetadata(ctx, address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to lookup token metadata: %w", err)
	}

	if token.Supply == "" || token.Supply == "0" {
		tokenSupply := rpc.ValueWithContext[client.TokenAmount]{}
		tokenSupply, err = tf.solSvc.GetTokenSupplyAndContext(ctx, address)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to lookup token supply: %w", err)
		}
		amount := strconv.FormatUint(tokenSupply.Value.Amount, 10)
		err = tf.repo.UpdateTokenSupply(ctx, address, amount, "mint")
		if err != nil {
			return nil, nil, err
		}

	}

	if metadata.Name != "" && metadata.Symbol != "" {
		err = tf.repo.UpdateTokenInfo(ctx, address, metadata)
		if err != nil {
			return nil, nil, err
		}
		token, err = tf.repo.FindToken(ctx, address)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to lookup repo token: %w", err)
		}
		tf.cache.PutToken(token.Address, *token)
		return token, &pairs, nil
	}

	return nil, nil, fmt.Errorf("token not found")

}

func (tf *TokenFinder) NewTokenProcessor() {
	tf.processor = &TokenProcessor{
		queue: make(chan string, 5000),
		seen:  sync.Map{},
		wg:    sync.WaitGroup{},
	}

	for i := 0; i < workers; i++ {
		tf.processor.wg.Add(1)
		go tf.worker()
	}

}

func (tf *TokenFinder) worker() {
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
