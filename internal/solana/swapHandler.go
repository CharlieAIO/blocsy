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

	//processInstructionData := types.ProcessInstructionData{
	//	AccountKeys:           accountKeys,
	//	Transfers:             transfers,
	//	InnerInstructionIndex: -1,
	//	TokenAccountMap:       CreateTokenAccountMap(tx),
	//}

	for _, transfer := range transfers {

	}

	//for _, transfer := range processInstructionData.Transfers {
	//	if transfer.Type == "token" {
	//		transferSwap := types.SolSwap{
	//			TokenIn:   transfer.Mint,
	//			Wallet:    transfer.ToUserAccount,
	//			AmountIn:  transfer.Amount,
	//			AmountOut: "0",
	//		}
	//		log.Printf("transferSwap: %+v", transferSwap)
	//		swaps = append(swaps, transferSwap)
	//	}
	//}

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
		if _, found := QuoteTokens[swap.TokenOut]; found {
			token = swap.TokenIn
			action = "BUY"
		} else if _, found := QuoteTokens[swap.TokenIn]; found {
			token = swap.TokenOut
			action = "SELL"
		} else {
			token = swap.TokenIn
			action = "TRANSFER"
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

func processInstruction(instructionData *types.ProcessInstructionData) types.SolSwap {
	type handlerFunc func(data *types.ProcessInstructionData) types.SolSwap
	handlers := map[string]handlerFunc{
		RAYDIUM_LIQ_POOL_V4:   dex.HandleRaydiumSwaps,
		ORCA_WHIRL_PROGRAM_ID: dex.HandleOrcaSwaps,
		METEORA_DLMM_PROGRAM:  dex.HandleMeteoraSwaps,
		METEORA_POOLS_PROGRAM: dex.HandleMeteoraSwaps,
		PUMPFUN:               dex.HandlePumpFunSwaps,
		//TOKEN_PROGRAM:         dex.HandleTokenSwaps,
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
