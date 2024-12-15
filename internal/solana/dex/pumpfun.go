package dex

import (
	"blocsy/internal/types"
	"encoding/base64"
	"strings"
)

//const TRADE_EVENT_DISCRIMINATOR = "bddb7fd34ee661ee"

func HandlePumpFunSwaps(index int, transfers []types.SolTransfer, accountKeys []string) (types.SolSwap, int) {
	currentTransfer := transfers[index]
	nextTransfer := transfers[index+1]
	if currentTransfer.ParentProgramId != nextTransfer.ParentProgramId {
		return types.SolSwap{}, 0
	}

	if len(currentTransfer.IxAccounts) < 4 || len(accountKeys) < (currentTransfer.IxAccounts)[3] {
		return types.SolSwap{}, 0
	}

	s := types.SolSwap{
		Pair:      accountKeys[(currentTransfer.IxAccounts)[3]],
		Exchange:  "PUMPFUN",
		Wallet:    accountKeys[(currentTransfer.IxAccounts)[6]],
		TokenOut:  nextTransfer.Mint,
		TokenIn:   currentTransfer.Mint,
		AmountIn:  currentTransfer.Amount,
		AmountOut: nextTransfer.Amount,
	}

	return s, 1

}

func HandlePumpFunNewToken(parsedLogs []types.LogDetails, programId string) []types.PumpFunCreation {

	var pfLogs []string

	var checkLogs func(logs []types.LogDetails)
	checkLogs = func(logs []types.LogDetails) {
		for _, logDetail := range logs {
			for _, log_ := range logDetail.Logs {
				if strings.Contains(log_, "Program data:") && logDetail.Program == programId {
					pfLogs = append(pfLogs, log_)
				}
			}
			checkLogs(logDetail.SubLogs)
		}
	}
	checkLogs(parsedLogs)
	if len(pfLogs) == 0 {
		return nil
	}

	var tokens []types.PumpFunCreation
	for _, pLog := range pfLogs {
		splitStr := strings.Split(pLog, "Program data: ")[1]
		bytesData, err := base64.StdEncoding.DecodeString(splitStr)
		if err != nil {
			continue
		}
		newToken := types.PumpFunCreation{}
		err = newToken.Decode(bytesData)
		if err != nil {
			continue
		}
		if newToken.Mint.String() == "" || newToken.Symbol == "" {
			continue
		}

		tokens = append(tokens, newToken)

	}

	return tokens

}
