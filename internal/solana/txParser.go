package solana

import (
	"blocsy/internal/types"
	"math"
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
			Amount:   strconv.FormatUint(tx.Meta.PostBalances[i], 10),
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

		postAmount := new(big.Float).SetInt64(int64(tx.Meta.PostBalances[i]))

		preAmount := new(big.Float).SetInt64(int64(tx.Meta.PreBalances[i]))

		difference := new(big.Float).Sub(postAmount, preAmount)
		swap.Amount = difference.Text('f', -1)
		balanceDiffMap[i] = swap
	}
	return balanceDiffMap
}

func ParseTransaction(tx *types.SolanaTx) ([]types.SolTransfer, []types.SolTransfer, []types.SolTransfer, []types.Token) {
	accountKeys := getAllAccountKeys(tx)
	AccountKeysMap := make(map[string]int, len(accountKeys))
	for i := range accountKeys {
		AccountKeysMap[accountKeys[i]] = i
	}

	balanceDiffMap := GetTokenBalanceDiffs(tx)
	nativeBalanceDiffMap := GetNativeBalanceDiffs(tx)

	transfers := make([]types.SolTransfer, 0)
	burns := make([]types.SolTransfer, 0)
	tokenMints := make([]types.SolTransfer, 0)
	tokensCreated := make([]types.Token, 0)

	for instructionIndex := range tx.Transaction.Message.Instructions {
		instruction := tx.Transaction.Message.Instructions[instructionIndex]
		processedOuter, found := processInstruction(instruction, AccountKeysMap, balanceDiffMap, nativeBalanceDiffMap, tx, -1, instructionIndex)
		if found {
			parentProgramId, parentAccounts := findParentProgram(instructionIndex, tx, -1, -1, accountKeys)
			processedOuter.IxAccounts = parentAccounts
			processedOuter.ParentProgramId = parentProgramId

			if processedOuter.Amount != "" && processedOuter.Amount != "0" {
				if processedOuter.Type == "burn" {
					burns = append(burns, processedOuter)
				} else if processedOuter.Type == "mint" {
					tokenMints = append(tokenMints, processedOuter)
				}
				transfers = append(transfers, processedOuter)

			}
		} else {
			if processedOuter.Type == "initMint" {
				token := types.Token{
					Address:  processedOuter.Mint,
					Decimals: uint8(processedOuter.Decimals),
					Network:  "solana",
					Supply:   "0",
				}

				name, symbol, uri, foundMetadata := findMetaplexInstruction(tx, processedOuter.Mint)
				if foundMetadata {
					token.Metadata = &uri
					token.Name = name
					token.Symbol = symbol
				}

				tokensCreated = append(tokensCreated, token)
			}
		}

		// Now check if there are anny inner instructions for this instruction
		for innerIxIndex := range tx.Meta.InnerInstructions {
			if tx.Meta.InnerInstructions[innerIxIndex].Index != instructionIndex {
				continue
			}
			for ixIndex := range tx.Meta.InnerInstructions[innerIxIndex].Instructions {
				innerInstruction := tx.Meta.InnerInstructions[innerIxIndex].Instructions[ixIndex]
				processedInner, foundInner := processInstruction(innerInstruction, AccountKeysMap, balanceDiffMap, nativeBalanceDiffMap, tx, instructionIndex, ixIndex)

				if foundInner {
					parentProgramId, parentAccounts := findParentProgram(instructionIndex, tx, innerIxIndex, ixIndex, accountKeys)
					processedInner.IxAccounts = parentAccounts
					processedInner.ParentProgramId = parentProgramId
					processedInner.EventData = findPumpFunSwapEvent(instructionIndex, tx, innerIxIndex, ixIndex, accountKeys)

					if processedInner.Amount != "" && processedInner.Amount != "0" {
						if processedInner.Type == "burn" {
							burns = append(burns, processedInner)
						} else if processedInner.Type == "mint" {
							tokenMints = append(tokenMints, processedInner)
						}
						transfers = append(transfers, processedInner)
					}
				} else {
					if processedInner.Type == "initMint" {
						token := types.Token{
							Address:  processedInner.Mint,
							Decimals: uint8(processedInner.Decimals),
							Network:  "solana",
							Supply:   "0",
						}
						name, symbol, uri, foundMetadata := findMetaplexInstruction(tx, processedInner.Mint)
						if foundMetadata {
							token.Metadata = &uri
							token.Name = name
							token.Symbol = symbol
						}

						tokensCreated = append(tokensCreated, token)
					}
				}
			}
		}
	}

	return transfers, burns, tokenMints, tokensCreated
}

