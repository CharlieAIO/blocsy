package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// PairLookupHandler godoc
//
//	@Summary		Lookup a pair
//	@Description	Retrieve pair information for a given pair address
//
//	@Security		ApiKeyAuth
//
//	@Tags			Pair
//	@Accept			json
//	@Produce		json
//	@Param			pair	path		string	true	"Pair address"
//	@Success		200		{object}	types.PairLookupResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/pair/{pair} [get]
func (h *Handler) PairLookupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "pair")

	pair, quoteToken, err := h.pairFinder.FindPair(ctx, address, nil)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.PairLookupResponse{Pair: *pair, QuoteToken: *quoteToken})
	return

}
