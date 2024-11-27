package solana

import (
	"blocsy/internal/types"
	"context"
	"fmt"
	"github.com/blocto/solana-go-sdk/client"
)

func NewPairsService(cache PairsCache, tf *TokenFinder, solSvc *SolanaService, repo PairsRepo) *PairsService {
	return &PairsService{
		cache:       cache,
		tokenFinder: tf,
		solSvc:      solSvc,
		repo:        repo,
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

	pair, _, quoteToken, _ := ps.repo.LookupByPair(ctx, address, "solana", *ps.tokenFinder)

	if pair.Address != "" && quoteToken.Address != "" && pair.Token != "" {
		ps.cache.PutPair(pair.Address, pair)
		return &pair, &quoteToken, nil
	}

	pair, err := ps.lookupPair(ctx, address, token_)
	if err != nil {
		return nil, nil, fmt.Errorf("(%s) failed to lookup pair: %w", address, err)
	}

	if err := ps.repo.StorePair(ctx, pair); err != nil {
		return nil, nil, fmt.Errorf("failed to add pair: %w", err)
	}

	if _, found := QuoteTokens[pair.QuoteToken.Address]; !found {
		return nil, nil, fmt.Errorf("unsupported quote token: %s", pair.QuoteToken.Address)
	}

	quoteTokenLookup, err := ps.tokenFinder.FindToken(ctx, pair.QuoteToken.Address)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to find quote token: %w", err)
	}

	if quoteTokenLookup == nil {
		return nil, nil, nil
	}

	return &pair, &types.QuoteToken{
		Identifier: pair.QuoteToken.Identifier,
		Name:       quoteTokenLookup.Name,
		Symbol:     quoteTokenLookup.Symbol,
		Address:    pair.QuoteToken.Address,
		Decimals:   quoteTokenLookup.Decimals,
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
		CreatedTimestamp: 0,
	}

	return pair, nil
}

func identifyPair(owner string, accInfo client.AccountInfo, token_ *string) (string, string, string, string, error) {
	var exchange, baseMint, tokenMint, baseMintIdentifier string

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
