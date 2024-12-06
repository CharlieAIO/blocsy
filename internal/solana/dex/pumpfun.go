package dex

import (
	"blocsy/internal/types"
	"encoding/hex"
	"github.com/mr-tron/base58"
	"log"
	"math"
	"math/big"
	"strconv"
	"strings"
)

const TRADE_EVENT_DISCRIMINATOR = "bddb7fd34ee661ee"

func HandlePumpFunSwaps(instructionData *types.ProcessInstructionData) types.SolSwap {

	if len(*instructionData.Accounts) < 4 || len(instructionData.AccountKeys) < (*instructionData.Accounts)[3] {
		log.Printf("Not enough accounts for PUMPFUN swap")
		return types.SolSwap{}
	}

	s := types.SolSwap{}
	var tokenOutDecimals, tokenInDecimals int

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
			tokenOutDecimals = 9
			tokenInDecimals = 6
			s.TokenIn = swap_.Mint.String()
			s.AmountOut = strconv.FormatUint(swap_.SolAmount, 10)
			s.AmountIn = strconv.FormatUint(swap_.TokenAmount, 10)
		} else {
			tokenOutDecimals = 6
			tokenInDecimals = 9
			s.TokenIn = "So11111111111111111111111111111111111111112"
			s.TokenOut = swap_.Mint.String()
			s.AmountOut = strconv.FormatUint(swap_.TokenAmount, 10)
			s.AmountIn = strconv.FormatUint(swap_.SolAmount, 10)
		}
	} else {
		return types.SolSwap{}
	}

	amountOutFloat, ok := new(big.Float).SetString(s.AmountOut)
	if !ok || amountOutFloat.Cmp(big.NewFloat(0)) == 0 {
		return types.SolSwap{}
	}

	amountInFloat, ok := new(big.Float).SetString(s.AmountIn)
	if !ok || amountInFloat.Cmp(big.NewFloat(0)) == 0 {
		return types.SolSwap{}
	}

	instructionData.Transfers = removeTransfer(instructionData.Transfers, *instructionData.InnerIndex)

	amountOutFloat.Quo(amountOutFloat, new(big.Float).SetFloat64(math.Pow10(tokenOutDecimals)))
	amountInFloat.Quo(amountInFloat, new(big.Float).SetFloat64(math.Pow10(tokenInDecimals)))
	s.Pair = instructionData.AccountKeys[(*instructionData.Accounts)[3]]
	s.AmountOut = amountOutFloat.String()
	s.AmountIn = amountInFloat.String()

	return s
}
