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
		if found, _ := IgnoreToUsers[transfer.ToUserAccount]; found {
			continue
		}
		if found, _ := IgnorePrograms[transfer.ParentProgramId]; found {
			continue
		}
		swap, inc := processInstruction(i, transfers, accountKeys)
		if found, _ := IgnoreTokens[swap.TokenOut]; found || IgnoreTokens[swap.TokenIn] {
			continue
		}

		if swap.Wallet != "" && swap.Pair != "" && validateSupportedDex(transfer.ParentProgramId) {
			swaps = append(swaps, swap)
		} else if transfer.Type != "native" && !validateProgramIsDex(transfer.ParentProgramId) {
			transferSwap := types.SolSwap{
				TokenIn:   transfer.Mint,
				Wallet:    transfer.ToUserAccount,
				AmountIn:  transfer.Amount,
				AmountOut: "0",
			}
			swaps = append(swaps, transferSwap)
		}
		i += inc
	}

	builtSwaps := make([]types.SwapLog, 0)
	for _, swap := range swaps {
		if swap.Wallet == "" || swap.TokenIn == "" {
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
		if amountOutF == 0 {
			if _, found := QuoteTokens[swap.TokenIn]; found {
				continue
			}
			token = swap.TokenIn
			action = "TRANSFER"
		} else if _, found := QuoteTokens[swap.TokenOut]; found {
			token = swap.TokenIn
			action = "BUY"
		} else if _, found := QuoteTokens[swap.TokenIn]; found {
			token = swap.TokenOut
			action = "SELL"
		} else {
			action = "UNKNOWN"
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
	}

	//if len(builtSwaps) == 0 {
	//	log.Printf("No swaps found for tx: %s", tx.Transaction.Signatures[0])
	//}

	return builtSwaps
}

func processInstruction(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int) {
	if index+1 >= len(transfers) {
		return types.SolSwap{}, 0
	}
	type handlerFunc func(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int)
	handlers := map[string]handlerFunc{
		RAYDIUM_LIQ_POOL_V4:   dex.HandleRaydiumSwaps,
		ORCA_WHIRL_PROGRAM_ID: dex.HandleOrcaSwaps,
		METEORA_DLMM_PROGRAM:  dex.HandleMeteoraSwaps,
		//METEORA_POOLS_PROGRAM: dex.HandleMeteoraSwaps,
		PUMPFUN: dex.HandlePumpFunSwaps,
	}

	accountsLen := len(transfers[index].IxAccounts)
	programId := transfers[index].ParentProgramId

	handler, exists := handlers[programId]
	if !exists {
		return types.SolSwap{}, 0
	}

	if programId == ORCA_WHIRL_PROGRAM_ID {
		if (accountsLen != 15 && accountsLen != 11) || accountKeys[transfers[index].IxAccounts[0]] != TOKEN_PROGRAM {
			return types.SolSwap{}, 0
		}
	} else if programId == RAYDIUM_LIQ_POOL_V4 && accountsLen != 18 && accountsLen != 17 {
		return types.SolSwap{}, 0
	} else if programId == METEORA_DLMM_PROGRAM && accountsLen != 18 && accountsLen != 17 {
		return types.SolSwap{}, 0
	}

	return handler(index, transfers, accountKeys)

}
