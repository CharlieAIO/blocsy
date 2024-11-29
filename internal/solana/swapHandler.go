package solana

import (
	"blocsy/internal/solana/dex"
	"blocsy/internal/types"
	"context"
	"errors"
	"log"
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
	source := "UNKNOWN"
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

		processInstructionData.ProgramId = &programId
		processInstructionData.Accounts = &accounts
		processInstructionData.InnerIndex = &innerIdx
		processInstructionData.Data = &instruction.Data

		processInstructionData.InnerInstructionIndex = -1

		swap := processInstruction(processInstructionData)
		if swap.Pair != "" || swap.TokenOut != "" || swap.TokenIn != "" {
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
				innerSwaps = append(innerSwaps, s)
			}
		}
		swaps = append(swaps, innerSwaps...)

	}

	builtSwaps := make([]types.SwapLog, 0)
	for _, swap := range swaps {
		if swap.Wallet == "" || swap.Pair == "" {
			log.Printf("%s ~ missing ... swap: %+v", tx.Transaction.Signatures[0], swap)
			continue
		}
		action := ""

		amountOutFloat, ok := new(big.Float).SetString(swap.AmountOut)
		if !ok || amountOutFloat.Cmp(big.NewFloat(0)) == 0 {
			continue
		}

		amountInFloat, ok := new(big.Float).SetString(swap.AmountIn)
		if !ok || amountInFloat.Cmp(big.NewFloat(0)) == 0 {
			continue
		}

		timeoutCtx, cancelCtx := context.WithTimeout(ctx, 30*time.Second)
		defer cancelCtx()

		var token_ *string

		if swap.Exchange == "PUMPFUN" {
			if swap.TokenOut == "So11111111111111111111111111111111111111112" {
				token_ = &swap.TokenIn
			} else {
				token_ = &swap.TokenOut
			}
		}

		pairDetails, _, err := sh.pf.FindPair(timeoutCtx, swap.Pair, token_)
		if err != nil {
			log.Printf("%s error finding pair: %v", tx.Transaction.Signatures[0], err)
			if errors.Is(err, types.TokenNotFound) {
				log.Println("pair not found:", err)
			}
			cancelCtx()
			continue
		}

		quoteTokenAddress := pairDetails.QuoteToken.Address

		_, err = sh.tf.FindToken(timeoutCtx, pairDetails.Token)
		if err != nil {
			log.Println("error finding token:", err)
			cancelCtx()
			continue
		}
		cancelCtx()

		if swap.TokenOut == quoteTokenAddress {
			action = "BUY"
		} else {
			action = "SELL"
		}

		amountOutF, _ := amountOutFloat.Float64()
		amountInF, _ := amountInFloat.Float64()

		price := 0.0
		if action == "BUY" {
			priceFloat := new(big.Float).Quo(amountOutFloat, amountInFloat)
			price, _ = priceFloat.Float64()
		} else {
			priceFloat := new(big.Float).Quo(amountInFloat, amountOutFloat)
			price, _ = priceFloat.Float64()
		}

		s := types.SwapLog{
			ID:          tx.Transaction.Signatures[0],
			Wallet:      swap.Wallet,
			Network:     "solana",
			Exchange:    source,
			BlockNumber: block,
			BlockHash:   "",
			Timestamp:   time.Unix(timestamp, 0),
			Type:        action,
			AmountOut:   amountOutF,
			AmountIn:    amountInF,
			Price:       price,
			Pair:        swap.Pair,
			LogIndex:    "0",
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

	//log.Printf("Program ID: %s | %d", programId, accountsLen)

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
