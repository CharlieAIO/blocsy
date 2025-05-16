package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"sort"
	"strconv"
	"sync"
	"time"
)

// TokenPnlHandler godoc
//
//	@Summary		Wallet Token PnL
//	@Description	Retrieve all token pnl for a wallet
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

			hasBuyOrSell := false
			for _, swap := range swapLogs {
				if swap.Action == "BUY" || swap.Action == "SELL" {
					hasBuyOrSell = true
					break
				}
			}

			if !hasBuyOrSell {
				return
			}

			pnlResults, _, _, _, _, _, _ := CalculateTokenPnL(
				ctx,
				swapLogs,
				usdPrice,
				h.swapsRepo.FindLatestSwap,
			)

			tokenSymbol := ""
			if swapLogs[0].TokenSymbol != nil {
				tokenSymbol = *swapLogs[0].TokenSymbol
			}

			result := types.TokenAndPnl{
				Token:       swapLogs[0].Token,
				TokenSymbol: tokenSymbol,
				PnL:         pnlResults,
			}
			if swapLogs[0].Token != "" {
				resultsMu.Lock()
				results = append(results, result)
				resultsMu.Unlock()
			}

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
