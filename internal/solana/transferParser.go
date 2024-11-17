package solana

import (
	"defi-intel/internal/types"
	"github.com/blocto/solana-go-sdk/common"
	"math/big"
	"strconv"
)

func findAccountKeyIndex(keyMap map[string]int, key string) (int, bool) {
	if i, ok := keyMap[key]; ok {
		return i, true
	}

	return -1, false
}

func GetAllTransfers(tx *types.SolanaTx) []types.SolTransfer {
	accountKeys := getAllAccountKeys(tx)
	AccountKeysMap := make(map[string]int, len(accountKeys))
	for i := range accountKeys {
		AccountKeysMap[accountKeys[i]] = i
	}

	balanceDiffMap := make(map[int]types.SolBalanceDiff, len(tx.Meta.PostTokenBalances))
	nativeBalanceDiffMap := make(map[int]types.SolBalanceDiff, len(tx.Meta.PostBalances))

	// Calculate token balance differences
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

	for i := range tx.Meta.PostBalances {
		swap := types.SolBalanceDiff{
			Mint:     "",
			Amount:   strconv.FormatInt(tx.Meta.PostBalances[i], 10),
			Decimals: 9,
			Owner:    "",
		}
		nativeBalanceDiffMap[i] = swap
	}

	for i := range tx.Meta.PreBalances {
		swap, ok := nativeBalanceDiffMap[i]
		if !ok {
			continue
		}

		postAmount := new(big.Float).SetInt64(tx.Meta.PostBalances[i])

		preAmount := new(big.Float).SetInt64(tx.Meta.PreBalances[i])

		difference := new(big.Float).Sub(postAmount, preAmount)
		swap.Amount = difference.Text('f', -1)
		nativeBalanceDiffMap[i] = swap
	}

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

func buildTransfer(ix types.Instruction, AccountKeysMap map[string]int, balanceDiffMap map[int]types.SolBalanceDiff, nativeBalanceDiffMap map[int]types.SolBalanceDiff, tx *types.SolanaTx, innerIndex int, ixIndex int) (types.SolTransfer, bool) {
	if ix.Data == "" {
		return types.SolTransfer{}, false
	}
	accountKeys := getAllAccountKeys(tx)

	if len(accountKeys)-1 < ix.ProgramIdIndex {
		return types.SolTransfer{}, false
	}
	programId := accountKeys[ix.ProgramIdIndex]

	//System program
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
		//amount := amountFromData
		amount := ""
		if balanceDiff, ok := findAccountKeyIndex(AccountKeysMap, destination); ok {
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
			Mint:            "", //So11111111111111111111111111111111111111112
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

			if balanceDiff, ok := findAccountKeyIndex(AccountKeysMap, destination); ok {
				if balanceDiffMap[balanceDiff].Owner != "" {
					toUserAccount = balanceDiffMap[balanceDiff].Owner
					mint = balanceDiffMap[balanceDiff].Mint
					decimals = balanceDiffMap[balanceDiff].Decimals
					amount = balanceDiffMap[balanceDiff].Amount
				}
			}
			if toUserAccount == "" {
				associatedAccount, mintAddress, found := findUserAccount(destination, tx)
				if found {
					toUserAccount = associatedAccount
					mint = mintAddress
				}
			}

			fromUserAccount := ""
			if balanceDiffSource, ok := findAccountKeyIndex(AccountKeysMap, source); ok {
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
				if uAccount, mintAddress, found := findUserAccount(source, tx); found {
					fromUserAccount = uAccount
					mint = mintAddress
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
			}

			if transfer.Amount == "" {
				return types.SolTransfer{}, false
			}

			return transfer, true
		}
	}

	return types.SolTransfer{}, false
}

func findUserAccount(tokenAccount string, tx *types.SolanaTx) (string, string, bool) {
	accountKeys := getAllAccountKeys(tx)
	for i := range tx.Meta.InnerInstructions {
		for j := range tx.Meta.InnerInstructions[i].Instructions {
			acc, mint, ok := findAccount(tx.Meta.InnerInstructions[i].Instructions[j], tokenAccount, accountKeys)
			if ok {
				return acc, mint, true
			}
		}
	}

	for i := range tx.Transaction.Message.Instructions {
		acc, mint, found := findAccount(tx.Transaction.Message.Instructions[i], tokenAccount, accountKeys)
		if found {
			return acc, mint, true
		}
	}

	return "", "", false
}

func findAccount(ix types.Instruction, tokenAccount string, accountKeys []string) (string, string, bool) {
	//parsedInfo := ix.Parsed.Info
	if len(accountKeys)-1 < ix.ProgramIdIndex {
		return "", "", false
	}

	programId := accountKeys[ix.ProgramIdIndex]

	// Associated Token Account Program | createAssociatedTokenAccount
	if programId == ASSOCIATED_TOKEN_PROGRAM {
		if len(ix.Accounts) == 6 {

			for _, accountIndex := range ix.Accounts {
				if accountIndex > len(accountKeys)-1 {
					return "", "", false
				}
			}
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

			for _, accountIndex := range ix.Accounts {
				if accountIndex > len(accountKeys)-1 {
					return "", "", false
				}
			}

			account := accountKeys[ix.Accounts[0]]
			mint := accountKeys[ix.Accounts[1]]
			owner := accountKeys[ix.Accounts[2]]
			//rentSysvar := accountKeys[ix.Accounts[3]]
			if account == tokenAccount {
				return owner, mint, true
			}
		}
	}
	return "", "", false
}

func getAssociatedTokenAddress(walletPubKey, mintPubKey common.PublicKey) (common.PublicKey, error) {
	associatedTokenProgramID := common.PublicKeyFromString(ASSOCIATED_TOKEN_PROGRAM)
	tokenProgramID := common.PublicKeyFromString(TOKEN_PROGRAM)

	seeds := [][]byte{
		walletPubKey.Bytes(),
		tokenProgramID.Bytes(),
		mintPubKey.Bytes(),
	}

	associatedTokenAddress, _, err := common.FindProgramAddress(seeds, associatedTokenProgramID)
	if err != nil {
		return common.PublicKey{}, err
	}

	return associatedTokenAddress, nil
}
