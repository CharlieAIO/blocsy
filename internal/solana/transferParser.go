package solana

import (
	"blocsy/internal/types"
	"math/big"
	"strconv"
)

func GetTokenBalanceDiffs(tx *types.SolanaTx) map[int]types.SolBalanceDiff {
	balanceDiffMap := make(map[int]types.SolBalanceDiff, len(tx.Meta.PostTokenBalances))

	for i := range tx.Meta.PostTokenBalances {
		swap := types.SolBalanceDiff{
			Mint:     tx.Meta.PostTokenBalances[i].Mint,
			Amount:   tx.Meta.PostTokenBalances[i].UITokenAmount.UiAmountString,
			Decimals: tx.Meta.PostTokenBalances[i].UITokenAmount.Decimals,
			Owner:    tx.Meta.PostTokenBalances[i].Owner,
		}

		balanceDiffMap[tx.Meta.PostTokenBalances[i].AccountIndex] = swap
	}

	for i := range tx.Meta.PreTokenBalances {
		swap, ok := balanceDiffMap[tx.Meta.PreTokenBalances[i].AccountIndex]
		if !ok {
			continue
		}

		amount, ok := new(big.Float).SetString(swap.Amount)
		if !ok {
			continue
		}

		preAmount, ok := new(big.Float).SetString(tx.Meta.PreTokenBalances[i].UITokenAmount.UiAmountString)
		if !ok {
			continue
		}

		amount.Sub(amount, preAmount)
		swap.Amount = amount.String()
		balanceDiffMap[tx.Meta.PreTokenBalances[i].AccountIndex] = swap

	}
	return balanceDiffMap
}

func GetNativeBalanceDiffs(tx *types.SolanaTx) map[int]types.SolBalanceDiff {
	balanceDiffMap := make(map[int]types.SolBalanceDiff, len(tx.Meta.PostBalances))

	for i := range tx.Meta.PostBalances {
		swap := types.SolBalanceDiff{
			Mint:     "",
			Amount:   strconv.FormatInt(tx.Meta.PostBalances[i], 10),
			Decimals: 9,
			Owner:    "",
		}
		balanceDiffMap[i] = swap
	}

	for i := range tx.Meta.PreBalances {
		swap, ok := balanceDiffMap[i]
		if !ok {
			continue
		}

		postAmount := new(big.Float).SetInt64(tx.Meta.PostBalances[i])

		preAmount := new(big.Float).SetInt64(tx.Meta.PreBalances[i])

		difference := new(big.Float).Sub(postAmount, preAmount)
		swap.Amount = difference.Text('f', -1)
		balanceDiffMap[i] = swap
	}
	return balanceDiffMap
}

func GetAllTransfers(tx *types.SolanaTx) []types.SolTransfer {
	accountKeys := getAllAccountKeys(tx)
	AccountKeysMap := make(map[string]int, len(accountKeys))
	for i := range accountKeys {
		AccountKeysMap[accountKeys[i]] = i
	}

	balanceDiffMap := GetTokenBalanceDiffs(tx)
	nativeBalanceDiffMap := GetNativeBalanceDiffs(tx)

	transfers := make([]types.SolTransfer, 0)
	for i := range tx.Meta.InnerInstructions {
		for ixIndex := range tx.Meta.InnerInstructions[i].Instructions {
			transfer, found := buildTransfer(tx.Meta.InnerInstructions[i].Instructions[ixIndex], AccountKeysMap, balanceDiffMap, nativeBalanceDiffMap, tx, tx.Meta.InnerInstructions[i].Index, ixIndex)
			if found {
				transfers = append(transfers, transfer)
			}
		}
	}

	for i := range tx.Transaction.Message.Instructions {
		transfer, found := buildTransfer(tx.Transaction.Message.Instructions[i], AccountKeysMap, balanceDiffMap, nativeBalanceDiffMap, tx, -1, i)
		if found {
			transfers = append(transfers, transfer)
		}
	}

	return transfers
}