func findParentProgram(ixIndex int, tx *types.SolanaTx, innerIxIndex int, innerInstructionIxIndex int, accountKeys []string) (string, []int) {
	if innerIxIndex >= 0 {
		// If inner instruction, traverse backwards within inner instructions
		for innerI := innerInstructionIxIndex; innerI >= 0; innerI-- {
			ix := tx.Meta.InnerInstructions[innerIxIndex].Instructions[innerI]

			//if validateParentProgram(accountKeys[ix.ProgramIdIndex]) && len(ix.Accounts) <= 2 {
			//	continue
			//}

			if validateParentProgram(accountKeys[ix.ProgramIdIndex]) {
				var accs []int
				if validateDexInstruction(accountKeys[ix.ProgramIdIndex], ix.Accounts, accountKeys) {
					accs = ix.Accounts
				}
				return accountKeys[ix.ProgramIdIndex], accs
			}
		}
	}

	baseIx := tx.Transaction.Message.Instructions[ixIndex]
	if validateParentProgram(accountKeys[baseIx.ProgramIdIndex]) {
		var accs []int
		if validateDexInstruction(accountKeys[baseIx.ProgramIdIndex], baseIx.Accounts, accountKeys) {
			accs = baseIx.Accounts
		}
		return accountKeys[baseIx.ProgramIdIndex], accs
	}

	// Fallback: Default to empty if no parent program is found
	return "", nil
}

