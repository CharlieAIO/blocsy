package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
)

// HoldingsLookupHandler godoc
//
//	@Summary		Wallet Token Holdings
//	@Description	Retrieve the token holdings for a given wallet address
//	@Security		ApiKeyAuth
//
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			wallet	path		string	true	"Wallet address"
//	@Param			token	path		string	true	"Token address"
//	@Success		200		{object}	types.HoldingsLookupResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/holdings/{wallet}/{token} [get]
func (h *Handler) HoldingsLookupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	wallet := chi.URLParam(r, "wallet")

	holdings, err := h.swapsRepo.FindWalletTokenHoldings(ctx, token, wallet)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	response := types.HoldingsLookupResponse{
		Results: holdings,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	return

}
