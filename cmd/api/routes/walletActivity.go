package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

// WalletActivityHandler godoc
//
//	@Summary		Get wallet activity
//	@Description	Retrieve wallet activity for a given wallet address
//
//	@Security		ApiKeyAuth
//
//	@Tags			Wallet
//	@Accept			json
//	@Produce		json
//	@Param			wallet	path		string	true	"Wallet Address"
//	@Param			limit	query		int		false	"Limit of records"		default(100)
//	@Param			offset	query		int		false	"Offset for pagination"	default(0)
//	@Success		200		{object}	types.WalletActivityResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/wallet/{wallet}/activity [get]
func (h *Handler) WalletActivityHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "wallet")
	limit_ := r.URL.Query().Get("limit")
	offset_ := r.URL.Query().Get("offset")
	if limit_ == "" || offset_ == "" {
		limit_ = "100"
		offset_ = "0"
	}

	limit, err := strconv.ParseInt(limit_, 10, 32)
	if err != nil {
		return
	}
	offset, err := strconv.ParseInt(offset_, 10, 32)
	if err != nil {
		return
	}

	swaps, err := h.swapsRepo.GetAllWalletSwaps(ctx, address, limit, offset)
	if err != nil {
		log.Printf("Failed to GetAllWalletSwaps: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.WalletActivityResponse{
		Results: swaps,
	})
	return

}
