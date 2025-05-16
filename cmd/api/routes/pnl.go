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
	var totalHeldTimeAcrossTokens time.Duration
	var totalSoldAmountAcrossTokens = big.NewFloat(0)
	var totalActivePositionsUSD = new(big.Float)

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
			mu.Unlock()
			if !ok {
				usdPrice = h.pricer.GetUSDPrice(quoteTokenSymbol)
				if usdPrice > 0 {
					mu.Lock()
					priceCache[quoteTokenSymbol] = usdPrice
					mu.Unlock()
				} else {
					return
				}
			}

			hasBuyOrSell := false
			for _, swap := range swapLogs {
				if swap.Action == "BUY" {
					hasBuyOrSell = true
					totalBuys++
				} else if swap.Action == "SELL" {
					hasBuyOrSell = true
					totalSells++
				}
			}

			if !hasBuyOrSell {
				return
			}

			tokenPnL, totalBuyValue, totalSellValue, _, _, totalSoldAmount, totalHeldTime, currentValue := CalculateTokenPnL(
				ctx,
				swapLogs,
				usdPrice,
				h.swapsRepo.FindLatestSwap,
			)

			mu.Lock()
			defer mu.Unlock()

			buyVolumeUSD.Add(buyVolumeUSD, totalBuyValue)
			sellVolumeUSD.Add(sellVolumeUSD, totalSellValue)

			totalHeldTimeAcrossTokens += totalHeldTime
			totalSoldAmountAcrossTokens.Add(totalSoldAmountAcrossTokens, totalSoldAmount)
			totalActivePositionsUSD.Add(totalActivePositionsUSD, currentValue)

			pnlResults.RealizedPnLUSD += tokenPnL.RealizedPnLUSD
			pnlResults.UnrealizedPnLUSD += tokenPnL.UnrealizedPnLUSD
			pnlResults.RealizedROI += tokenPnL.RealizedROI
			pnlResults.UnrealizedROI += tokenPnL.UnrealizedROI

			if totalBuyValue.Cmp(big.NewFloat(0)) > 0 {
				totalROI := big.NewFloat(tokenPnL.ROI / 100)
				weightedROINumerator.Add(weightedROINumerator, new(big.Float).Mul(totalROI, totalBuyValue))
				totalInvestment.Add(totalInvestment, totalBuyValue)
			}

			if tokenPnL.RealizedPnLUSD > 0 {
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
	pnlResults.TotalActivePositionsUSD, _ = totalActivePositionsUSD.Float64()
	// Calculate average hold time based on the aggregated values
	if totalSoldAmountAcrossTokens.Cmp(big.NewFloat(0)) > 0 {
		totalSoldAmountFloat, _ := totalSoldAmountAcrossTokens.Float64()
		if totalSoldAmountFloat > 0 {
			pnlResults.AverageHoldTime = time.Duration(float64(totalHeldTimeAcrossTokens) / totalSoldAmountFloat)
		}
	}

	response := types.AggregatedPnLResponse{
		Results: pnlResults,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
