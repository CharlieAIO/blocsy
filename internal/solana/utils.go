package solana

import (
	"blocsy/internal/types"
	"github.com/mr-tron/base58"
	pb "github.com/rpcpool/yellowstone-grpc/examples/golang/proto"
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

func validateSupportedDex(programId string) bool {
	switch programId {
	case
		PUMPFUN,
		METEORA_DLMM_PROGRAM,
		RAYDIUM_LIQ_POOL_V4,
		PUMPFUN_AMM,
		RAYDIUM_CONCENTRATED_LIQ,
		RAYDIUM_LAUNCHPAD,
		RAYDIUM_CPMM,
		ORCA_WHIRL_PROGRAM_ID:
		return true
	}
	return false
}

func validateDexInstruction(program string, accounts []int, accountKeys []string) bool {
	if program == ORCA_WHIRL_PROGRAM_ID {
		if len(accounts) == 15 || (len(accounts) == 11 && accountKeys[accounts[0]] == TOKEN_PROGRAM) {
			return true
		}
	}
	if program == RAYDIUM_LIQ_POOL_V4 {
		if len(accounts) == 18 || len(accounts) == 17 && accountKeys[accounts[0]] == TOKEN_PROGRAM {
			return true
		}
	}
	if program == METEORA_DLMM_PROGRAM {
		if len(accounts) >= 15 && accountKeys[accounts[14]] == METEORA_DLMM_PROGRAM {
			return true
		}
	}
	if program == PUMPFUN {
		if len(accounts) >= 12 && accountKeys[accounts[11]] == PUMPFUN {
			return true
		}
	}
	if program == PUMPFUN_AMM {
		if len(accounts) >= 17 && accountKeys[accounts[16]] == PUMPFUN_AMM {
			return true
		}
	}
	if program == RAYDIUM_CONCENTRATED_LIQ {
		if len(accounts) >= 10 && accountKeys[accounts[8]] == TOKEN_PROGRAM {
			return true
		}
	}
	if program == RAYDIUM_CPMM {
		if len(accounts) == 13 {
			return true
		}
	}
	if program == RAYDIUM_LAUNCHPAD {
		if len(accounts) >= 15 && accountKeys[accounts[14]] == RAYDIUM_LAUNCHPAD {
			return true
		}
	}
	if program == PHOENIX {
		if len(accounts) >= 9 && accountKeys[accounts[8]] == TOKEN_PROGRAM && accountKeys[accounts[0]] == PHOENIX {
			return true
		}
	}
	if program == LIFINITY_SWAP_V2 {
		if len(accounts) >= 13 && accountKeys[accounts[9]] == TOKEN_PROGRAM {
			return true
		}
	}
	return false
}

func validateParentProgram(programId string) bool {
	switch programId {
	case PUMPFUN,
		METEORA_DLMM_PROGRAM, METEORA_POOLS_PROGRAM,
		RAYDIUM_LIQ_POOL_V4, RAYDIUM_CONCENTRATED_LIQ, RAYDIUM_CPMM, RAYDIUM_LAUNCHPAD,
		JUPITER_V6_AGGREGATOR,
		LIFINITY_SWAP_V2, PHOENIX,
		PUMPFUN_AMM,
		ORCA_WHIRL_PROGRAM_ID, ORCA_SWAP_V2, ORCA_SWAP:
		return true
	}
	return false
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

func convertToIntSlice(b []byte) []int {
	ints := make([]int, len(b))
	for i, v := range b {
		ints[i] = int(v)
	}
	return ints
}

func convertToBase58Strings(byteArrays [][]byte) []string {
	strings := make([]string, len(byteArrays))
	for i, b := range byteArrays {
		strings[i] = base58.Encode(b)
	}
	return strings
}

func convertToTokenBalanceSlice(input []*pb.TokenBalance) []types.TokenBalance {
	output := make([]types.TokenBalance, len(input))
	for i, v := range input {
		output[i] = types.TokenBalance{
			AccountIndex: int(v.AccountIndex),
			Mint:         v.Mint,
			Owner:        v.Owner,
			ProgramId:    v.ProgramId,
			UITokenAmount: types.UITokenAmount{
				Amount:         v.UiTokenAmount.Amount,
				Decimals:       int(v.UiTokenAmount.Decimals),
				UiAmount:       v.UiTokenAmount.UiAmount,
				UiAmountString: v.UiTokenAmount.UiAmountString,
			},
		}
	}
	return output
}

func convertToInstructions(instructions_ []*pb.CompiledInstruction) []types.Instruction {
	instructions := make([]types.Instruction, len(instructions_))
	for i, instr := range instructions_ {
		instructions[i] = types.Instruction{
			ProgramIdIndex: int(instr.ProgramIdIndex),
			Data:           base58.Encode(instr.Data),
			Accounts:       convertToIntSlice(instr.Accounts),
		}
	}
	return instructions
}

func convertToInnerInstructions(instructions_ []*pb.InnerInstruction) []types.Instruction {
	instructions := make([]types.Instruction, len(instructions_))
	for i, instr := range instructions_ {
		instructions[i] = types.Instruction{
			ProgramIdIndex: int(instr.ProgramIdIndex),
			Data:           base58.Encode(instr.Data),
			Accounts:       convertToIntSlice(instr.Accounts),
		}
	}
	return instructions
}
