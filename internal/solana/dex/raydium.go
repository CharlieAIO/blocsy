package dex

import (
	"blocsy/internal/types"
)

func HandleRaydiumSwaps(instructionData types.ProcessInstructionData) types.SolSwap {

	if len(*instructionData.InnerAccounts) < 2 || len(instructionData.AccountKeys) < (*instructionData.InnerAccounts)[1] {
		return types.SolSwap{}
	}

	transfer1, ok := FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, (instructionData.InnerInstructionIndex)+1)
	if !ok {
		return types.SolSwap{}
	}

	transfer2, ok := FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, (instructionData.InnerInstructionIndex)+2)
	if !ok {
		return types.SolSwap{}
	}

	wallet := transfer1.FromUserAccount
	if wallet == "" {
		wallet = transfer2.ToUserAccount
	}

	s := types.SolSwap{
		Pair:      instructionData.AccountKeys[(*instructionData.InnerAccounts)[1]],
		Exchange:  "RAYDIUM",
		Wallet:    wallet,
		TokenOut:  transfer1.Mint,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
		AmountOut: transfer1.Amount,
	}

	return s

}
