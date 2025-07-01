package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"log"
	"net/http"
)

// SearchQueryHandler godoc
//
//	@Summary		Query All
//	@Description	Lookup and return all pairs,tokens and wallets given a query
//
//	@Security		ApiKeyAuth
//
//	@Tags			Lookup
//	@Accept			json
//	@Produce		json
//	@Param			q	query		string	true	"Search query using token address, pair address, wallet address or token name/ token symbol"
//	@Success		200	{object}	types.SearchQueryResponse
//	@Failure		400	{object}	map[string]interface{}
//	@Failure		500	{object}	map[string]interface{}
//	@Router			/v1/search [get]
func (h *Handler) SearchQueryHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	searchQuery := r.URL.Query().Get("q")

	results, err := h.swapsRepo.QueryAll(ctx, searchQuery)
	if err != nil {
		log.Printf("Failed to QueryAll: %v", err)
		http.Error(w, "", http.StatusInternalServerError)
		return
	}

	// Convert QueryAll results to QueryAllResponse
	responseResults := make([]types.QueryAllResponse, len(results))
	for i, result := range results {
		responseResults[i] = types.ConvertToResponse(result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(types.SearchQueryResponse{Results: responseResults})
	return

}
