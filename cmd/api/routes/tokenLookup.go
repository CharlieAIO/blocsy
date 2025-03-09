package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// TokenLookupHandler godoc
//
//	@Summary		Token Lookup
//	@Description	Retrieve token information for a given token address
//
//	@Security		ApiKeyAuth
//
//	@Tags			Lookup
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Token address"
//	@Success		200		{object}	types.TokenLookupResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/token/{token} [get]
func (h *Handler) TokenLookupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "token")

	token, pairs, err := h.tokenFinder.FindToken(ctx, address, false)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	if token == nil {
		http.Error(w, "", http.StatusBadRequest)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.TokenLookupResponse{Token: *token, Pairs: *pairs})
	return

}
