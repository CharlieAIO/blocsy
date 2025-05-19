package routes

import (
	"blocsy/internal/types"
	"fmt"
	"github.com/goccy/go-json"
	"log"
	"math/big"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
)

// ActivePositionsHandler godoc
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
func (h *Handler) ActivePositionsHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	wallet := chi.URLParam(r, "wallet")

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

	response := types.AggregatedPnLResponse{
		Results: pnlResults,
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
