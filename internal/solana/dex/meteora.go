package dex

import "blocsy/internal/types"

func HandleMeteoraSwaps(instructionData types.ProcessInstructionData) types.SolSwap {
	if len(*instructionData.Accounts) == 0 || len(instructionData.AccountKeys) < (*instructionData.Accounts)[0] {
		return types.SolSwap{}
	}

	tf1Index := 1
	tf2Index := 2

	transfer1, ok := FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, (instructionData.InnerInstructionIndex)+tf1Index)
	if !ok {
		return types.SolSwap{}
	}
	for transfer1.Type != "token" {
		tf1Index++
		tf2Index++

		transfer1, ok = FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, (instructionData.InnerInstructionIndex)+tf1Index)
		if !ok {
			return types.SolSwap{}
		}
	}

	transfer2, ok := FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, tf2Index+(instructionData.InnerInstructionIndex))
	if !ok {
		return types.SolSwap{}
	}
	for transfer2.Type != "token" {
		tf2Index++

		transfer2, ok = FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, tf2Index+(instructionData.InnerInstructionIndex))
		if !ok {
			return types.SolSwap{}
		}
	}

	s := types.SolSwap{
		Pair:      instructionData.AccountKeys[(*instructionData.Accounts)[0]],
		Exchange:  "METEORA",
		Wallet:    transfer1.FromUserAccount,
		TokenOut:  transfer1.Mint,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
		AmountOut: transfer1.Amount,
	}
	return s
}
