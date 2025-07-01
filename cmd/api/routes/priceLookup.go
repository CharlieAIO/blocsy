package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// PriceLookupHandler godoc
//
//	@Summary		Query token price
//	@Description	Retrieve the price of a token in USD
//
//	@Security		ApiKeyAuth
//
//	@Tags			Lookup
//	@Accept			json
//	@Produce		json
//	@Param			symbol	path		string	true	"Token symbol"
//	@Success		200		{object}	types.PriceLookupResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/v1/price/{symbol} [get]
func (h *Handler) PriceLookupHandler(w http.ResponseWriter, r *http.Request) {
	symbol := chi.URLParam(r, "symbol")

	usdPrice := h.pricer.GetUSDPrice(symbol)

	response := types.PriceLookupResponse{
		Price:    usdPrice,
		Symbol:   symbol,
		Currency: "USD",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
	return

}
