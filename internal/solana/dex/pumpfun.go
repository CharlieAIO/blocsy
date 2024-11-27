package dex

import (
	"blocsy/internal/types"
	"encoding/hex"
	"github.com/mr-tron/base58"
	"strconv"
	"strings"
)

const TRADE_EVENT_DISCRIMINATOR = "bddb7fd34ee661ee"

//func HandlePumpFunSwaps(tx *types.SolanaTx, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
//	increment := 0
//	s := types.SolSwap{}
//
//	for _, innerIx := range tx.Meta.InnerInstructions {
//		if innerIx.Index == innerIndex {
//			for index, ix := range innerIx.Instructions {
//				if index < ixIndex {
//					continue
//				}
//				increment++
//
//				bytesData, _ := base58.Decode(ix.Data)
//				if bytesData == nil {
//					continue
//				}
//
//				hexData := hex.EncodeToString(bytesData)
//				pos := strings.Index(hexData, TRADE_EVENT_DISCRIMINATOR)
//				if pos != -1 {
//					log.Println(innerIndex, ixIndex, index)
//					slicedHexData := hexData[pos+len(TRADE_EVENT_DISCRIMINATOR):]
//					decodedBytes, err := hex.DecodeString(slicedHexData)
//					if err != nil {
//						continue
//					}
//
//					log.Printf("sliceHexData: %s", slicedHexData)
//
//					swap_ := types.PumpFunSwap{}
//					swap_.Decode(decodedBytes)
//					if swap_.Mint.String() == "" {
//						continue
//					}
//
//					s.TokenIn = swap_.Mint.String()
//					s.Wallet = swap_.User.String()
//					s.Pair = swap_.Mint.String()
//					s.Exchange = "PUMPFUN"
//
//					if swap_.IsBuy {
//						s.TokenOut = "So11111111111111111111111111111111111111112"
//						s.TokenIn = swap_.Mint.String()
//						s.AmountOut = strconv.FormatUint(swap_.SolAmount, 10)
//						s.AmountIn = strconv.FormatUint(swap_.TokenAmount, 10)
//					} else {
//						s.TokenIn = "So11111111111111111111111111111111111111112"
//						s.TokenOut = swap_.Mint.String()
//						s.AmountOut = strconv.FormatUint(swap_.TokenAmount, 10)
//						s.AmountIn = strconv.FormatUint(swap_.SolAmount, 10)
//					}
//					break
//				} else {
//					continue
//				}
//
//			}
//
//		}
//	}
//	return s, increment
//}

func HandlePumpFunSwaps(ixData string, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
	increment := 1
	s := types.SolSwap{}

	bytesData, _ := base58.Decode(ixData)
	if bytesData == nil {
		return types.SolSwap{}, 0
	}

	hexData := hex.EncodeToString(bytesData)
	pos := strings.Index(hexData, TRADE_EVENT_DISCRIMINATOR)
	if pos != -1 {
		slicedHexData := hexData[pos+len(TRADE_EVENT_DISCRIMINATOR):]
		decodedBytes, err := hex.DecodeString(slicedHexData)
		if err != nil {
			return types.SolSwap{}, 0
		}

		swap_ := types.PumpFunSwap{}
		swap_.Decode(decodedBytes)
		if swap_.Mint.String() == "" {
			return types.SolSwap{}, 0
		}

		s.TokenIn = swap_.Mint.String()
		s.Wallet = swap_.User.String()
		s.Pair = swap_.Mint.String()
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
		return types.SolSwap{}, 0
	}

	return s, increment
}
