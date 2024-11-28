package dex

import (
	"blocsy/internal/types"
)

func HandleOrcaSwaps(instructionData types.ProcessInstructionData) types.SolSwap {
	if len(*instructionData.Accounts) < 3 || len(instructionData.AccountKeys) < (*instructionData.Accounts)[2] {
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

	pair := ""
	if len(*instructionData.Accounts) == 15 {
		pair = instructionData.AccountKeys[(*instructionData.Accounts)[4]]
	} else {
		pair = instructionData.AccountKeys[(*instructionData.Accounts)[2]]
	}

	s := types.SolSwap{
		Pair:      pair,
		Exchange:  "ORCA",
		Wallet:    transfer1.FromUserAccount,
		TokenOut:  transfer1.Mint,
		AmountOut: transfer1.Amount,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
	}
	return s
}
