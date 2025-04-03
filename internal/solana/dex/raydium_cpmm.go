package dex

import (
	"blocsy/internal/types"
)

func HandleRaydiumCPMMSwaps(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int) {
	if index+1 >= len(transfers) {
		return types.SolSwap{}, 0
	}

	currentTransfer := transfers[index]
	nextTransfer := transfers[index+1]
	if currentTransfer.ParentProgramId != nextTransfer.ParentProgramId {
		return types.SolSwap{}, 0
	}

	if len(currentTransfer.IxAccounts) < 2 || len(accountKeys) < (currentTransfer.IxAccounts)[1] {
		return types.SolSwap{}, 0
	}

	pair := accountKeys[(currentTransfer.IxAccounts)[3]]
	wallet := accountKeys[(currentTransfer.IxAccounts)[0]]

	if currentTransfer.Authority != wallet {
		currentTransfer = transfers[index+1]
		nextTransfer = transfers[index]
	}

	s := types.SolSwap{
		Pair:      pair,
		Exchange:  "RAYDIUM_CPMM",
		Wallet:    wallet,
		TokenOut:  currentTransfer.Mint,
		TokenIn:   nextTransfer.Mint,
		AmountIn:  nextTransfer.Amount,
		AmountOut: currentTransfer.Amount,
	}

	return s, 1

}
