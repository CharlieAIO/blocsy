package dex

import (
	"defi-intel/internal/types"
)

func HandleFluxbeamSwaps(tx *types.SolanaTx, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
	transfer1, ok := FindTransfer(transfers, innerIndex, ixIndex+1)
	if !ok {
		return types.SolSwap{}, 0
	}

	tf2Index := ixIndex + 2
	transfer2, ok := FindTransfer(transfers, innerIndex, tf2Index)
	if !ok {
		return types.SolSwap{}, 0
	}

	for transfer2.Type == "mint" {
		tf2Index++
		transfer2, ok = FindTransfer(transfers, innerIndex, tf2Index)
		if !ok {
			return types.SolSwap{}, 0
		}

		if transfer2.Mint == "" {
			break
		}
	}

	s := types.SolSwap{
		Exchange:  "FLUXBEAM",
		Wallet:    transfer1.FromUserAccount,
		TokenOut:  transfer1.Mint,
		AmountOut: transfer1.Amount,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
	}
	return s, tf2Index - ixIndex
}
