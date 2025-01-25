package solana

import (
	"blocsy/internal/types"
	"github.com/mr-tron/base58"
	"github.com/near/borsh-go"
)

func DecodeMetaplexData(encodedData string) (*struct {
	Instruction       uint8
	Data              types.MetaplexData
	IsMutable         bool
	CollectionDetails *borsh.Enum
}, error) {

	decodedBytes, err := base58.Decode(encodedData)
	if err != nil {
		return nil, err
	}

	deserialized := struct {
		Instruction       uint8
		Data              types.MetaplexData
		IsMutable         bool
		CollectionDetails *borsh.Enum
	}{}

	err = borsh.Deserialize(&deserialized, decodedBytes)
	if err != nil {
		return nil, err
	}

	return &deserialized, nil
}
