package solana

import (
	"blocsy/internal/types"
	"encoding/binary"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/mr-tron/base58"
	"log"
)

func determineTokenInstructionType(tag byte) string {
	switch tag {
	case 0:
		return "InitializeMint"
	case 1:
		return "InitializeAccount"
	case 2:
		return "InitializeMultisig"
	case 3:
		return "Transfer"
	case 4:
		return "Approve"
	case 5:
		return "Revoke"
	case 6:
		return "SetAuthority"
	case 7:
		return "MintTo"
	case 8:
		return "Burn"
	case 9:
		return "CloseAccount"
	case 10:
		return "FreezeAccount"
	case 11:
		return "ThawAccount"
	case 12:
		return "TransferChecked"
	case 13:
		return "ApproveChecked"
	case 14:
		return "MintToChecked"
	case 15:
		return "BurnChecked"
	case 16:
		return "InitializeAccount2"
	case 17:
		return "SyncNative"
	case 18:
		return "InitializeAccount3"
	case 19:
		return "InitializeMultisig2"
	case 20:
		return "InitializeMint2"
	case 21:
		return "GetAccountDataSize"
	case 22:
		return "InitializeImmutableOwner"
	case 23:
		return "AmountToUiAmount"
	case 24:
		return "UiAmountToAmount"
	default:
		return "InvalidInstruction"
	}
}

func unpackU64(data []byte) (uint64, []byte) {
	if len(data) < 8 {
		log.Fatalf("Insufficient data to decode u64")
	}
	value := binary.LittleEndian.Uint64(data[:8])
	return value, data[8:]
}

func DecodeTokenProgramData(encodedData string) types.TokenProgramData {
	data := types.TokenProgramData{}

	decodedBytes, err := base58.Decode(encodedData)
	if err != nil {
		return data
	}

	// Extract instruction tag
	instructionTag := decodedBytes[0]
	remainingBytes := decodedBytes[1:]

	data.RawType = instructionTag
	data.Type = determineTokenInstructionType(instructionTag)

	// Decode based on instruction tag
	switch instructionTag {
	case 3, 4, 7, 8: // Transfer, Approve, MintTo, Burn
		amount, _ := unpackU64(remainingBytes)
		data.Amount = amount

	case 6: // SetAuthority
		if len(remainingBytes) < 1 {
			return data
		}
		authorityType := remainingBytes[0]
		remainingBytes = remainingBytes[1:]
		data.AuthorityType = authorityType

		if len(remainingBytes) >= 32 {
			data.NewAuthority = common.PublicKeyFromBytes(remainingBytes[:32])
		}

	case 12, 13, 14, 15: // TransferChecked,ApproveChecked,MintToChecked,BurnChecked
		amount, rest := unpackU64(remainingBytes)
		if len(rest) < 1 {
			return data
		}
		decimals := rest[0]
		data.Decimals = int(decimals)
		data.Amount = amount

	case 16, 18: // InitializeAccount2,InitializeAccount3
		if len(remainingBytes) < 32 {
			return data
		}
		data.Owner = common.PublicKeyFromBytes(remainingBytes[:32])

	case 20, 0: // InitializeMint2,InitializeMint
		if len(remainingBytes) < 1 {
			return data
		}
		decimals := remainingBytes[0]
		remainingBytes = remainingBytes[1:]
		data.Decimals = int(decimals)

		if len(remainingBytes) < 32 {
			return data
		}
		data.MintAuthority = common.PublicKeyFromBytes(remainingBytes[:32])

		remainingBytes = remainingBytes[32:]

		if len(remainingBytes) >= 32 {
			data.FreezeAuthority = common.PublicKeyFromBytes(remainingBytes[:32])

		}

	}

	return data
}
