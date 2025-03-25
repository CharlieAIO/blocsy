package dex

import (
	"blocsy/internal/types"
)

func HandlePumpFunAmmSwaps(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int) {
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

	wallet := accountKeys[(currentTransfer.IxAccounts)[1]]

	s := types.SolSwap{
		Pair:      accountKeys[(currentTransfer.IxAccounts)[0]],
		Exchange:  "PUMPFUN_AMM",
		Wallet:    wallet,
		TokenOut:  currentTransfer.Mint,
		TokenIn:   nextTransfer.Mint,
		AmountIn:  nextTransfer.Amount,
		AmountOut: currentTransfer.Amount,
	}

	return s, 1

}
