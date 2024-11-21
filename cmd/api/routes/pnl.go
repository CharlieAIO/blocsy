package routes

import (
	"blocsy/internal/types"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/goccy/go-json"

	"github.com/go-chi/chi/v5"
)

func (h *Handler) PnlHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	wallet := chi.URLParam(r, "wallet")

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	page := 1
	pageSize := 5

	if pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	swaps, err := h.swapsRepo.GetAllWalletSwaps(ctx, wallet)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if len(swaps) == 0 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results":  []types.PNLInfo{},
			"nextPage": nil,
		})
		return
	}

	pairSwaps := make(map[string][]types.SwapLog)
	for _, swap := range swaps {
		pairSwaps[swap.Pair] = append(pairSwaps[swap.Pair], swap)
	}

	allPairs := make([]string, 0, len(pairSwaps))
	for pair := range pairSwaps {
		allPairs = append(allPairs, pair)
	}

	totalPairs := len(allPairs)
	startIndex := (page - 1) * pageSize
	endIndex := startIndex + pageSize

	if startIndex >= totalPairs {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results":  []types.PNLInfo{},
			"nextPage": nil,
		})
		return
	}

	if endIndex > totalPairs {
		endIndex = totalPairs
	}

	currentPagePairs := allPairs[startIndex:endIndex]
	hasMore := endIndex < totalPairs

	pnlMap := make(map[string]types.PNLInfo)
	for _, pair := range currentPagePairs {
		swapLogs := pairSwaps[pair]

		var tokenAddress, quoteTokenSymbol, network, exchange string
		swap := swapLogs[0]

		if swap.Exchange == "PUMPFUN" {
			quoteTokenLookup, err := h.tokenFinder.FindToken(ctx, "So11111111111111111111111111111111111111112")
			if err != nil {
				log.Println("Error finding token:", err)
				continue
			}
			quoteTokenSymbol = quoteTokenLookup.Symbol
			tokenAddress = swap.Pair
		} else {
			pairLookup, _, err := h.pairFinder.FindPair(ctx, swap.Pair)
			if err != nil {
				if errors.Is(err, types.TokenNotFound) {
					continue
				}
				log.Println("Error finding pair:", err)
				return
			}

			if pairLookup.Address == "" {
				continue
			}

			quoteTokenLookup, err := h.tokenFinder.FindToken(ctx, pairLookup.QuoteToken.Address)
			if err != nil {
				log.Println("Error finding token:", err)
				continue
			}

			quoteTokenSymbol = quoteTokenLookup.Symbol
			tokenAddress = pairLookup.Token
			network = pairLookup.Network
			exchange = pairLookup.Exchange
		}

		usdPrice := h.pricer.GetUSDPrice(quoteTokenSymbol)
		info := types.PNLInfo{
			Pair:                    swap.Pair,
			QuoteTokenSymbol:        quoteTokenSymbol,
			Token:                   tokenAddress,
			Network:                 network,
			Exchange:                exchange,
			Swaps:                   len(swapLogs),
			TotalBuyVolume:          "0",
			TotalBuyVolumeUSD:       "0",
			TotalSellVolume:         "0",
			TotalSellVolumeUSD:      "0",
			TotalPnL:                "0",
			TotalPnLUSD:             "0",
			UnrealizedPnL:           "0",
			UnrealizedPnLUSD:        "0",
			RoiPercentage:           "0%",
			UnrealizedRoiPercentage: "0%",
		}

		for _, swap := range swapLogs {
			if swap.Type == "BUY" {
				totalBuyVolume, _ := strconv.ParseFloat(info.TotalBuyVolume, 64)
				info.TotalBuyVolume = fmt.Sprintf("%.20f", totalBuyVolume+swap.AmountOut)
				totalBuyVolumeUSD, _ := strconv.ParseFloat(info.TotalBuyVolumeUSD, 64)
				info.TotalBuyVolumeUSD = fmt.Sprintf("%.2f", totalBuyVolumeUSD+(swap.AmountOut*usdPrice))
			} else {
				totalSellVolume, _ := strconv.ParseFloat(info.TotalSellVolume, 64)
				info.TotalSellVolume = fmt.Sprintf("%.20f", totalSellVolume+swap.AmountIn)
				totalSellVolumeUSD, _ := strconv.ParseFloat(info.TotalSellVolumeUSD, 64)
				info.TotalSellVolumeUSD = fmt.Sprintf("%.2f", totalSellVolumeUSD+(swap.AmountIn*usdPrice))
			}
		}

		// Calculate final PnL, ROI, and Unrealized PnL
		totalBuyVolume, _ := strconv.ParseFloat(info.TotalBuyVolume, 64)
		totalSellVolume, _ := strconv.ParseFloat(info.TotalSellVolume, 64)
		totalBuyVolumeUSD, _ := strconv.ParseFloat(info.TotalBuyVolumeUSD, 64)
		totalSellVolumeUSD, _ := strconv.ParseFloat(info.TotalSellVolumeUSD, 64)

		totalPnL := totalSellVolume - totalBuyVolume
		totalPnLUSD := totalSellVolumeUSD - totalBuyVolumeUSD
		netTokens := totalBuyVolume - totalSellVolume
		unrealizedPnLUSD := netTokens * usdPrice

		info.TotalPnL = fmt.Sprintf("%.20f", totalPnL)
		info.TotalPnLUSD = fmt.Sprintf("%.2f", totalPnLUSD)
		info.UnrealizedPnL = fmt.Sprintf("%.20f", unrealizedPnLUSD)
		info.UnrealizedPnLUSD = fmt.Sprintf("%.2f", unrealizedPnLUSD)

		if totalBuyVolumeUSD != 0 {
			info.RoiPercentage = fmt.Sprintf("%.2f%%", (totalPnLUSD/totalBuyVolumeUSD)*100)
			info.UnrealizedRoiPercentage = fmt.Sprintf("%.2f%%", (unrealizedPnLUSD/totalBuyVolumeUSD)*100)
		}

		pnlMap[pair] = info
	}

	var pnlResults []types.PNLInfo
	for _, info := range pnlMap {
		pnlResults = append(pnlResults, info)
	}

	nextPage := page + 1
	response := map[string]interface{}{
		"results":  pnlResults,
		"nextPage": nil,
	}
	if hasMore {
		response["nextPage"] = nextPage
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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
	pairSwaps := make(map[string][]types.SwapLog)
	tokensTraded := make(map[string]bool)
	winCount := 0

	var mu sync.Mutex
	var wg sync.WaitGroup
	sem := make(chan struct{}, 100) // Limit to 100 goroutines

	for _, swap := range swaps {
		pairSwaps[swap.Pair] = append(pairSwaps[swap.Pair], swap)
	}

	for pair, swapLogs := range pairSwaps {
		wg.Add(1)
		sem <- struct{}{}

		go func(pair string, swapLogs []types.SwapLog) {
			defer wg.Done()
			defer func() { <-sem }()

			var quoteTokenSymbol string
			if swapLogs[0].Exchange == "PUMPFUN" {
				quoteTokenSymbol = "SOL"
			} else {
				_, qt, err := h.pairFinder.FindPair(ctx, pair)
				if err != nil {
					log.Println("Error finding pair:", err)
					return
				}
				quoteTokenSymbol = qt.Symbol
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

			totalBuyAmount, totalBuyValue := 0.0, 0.0
			totalSellAmount, totalSellValue := 0.0, 0.0

			for _, swap := range swapLogs {
				if swap.Type == "BUY" {
					totalBuyAmount += swap.AmountOut
					totalBuyValue += swap.AmountOut * swap.Price
				} else if swap.Type == "SELL" {
					totalSellAmount += swap.AmountIn
					totalSellValue += swap.AmountIn * swap.Price
				}
			}

			realizedPNL := totalSellValue - totalBuyValue
			unrealizedPNL := 0.0
			remainingAmount := 0.0

			if totalBuyAmount > totalSellAmount {
				remainingAmount = totalBuyAmount - totalSellAmount
				unrealizedPNL = remainingAmount * usdPrice
			}

			mu.Lock()
			defer mu.Unlock()

			pnlResults.RealizedPnLUSD += realizedPNL
			pnlResults.UnrealizedPnLUSD += unrealizedPNL
			if realizedPNL > 0 {
				winCount++
			}
			totalBuyValue += totalBuyValue

			if totalBuyValue > 0 {
				pnlResults.RealizedROI += (realizedPNL / totalBuyValue) * 100
			}

			if remainingAmount > 0 && totalBuyValue > 0 {
				pnlResults.UnrealizedROI += (unrealizedPNL / (remainingAmount * usdPrice)) * 100
			}

			tokensTraded[pair] = true
		}(pair, swapLogs)
	}

	wg.Wait()

	pnlResults.TokensTraded = len(tokensTraded)
	if pnlResults.TokensTraded > 0 {
		pnlResults.WinRate = (float64(winCount) / float64(pnlResults.TokensTraded)) * 100
	}

	// Send response
	response := map[string]interface{}{
		"results": pnlResults,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