func buildTransfer(
	ix types.Instruction,
	AccountKeysMap map[string]int,
	balanceDiffMap map[int]types.SolBalanceDiff,
	nativeBalanceDiffMap map[int]types.SolBalanceDiff,
	tx *types.SolanaTx,
	innerIndex int,
	ixIndex int) (types.SolTransfer, bool) {
	if ix.Data == "" {
		return types.SolTransfer{}, false
	}
	accountKeys := getAllAccountKeys(tx)

	if len(accountKeys)-1 < ix.ProgramIdIndex {
		return types.SolTransfer{}, false
	}
	programId := accountKeys[ix.ProgramIdIndex]

	if programId == "11111111111111111111111111111111" {
		if len(ix.Accounts) != 2 || len(ix.Data) > 30 {
			return types.SolTransfer{}, false
		}

		for _, accountIndex := range ix.Accounts {
			if accountIndex > len(accountKeys)-1 {
				return types.SolTransfer{}, false
			}
		}

		source := accountKeys[ix.Accounts[0]]
		destination := accountKeys[ix.Accounts[1]]
		if destination == "" || source == "" {
			return types.SolTransfer{}, false
		}
		amount := ""
		if balanceDiff, ok := FindAccountKeyIndex(AccountKeysMap, destination); ok {
			if balanceDiffMap[balanceDiff].Amount != "" {
				amount = balanceDiffMap[balanceDiff].Amount
			} else if nativeBalanceDiffMap[balanceDiff].Amount != "0" {
				amount = nativeBalanceDiffMap[balanceDiff].Amount
			}
		} else {
			return types.SolTransfer{}, false
		}

		transfer := types.SolTransfer{
			IxIndex:         ixIndex,
			InnerIndex:      innerIndex,
			ToUserAccount:   destination,
			FromUserAccount: source,
			Amount:          amount,
			Mint:            "So11111111111111111111111111111111111111112", //So11111111111111111111111111111111111111112
			Type:            "native",
		}
		if transfer.Amount == "" {

			return types.SolTransfer{}, false
		}
		return transfer, true
	}

	//Spl-token program
	if programId == TOKEN_PROGRAM {
		if len(ix.Accounts) < 2 {
			return types.SolTransfer{}, false
		}

		for _, accountIndex := range ix.Accounts {
			if accountIndex > len(accountKeys)-1 {
				return types.SolTransfer{}, false
			}
		}

		amount := ""
		source := ""
		destination := ""

		//transferChecked
		if len(ix.Accounts) == 4 {
			source = accountKeys[ix.Accounts[0]]
			//mint = accountKeys[ix.Accounts[1]]
			destination = accountKeys[ix.Accounts[2]]
			//authority := accountKeys[ix.Accounts[3]]
		}
		//transfer
		if len(ix.Accounts) == 3 {
			source = accountKeys[ix.Accounts[0]]
			destination = accountKeys[ix.Accounts[1]]
			//authority := accountKeys[ix.Accounts[2]]
		}
		tType := "token"

		if destination != "" {
			toUserAccount := ""
			mint := ""
			decimals := -1

			if balanceDiff, ok := FindAccountKeyIndex(AccountKeysMap, destination); ok {
				if balanceDiffMap[balanceDiff].Owner != "" {
					toUserAccount = balanceDiffMap[balanceDiff].Owner
					mint = balanceDiffMap[balanceDiff].Mint
					decimals = balanceDiffMap[balanceDiff].Decimals
					amount = balanceDiffMap[balanceDiff].Amount
				}
			}
			if toUserAccount == "" {
				tokenAccount, found := findUserAccount(destination, tx)
				if found {
					toUserAccount = tokenAccount.UserAccount
					mint = tokenAccount.MintAddress
					decimals = tokenAccount.Decimals
				}
			}

			fromUserAccount := ""
			if balanceDiffSource, ok := FindAccountKeyIndex(AccountKeysMap, source); ok {
				if balanceDiffMap[balanceDiffSource].Owner != "" {
					fromUserAccount = balanceDiffMap[balanceDiffSource].Owner
					if amount == "" || amount == "0" {
						amountFloat, ok := new(big.Float).SetString(balanceDiffMap[balanceDiffSource].Amount)
						if !ok {
							amountFloat = new(big.Float).SetInt64(0)
						}
						amountFloat.Abs(amountFloat)
						amount = amountFloat.Text('f', -1)
					}
				}
			}
			if fromUserAccount == "" {
				if tokenAccount, found := findUserAccount(source, tx); found {
					fromUserAccount = tokenAccount.UserAccount
					mint = tokenAccount.MintAddress
					decimals = tokenAccount.Decimals
				} else {
					tType = "mint"
				}
			}

			transfer := types.SolTransfer{
				IxIndex:          ixIndex,
				InnerIndex:       innerIndex,
				ToUserAccount:    toUserAccount,
				ToTokenAccount:   destination,
				FromTokenAccount: source,
				FromUserAccount:  fromUserAccount,
				Amount:           amount,
				Mint:             mint,
				Decimals:         decimals,
				Type:             tType,
				ProgramId:        programId,
			}

			if transfer.Amount == "" {
				return types.SolTransfer{}, false
			}

			return transfer, true
		}
	}

	return types.SolTransfer{}, false
}

