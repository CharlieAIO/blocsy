package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"net/http"
)

// TopTradersHandler godoc
//
//	@Summary		Get top traders by pnl
//	@Description	Retrieve the top traders by pnl for a given token
//
//	@Security		ApiKeyAuth
//
//	@Tags			Token
//	@Accept			json
//	@Produce		json
//	@Param			token	path		string	true	"Token address"
//	@Success		200		{object}	types.TopTradersResponse
//	@Failure		400		{object}	map[string]interface{}
//	@Failure		500		{object}	map[string]interface{}
//	@Router			/top-traders/{token} [get]
func (h *Handler) TopTradersHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "token")

	traders, err := h.swapsRepo.FindTopTraders(ctx, address)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.TopTradersResponse{
		Results: traders,
	})
	return

}
