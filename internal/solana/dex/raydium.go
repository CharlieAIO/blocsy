package dex

import (
	"blocsy/internal/types"
)

func HandleRaydiumSwaps(ixData string, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {

	transfer1, ok := FindTransfer(transfers, innerIndex, ixIndex+1)
	if !ok {
		return types.SolSwap{}, 0
	}

	transfer2, ok := FindTransfer(transfers, innerIndex, ixIndex+2)
	if !ok {
		return types.SolSwap{}, 0
	}

	wallet := transfer1.FromUserAccount
	if wallet == "" {
		wallet = transfer2.ToUserAccount
	}

	s := types.SolSwap{
		Exchange:  "RAYDIUM",
		Wallet:    wallet,
		TokenOut:  transfer1.Mint,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
		AmountOut: transfer1.Amount,
	}

	return s, 2

}
