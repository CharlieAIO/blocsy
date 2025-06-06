package solana

import (
	"blocsy/internal/types"
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
	"sync"
	"time"
)

func NewPairsService(cache PairsCache, tf *TokenFinder, solSvc *SolanaService, repo PairsRepo) *PairsService {
	return &PairsService{
		cache:       cache,
		tokenFinder: tf,
		solSvc:      solSvc,
		repo:        repo,

		processor: nil,
	}
}

func (ps *PairsService) FindPair(ctx context.Context, address string, token_ *string) (*types.Pair, *types.QuoteToken, error) {
	cachedPair, ok := ps.cache.GetPair(address)
	if ok {
		if cachedQT, ok := ps.cache.GetToken(address); ok {
			return cachedPair, &types.QuoteToken{
				Identifier: cachedPair.QuoteToken.Identifier,
				Name:       cachedQT.Name,
				Symbol:     cachedQT.Symbol,
				Address:    cachedPair.QuoteToken.Address,
				Decimals:   cachedQT.Decimals,
			}, nil
		}
	}

	//pair, _, quoteToken, _ := ps.repo.FindPair(ctx, address)
	pair, err := ps.repo.FindPair(ctx, address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find pair: %w", err)
	}
	if pair.Address == "" || pair.Token == "" || pair.QuoteToken.Address == "" {
		return nil, nil, fmt.Errorf("invalid pair found")
	}
	quoteToken, _, err := ps.tokenFinder.FindToken(ctx, pair.QuoteToken.Address, false)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find quote token: %w", err)
	}

	return pair, &types.QuoteToken{
		Identifier: pair.QuoteToken.Identifier,
		Name:       quoteToken.Name,
		Symbol:     quoteToken.Symbol,
		Address:    quoteToken.Address,
		Decimals:   quoteToken.Decimals,
	}, nil

}

func (ps *PairsService) lookupPair(ctx context.Context, address string, token_ *string) (types.Pair, error) {
	accInfo, err := ps.solSvc.GetAccountInfo(ctx, address)
	if err != nil {
		return types.Pair{}, fmt.Errorf("failed to get account info: %w", err)
	}

	owner := accInfo.Owner

	exchange, quoteToken, token, identifier, err := identifyPair(owner.String(), accInfo, token_)
	if err != nil {
		return types.Pair{}, fmt.Errorf("failed to identify pair: %w", err)
	}

	pair := types.Pair{
		Address:  address,
		Network:  "solana",
		Exchange: exchange,
		QuoteToken: types.QuoteTokenSimple{
			Identifier: identifier,
			Address:    quoteToken,
		},
		Token:            token,
		CreatedBlock:     0,
		CreatedTimestamp: time.Unix(0, 0),
	}

	return pair, nil
}

func (ps *PairsService) NewPairProcessor() {
	ps.processor = &PairProcessor{
		queue: make(chan PairProcessorQueue, 5000),
		seen:  sync.Map{},
		wg:    sync.WaitGroup{},
	}

	for i := 0; i < workers; i++ {
		ps.processor.wg.Add(1)
		go ps.worker(i)
	}

}

func (ps *PairsService) worker(id int) {
	defer ps.processor.wg.Done()

	for pair := range ps.processor.queue {
		timeOutCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		_, _, _ = ps.FindPair(timeOutCtx, pair.address, pair.token)
		cancel()
		ps.processor.seen.Delete(pair.address)
	}
}

func (ps *PairsService) AddToQueue(pair PairProcessorQueue) {
	if ps.processor == nil {
		return
	}

	if _, exists := ps.processor.seen.LoadOrStore(pair.address, struct{}{}); exists {
		return
	}

	select {
	case ps.processor.queue <- pair:
	default:
		ps.processor.seen.Delete(pair)
	}
}

func identifyPair(owner string, accInfo client.AccountInfo, token_ *string) (string, string, string, string, error) {
	var exchange, baseMint, tokenMint, baseMintIdentifier string

	if token_ == nil {
		token_ = new(string)
	}

	if owner == METEORA_DLMM_PROGRAM {
		met_ := types.MeteoraLayout{}
		err := met_.Decode(accInfo.Data)
		if err != nil {
			return "", "", "", "", err
		}
		exchange = "METEORA"
		baseMint = met_.TokenYMint.String()
		tokenMint = met_.TokenXMint.String()
		baseMintIdentifier = "tokenY"
	} else if owner == METEORA_POOLS_PROGRAM {
		metP_ := types.MeteoraPoolsLayout{}
		err := metP_.Decode(accInfo.Data)
		if err != nil {
			return "", "", "", "", err
		}
		exchange = "METEORA"
		baseMint = metP_.TokenBMint.String()
		tokenMint = metP_.TokenAMint.String()
		baseMintIdentifier = "tokenBMint"
	} else if owner == ORCA_WHIRL_PROGRAM_ID {
		whirl := types.OrcaWhirlpool{}
		err := whirl.Decode(accInfo.Data)
		if err != nil {
			return "", "", "", "", err
		}
		exchange = "ORCA"
		baseMint = whirl.TokenMintA.String()
		tokenMint = whirl.TokenMintB.String()
		baseMintIdentifier = "mintA"
	} else if owner == RAYDIUM_LIQ_POOL_V4 {
		rayV4 := types.RaydiumV4Layout{}
		err := rayV4.Decode(accInfo.Data)
		if err != nil {
			return "", "", "", "", err
		}
		exchange = "RAYDIUM"
		baseMint = rayV4.BaseMint.String()
		tokenMint = rayV4.QuoteMint.String()
		baseMintIdentifier = "baseMint"
	} else if owner == FLUXBEAM_PROGRAM {
		fbPool := types.FluxBeamPool{}
		fbPool.Decode(accInfo.Data)
		exchange = "FLUXBEAM"
		baseMint = fbPool.MintA.String()
		tokenMint = fbPool.MintB.String()
		baseMintIdentifier = "mintA"
	} else if owner == PUMPFUN {
		exchange = "PUMPFUN"
		baseMint = "So11111111111111111111111111111111111111112"
		tokenMint = *token_
		baseMintIdentifier = "N/A"
	} else {
		return "", "", "", "", fmt.Errorf("unknown program owner: %s", owner)
	}

	if _, isBaseMint := QuoteTokens[baseMint]; !isBaseMint {
		if _, isTokenMint := QuoteTokens[tokenMint]; isTokenMint {
			baseMint, tokenMint = tokenMint, baseMint
			baseMintIdentifier = "mintB"
		}
	}

	return exchange, baseMint, tokenMint, baseMintIdentifier, nil
}
