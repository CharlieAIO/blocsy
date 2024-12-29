package websocket

import "blocsy/internal/types"

func validateClientType(clientType string) bool {
	return clientType == "wallet" || clientType == "pf-tokens"
}

func filterRelevantSwaps(swaps []types.SwapLog, wallets []string) []types.SwapLog {
	var relevantSwaps []types.SwapLog
	walletSet := make(map[string]struct{}, len(wallets))
	for _, wallet := range wallets {
		walletSet[wallet] = struct{}{}
	}

	for _, swap := range swaps {
		if _, exists := walletSet[swap.Wallet]; exists {
			if swap.Action != "BUY" && swap.Action != "SELL" {
				continue
			}
			relevantSwaps = append(relevantSwaps, swap)
		}
	}
	return relevantSwaps
}
