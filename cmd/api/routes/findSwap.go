package routes

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strconv"
)

func stringToInt64(s string) int64 {
	i, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Fatalf("Error converting string to int64: %v", err)
	}
	return i
}

func stringToFloat64(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		log.Fatalf("Error converting string to float64: %v", err)
	}
	return f
}

func (h *Handler) FindSwapHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "token")
	amountStr := chi.URLParam(r, "amount")
	timestampStr := chi.URLParam(r, "timestamp")

	amount := stringToFloat64(amountStr)
	timestamp := stringToInt64(timestampStr)

	swap, err := h.swapsRepo.FindSwap(ctx, timestamp, address, amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"result": swap,
	})
	return

}
