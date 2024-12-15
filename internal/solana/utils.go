package solana

import (
	"blocsy/internal/types"
	"math/big"
	"strings"
)

func getAllAccountKeys(tx *types.SolanaTx) []string {
	if tx.Meta.LogMessages == nil {
		return tx.Transaction.Message.AccountKeys
	}
	keys := append(tx.Transaction.Message.AccountKeys, tx.Meta.LoadedAddresses.Writable...)
	keys = append(keys, tx.Meta.LoadedAddresses.Readonly...)
	return keys
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
func validateTX(tx *types.SolanaTx) bool {
	accountKeys := getAllAccountKeys(tx)
	//	validate tx to make sure it contains at least 1 address that we are interested in
	for _, key := range accountKeys {
		if key == PUMPFUN ||
			key == METEORA_DLMM_PROGRAM ||
			key == METEORA_POOLS_PROGRAM ||
			key == RAYDIUM_LIQ_POOL_V4 ||
			key == TOKEN_PROGRAM ||
			key == ORCA_WHIRL_PROGRAM_ID {
			return true
		}

	}
	return false
}

func validateProgramIsDex(programId string) bool {
	switch programId {
	case PUMPFUN,
		METEORA_DLMM_PROGRAM, METEORA_POOLS_PROGRAM,
		RAYDIUM_LIQ_POOL_V4, RAYDIUM_CONCENTRATED_LIQ,
		LIFINITY_SWAP_V2, PHOENIX,
		ORCA_WHIRL_PROGRAM_ID, ORCA_SWAP_V2, ORCA_SWAP:
		return true
	}
	return false
}

func FindInnerIx(instructions []types.InnerInstruction, idxMatch int) ([]types.Instruction, int) {
	for i := range instructions {
		if instructions[i].Index == idxMatch {
			return instructions[i].Instructions, instructions[i].Index
		}
	}
	return []types.Instruction{}, -1
}

func FindAccountKeyIndex(keyMap map[string]int, key string) (int, bool) {
	if i, ok := keyMap[key]; ok {
		return i, true
	}

	return -1, false
}

func GetLogs(logs []string) []types.LogDetails {
	details := make([]types.LogDetails, 0)
	var current types.LogDetails
	var stack []types.LogDetails

	for _, l := range logs {
		if strings.Contains(l, "invoke") {
			if current.Program != "" {
				stack = append(stack, current)
			}
			current = types.LogDetails{
				Program: strings.Fields(l)[1],
			}
		} else if strings.Contains(l, "Program log:") || strings.Contains(l, "Program data:") {
			current.Logs = append(current.Logs, l)
		} else if strings.Contains(l, "success") {
			if len(stack) > 0 {
				parent := stack[len(stack)-1]
				stack = stack[:len(stack)-1]
				parent.SubLogs = append(parent.SubLogs, current)
				current = parent
			} else {
				details = append(details, current)
				current = types.LogDetails{}
			}
		}
	}

	if current.Program != "" {
		details = append(details, current)
	}

	return details
}

func ABSValue(amount string) string {
	amountFloat, ok := new(big.Float).SetString(amount)
	if !ok {
		amountFloat = new(big.Float).SetInt64(0)
	}
	amountFloat.Abs(amountFloat)
	amount = amountFloat.Text('f', -1)
	return amount
}
