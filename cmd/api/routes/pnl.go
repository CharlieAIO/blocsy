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
	totalBuys, totalSells := int64(0), int64(0)
	sellVolumeUSD := new(big.Float)
	buyVolumeUSD := new(big.Float)
	durationsHeld := make(map[string]time.Duration)

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

			if swapLogs[0].QuoteTokenSymbol == nil {
				log.Printf("No quote token symbol for pair: %s", pair)
				return
			}
			quoteTokenSymbol := *swapLogs[0].QuoteTokenSymbol

			mu.Lock()
			usdPrice, ok := priceCache[quoteTokenSymbol]
			if !ok {
				usdPrice = h.pricer.GetUSDPrice(quoteTokenSymbol)
				if usdPrice > 0 {
					priceCache[quoteTokenSymbol] = usdPrice
				} else {
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

			var buyQueue []types.TokenLot
			var totalHeldTime time.Duration
			var totalSoldAmount = big.NewFloat(0)

			for _, swap := range swapLogs {
				amountOutFloat := new(big.Float).SetFloat64(swap.AmountOut)
				amountInFloat := new(big.Float).SetFloat64(swap.AmountIn)
				if swap.Action == "BUY" || swap.Action == "RECEIVE" {
					totalBuyTokens.Add(totalBuyTokens, amountInFloat)
					if swap.Action == "BUY" {
						totalBuys++
						hasBuyOrSell = true
						totalBuyValue.Add(totalBuyValue, new(big.Float).Mul(amountOutFloat, big.NewFloat(usdPrice)))
					}
					buyQueue = append(buyQueue, types.TokenLot{
						Amount:    new(big.Float).Set(amountInFloat),
						Timestamp: swap.Timestamp,
					})

				} else if swap.Action == "SELL" || swap.Action == "TRANSFER" {
					totalSellTokens.Add(totalSellTokens, amountOutFloat)
					if swap.Action == "SELL" {
						totalSells++
						hasBuyOrSell = true
						totalSellValue.Add(totalSellValue, new(big.Float).Mul(amountInFloat, big.NewFloat(usdPrice)))
					}
					toSell := new(big.Float).Set(amountOutFloat)

					for len(buyQueue) > 0 && toSell.Cmp(big.NewFloat(0)) > 0 {
						currentLot := &buyQueue[0]
						if currentLot.Amount.Cmp(toSell) <= 0 {
							heldDuration := swap.Timestamp.Sub(currentLot.Timestamp)
							amountFloat, _ := currentLot.Amount.Float64()
							totalHeldTime += time.Duration(float64(heldDuration.Nanoseconds()) * amountFloat)

							totalSoldAmount.Add(totalSoldAmount, currentLot.Amount)
							toSell.Sub(toSell, currentLot.Amount)
							buyQueue = buyQueue[1:] // remove lot
						} else {
							heldDuration := swap.Timestamp.Sub(currentLot.Timestamp)
							toSellFloat, _ := toSell.Float64()
							totalHeldTime += time.Duration(float64(heldDuration.Nanoseconds()) * toSellFloat)
							totalSoldAmount.Add(totalSoldAmount, toSell)
							currentLot.Amount.Sub(currentLot.Amount, toSell)
							toSell = big.NewFloat(0)
						}
					}
				}
			}
			if !hasBuyOrSell {
				return
			}

			durationsHeld[token] = totalHeldTime

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

			buyVolumeUSD.Add(buyVolumeUSD, new(big.Float).Mul(totalBuyValue, big.NewFloat(usdPrice)))
			sellVolumeUSD.Add(sellVolumeUSD, new(big.Float).Mul(totalSellValue, big.NewFloat(usdPrice)))

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
	pnlResults.TotalBuy = totalBuys
	pnlResults.TotalSell = totalSells

	pnlResults.TotalBuyVolumeUSD, _ = buyVolumeUSD.Float64()
	pnlResults.TotalSellVolumeUSD, _ = sellVolumeUSD.Float64()
	pnlResults.TotalVolumeUSD = pnlResults.TotalBuyVolumeUSD + pnlResults.TotalSellVolumeUSD
	pnlResults.AverageHoldTime = time.Duration(float64(pnlResults.TotalVolumeUSD) / float64(pnlResults.TotalSell) * float64(time.Hour))

	response := types.AggregatedPnLResponse{
		Results: pnlResults,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
