package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"log"
	"net/http"
)

// TopTokensHandler godoc
//
//	@Summary		Top Tokens
//	@Description	Retrieve the most recent top tokens based on Market Cap
//
//	@Security		ApiKeyAuth
//
//	@Tags			Analytics
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Token address"
//	@Success		200		{object}	types.TopRecentTokensResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/top-tokens [get]
func (h *Handler) TopTokensHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	results, err := h.swapsRepo.FindTopRecentTokens(ctx)
	if err != nil {
		log.Printf("Failed to find top tokens: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(types.TopRecentTokensResponse{
		Results: results,
	}); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}
}
