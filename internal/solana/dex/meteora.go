package dex

import "blocsy/internal/types"

func HandleMeteoraSwaps(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int) {
	currentTransfer := transfers[index]
	nextTransfer := transfers[index+1]
	if currentTransfer.ParentProgramId != nextTransfer.ParentProgramId {
		return types.SolSwap{}, 0
	}

	if len(currentTransfer.IxAccounts) == 0 || len(accountKeys) < (currentTransfer.IxAccounts)[0] {
		return types.SolSwap{}, 0
	}

	wallet := accountKeys[(currentTransfer.IxAccounts)[10]]

	s := types.SolSwap{
		Pair:      accountKeys[(currentTransfer.IxAccounts)[0]],
		Exchange:  "METEORA",
		Wallet:    wallet,
		TokenOut:  currentTransfer.Mint,
		TokenIn:   nextTransfer.Mint,
		AmountIn:  nextTransfer.Amount,
		AmountOut: currentTransfer.Amount,
	}

	return s, 1

}
