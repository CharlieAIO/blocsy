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
	swaps := make([]types.SolSwap, 0)
	source := ""
	accountKeys := getAllAccountKeys(tx)
	for index, instruction := range tx.Transaction.Message.Instructions {
		innerInstructions, innerIdx := FindInnerIx(tx.Meta.InnerInstructions, index)
		accounts := instruction.Accounts

		s := types.SolSwap{}

		if len(accountKeys)-1 < instruction.ProgramIdIndex {
			continue
		}

		programId := accountKeys[instruction.ProgramIdIndex]

		if programId == RAYDIUM_LIQ_POOL_V4 {
			if len(accounts) != 18 && len(accounts) != 17 {
				continue
			}
			source = "RAYDIUM"
			s, _ = dex.HandleRaydiumSwaps(tx, innerIdx, -1, transfers)
			if len(accounts) >= 2 && (len(accountKeys)-1) >= accounts[1] {
				s.Pair = accountKeys[accounts[1]]
			}
		} else if programId == ORCA_WHIRL_PROGRAM_ID {
			if len(accounts) != 11 {
				continue
			}
			if len(accountKeys) > accounts[0] && accountKeys[accounts[0]] != TOKEN_PROGRAM {
				continue
			}
			source = "ORCA"
			s, _ = dex.HandleOrcaSwaps(tx, innerIdx, -1, transfers)
			if len(accounts) >= 3 && (len(accountKeys)-1) >= accounts[2] {
				s.Pair = accountKeys[accounts[2]]
			}
		} else if programId == FLUXBEAM_PROGRAM {
			source = "FLUXBEAM"
			s, _ = dex.HandleFluxbeamSwaps(tx, innerIdx, -1, transfers)
			if len(accounts) >= 1 && (len(accountKeys)-1) >= accounts[0] {
				s.Pair = accountKeys[accounts[0]]
			}
		} else if programId == METEORA_DLMM_PROGRAM {
			if len(accounts) != 18 {
				continue
			}
			source = "METEORA"
			s, _ = dex.HandleMeteoraSwaps(tx, innerIdx, -1, transfers)
			if len(accounts) >= 1 && (len(accountKeys)-1) >= accounts[0] {
				s.Pair = accountKeys[accounts[0]]
			}
		} else if programId == METEORA_POOLS_PROGRAM {
			source = "METEORA"
			s, _ = dex.HandleMeteoraSwaps(tx, innerIdx, -1, transfers)
			if len(accounts) >= 1 && (len(accountKeys)-1) >= accounts[0] {
				s.Pair = accountKeys[accounts[0]]
			}
		} else if programId == PUMPFUN {
			source = "PUMPFUN"
			s, _ = dex.HandlePumpFunSwaps(tx, innerIdx, -1, transfers)
			if len(accounts) >= 2 && (len(accountKeys)-1) >= accounts[1] {
				s.Pair = accountKeys[accounts[1]]
			}
		} else {
			if programId == JUPITER_V6_AGGREGATOR {
				source = "JUPITER"
			}
			if programId == RAYDIUM_AMM_ROUTING {
				source = "RAYDIUM_ROUTING"
			}

			s_ := CheckInnerTx(tx, transfers, innerInstructions, innerIdx)
			swaps = append(swaps, s_...)
		}

		if s.TokenOut != "" || s.TokenIn != "" {
			swaps = append(swaps, s)
		}

	}
	if source == "" {
		source = "UNKNOWN"
	}

	builtSwaps := make([]types.SwapLog, 0)
	for _, swap := range swaps {
		//log.Println("Swap:", swap)
		if swap.Wallet == "" || swap.Pair == "" {
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

		if swap.Exchange == "PUMPFUN" {
			tokenOutDecimals := 0
			tokenInDecimals := 0
			if swap.TokenOut == "So11111111111111111111111111111111111111112" {
				tokenDetails, err := sh.tf.FindToken(ctx, swap.TokenIn)
				if err != nil {
					continue
				}
				if err != nil {
					if errors.Is(err, types.TokenNotFound) {
						continue
					}
					log.Println("Error finding token:", err)
					continue
				}

				swap.Pair = swap.TokenIn
				tokenOutDecimals = 9
				tokenInDecimals = int(tokenDetails.Decimals)
				action = "BUY"
			} else {
				tokenDetails, err := sh.tf.FindToken(ctx, swap.TokenOut)
				if err != nil {
					continue
				}
				swap.Pair = swap.TokenOut
				tokenOutDecimals = int(tokenDetails.Decimals)
				tokenInDecimals = 9
				action = "SELL"
			}

			amountOutFloat.Quo(amountOutFloat, new(big.Float).SetFloat64(math.Pow10(tokenOutDecimals)))
			amountInFloat.Quo(amountInFloat, new(big.Float).SetFloat64(math.Pow10(tokenInDecimals)))

		} else {
			timeoutCtx, cancel := context.WithTimeout(ctx, 45*time.Second)
			defer cancel()

			pairDetails, _, err := sh.pf.FindPair(timeoutCtx, swap.Pair)
			if err != nil {
				if errors.Is(err, types.TokenNotFound) {
					log.Println("pair not found:", err)
					continue
				}
				continue
			}

			quoteTokenAddress := pairDetails.QuoteToken.Address

			_, err = sh.tf.FindToken(ctx, pairDetails.Token)
			if err != nil {
				continue
			}

			ctx.Done()

			if swap.TokenOut == quoteTokenAddress {
				action = "BUY"
			} else {
				action = "SELL"
			}
		}

		amountOutF, _ := amountOutFloat.Float64()
		amountInF, _ := amountInFloat.Float64()

		price := 0.0
		if action == "BUY" {
			priceFloat := new(big.Float).Quo(amountOutFloat, amountInFloat)
			// price = priceFloat.Text('f', -1)
			price, _ = priceFloat.Float64()
		} else {
			priceFloat := new(big.Float).Quo(amountInFloat, amountOutFloat)
			// price = priceFloat.Text('f', -1)
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

		//if err := sh.sRepo.InsertSwap(ctx, s); err != nil {
		//	log.Println("Error inserting swap:", err)
		//	continue
		//}
	}

	return builtSwaps
}

func CheckInnerTx(tx *types.SolanaTx, transfers []types.SolTransfer, instructions []types.Instruction, innerIndex int) []types.SolSwap {
	swaps := make([]types.SolSwap, 0)

	for index := 0; index < len(instructions); index++ {
		ix := instructions[index]
		s, indexIncrement := HandleTxSwap(ix, tx, innerIndex, index, transfers)
		if s != nil {
			if s.TokenOut == "" || s.TokenIn == "" {
				continue
			}
			swaps = append(swaps, *s)
		}
		index += indexIncrement

	}
	return swaps

}

func HandleTxSwap(ix types.Instruction, tx *types.SolanaTx, innerIndex int, index int, transfers []types.SolTransfer) (*types.SolSwap, int) {
	s := types.SolSwap{}
	indexIncrement := 0
	accounts := ix.Accounts
	accountKeys := getAllAccountKeys(tx)

	if len(accounts) < 4 && ix.Data == "" {
		return nil, indexIncrement
	}

	if len(accountKeys)-1 < ix.ProgramIdIndex {
		return nil, indexIncrement
	}

	programId := accountKeys[ix.ProgramIdIndex]

	switch programId {
	case ORCA_WHIRL_PROGRAM_ID:
		if len(accounts) != 11 {
			return nil, indexIncrement
		}
		if accountKeys[accounts[0]] != TOKEN_PROGRAM {
			return nil, indexIncrement
		}
		s, indexIncrement = dex.HandleOrcaSwaps(tx, innerIndex, index, transfers)
		if len(accounts) >= 3 && (len(accountKeys)-1) >= accounts[2] {
			s.Pair = accountKeys[accounts[2]]
		}
	case RAYDIUM_LIQ_POOL_V4:
		if len(accounts) != 18 && len(accounts) != 17 {
			return nil, indexIncrement
		}
		s, indexIncrement = dex.HandleRaydiumSwaps(tx, innerIndex, index, transfers)
		if len(accounts) >= 2 && (len(accountKeys)-1) >= accounts[1] {
			s.Pair = accountKeys[accounts[1]]
		}
	case RAYDIUM_CONCENTRATED_LIQ:
		return &s, indexIncrement
	case METEORA_DLMM_PROGRAM:
		if len(accounts) != 18 {
			return nil, indexIncrement
		}
		s, indexIncrement = dex.HandleMeteoraSwaps(tx, innerIndex, index, transfers)
		if len(accounts) >= 1 && (len(accountKeys)-1) >= accounts[0] {
			s.Pair = accountKeys[accounts[0]]
		}
	case METEORA_POOLS_PROGRAM:
		s, indexIncrement = dex.HandleMeteoraSwaps(tx, innerIndex, index, transfers)
		if len(accounts) >= 1 && (len(accountKeys)-1) >= accounts[0] {
			s.Pair = accountKeys[accounts[0]]
		}
	case FLUXBEAM_PROGRAM:
		s, indexIncrement = dex.HandleFluxbeamSwaps(tx, innerIndex, index, transfers)
		if len(accounts) >= 1 && (len(accountKeys)-1) >= accounts[0] {
			s.Pair = accountKeys[accounts[0]]
		}
	case PUMPFUN:
		s, indexIncrement = dex.HandlePumpFunSwaps(tx, innerIndex, index, transfers)
	default:
		return nil, indexIncrement
	}

	return &s, indexIncrement
}

func FindInnerIx(instructions []types.InnerInstruction, idxMatch int) ([]types.Instruction, int) {
	for i := range instructions {
		if instructions[i].Index == idxMatch {
			return instructions[i].Instructions, instructions[i].Index
		}
	}
	return []types.Instruction{}, -1
}
