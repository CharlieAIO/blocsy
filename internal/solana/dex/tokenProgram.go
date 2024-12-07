package dex

import (
	"blocsy/internal/types"
)

func HandleTokenSwaps(instructionData *types.ProcessInstructionData) types.SolSwap {
	if len(*instructionData.Accounts) < 2 || len(instructionData.AccountKeys) < (*instructionData.Accounts)[1] {
		return types.SolSwap{}
	}

	transfer1, ok := FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, (instructionData.InnerInstructionIndex)+1)
	if !ok {
		return types.SolSwap{}
	}

	//log.Printf("finding transfer %d % d", -1, instructionData.InnerInstructionIndex+2)
	//for _, acc := range *instructionData.Accounts {
	//	log.Printf("account: %s", instructionData.AccountKeys[acc])
	//}

	s := types.SolSwap{
		Pair:      "",
		Exchange:  "",
		Wallet:    instructionData.AccountKeys[(*instructionData.Accounts)[2]],
		TokenOut:  "",
		TokenIn:   instructionData.AccountKeys[(*instructionData.Accounts)[3]],
		AmountIn:  transfer1.Amount,
		AmountOut: "",
	}

	return s

}
