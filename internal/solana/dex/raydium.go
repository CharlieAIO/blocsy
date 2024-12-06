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

func decodeLog(log string) types.RaySwapLog {
	decoded, err := base64.StdEncoding.DecodeString(log)
	if err != nil {
		return types.RaySwapLog{}
	}

	logType := types.LogType(decoded[0])

	decodedBytes, err := base64.StdEncoding.DecodeString(log)
	if err != nil {
		return types.RaySwapLog{}
	}

	switch logType {
	case types.SWAP_BASE_IN:
		baseIn := types.RaySwapBaseIn{}
		err := baseIn.Decode(decodedBytes)
		if err != nil {
			return types.RaySwapLog{}
		}
		return types.RaySwapLog{
			LogType:    types.LogType(baseIn.LogType),
			AmountIn:   baseIn.AmountIn,
			MinimumOut: baseIn.MinimumAmountOut,
			Direction:  baseIn.Direction,
			UserSource: baseIn.UserSource,
			PoolCoin:   baseIn.PoolCoin,
			PoolPC:     baseIn.PoolPc,
			OutAmount:  baseIn.OutAmount,
		}

	case types.SWAP_BASE_OUT:
		baseOut := types.RaySwapBaseOut{}
		err := baseOut.Decode(decodedBytes)
		if err != nil {
			return types.RaySwapLog{}
		}
		return types.RaySwapLog{
			LogType:    types.LogType(baseOut.LogType),
			MaxIn:      baseOut.MaximumAmountIn,
			AmountOut:  baseOut.AmountOut,
			Direction:  baseOut.Direction,
			UserSource: baseOut.UserSource,
			PoolCoin:   baseOut.PoolCoin,
			PoolPC:     baseOut.PoolPc,
			DeductIn:   baseOut.DeductIn,
		}

	default:
		return types.RaySwapLog{}
	}
}

func parseRayLogs(logs []string) []types.RaySwapLog {
	regex := regexp.MustCompile(`ray_log: ([A-Za-z0-9+/=]+)`)
	var rayLogs []types.RaySwapLog

	for _, L := range logs {
		matches := regex.FindAllStringSubmatch(L, -1)
		for _, match := range matches {
			if len(match) > 1 {
				rayLogs = append(rayLogs, decodeLog(match[1]))

			}
		}
	}

	return rayLogs
}

//func HandleRaydiumSwaps(instructionData types.ProcessInstructionData) types.SolSwap {
//	if len(*instructionData.Accounts) < 2 || len(instructionData.AccountKeys) < (*instructionData.Accounts)[1] {
//		log.Printf("Not enough accounts for RAYDIUM swap")
//		return types.SolSwap{}
//	}
//
//	rayLogs := parseRayLogs(instructionData.Logs)
//	if len(rayLogs) == 0 {
//		return types.SolSwap{}
//	}
//
//	sourceTokenAccount := instructionData.AccountKeys[(*instructionData.Accounts)[14]]
//	sourceTokenAccountMint := instructionData.TokenAccountMap[sourceTokenAccount].MintAddress
//	destTokenAccount := instructionData.AccountKeys[(*instructionData.Accounts)[15]]
//	destTokenAccountMint := instructionData.TokenAccountMap[destTokenAccount].MintAddress
//	wallet := instructionData.AccountKeys[(*instructionData.Accounts)[16]]
//
//	if len(*instructionData.Accounts) == 18 {
//		sourceTokenAccount = instructionData.AccountKeys[(*instructionData.Accounts)[15]]
//		destTokenAccount = instructionData.AccountKeys[(*instructionData.Accounts)[16]]
//		wallet = instructionData.AccountKeys[(*instructionData.Accounts)[17]]
//	}
//
//	//tf1, ok := FindTransferFromAccounts(instructionData.Transfers, sourceTokenAccount)
//	//if !ok {
//	//	log.Printf("Could not find transfer for RAYDIUM swap")
//	//	return types.SolSwap{}
//	//}
//	//tf2, ok := FindTransferFromAccounts(instructionData.Transfers, destTokenAccount)
//	//if !ok {
//	//	log.Printf("Could not find transfer for RAYDIUM swap")
//	//	return types.SolSwap{}
//	//}
//
//	s := types.SolSwap{
//		Pair:     instructionData.AccountKeys[(*instructionData.Accounts)[1]],
//		Exchange: "RAYDIUM",
//		Wallet:   wallet,
//	}
//	if rayLogs[0].Direction == 2 {
//		s.TokenOut = sourceTokenAccountMint //tf1.Decimals
//		s.TokenIn = destTokenAccountMint    //tf2.Decimals
//
//		s.AmountOut = strconv.FormatUint(rayLogs[0].AmountIn, 10)
//		s.AmountIn = strconv.FormatUint(rayLogs[0].OutAmount, 10)
//	} else {
//		s.TokenIn = sourceTokenAccountMint //tf1.Decimals
//		s.TokenOut = destTokenAccountMint  //tf2.Decimals
//
//		s.AmountIn = strconv.FormatUint(rayLogs[0].AmountIn, 10)
//		s.AmountOut = strconv.FormatUint(rayLogs[0].OutAmount, 10)
//	}
//
//	return s
//
//}

func HandleRaydiumSwaps(instructionData *types.ProcessInstructionData) types.SolSwap {

	if len(*instructionData.Accounts) < 2 || len(instructionData.AccountKeys) < (*instructionData.Accounts)[1] {
		return types.SolSwap{}
	}

	transfer1, ok := FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, (instructionData.InnerInstructionIndex)+1)
	if !ok {
		return types.SolSwap{}
	}

	transfer2, ok := FindTransfer(instructionData.Transfers, *instructionData.InnerIndex, (instructionData.InnerInstructionIndex)+2)
	if !ok {
		return types.SolSwap{}
	}
	instructionData.Transfers = removeTransfer(instructionData.Transfers, *instructionData.InnerIndex)

	//wallet := transfer1.FromUserAccount
	//if wallet == "" {
	//	wallet = transfer2.ToUserAccount
	//}
	wallet := instructionData.AccountKeys[(*instructionData.Accounts)[16]]
	if len(*instructionData.Accounts) == 18 {
		wallet = instructionData.AccountKeys[(*instructionData.Accounts)[17]]
	}

	s := types.SolSwap{
		Pair:      instructionData.AccountKeys[(*instructionData.Accounts)[1]],
		Exchange:  "RAYDIUM",
		Wallet:    wallet,
		TokenOut:  transfer1.Mint,
		TokenIn:   transfer2.Mint,
		AmountIn:  transfer2.Amount,
		AmountOut: transfer1.Amount,
	}

	return s

}
