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

// TokenPnlHandler godoc
//
//	@Summary		Lookup token pnl for a wallet
//	@Description	Returns all token pnl for a wallet
//
//	@Security		ApiKeyAuth
//
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			wallet		path		string	true	"Wallet address"
//	@Param			timeframe	query		string	false	"Timeframe (1d, 7d, 14d, 30d)"
//	@Param			page		query		int		false	"Page number"
//	@Success		200			{object}	types.TokenPNLResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/token-pnl/{wallet} [get]
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

	// Group swaps by pair.
	tokenSwaps := make(map[string][]types.SwapLog)
	for _, swap := range swaps {
		tokenSwaps[swap.Token] = append(tokenSwaps[swap.Token], swap)
	}

	var results []types.TokenAndPnl
	var resultsMu sync.Mutex
	priceCache := make(map[string]float64)
	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 100) // limit to 100 concurrent goroutines

	for token, swapLogs := range tokenSwaps {
		wg.Add(1)
		sem <- struct{}{}

		go func(token string, swapLogs []types.SwapLog) {
			defer wg.Done()
			defer func() { <-sem }()

			var pair = swapLogs[0].Pair

			var quoteTokenAddress string
			if swapLogs[0].Source == "PUMPFUN" {
				quoteTokenAddress = "SOL"
			} else {
				start := time.Now()
				_, qt, err := h.pairFinder.FindPair(ctx, pair, nil)
				if err != nil {
					quoteTokenAddress = "SOL"
				} else {
					quoteTokenAddress = qt.Address // use the token address instead of symbol
				}
				log.Printf("%s | findPair took %s", pair, time.Since(start))
			}

			if quoteTokenAddress == "" {
				log.Printf("No token address for pair: %s", pair)
				return
			}

			mu.Lock()
			usdPrice, ok := priceCache[quoteTokenAddress]
			if !ok {
				usdPrice = h.pricer.GetUSDPrice(quoteTokenAddress)
				if usdPrice > 0 {
					priceCache[quoteTokenAddress] = usdPrice
				} else {
					log.Printf("Missing price for token: %s", quoteTokenAddress)
					mu.Unlock()
					return
				}
			}
			mu.Unlock()

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

			realizedPNL := new(big.Float)
			if totalSellTokens.Cmp(big.NewFloat(0)) != 0 {
				realizedPNL.Sub(totalSellValue, totalBuyValue)
			} else {
				realizedPNL.SetFloat64(0)
			}

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

			var pnlResults types.TokenPnL
			realizedPNLFloatUSD, _ := realizedPNL.Float64()
			pnlResults.RealizedPnLUSD = realizedPNLFloatUSD

			unrealizedPNLFloatUSD, _ := unrealizedPNL.Float64()
			pnlResults.UnrealizedPnLUSD = unrealizedPNLFloatUSD

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
			pnlResults.TotalTrades = len(swapLogs)

			result := types.TokenAndPnl{
				Token: swapLogs[0].Token,
				PnL:   pnlResults,
			}
			resultsMu.Lock()
			results = append(results, result)
			resultsMu.Unlock()
		}(token, swapLogs)
	}

	wg.Wait()

	pageStr := r.URL.Query().Get("page")
	pageNum := 1
	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			pageNum = p
		}
	}
	pageSize := 100

	sort.Slice(results, func(i, j int) bool {
		return results[i].Token < results[j].Token
	})

	totalResults := len(results)
	totalPages := (totalResults + pageSize - 1) / pageSize
	startIndex := (pageNum - 1) * pageSize
	endIndex := startIndex + pageSize
	if startIndex > totalResults {
		startIndex = totalResults
	}
	if endIndex > totalResults {
		endIndex = totalResults
	}
	paginatedResults := results[startIndex:endIndex]

	response := types.TokenPNLResponse{
		Tokens: paginatedResults,
		Pagination: types.Pagination{
			Page:       pageNum,
			PageSize:   pageSize,
			Total:      totalResults,
			TotalPages: totalPages,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}
