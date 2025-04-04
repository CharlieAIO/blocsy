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

	for i := 0; i < len(transfers); i++ {
		transfer := transfers[i]
		if found, _ := IgnorePrograms[transfer.ParentProgramId]; found {
			continue
		}
		swap, inc := processTransfer(i, transfers, accountKeys)
		if found, _ := IgnoreTokens[swap.TokenOut]; found || IgnoreTokens[swap.TokenIn] {
			continue
		}

		if source := Programs[transfer.ParentProgramId]; source != "" {
			swap.Source = source
		}

		if swap.Wallet != "" && swap.Pair != "" && validateSupportedDex(transfer.ParentProgramId) {
			if swap.TokenIn == swap.TokenOut {
				continue
			}
			swaps = append(swaps, swap)
		} else if transfer.Type != "native" && (validateSupportedDex(transfer.ParentProgramId) || transfer.ParentProgramId == "") {
			if _, found := QuoteTokens[transfer.Mint]; found {
				continue
			}
			transferSwap := types.SolSwap{
				TokenIn:   transfer.Mint,
				Wallet:    transfer.ToUserAccount,
				AmountIn:  transfer.Amount,
				AmountOut: "0",
			}
			transferSwap2 := types.SolSwap{
				TokenOut:  transfer.Mint,
				Wallet:    transfer.FromUserAccount,
				AmountIn:  "0",
				AmountOut: transfer.Amount,
			}

			swaps = append(swaps, transferSwap, transferSwap2)
		}
		i += inc
	}

	builtSwaps := make([]types.SwapLog, 0)
	balanceSheet := map[string]map[string]float64{}
	for _, swap := range swaps {
		if swap.Wallet == "" || (swap.TokenIn == "" && swap.TokenOut == "") {
			continue
		}

		amountOutFloat, ok := new(big.Float).SetString(swap.AmountOut)
		if !ok {
			continue
		}

		amountInFloat, ok := new(big.Float).SetString(swap.AmountIn)
		if !ok {
			continue
		}

		amountOutF, _ := amountOutFloat.Float64()
		amountInF, _ := amountInFloat.Float64()

		token := ""
		action := ""

		if _, found := QuoteTokens[swap.TokenOut]; found {
			token = swap.TokenIn
			action = "BUY"
		} else if _, found = QuoteTokens[swap.TokenIn]; found {
			token = swap.TokenOut
			action = "SELL"
		} else {
			action = "UNKNOWN"
		}

		if action == "UNKNOWN" {
			if amountOutF == 0 {
				if _, foundIn := QuoteTokens[swap.TokenIn]; foundIn {
					continue
				}
				token = swap.TokenIn
				action = "RECEIVE"
			}
			if amountInF == 0 {
				if _, foundOut := QuoteTokens[swap.TokenOut]; foundOut {
					continue
				}
				token = swap.TokenOut
				action = "TRANSFER"
			}
		}

		sh.tf.AddToQueue(token)
		if swap.Pair != "" {
			sh.pf.AddToQueue(PairProcessorQueue{address: swap.Pair, token: &token})
		}

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

		if _, found := balanceSheet[swap.Wallet]; !found {
			balanceSheet[swap.Wallet] = map[string]float64{
				token: amountInF - amountOutF,
			}
		} else {
			balanceSheet[swap.Wallet][token] += amountInF - amountOutF
		}
	}

	finalSwaps := make([]types.SwapLog, 0)
	for _, swap := range builtSwaps {
		if balanceSheet[swap.Wallet][swap.Token] == 0 {
			continue
		}
		if _, found := IgnoreToUsers[swap.Wallet]; found {
			continue
		}
		finalSwaps = append(finalSwaps, swap)
	}

	return finalSwaps
}

func processTransfer(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int) {
	type handlerFunc func(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int)
	handlers := map[string]handlerFunc{
		RAYDIUM_LIQ_POOL_V4:      dex.HandleRaydiumSwaps,
		RAYDIUM_CONCENTRATED_LIQ: dex.HandleRaydiumConcentratedSwaps,
		RAYDIUM_CPMM:             dex.HandleRaydiumCPMMSwaps,
		ORCA_WHIRL_PROGRAM_ID:    dex.HandleOrcaSwaps,
		METEORA_DLMM_PROGRAM:     dex.HandleMeteoraSwaps,
		//METEORA_POOLS_PROGRAM: dex.HandleMeteoraSwaps,
		PUMPFUN:     dex.HandlePumpFunSwaps,
		PUMPFUN_AMM: dex.HandlePumpFunAmmSwaps,
	}

	programId := transfers[index].ParentProgramId

	handler, exists := handlers[programId]
	if !exists {
		return types.SolSwap{}, 0
	}

	if !validateDexInstruction(programId, transfers[index].IxAccounts, accountKeys) {
		return types.SolSwap{}, 0
	}

	return handler(index, transfers, accountKeys)

}
