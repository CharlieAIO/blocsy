package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
)

// TopTradersHandler godoc
//
//	@Summary		Top Traders
//	@Description	Retrieve the top traders based on pnl for a given token
//
//	@Security		ApiKeyAuth
//
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Token address"
//	@Success		200		{object}	types.TopTradersResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/v1/top-traders/{token} [get]
func (h *Handler) TopTradersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "token")
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "100"
	}
	limitInt, err := strconv.ParseInt(limit, 10, 32)

	traders, err := h.swapsRepo.FindTopTraders(ctx, address, limitInt)
	if err != nil {
		log.Printf("Failed to find top traders: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(types.TopTradersResponse{
		Results: traders,
	}); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
