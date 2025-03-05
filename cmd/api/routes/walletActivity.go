package routes

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

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
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": swaps,
	})
	return

}
