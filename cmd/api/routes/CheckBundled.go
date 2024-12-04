package routes

import (
	"blocsy/internal/types"
	"encoding/json"
	"github.com/go-chi/chi/v5"
	"log"
	"net/http"
	"strings"
)

func isTokenCreation(parsedLogs []types.LogDetails) bool {
	var containsInitializeMint, containsMintTo bool
	var keywords = []string{"InitializeMint", "MintTo"}

	var checkLogs func(logs []types.LogDetails)
	checkLogs = func(logs []types.LogDetails) {
		for _, logDetail := range logs {
			for _, log := range logDetail.Logs {
				for _, keyword := range keywords {
					if strings.Contains(log, keyword) {
						if keyword == "InitializeMint" {
							containsInitializeMint = true
						}
						if keyword == "MintTo" {
							containsMintTo = true
						}
					}
				}
			}
			checkLogs(logDetail.SubLogs)
		}
	}

	checkLogs(parsedLogs)

	return containsInitializeMint && containsMintTo
}

func (h *Handler) CheckBundledHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	address := chi.URLParam(r, "token")

	_, _, err := h.tokenFinder.FindToken(ctx, address, false)
	if err != nil {
		http.Error(w, "token not found", http.StatusInternalServerError)
		return
	}

	log.Printf("Checking bundled for token: %s", address)
	swaps, err := h.swapsRepo.FindFirstTokenSwaps(ctx, address)
	if err != nil || len(swaps) == 0 {
		log.Printf("Error finding swaps: %v", err)
		http.Error(w, "swap history not found", http.StatusInternalServerError)
		return
	}

	var firstBlockSwaps []types.SwapLog
	for _, swap := range swaps {
		if swap.BlockNumber == swaps[0].BlockNumber {
			firstBlockSwaps = append(firstBlockSwaps, swap)
		}
	}

	if len(firstBlockSwaps) == 1 {
		http.Error(w, "not bundled", http.StatusInternalServerError)
		return
	}

	var hasTokenCreation bool
	for _, swap := range firstBlockSwaps {
		tx, err := h.nodes[0].GetTx(ctx, swap.ID)
		if err != nil {
			return
		}
		parsedLogs := h.nodes[0].GetParsedLogs(tx.Meta.LogMessages)
		if isTokenCreation(parsedLogs) {
			hasTokenCreation = true
			break
		}
	}

	if !hasTokenCreation {
		http.Error(w, "not bundled", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"results": firstBlockSwaps,
	})
	return

}
