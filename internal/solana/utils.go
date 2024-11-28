package solana

import (
	"blocsy/internal/types"
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
			key == ORCA_WHIRL_PROGRAM_ID {
			return true
		}

	}
	return false
}

func validateProgramId(programId string) bool {
	switch programId {
	case PUMPFUN, METEORA_DLMM_PROGRAM, METEORA_POOLS_PROGRAM, RAYDIUM_LIQ_POOL_V4, ORCA_WHIRL_PROGRAM_ID:
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
