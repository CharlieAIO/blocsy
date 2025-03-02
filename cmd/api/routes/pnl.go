package routes

import (
	"blocsy/internal/types"
	"github.com/goccy/go-json"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) AggregatedPnlHandler(w http.ResponseWriter, r *http.Request) {
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
			"results": types.AggregatedPnL{},
		})
		return
	}

	//pnlResults := types.AggregatedPnL{}
	priceCache := make(map[string]float64)
	pairSwaps := make(map[string][]types.SwapLog)
	tokensTraded := make(map[string]bool)
	winCount := 0

	totalInvestment := new(big.Float)
	weightedROINumerator := new(big.Float)

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 100) // Limit to 100 goroutines

	for _, swap := range swaps {
		pairSwaps[swap.Pair] = append(pairSwaps[swap.Pair], swap)
	}

	individualPairPnL := make(map[string]types.AggregatedPnL)

	for pair, swapLogs := range pairSwaps {
		wg.Add(1)
		sem <- struct{}{}

		pnlResults := types.AggregatedPnL{}

		go func(pair string, swapLogs []types.SwapLog) {
			defer wg.Done()
			defer func() { <-sem }()

			var quoteTokenSymbol string
			if swapLogs[0].Source == "PUMPFUN" {
				quoteTokenSymbol = "SOL"
			} else {
				start := time.Now()
				_, qt, err := h.pairFinder.FindPair(ctx, pair, nil)
				if err != nil {
					quoteTokenSymbol = "SOL"
				} else {
					quoteTokenSymbol = qt.Symbol
				}
				log.Printf("%s | findPair took %s", pair, time.Since(start))
			}

			if quoteTokenSymbol == "" {
				log.Printf("No quote token symbol for pair: %s", pair)
				return
			}

			mu.Lock()
			usdPrice, ok := priceCache[quoteTokenSymbol]
			if !ok {
				usdPrice = h.pricer.GetUSDPrice(quoteTokenSymbol)
				if usdPrice > 0 {
					priceCache[quoteTokenSymbol] = usdPrice
				} else {
					log.Printf("Missing price for token: %s", quoteTokenSymbol)
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
				if err == nil {
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

			mu.Lock()
			defer mu.Unlock()

			realizedPNLFloatUSD, _ := realizedPNL.Float64()
			pnlResults.RealizedPnLUSD += realizedPNLFloatUSD

			unrealizedPNLFloatUSD, _ := unrealizedPNL.Float64()
			pnlResults.UnrealizedPnLUSD += unrealizedPNLFloatUSD

			// Realized ROI
			if totalSellValue.Cmp(big.NewFloat(0)) > 0 {
				realizedROI := new(big.Float).Quo(realizedPNL, totalSellValue)
				realizedROIFloat, _ := realizedROI.Float64()
				pnlResults.RealizedROI += realizedROIFloat * 100
			}

			// Unrealized ROI
			if remainingAmount.Cmp(big.NewFloat(0)) > 0 {
				avgBuyPrice := new(big.Float).Quo(totalBuyValue, totalBuyTokens)
				remainingCost := new(big.Float).Mul(avgBuyPrice, remainingAmount)

				if remainingCost.Cmp(big.NewFloat(0)) > 0 {
					unrealizedROI := new(big.Float).Quo(unrealizedPNL, remainingCost)
					unrealizedROIFloat, _ := unrealizedROI.Float64()
					pnlResults.UnrealizedROI += unrealizedROIFloat * 100
				}
			}

			if totalBuyValue.Cmp(big.NewFloat(0)) > 0 {
				totalROI := new(big.Float).Quo(new(big.Float).Add(realizedPNL, unrealizedPNL), totalBuyValue)
				weightedROINumerator.Add(weightedROINumerator, new(big.Float).Mul(totalROI, totalBuyValue))
				totalInvestment.Add(totalInvestment, totalBuyValue)
			}

			if realizedPNL.Cmp(big.NewFloat(0)) > 0 {
				winCount++
			}

			tokensTraded[pair] = true
		}(pair, swapLogs)

		individualPairPnL[pair] = pnlResults
	}

	wg.Wait()

	combinedPnL := types.AggregatedPnL{}
	for _, pnl := range individualPairPnL {
		combinedPnL.RealizedPnLUSD += pnl.RealizedPnLUSD
		combinedPnL.UnrealizedPnLUSD += pnl.UnrealizedPnLUSD
		combinedPnL.RealizedROI += pnl.RealizedROI
		combinedPnL.UnrealizedROI += pnl.UnrealizedROI
	}

	if totalInvestment.Cmp(big.NewFloat(0)) > 0 {
		finalROI := new(big.Float).Quo(weightedROINumerator, totalInvestment)
		finalROIFloat, _ := finalROI.Float64()
		combinedPnL.ROI = finalROIFloat * 100
	}

	combinedPnL.PnLUSD = combinedPnL.RealizedPnLUSD + combinedPnL.UnrealizedPnLUSD

	combinedPnL.TokensTraded = len(tokensTraded)
	if combinedPnL.TokensTraded > 0 {
		combinedPnL.WinRate = (float64(winCount) / float64(combinedPnL.TokensTraded)) * 100
	}

	response := map[string]interface{}{
		"overall": combinedPnL,
		"results": individualPairPnL,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}
}
