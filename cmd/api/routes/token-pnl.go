package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"math/big"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

func (h *Handler) TokenPnlHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	wallet := chi.URLParam(r, "wallet")
	timeframe := r.URL.Query().Get("timeframe")
	var days int
	switch timeframe {
	case "1d":
		days = 1
	case "7d":
		days = 7
	case "14d":
		days = 14
	case "30d":
		days = 30
	default:
		days = 30
	}

	startDate := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -days)
	swaps, err := h.swapsRepo.GetSwapsOnDate(ctx, wallet, startDate)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if len(swaps) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"tokens": []interface{}{},
			"pagination": map[string]interface{}{
				"page":       1,
				"pageSize":   100,
				"total":      0,
				"totalPages": 0,
			},
		})
		return
	}

	// Group swaps by pair as they are stored
	pairSwaps := make(map[string][]types.SwapLog)
	for _, swap := range swaps {
		pairSwaps[swap.Pair] = append(pairSwaps[swap.Pair], swap)
	}

	// We'll aggregate PnL for each token (by token address)
	individualTokenPnL := make(map[string]types.AggregatedPnL)
	priceCache := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 100) // limit to 100 concurrent goroutines

	for pair, swapLogs := range pairSwaps {
		wg.Add(1)
		sem <- struct{}{}

		go func(pair string, swapLogs []types.SwapLog) {
			defer wg.Done()
			defer func() { <-sem }()

			// Determine the token address for this pair.
			var tokenAddress string
			if swapLogs[0].Source == "PUMPFUN" {
				tokenAddress = "SOL"
			} else {
				start := time.Now()
				_, qt, err := h.pairFinder.FindPair(ctx, pair, nil)
				if err != nil {
					tokenAddress = "SOL"
				} else {
					tokenAddress = qt.Address // use the token address instead of symbol
				}
				log.Printf("%s | findPair took %s", pair, time.Since(start))
			}

			if tokenAddress == "" {
				log.Printf("No token address for pair: %s", pair)
				return
			}

			// Get USD price for the token (using a cache)
			mu.Lock()
			usdPrice, ok := priceCache[tokenAddress]
			if !ok {
				usdPrice = h.pricer.GetUSDPrice(tokenAddress)
				if usdPrice > 0 {
					priceCache[tokenAddress] = usdPrice
				} else {
					log.Printf("Missing price for token: %s", tokenAddress)
					mu.Unlock()
					return
				}
			}
			mu.Unlock()

			// Aggregate swap values for this pair
			totalBuyTokens := new(big.Float)
			totalSellTokens := new(big.Float)
			totalBuyValue := new(big.Float)
			totalSellValue := new(big.Float)

			for _, swap := range swapLogs {
				amountOutFloat := new(big.Float).SetFloat64(swap.AmountOut)
				amountInFloat := new(big.Float).SetFloat64(swap.AmountIn)
				if swap.Action == "BUY" {
					totalBuyTokens.Add(totalBuyTokens, amountInFloat)
					totalBuyValue.Add(totalBuyValue, new(big.Float).Mul(amountOutFloat, big.NewFloat(usdPrice)))
				} else if swap.Action == "SELL" {
					totalSellTokens.Add(totalSellTokens, amountOutFloat)
					totalSellValue.Add(totalSellValue, new(big.Float).Mul(amountInFloat, big.NewFloat(usdPrice)))
				}
			}

			// Calculate realized PnL
			realizedPNL := new(big.Float)
			if totalSellTokens.Cmp(big.NewFloat(0)) != 0 {
				realizedPNL.Sub(totalSellValue, totalBuyValue)
			} else {
				realizedPNL.SetFloat64(0)
			}

			// Calculate unrealized PnL based on remaining tokens
			unrealizedPNL := new(big.Float)
			remainingAmount := new(big.Float).Sub(totalBuyTokens, totalSellTokens)
			if remainingAmount.Cmp(big.NewFloat(0)) > 0 {
				mostRecentPrice := new(big.Float)
				mostRecentSwap, err := h.swapsRepo.FindLatestSwap(ctx, pair)
				if err == nil && len(mostRecentSwap) > 0 {
					amountOutFloat := new(big.Float).SetFloat64(mostRecentSwap[0].AmountOut)
					amountInFloat := new(big.Float).SetFloat64(mostRecentSwap[0].AmountIn)
					if mostRecentSwap[0].Action == "BUY" {
						mostRecentPrice = new(big.Float).Quo(amountOutFloat, amountInFloat)
					} else {
						mostRecentPrice = new(big.Float).Quo(amountInFloat, amountOutFloat)
					}
				}
				unrealizedPNL.Mul(remainingAmount, mostRecentPrice)
			}

			var pnlResults types.AggregatedPnL
			realizedPNLFloatUSD, _ := realizedPNL.Float64()
			pnlResults.RealizedPnLUSD = realizedPNLFloatUSD

			unrealizedPNLFloatUSD, _ := unrealizedPNL.Float64()
			pnlResults.UnrealizedPnLUSD = unrealizedPNLFloatUSD

			// Compute ROI values
			if totalSellValue.Cmp(big.NewFloat(0)) > 0 {
				realizedROI := new(big.Float).Quo(realizedPNL, totalSellValue)
				realizedROIFloat, _ := realizedROI.Float64()
				pnlResults.RealizedROI = realizedROIFloat * 100
			}
			if remainingAmount.Cmp(big.NewFloat(0)) > 0 {
				avgBuyPrice := new(big.Float).Quo(totalBuyValue, totalBuyTokens)
				remainingCost := new(big.Float).Mul(avgBuyPrice, remainingAmount)
				if remainingCost.Cmp(big.NewFloat(0)) > 0 {
					unrealizedROI := new(big.Float).Quo(unrealizedPNL, remainingCost)
					unrealizedROIFloat, _ := unrealizedROI.Float64()
					pnlResults.UnrealizedROI = unrealizedROIFloat * 100
				}
			}
			if totalBuyValue.Cmp(big.NewFloat(0)) > 0 {
				totalROI := new(big.Float).Quo(new(big.Float).Add(realizedPNL, unrealizedPNL), totalBuyValue)
				finalROI, _ := totalROI.Float64()
				pnlResults.ROI = finalROI * 100
			}

			pnlResults.PnLUSD = pnlResults.RealizedPnLUSD + pnlResults.UnrealizedPnLUSD
			pnlResults.TokensTraded = len(swapLogs)
			if pnlResults.TokensTraded > 0 {
				if realizedPNL.Cmp(big.NewFloat(0)) > 0 {
					pnlResults.WinRate = 100
				} else {
					pnlResults.WinRate = 0
				}
			}

			// Aggregate results by token address
			mu.Lock()
			existing := individualTokenPnL[tokenAddress]
			existing.RealizedPnLUSD += pnlResults.RealizedPnLUSD
			existing.UnrealizedPnLUSD += pnlResults.UnrealizedPnLUSD
			existing.RealizedROI += pnlResults.RealizedROI
			existing.UnrealizedROI += pnlResults.UnrealizedROI
			if existing.ROI == 0 {
				existing.ROI = pnlResults.ROI
			} else {
				existing.ROI = (existing.ROI + pnlResults.ROI) / 2
			}
			existing.PnLUSD += pnlResults.PnLUSD
			existing.TokensTraded += pnlResults.TokensTraded
			if pnlResults.WinRate > 0 {
				existing.WinRate = 100
			}
			individualTokenPnL[tokenAddress] = existing
			mu.Unlock()

		}(pair, swapLogs)
	}

	wg.Wait()

	// Pagination: up to 100 token pnls per page.
	// Read page from query parameters (defaulting to page 1).
	pageStr := r.URL.Query().Get("page")
	pageNum := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	pageSize := 100

	// Convert the map to a slice for consistent ordering.
	type tokenPnl struct {
		Token string              `json:"token"`
		PnL   types.AggregatedPnL `json:"pnl"`
	}
	tokens := make([]tokenPnl, 0, len(individualTokenPnL))
	for token, pnl := range individualTokenPnL {
		tokens = append(tokens, tokenPnl{
			Token: token,
			PnL:   pnl,
		})
	}

	// Sort tokens by token address.
	sort.Slice(tokens, func(i, j int) bool {
		return tokens[i].Token < tokens[j].Token
	})

	totalTokens := len(tokens)
	totalPages := (totalTokens + pageSize - 1) / pageSize
	startIndex := (pageNum - 1) * pageSize
	endIndex := startIndex + pageSize
	if startIndex > totalTokens {
		startIndex = totalTokens
	}
	if endIndex > totalTokens {
		endIndex = totalTokens
	}
	paginatedTokens := tokens[startIndex:endIndex]

	response := map[string]interface{}{
		"tokens": paginatedTokens,
		"pagination": map[string]interface{}{
			"page":       pageNum,
			"pageSize":   pageSize,
			"total":      totalTokens,
			"totalPages": totalPages,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}
