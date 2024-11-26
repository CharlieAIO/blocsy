package dex

import (
	"blocsy/internal/types"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"regexp"
)

func newSwapBaseInLogFromBytes(data []byte) (*types.SwapBaseInLog, error) {
	if len(data) < 56 {
		return nil, fmt.Errorf("data too short for SwapBaseInLog, got %d bytes", len(data))
	}

	return &types.SwapBaseInLog{
		LogType:    types.SWAP_BASE_IN,
		AmountIn:   binary.BigEndian.Uint64(data[1:9]),
		MinimumOut: binary.BigEndian.Uint64(data[9:17]),
		Direction:  binary.BigEndian.Uint64(data[17:25]),
		UserSource: binary.BigEndian.Uint64(data[25:33]),
		PoolCoin:   binary.BigEndian.Uint64(data[33:41]),
		PoolPC:     binary.BigEndian.Uint64(data[41:49]),
		OutAmount:  binary.BigEndian.Uint64(data[49:57]),
	}, nil
}

func newSwapBaseOutLogFromBytes(data []byte) (*types.SwapBaseOutLog, error) {
	if len(data) < 56 {
		return nil, fmt.Errorf("data too short for SwapBaseOutLog")
	}
	return &types.SwapBaseOutLog{
		LogType:    types.SWAP_BASE_OUT,
		MaxIn:      binary.BigEndian.Uint64(data[1:9]),
		AmountOut:  binary.BigEndian.Uint64(data[9:17]),
		Direction:  binary.BigEndian.Uint64(data[17:25]),
		UserSource: binary.BigEndian.Uint64(data[25:33]),
		PoolCoin:   binary.BigEndian.Uint64(data[33:41]),
		PoolPC:     binary.BigEndian.Uint64(data[41:49]),
		DeductIn:   binary.BigEndian.Uint64(data[49:57]),
	}, nil
}

func parseRayLogs(logs []string) []string {
	regex := regexp.MustCompile(`ray_log: ([A-Za-z0-9+/=]+)`)
	var rayLogs []string

	for _, L := range logs {
		matches := regex.FindAllStringSubmatch(L, -1)
		for _, match := range matches {
			if len(match) > 1 {
				rayLogs = append(rayLogs, match[1])
			}
		}
	}

	return rayLogs
}

func decodeRayLog(log string) (interface{}, types.LogType) {
	decoded, err := base64.StdEncoding.DecodeString(log)
	if err != nil {
		return nil, 0 // Use an appropriate default or error value
	}

	logType := types.LogType(decoded[0])

	switch logType {
	case types.SWAP_BASE_IN:
		log, err := newSwapBaseInLogFromBytes(decoded)
		if err != nil {
			println("Error decoding SwapBaseInLog:", err)
			return nil, logType
		}
		return log, logType

	case types.SWAP_BASE_OUT:
		log, err := newSwapBaseOutLogFromBytes(decoded)
		if err != nil {
			println("Error decoding SwapBaseOutLog:", err)
			return nil, logType
		}
		return log, logType

	default:
		return nil, logType
	}
}

func convertToSwapLog(s interface{}) *types.RaySwapLog {
	switch v := s.(type) {
	case *types.SwapBaseInLog:
		return &types.RaySwapLog{
			LogType:    v.LogType,
			AmountIn:   v.AmountIn,
			MinimumOut: v.MinimumOut,
			Direction:  v.Direction,
			UserSource: v.UserSource,
			PoolCoin:   v.PoolCoin,
			PoolPC:     v.PoolPC,
			OutAmount:  v.OutAmount,
		}
	case *types.SwapBaseOutLog:
		return &types.RaySwapLog{
			LogType:    v.LogType,
			MaxIn:      v.MaxIn,
			AmountOut:  v.AmountOut,
			Direction:  v.Direction,
			UserSource: v.UserSource,
			PoolCoin:   v.PoolCoin,
			PoolPC:     v.PoolPC,
			DeductIn:   v.DeductIn,
		}
	default:
		// Handle the case where s is not *SwapBaseInLog or *SwapBaseOutLog
		return nil
	}
}

func HandleRaydiumSwaps(tx *types.SolanaTx, innerIndex int, ixIndex int, transfers []types.SolTransfer) (types.SolSwap, int) {
	//signer := tx.Transaction.Message.AccountKeys[0].Pubkey

	transfer1, ok := FindTransfer(transfers, innerIndex, ixIndex+1)
	if !ok {
		return types.SolSwap{}, 0
	}

	transfer2, ok := FindTransfer(transfers, innerIndex, ixIndex+2)
	if !ok {
		return types.SolSwap{}, 0
	}

	wallet := transfer1.FromUserAccount
	if wallet == "" {
		wallet = transfer2.ToUserAccount
	}

	s := types.SolSwap{
		Exchange:  "RAYDIUM",
		Wallet:    wallet,
		TokenOut:  transfer1.Mint,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
		AmountOut: transfer1.Amount,
	}

	return s, 2

}
