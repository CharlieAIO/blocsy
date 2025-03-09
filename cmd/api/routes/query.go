package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"net/http"
)

// QueryHandler godoc
//
//	@Summary		Query All
//	@Description	Lookup and return all pairs,tokens and wallets given a query
//
//	@Security		ApiKeyAuth
//
//	@Tags			Lookup
//	@Accept			json
//	@Produce		json
//	@Param			q	query		string		true	"Search query using token address, pair address, wallet address or token name/ token symbol"
//	@Success		200		{object}	types.TokenLookupResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/token/{token} [get]
func (h *Handler) QueryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := r.URL.Query().Get("q")

	token, pairs, err := h.tokenFinder.FindToken(ctx, address, false)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.TokenLookupResponse{Token: *token, Pairs: *pairs})
	return

}
