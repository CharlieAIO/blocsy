package routes

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *Handler) HoldingsLookupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	token := chi.URLParam(r, "token")
	wallet := chi.URLParam(r, "wallet")

	holdings, err := h.swapsRepo.FindWalletTokenHoldings(ctx, token, wallet)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": holdings,
	})
	return

}
