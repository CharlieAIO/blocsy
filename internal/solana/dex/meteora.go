package dex

import (
	"blocsy/internal/types"
)

func HandleMeteoraSwaps(ixData string, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
	tf1Index := 1
	tf2Index := 2

	transfer1, ok := FindTransfer(transfers, innerIndex, ixIndex+tf1Index)
	if !ok {
		return types.SolSwap{}, 0
	}
	for transfer1.Type != "token" {
		tf1Index++
		tf2Index++

		transfer1, ok = FindTransfer(transfers, innerIndex, ixIndex+tf1Index)
		if !ok {
			return types.SolSwap{}, 0
		}
	}

	transfer2, ok := FindTransfer(transfers, innerIndex, tf2Index+ixIndex)
	if !ok {
		return types.SolSwap{}, 0
	}
	for transfer2.Type != "token" {
		tf2Index++

		transfer2, ok = FindTransfer(transfers, innerIndex, tf2Index+ixIndex)
		if !ok {
			return types.SolSwap{}, 0
		}
	}

	s := types.SolSwap{
		Exchange:  "METEORA",
		Wallet:    transfer1.FromUserAccount,
		TokenOut:  transfer1.Mint,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
		AmountOut: transfer1.Amount,
	}
	return s, tf2Index
}
