package dex

import (
	"defi-intel/internal/types"
)

func HandleLifinitySwaps(tx *types.SolanaTx, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
	tf2Index := ixIndex + 2

	transfer1, ok := FindTransfer(transfers, innerIndex, ixIndex+1)
	if !ok {
		return types.SolSwap{}, 0
	}

	transfer2, ok := FindTransfer(transfers, innerIndex, tf2Index)
	if !ok {
		return types.SolSwap{}, 0
	}

	if transfer2.Mint != "" {
		for transfer2.Type == "Mint" {
			tf2Index++
			transfer2, ok = FindTransfer(transfers, innerIndex, tf2Index)
			if !ok {
				return types.SolSwap{}, 0
			}
		}
	}

	s := types.SolSwap{
		Exchange:  "LIFINITY",
		Wallet:    transfer1.FromUserAccount,
		TokenOut:  transfer1.Mint,
		AmountOut: transfer1.Amount,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
	}
	return s, tf2Index - ixIndex
}
