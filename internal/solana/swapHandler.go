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

	for index, instruction := range tx.Transaction.Message.Instructions {
		accounts := instruction.Accounts
		innerInstructions, innerIdx := FindInnerIx(tx.Meta.InnerInstructions, index)

		if len(accountKeys)-1 < instruction.ProgramIdIndex {
			continue
		}

		programId := accountKeys[instruction.ProgramIdIndex]

		swap, _ := processInstruction(instruction.Data, accountKeys, programId, accounts, innerIdx, -1, transfers, accounts)
		if swap.Pair != "" || swap.TokenOut != "" || swap.TokenIn != "" {
			swaps = append(swaps, swap)
		}

		innerSwaps := CheckInnerTx(accountKeys, transfers, innerInstructions, innerIdx, accounts)
		if len(innerSwaps) > 0 {
			swaps = append(swaps, innerSwaps...)
		}

	}

	builtSwaps := make([]types.SwapLog, 0)
	for _, swap := range swaps {
		log.Printf("processing swap: %+v", swap)
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
			log.Println("error finding pair:", err)
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

func CheckInnerTx(accountKeys []string, transfers []types.SolTransfer, instructions []types.Instruction, innerIndex int, ixAccounts []int) []types.SolSwap {
	swaps := make([]types.SolSwap, 0)

	for index := 0; index < len(instructions); index++ {
		ix := instructions[index]
		indexIncrement := 0
		accounts := ix.Accounts
		programId := accountKeys[ix.ProgramIdIndex]

		if len(accounts) < 4 && ix.Data == "" {
			continue
		}

		if len(accountKeys)-1 < ix.ProgramIdIndex {
			continue
		}

		s, indexIncrement := processInstruction(ix.Data, accountKeys, programId, accounts, innerIndex, index, transfers, ixAccounts)
		if s.Pair == "" || s.TokenOut == "" || s.TokenIn == "" {
			continue
		}
		swaps = append(swaps, s)

		index += indexIncrement

	}
	return swaps

}

func processInstruction(ixData string, accountKeys []string, programId string, accounts []int, innerIndex int, index int, transfers []types.SolTransfer, ixAccounts []int) (types.SolSwap, int) {
	type handlerFunc func(string, int, int, []types.SolTransfer) (types.SolSwap, int)
	handlers := map[string]handlerFunc{
		RAYDIUM_LIQ_POOL_V4:   dex.HandleRaydiumSwaps,
		ORCA_WHIRL_PROGRAM_ID: dex.HandleOrcaSwaps,
		METEORA_DLMM_PROGRAM:  dex.HandleMeteoraSwaps,
		METEORA_POOLS_PROGRAM: dex.HandleMeteoraSwaps,
		//PHOENIX: dex.HandlePhoenixSwaps,
		//FLUXBEAM_PROGRAM:      dex.HandleFluxbeamSwaps,
		//LIFINITY_SWAP_V2: dex.HandleLifinitySwaps,
		PUMPFUN: dex.HandlePumpFunSwaps,
	}

	handler, exists := handlers[programId]
	if !exists {
		return types.SolSwap{}, 0
	}

	if programId == ORCA_WHIRL_PROGRAM_ID {
		if len(accounts) != 11 || accountKeys[accounts[0]] != TOKEN_PROGRAM {
			return types.SolSwap{}, 0
		}
	} else if programId == RAYDIUM_LIQ_POOL_V4 && len(accounts) != 18 && len(accounts) != 17 {
		return types.SolSwap{}, 0
	} else if programId == METEORA_DLMM_PROGRAM && len(accounts) != 18 {
		return types.SolSwap{}, 0
	}

	s, indexIncrement := handler(ixData, innerIndex, index, transfers)

	setPairField(&s, programId, accounts, ixAccounts, accountKeys)

	return s, indexIncrement
}

func setPairField(s *types.SolSwap, programId string, accounts []int, ixAccounts []int, accountKeys []string) {
	switch programId {
	case ORCA_WHIRL_PROGRAM_ID:
		if len(accounts) >= 3 && len(accountKeys) > accounts[2] {
			s.Pair = accountKeys[accounts[2]]
		}
	case RAYDIUM_LIQ_POOL_V4:
		if len(accounts) >= 2 && len(accountKeys) > accounts[1] {
			s.Pair = accountKeys[accounts[1]]
		}
	case METEORA_DLMM_PROGRAM, METEORA_POOLS_PROGRAM, FLUXBEAM_PROGRAM:
		if len(accounts) >= 1 && len(accountKeys) > accounts[0] {
			s.Pair = accountKeys[accounts[0]]
		}
	case PUMPFUN:
		if len(accounts) >= 3 && len(accountKeys) > accounts[3] {
			log.Printf("setting pumpfun pair: %s", accountKeys[accounts[3]])
			s.Pair = accountKeys[ixAccounts[3]]
		}
	}
}

func FindInnerIx(instructions []types.InnerInstruction, idxMatch int) ([]types.Instruction, int) {
	for i := range instructions {
		if instructions[i].Index == idxMatch {
			return instructions[i].Instructions, instructions[i].Index
		}
	}
	return []types.Instruction{}, -1
}
