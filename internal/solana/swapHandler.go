package solana

import (
	"blocsy/internal/solana/dex"
	"blocsy/internal/types"
	"context"
	"errors"
	"log"
	"math"
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
	}

	for index, instruction := range tx.Transaction.Message.Instructions {
		accounts := instruction.Accounts
		innerInstructions, innerIdx := FindInnerIx(tx.Meta.InnerInstructions, index)

		if len(accountKeys)-1 < instruction.ProgramIdIndex {
			continue
		}

		programId := accountKeys[instruction.ProgramIdIndex]

		processInstructionData.ProgramId = &programId
		processInstructionData.InstructionAccounts = &accounts
		processInstructionData.InnerAccounts = &accounts
		processInstructionData.InnerIndex = &innerIdx
		processInstructionData.InnerInstructionIndex = -1
		processInstructionData.Data = &instruction.Data

		swap := processInstruction(processInstructionData, accounts)
		if swap.Pair != "" || swap.TokenOut != "" || swap.TokenIn != "" {
			swaps = append(swaps, swap)
		}

		innerSwaps := make([]types.SolSwap, 0)
		for innerIxIndex, innerIx := range innerInstructions {
			if len(accountKeys)-1 < innerIx.ProgramIdIndex || innerIx.Data == "" {
				continue
			}
			processInstructionData.InnerInstructionIndex = innerIxIndex
			processInstructionData.InnerAccounts = &innerIx.Accounts
			processInstructionData.ProgramId = &accountKeys[innerIx.ProgramIdIndex]
			processInstructionData.Data = &innerIx.Data

			s := processInstruction(processInstructionData, innerIx.Accounts)
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

		timeoutCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
		defer cancel()

		var token_ *string
		if swap.Exchange == "PUMPFUN" {
			var tokenInDecimals, tokenOutDecimals int
			if swap.TokenOut == "So11111111111111111111111111111111111111112" {
				tokenInDecimals = 6
				tokenOutDecimals = 9
				token_ = &swap.TokenIn
			} else {
				tokenInDecimals = 9
				tokenOutDecimals = 6
				token_ = &swap.TokenOut
			}
			amountOutFloat.Quo(amountOutFloat, new(big.Float).SetFloat64(math.Pow10(tokenOutDecimals)))
			amountInFloat.Quo(amountInFloat, new(big.Float).SetFloat64(math.Pow10(tokenInDecimals)))
		}

		pairDetails, _, err := sh.pf.FindPair(timeoutCtx, swap.Pair, token_)
		if err != nil {
			log.Printf("%s error finding pair: %v", tx.Transaction.Signatures[0], err)
			if errors.Is(err, types.TokenNotFound) {
				log.Println("pair not found:", err)
				continue
			}
			continue
		}

		quoteTokenAddress := pairDetails.QuoteToken.Address

		_, err = sh.tf.FindToken(ctx, pairDetails.Token)
		if err != nil {
			log.Println("error finding token:", err)
			continue
		}

		ctx.Done()

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

	return builtSwaps
}

func processInstruction(instructionData types.ProcessInstructionData, accounts []int) types.SolSwap {
	type handlerFunc func(data types.ProcessInstructionData) types.SolSwap
	handlers := map[string]handlerFunc{
		RAYDIUM_LIQ_POOL_V4:   dex.HandleRaydiumSwaps,
		ORCA_WHIRL_PROGRAM_ID: dex.HandleOrcaSwaps,
		METEORA_DLMM_PROGRAM:  dex.HandleMeteoraSwaps,
		METEORA_POOLS_PROGRAM: dex.HandleMeteoraSwaps,
		PUMPFUN:               dex.HandlePumpFunSwaps,
	}

	accountsLen := len(accounts)
	programId := *instructionData.ProgramId

	handler, exists := handlers[programId]
	if !exists {
		return types.SolSwap{}
	}

	if programId == ORCA_WHIRL_PROGRAM_ID {
		if accountsLen != 11 || instructionData.AccountKeys[(accounts)[0]] != TOKEN_PROGRAM {
			return types.SolSwap{}
		}
	} else if programId == RAYDIUM_LIQ_POOL_V4 && accountsLen != 18 && accountsLen != 17 {
		return types.SolSwap{}
	} else if programId == METEORA_DLMM_PROGRAM && accountsLen != 18 {
		return types.SolSwap{}
	}

	return handler(instructionData)

}
