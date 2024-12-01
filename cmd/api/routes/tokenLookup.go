package routes

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func (h *Handler) TokenLookupHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "token")

	token, pairs, err := h.tokenFinder.FindToken(ctx, address, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": map[string]interface{}{"token": token, "pairs": pairs},
	})
	return

}
