package dex

import (
	"blocsy/internal/types"
	"encoding/hex"
	"github.com/mr-tron/base58/base58"
	"strconv"
	"strings"
)

const TRADE_EVENT_DISCRIMINATOR = "bddb7fd34ee661ee"

func HandlePumpFunSwaps(instructionData types.ProcessInstructionData) types.SolSwap {

	if len(*instructionData.InstructionAccounts) < 4 || len(instructionData.AccountKeys) < (*instructionData.InstructionAccounts)[3] {
		return types.SolSwap{}
	}

	s := types.SolSwap{}

	bytesData, _ := base58.Decode(*instructionData.Data)
	if bytesData == nil {
		return types.SolSwap{}
	}

	hexData := hex.EncodeToString(bytesData)
	pos := strings.Index(hexData, TRADE_EVENT_DISCRIMINATOR)
	if pos != -1 {
		slicedHexData := hexData[pos+len(TRADE_EVENT_DISCRIMINATOR):]
		decodedBytes, err := hex.DecodeString(slicedHexData)
		if err != nil {
			return types.SolSwap{}
		}

		swap_ := types.PumpFunSwap{}
		swap_.Decode(decodedBytes)
		if swap_.Mint.String() == "" {
			return types.SolSwap{}
		}

		s.TokenIn = swap_.Mint.String()
		s.Wallet = swap_.User.String()
		s.Exchange = "PUMPFUN"

		if swap_.IsBuy {
			s.TokenOut = "So11111111111111111111111111111111111111112"
			s.TokenIn = swap_.Mint.String()
			s.AmountOut = strconv.FormatUint(swap_.SolAmount, 10)
			s.AmountIn = strconv.FormatUint(swap_.TokenAmount, 10)
		} else {
			s.TokenIn = "So11111111111111111111111111111111111111112"
			s.TokenOut = swap_.Mint.String()
			s.AmountOut = strconv.FormatUint(swap_.TokenAmount, 10)
			s.AmountIn = strconv.FormatUint(swap_.SolAmount, 10)
		}
	} else {
		return types.SolSwap{}
	}

	s.Pair = instructionData.AccountKeys[(*instructionData.InstructionAccounts)[3]]

	return s
}
