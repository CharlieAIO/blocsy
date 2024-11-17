package dex

import (
	"defi-intel/internal/types"
)

func HandleOrcaSwaps(tx *types.SolanaTx, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
	transfer1, ok := FindTransfer(transfers, innerIndex, ixIndex+1)
	if !ok {
		return types.SolSwap{}, 0
	}

	transfer2, ok := FindTransfer(transfers, innerIndex, ixIndex+2)
	if !ok {
		return types.SolSwap{}, 0
	}

	s := types.SolSwap{
		Exchange:  "ORCA",
		Wallet:    transfer1.FromUserAccount,
		TokenOut:  transfer1.Mint,
		AmountOut: transfer1.Amount,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
	}
	return s, 2
}
