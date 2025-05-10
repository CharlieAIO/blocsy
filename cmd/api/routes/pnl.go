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

// AggregatedPnlHandler godoc
//
//	@Summary		Wallet Aggregated PnL
//	@Description	Retrieve aggregated pnl for a given wallet address
//	@Security		ApiKeyAuth
//
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			wallet		path		string	true	"Wallet address"
//	@Param			timeframe	query		string	false	"Timeframe (1d, 7d, 14d, 30d)"
//	@Success		200			{object}	types.AggregatedPnLResponse
//	@Failure		400			{object}	map[string]interface{}
//	@Failure		500			{object}	map[string]interface{}
//	@Router			/pnl/{wallet} [get]
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
		log.Printf("Failed to get swaps for wallet: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(swaps) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": types.AggregatedPnL{},
		})
		return
	}

	pnlResults := types.AggregatedPnL{}
	priceCache := make(map[string]float64)
	tokenSwaps := make(map[string][]types.SwapLog)
	tokensTraded := make(map[string]bool)
	winCount := 0

	totalInvestment := new(big.Float)
	weightedROINumerator := new(big.Float)

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 100) // Limit to 100 goroutines

	for _, swap := range swaps {
		tokenSwaps[swap.Token] = append(tokenSwaps[swap.Token], swap)
	}

	for token, swapLogs := range tokenSwaps {
		wg.Add(1)
		sem <- struct{}{}

		go func(token string, swapLogs []types.SwapLog) {
			defer wg.Done()
			defer func() { <-sem }()

			var pair = swapLogs[0].Pair

			var quoteTokenSymbol string
			if swapLogs[0].Source == "PUMPFUN" {
				quoteTokenSymbol = "SOL"
			} else {
				_, qt, err := h.pairFinder.FindPair(ctx, pair, nil)
				if err != nil {
					quoteTokenSymbol = "SOL"
				} else {
					quoteTokenSymbol = qt.Symbol
				}
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
					//log.Printf("Missing price for token: %s", quoteTokenSymbol)
					mu.Unlock()
					return
				}
			}
			mu.Unlock()

			totalBuyTokens := new(big.Float)
			totalSellTokens := new(big.Float)
			totalBuyValue := new(big.Float)
			totalSellValue := new(big.Float)

			hasBuyOrSell := false

			for _, swap := range swapLogs {
				amountOutFloat := new(big.Float).SetFloat64(swap.AmountOut)
				amountInFloat := new(big.Float).SetFloat64(swap.AmountIn)
				if swap.Action == "BUY" || swap.Action == "RECEIVE" {
					totalBuyTokens.Add(totalBuyTokens, amountInFloat)
					if swap.Action == "BUY" {
						hasBuyOrSell = true
						totalBuyValue.Add(totalBuyValue, new(big.Float).Mul(amountOutFloat, big.NewFloat(usdPrice)))
					}
				} else if swap.Action == "SELL" || swap.Action == "TRANSFER" {
					totalSellTokens.Add(totalSellTokens, amountOutFloat)
					if swap.Action == "SELL" {
						hasBuyOrSell = true
						totalSellValue.Add(totalSellValue, new(big.Float).Mul(amountInFloat, big.NewFloat(usdPrice)))
					}
				}
			}
			if !hasBuyOrSell {
				return
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

			if unrealizedPNL.IsInf() {
				pnlResults.UnrealizedPnLUSD = 0
			}

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

			//log.Printf("usdPrice: %f | totalBuyValue: %f | totalSellValue: %f | totalBuyTokens: %f | totalSellTokens: %f | remainingAmount: %f | realizedPNL: %f | unrealizedPNL: %f | totalInvestment: %f | weightedROINumerator: %f | winCount: %d", usdPrice, totalBuyValue, totalSellValue, totalBuyTokens, totalSellTokens, remainingAmount, realizedPNL, unrealizedPNL, totalInvestment, weightedROINumerator, winCount)
			tokensTraded[pair] = true
		}(token, swapLogs)
	}

	wg.Wait()

	if totalInvestment.Cmp(big.NewFloat(0)) > 0 {
		finalROI := new(big.Float).Quo(weightedROINumerator, totalInvestment)
		finalROIFloat, _ := finalROI.Float64()
		pnlResults.ROI = finalROIFloat * 100
	}

	pnlResults.PnLUSD = pnlResults.RealizedPnLUSD + pnlResults.UnrealizedPnLUSD

	pnlResults.TokensTraded = len(tokensTraded)
	if pnlResults.TokensTraded > 0 {
		pnlResults.WinRate = (float64(winCount) / float64(pnlResults.TokensTraded)) * 100
	}

	//log.Printf("pnl results: %+v", pnlResults)

	response := types.AggregatedPnLResponse{
		Results: pnlResults,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
