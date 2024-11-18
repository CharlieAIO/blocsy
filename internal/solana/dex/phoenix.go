package dex

import (
	"blocsy/internal/types"
)

func HandlePhoenixSwaps(innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
	transfer1, ok := FindTransfer(transfers, innerIndex, ixIndex+1)
	if !ok {
		return types.SolSwap{}, 0
	}

	transfer2, ok := FindTransfer(transfers, innerIndex, ixIndex+2)
	if !ok {
		return types.SolSwap{}, 0
	}

	s := types.SolSwap{
		Exchange:  "PHOENIX",
		Wallet:    transfer2.FromUserAccount,
		TokenOut:  transfer2.Mint,
		TokenIn:   transfer1.Mint,
		AmountIn:  transfer1.Amount,
		AmountOut: transfer2.Amount,
	}
	return s, 3
}