func CreateTokenAccountMap(tx *types.SolanaTx) map[string]types.TokenAccountDetails {
	accountMap := make(map[string]types.TokenAccountDetails)

	accountKeys := getAllAccountKeys(tx)

	for _, account := range accountKeys {
		if tokenAccount, found := findUserAccount(account, tx); found {
			accountMap[account] = tokenAccount
		}
	}

	return accountMap
}

func findUserAccount(tokenAccount string, tx *types.SolanaTx) (types.TokenAccountDetails, bool) {
	accountKeys := getAllAccountKeys(tx)

	currentInfo := types.TokenAccountDetails{
		UserAccount: "",
		MintAddress: "",
		Decimals:    -1,
	}
	foundInfo := false

	for i := range tx.Transaction.Message.Instructions {
		userAccount, mint, found := findAccount(tx.Transaction.Message.Instructions[i], tokenAccount, accountKeys)
		if found {
			foundInfo = true
			if currentInfo.UserAccount == "" {
				currentInfo.UserAccount = userAccount
			}
			if currentInfo.MintAddress == "" {
				currentInfo.MintAddress = mint
			}
		}

		for innerI := range tx.Meta.InnerInstructions {
			if tx.Meta.InnerInstructions[innerI].Index != i {
				continue
			}

			for innerIxIndex := range tx.Meta.InnerInstructions[innerI].Instructions {
				userAccount, mint, found := findAccount(tx.Meta.InnerInstructions[innerI].Instructions[innerIxIndex], tokenAccount, accountKeys)

				if found {
					foundInfo = true
					if currentInfo.UserAccount == "" {
						currentInfo.UserAccount = userAccount
					}
					if currentInfo.MintAddress == "" {
						currentInfo.MintAddress = mint
					}
				}
			}
		}

	}

	if currentInfo.Decimals == -1 {
		decimals := findMintInBalances(tx, currentInfo.MintAddress)
		if decimals != -1 {
			currentInfo.Decimals = decimals
		}
	}

	return currentInfo, foundInfo
}

func findAccount(ix types.Instruction, tokenAccount string, accountKeys []string) (string, string, bool) {
	if len(accountKeys)-1 < ix.ProgramIdIndex {
		return "", "", false
	}

	for _, accountIndex := range ix.Accounts {
		if accountIndex > len(accountKeys)-1 {
			return "", "", false
		}
	}

	programId := accountKeys[ix.ProgramIdIndex]

	// Associated Token Account Program | createAssociatedTokenAccount
	if programId == ASSOCIATED_TOKEN_PROGRAM {
		if len(ix.Accounts) == 6 {

			//tokenProgram := accountKeys[ix.Accounts[5]]
			//systemProgram := accountKeys[ix.Accounts[4]]
			mint := accountKeys[ix.Accounts[3]]
			//source := accountKeys[ix.Accounts[2]]
			account := accountKeys[ix.Accounts[1]]
			wallet := accountKeys[ix.Accounts[0]] // same as source
			if account == tokenAccount {
				return wallet, mint, true
			}

		}

	}
	//spl-token program | initializeAccount, initializeAccount2, initializeAccount3
	if programId == TOKEN_PROGRAM {

		if len(ix.Accounts) == 4 {
			account := accountKeys[ix.Accounts[0]]
			mint := accountKeys[ix.Accounts[1]]
			owner := accountKeys[ix.Accounts[2]]
			//rentSysvar := accountKeys[ix.Accounts[3]]
			if account == tokenAccount {
				return owner, mint, true
			}
		}
		//if len(ix.Accounts) == 3 {
		//	//source := accountKeys[ix.Accounts[0]]
		//	destination := accountKeys[ix.Accounts[1]]
		//	//authority := accountKeys[ix.Accounts[2]]
		//	//log.Printf("authority: %s | destination: %s | source: %s", authority, destination, source)
		//	if destination == tokenAccount {
		//		return "", "", true
		//	}
		//}
		if len(ix.Accounts) == 2 {
			account := accountKeys[ix.Accounts[0]]
			mint := accountKeys[ix.Accounts[1]]
			if account == tokenAccount {
				return "", mint, true
			}
		}
	}

	if programId == SYSTEM_PROGRAM {
		if len(ix.Accounts) == 2 {
			source := accountKeys[ix.Accounts[0]]
			newAccount := accountKeys[ix.Accounts[1]]
			if newAccount == tokenAccount {
				return source, "", true
			}
		}
	}
	return "", "", false
}

func findMintInBalances(tx *types.SolanaTx, mint string) int {

	for _, tokenBalance := range tx.Meta.PostTokenBalances {
		if tokenBalance.Mint == mint {
			return tokenBalance.UITokenAmount.Decimals
		}
	}
	return -1
}
