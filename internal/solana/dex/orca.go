package dex

import (
	"blocsy/internal/types"
)

func HandleOrcaSwaps(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int) {
	if index+1 >= len(transfers) {
		return types.SolSwap{}, 0
	}

	currentTransfer := transfers[index]
	nextTransfer := transfers[index+1]
	if currentTransfer.ParentProgramId != nextTransfer.ParentProgramId {
		return types.SolSwap{}, 0
	}

	if len(currentTransfer.IxAccounts) < 3 || len(accountKeys) < (currentTransfer.IxAccounts)[2] {
		return types.SolSwap{}, 0
	}

	wallet := currentTransfer.FromUserAccount

	pair := ""
	if len(currentTransfer.IxAccounts) == 15 {
		pair = accountKeys[(currentTransfer.IxAccounts)[4]]
	} else {
		pair = accountKeys[(currentTransfer.IxAccounts)[2]]
	}

	s := types.SolSwap{
		Pair:      pair,
		Exchange:  "ORCA",
		Wallet:    wallet,
		TokenOut:  currentTransfer.Mint,
		TokenIn:   nextTransfer.Mint,
		AmountIn:  nextTransfer.Amount,
		AmountOut: currentTransfer.Amount,
	}

	return s, 1

}
