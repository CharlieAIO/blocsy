package routes

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *Handler) PriceLookupHandler(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	usdPrice := h.pricer.GetUSDPrice(symbol)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": usdPrice,
	})
	return

}
