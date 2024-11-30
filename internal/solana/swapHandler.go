package solana

import (
	"blocsy/internal/solana/dex"
	"blocsy/internal/types"
	"context"
	"math/big"
	"time"
)

func NewSwapHandler(tf SolanaTokenFinder, pf SolanaPairFinder) *SwapHandler {
	return &SwapHandler{
		tf: tf,
		pf: pf,
	}
}

func (sh *SwapHandler) HandleSwaps(ctx context.Context, transfers []types.SolTransfer, tx *types.SolanaTx, timestamp int64, block uint64) []types.SwapLog {
	if !validateTX(tx) {
		return []types.SwapLog{}
	}
	swaps := make([]types.SolSwap, 0)
	accountKeys := getAllAccountKeys(tx)

	processInstructionData := types.ProcessInstructionData{
		AccountKeys:           accountKeys,
		Transfers:             transfers,
		InnerInstructionIndex: -1,
		TokenAccountMap:       CreateTokenAccountMap(tx),
	}

	logs := GetLogs(tx.Meta.LogMessages)
	//for _, x := range transfers {
	//	log.Printf("transfer: %+v", x)
	//}

	for index, instruction := range tx.Transaction.Message.Instructions {
		accounts := instruction.Accounts
		innerInstructions, innerIdx := FindInnerIx(tx.Meta.InnerInstructions, index)

		if index < len(logs) {
			processInstructionData.Logs = logs[index].Logs
		}

		if len(accountKeys)-1 < instruction.ProgramIdIndex {
			continue
		}

		programId := accountKeys[instruction.ProgramIdIndex]
		instructionSource := identifySource(programId)

		processInstructionData.ProgramId = &programId
		processInstructionData.Accounts = &accounts
		processInstructionData.InnerIndex = &innerIdx
		processInstructionData.Data = &instruction.Data

		processInstructionData.InnerInstructionIndex = -1

		swap := processInstruction(processInstructionData)
		if swap.Pair != "" || swap.TokenOut != "" || swap.TokenIn != "" {
			swap.Source = instructionSource
			swaps = append(swaps, swap)
		}

		innerSwaps := make([]types.SolSwap, 0)
		for innerIxIndex, innerIx := range innerInstructions {

			if index < len(logs) && innerIxIndex < len(logs[index].SubLogs) {
				processInstructionData.Logs = logs[index].SubLogs[innerIxIndex].Logs
			}

			if len(accountKeys)-1 < innerIx.ProgramIdIndex || innerIx.Data == "" {
				continue
			}

			if validateProgramId(accountKeys[innerIx.ProgramIdIndex]) && len(innerIx.Accounts) > 5 {
				accountsCopy := make([]int, len(innerIx.Accounts))
				copy(accountsCopy, innerIx.Accounts)
				processInstructionData.ProgramId = &accountKeys[innerIx.ProgramIdIndex]
				processInstructionData.Accounts = &accountsCopy
			}

			processInstructionData.InnerInstructionIndex = innerIxIndex
			processInstructionData.InnerAccounts = &innerIx.Accounts
			processInstructionData.Data = &innerIx.Data

			s := processInstruction(processInstructionData)
			if s.Pair != "" && s.TokenOut != "" && s.TokenIn != "" {
				s.Source = instructionSource
				innerSwaps = append(innerSwaps, s)
			}
		}
		swaps = append(swaps, innerSwaps...)

	}

	builtSwaps := make([]types.SwapLog, 0)
	for _, swap := range swaps {
		if swap.Wallet == "" || swap.Pair == "" {
			continue
		}

		amountOutFloat, ok := new(big.Float).SetString(swap.AmountOut)
		if !ok || amountOutFloat.Cmp(big.NewFloat(0)) == 0 {
			continue
		}

		amountInFloat, ok := new(big.Float).SetString(swap.AmountIn)
		if !ok || amountInFloat.Cmp(big.NewFloat(0)) == 0 {
			continue
		}

		amountOutF, _ := amountOutFloat.Float64()
		amountInF, _ := amountInFloat.Float64()

		token := ""
		action := ""
		if _, found := QuoteTokens[swap.TokenOut]; found {
			token = swap.TokenIn
			action = "BUY"
		} else if _, found := QuoteTokens[swap.TokenIn]; found {
			token = swap.TokenOut
			action = "SELL"
		}

		sh.tf.AddToQueue(token)
		sh.pf.AddToQueue(PairProcessorQueue{address: swap.Pair, token: &token})

		s := types.SwapLog{
			ID:          tx.Transaction.Signatures[0],
			Wallet:      swap.Wallet,
			Source:      swap.Source,
			BlockNumber: block,
			Timestamp:   time.Unix(timestamp, 0),
			AmountOut:   amountOutF,
			AmountIn:    amountInF,
			Action:      action,
			Pair:        swap.Pair,
			Token:       token,
			Processed:   false,
		}
		builtSwaps = append(builtSwaps, s)
	}

	//if len(builtSwaps) == 0 {
	//	log.Printf("No swaps found for tx: %s", tx.Transaction.Signatures[0])
	//}

	return builtSwaps
}

func processInstruction(instructionData types.ProcessInstructionData) types.SolSwap {
	type handlerFunc func(data types.ProcessInstructionData) types.SolSwap
	handlers := map[string]handlerFunc{
		RAYDIUM_LIQ_POOL_V4:   dex.HandleRaydiumSwaps,
		ORCA_WHIRL_PROGRAM_ID: dex.HandleOrcaSwaps,
		METEORA_DLMM_PROGRAM:  dex.HandleMeteoraSwaps,
		METEORA_POOLS_PROGRAM: dex.HandleMeteoraSwaps,
		PUMPFUN:               dex.HandlePumpFunSwaps,
	}

	accountsLen := len(*instructionData.Accounts)
	programId := *instructionData.ProgramId

	handler, exists := handlers[programId]
	if !exists {
		return types.SolSwap{}
	}

	if programId == ORCA_WHIRL_PROGRAM_ID {
		if (accountsLen != 15 && accountsLen != 11) || instructionData.AccountKeys[(*instructionData.Accounts)[0]] != TOKEN_PROGRAM {
			return types.SolSwap{}
		}
	} else if programId == RAYDIUM_LIQ_POOL_V4 && accountsLen != 18 && accountsLen != 17 {
		return types.SolSwap{}
	} else if programId == METEORA_DLMM_PROGRAM && accountsLen != 18 && accountsLen != 17 {
		return types.SolSwap{}
	}

	return handler(instructionData)

}

func identifySource(programId string) string {
	switch programId {
	case RAYDIUM_LIQ_POOL_V4:
		return "RAYDIUM"
	case ORCA_WHIRL_PROGRAM_ID:
		return "ORCA"
	case METEORA_DLMM_PROGRAM:
		return "METEORA"
	case METEORA_POOLS_PROGRAM:
		return "METEORA"
	case PUMPFUN:
		return "PUMPFUN"
	case JUPITER_V6_AGGREGATOR:
		return "JUPITER"
	default:
		return "UNKNOWN"
	}
}