func processInstruction(
	ix types.Instruction,
	AccountKeysMap map[string]int,
	balanceDiffMap map[int]types.SolBalanceDiff,
	nativeBalanceDiffMap map[int]types.SolBalanceDiff,
	tx *types.SolanaTx,
	innerIndex int,
	ixIndex int) (types.SolTransfer, bool) {
	accountKeys := getAllAccountKeys(tx)

	if len(accountKeys)-1 < ix.ProgramIdIndex {
		return types.SolTransfer{}, false
	}
	programId := accountKeys[ix.ProgramIdIndex]

	if programId == SYSTEM_PROGRAM {

		var instructionData = DecodeSystemProgramData(ix.Data)

		if instructionData.Type != "Transfer" {
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
		//amount := ""
		//if balanceDiff, ok := FindAccountKeyIndex(AccountKeysMap, destination); ok {
		//	if balanceDiffMap[balanceDiff].Amount != "" {
		//		amount = balanceDiffMap[balanceDiff].Amount
		//	} else if nativeBalanceDiffMap[balanceDiff].Amount != "0" {
		//		amount = nativeBalanceDiffMap[balanceDiff].Amount
		//	}
		//} else {
		//	return types.SolTransfer{}, false
		//}
		amount := new(big.Float).Quo(new(big.Float).SetUint64(instructionData.Lamports), big.NewFloat(1e9)).Text('f', -1)

		transfer := types.SolTransfer{
			IxIndex:         ixIndex,
			InnerIndex:      innerIndex,
			ToUserAccount:   destination,
			FromUserAccount: source,
			Amount:          amount,
			Mint:            "So11111111111111111111111111111111111111112", // this is WSOL address
			Type:            "native",
			Authority:       "",
		}
		if transfer.Amount == "" {

			return types.SolTransfer{}, false
		}
		return transfer, true
	}

	//Spl-token program
	if programId == TOKEN_PROGRAM {
		for _, accountIndex := range ix.Accounts {
			if accountIndex > len(accountKeys)-1 {
				return types.SolTransfer{}, false
			}
		}

		var amount, source, destination, authority, mint, toUserAccount string
		var decimals = -1

		var instructionData = DecodeTokenProgramData(ix.Data)
		tType := "token"

		if instructionData.Type == "InitializeMint" || instructionData.Type == "InitializeMint2" {
			mint = accountKeys[ix.Accounts[0]]
			decimals = instructionData.Decimals
			return types.SolTransfer{
				Type:     "initMint",
				Mint:     mint,
				Decimals: decimals,
			}, false
		}

		if instructionData.Type == "TransferChecked" {
			source = accountKeys[ix.Accounts[0]]
			mint = accountKeys[ix.Accounts[1]]
			destination = accountKeys[ix.Accounts[2]]
			authority = accountKeys[ix.Accounts[3]]
		} else if instructionData.Type == "Transfer" {
			decimals = instructionData.Decimals
			source = accountKeys[ix.Accounts[0]]
			destination = accountKeys[ix.Accounts[1]]
			authority = accountKeys[ix.Accounts[2]]
		} else if instructionData.Type == "MintTo" {
			tType = "mint"
			mint = accountKeys[ix.Accounts[0]]
			destination = accountKeys[ix.Accounts[1]] //account
			authority = accountKeys[ix.Accounts[2]]
		} else if instructionData.Type == "Burn" {
			tType = "burn"
			source = accountKeys[ix.Accounts[0]]
			mint = accountKeys[ix.Accounts[1]]
			authority = accountKeys[ix.Accounts[2]]
		} else {
			return types.SolTransfer{}, false
		}

		fromUserAccount := ""
		if balanceDiffSource, ok := FindAccountKeyIndex(AccountKeysMap, source); ok {
			if balanceDiffMap[balanceDiffSource].Owner != "" {
				fromUserAccount = balanceDiffMap[balanceDiffSource].Owner
			}
		}
		if fromUserAccount == "" {
			if tokenAccount, found := findUserAccount(source, tx); found {
				fromUserAccount = tokenAccount.UserAccount
				if mint == "" {
					mint = tokenAccount.MintAddress
				}
				decimals = tokenAccount.Decimals
			}
		}

		if balanceDiff, ok := FindAccountKeyIndex(AccountKeysMap, destination); ok {
			if balanceDiffMap[balanceDiff].Owner != "" {
				toUserAccount = balanceDiffMap[balanceDiff].Owner
				if mint == "" {
					mint = balanceDiffMap[balanceDiff].Mint
				}
				decimals = balanceDiffMap[balanceDiff].Decimals
			}
		}
		if toUserAccount == "" {
			tokenAccount, found := findUserAccount(destination, tx)
			if found {
				toUserAccount = tokenAccount.UserAccount
				if mint == "" {
					mint = tokenAccount.MintAddress
				}
				decimals = tokenAccount.Decimals
			}
		}
		amount = new(big.Float).Quo(new(big.Float).SetUint64(instructionData.Amount), new(big.Float).SetFloat64(math.Pow10(decimals))).Text('f', -1)
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
			Authority:        authority,
		}

		return transfer, true
	}

	return types.SolTransfer{}, false
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
				userAccount, mint, found = findAccount(tx.Meta.InnerInstructions[innerI].Instructions[innerIxIndex], tokenAccount, accountKeys)

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
	var userAccount, mint, foundTokenAccount string

	if len(accountKeys)-1 < ix.ProgramIdIndex {
		return userAccount, mint, false
	}

	for _, accountIndex := range ix.Accounts {
		if accountIndex > len(accountKeys)-1 {
			return userAccount, mint, false
		}
	}

	programId := accountKeys[ix.ProgramIdIndex]

	if programId == ASSOCIATED_TOKEN_PROGRAM {
		if len(ix.Accounts) == 6 {

			//tokenProgram := accountKeys[ix.Accounts[5]]
			//systemProgram := accountKeys[ix.Accounts[4]]
			mint = accountKeys[ix.Accounts[3]]
			//source := accountKeys[ix.Accounts[2]]
			foundTokenAccount = accountKeys[ix.Accounts[1]]
			userAccount = accountKeys[ix.Accounts[0]] // same as source

		}

	}
	if programId == TOKEN_PROGRAM {
		var instructionData = DecodeTokenProgramData(ix.Data)
		if instructionData.Type == "InitializeAccount3" {
			foundTokenAccount = accountKeys[ix.Accounts[0]]
			mint = accountKeys[ix.Accounts[1]]
			userAccount = instructionData.Owner.String()
		} else if instructionData.Type == "InitializeAccount" {
			foundTokenAccount = accountKeys[ix.Accounts[0]]
			mint = accountKeys[ix.Accounts[1]]
			userAccount = accountKeys[ix.Accounts[2]]
		} else if instructionData.Type == "InitializeAccount2" {
			foundTokenAccount = accountKeys[ix.Accounts[0]]
			mint = accountKeys[ix.Accounts[1]]
			userAccount = instructionData.Owner.String()
		} else if instructionData.Type == "CloseAccount" {
			foundTokenAccount = accountKeys[ix.Accounts[0]]
			userAccount = accountKeys[ix.Accounts[2]]
		}

	}
	if programId == SYSTEM_PROGRAM {
		var instructionData = DecodeSystemProgramData(ix.Data)

		if instructionData.Type == "CreateAccount" {
			source := accountKeys[ix.Accounts[0]]
			newAccount := accountKeys[ix.Accounts[1]]
			if newAccount == tokenAccount {
				return source, "", true
			}
		}
	}

	if foundTokenAccount == tokenAccount {
		return userAccount, mint, true
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

func findPumpFunSwapEvent(ixIndex int, tx *types.SolanaTx, innerIxIndex int, innerInstructionIxIndex int, accountKeys []string) string {
	if innerIxIndex >= 0 {
		// If inner instruction, traverse backwards within inner instructions
		for innerI := innerInstructionIxIndex; innerI < len(tx.Meta.InnerInstructions[innerIxIndex].Instructions); innerI++ {
			ix := tx.Meta.InnerInstructions[innerIxIndex].Instructions[innerI]
			if accountKeys[ix.ProgramIdIndex] == PUMPFUN {
				return ix.Data
			}
		}
	}
	return ""
}

func findMetaplexInstruction(tx *types.SolanaTx, mint string) (string, string, string, bool) {
	accountKeys := getAllAccountKeys(tx)

	for _, instruction := range tx.Transaction.Message.Instructions {
		if accountKeys[instruction.ProgramIdIndex] == METAPLEX_TOKEN_METDATA {
			if len(instruction.Accounts) > 2 {
				if accountKeys[instruction.Accounts[1]] == mint {
					metadataAccount, err := DecodeMetaplexData(instruction.Data)
					if err != nil {
						continue
					}
					return metadataAccount.Data.Name, metadataAccount.Data.Symbol, metadataAccount.Data.Uri, true
				}
			}
		}
	}

	for _, innerInstruction := range tx.Meta.InnerInstructions {
		for _, instruction := range innerInstruction.Instructions {
			if accountKeys[instruction.ProgramIdIndex] == METAPLEX_TOKEN_METDATA {
				if len(instruction.Accounts) > 2 {
					if accountKeys[instruction.Accounts[1]] == mint {
						metadataAccount, err := DecodeMetaplexData(instruction.Data)
						if err != nil {
							continue
						}
						return metadataAccount.Data.Name, metadataAccount.Data.Symbol, metadataAccount.Data.Uri, true

					}
				}
			}
		}
	}

	return "", "", "", false
}
