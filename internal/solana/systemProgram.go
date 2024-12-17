package solana

import (
	"blocsy/internal/types"
	"github.com/blocto/solana-go-sdk/common"
	"github.com/mr-tron/base58"
	"log"
)

func determineSystemInstructionType(tag byte) string {
	switch tag {
	case 0:
		return "CreateAccount"
	case 2:
		return "Transfer"
	case 8:
		return "Allocate"
	case 1:
		return "Assign"
	default:
		return "InvalidInstruction"
	}

}

func DecodeSystemProgramData(encodedData string) types.SystemProgramData {
	decodedBytes, err := base58.Decode(encodedData)
	if err != nil {
		log.Fatalf("Error decoding Base58: %v", err)
	}

	data := types.SystemProgramData{}
	data.RawType = decodedBytes[0]
	data.Type = determineSystemInstructionType(decodedBytes[0])

	switch data.RawType {
	case 0:
		if len(decodedBytes) < 52 {
			return data
		}
		lamports, _ := unpackU64(decodedBytes[4:12])
		data.Lamports = lamports

		space, _ := unpackU64(decodedBytes[12:20])
		data.Space = space

		data.ProgramID = common.PublicKeyFromBytes(decodedBytes[20:52])
	case 2, 8:
		if len(decodedBytes) < 12 {
			return data
		}

		lamports, _ := unpackU64(decodedBytes[4:])
		data.Lamports = lamports
	case 1:
		if len(decodedBytes) < 36 {
			return data
		}
		data.ProgramID = common.PublicKeyFromBytes(decodedBytes[4:])

	}

	return data
}
