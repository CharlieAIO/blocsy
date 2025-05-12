package types

import (
	"database/sql"
	"time"
)

type AppError string

func (e AppError) Error() string {
	return string(e)
}

type ProcessInstructionData struct {
	AccountKeys         []string
	InstructionAccounts *[]int

	Accounts *[]int

	Transfers      []SolTransfer
	ProgramId      *string
	InnerProgramId *string

	InnerInstructionIndex int
	InnerIndex            *int
	Data                  *string

	InnerAccounts   *[]int
	Logs            []string
	TokenAccountMap map[string]TokenAccountDetails
}

type TokenAccountDetails struct {
	UserAccount string
	MintAddress string
	Decimals    int
}

//easyjson:json
type Token struct {
	Name             string    `json:"name" db:"name"`
	Symbol           string    `json:"symbol" db:"symbol"`
	Decimals         uint8     `json:"decimals" db:"decimals"`
	Address          string    `json:"address" db:"address"`
	Supply           string    `json:"supply" db:"supply"`
	CreatedBlock     int64     `json:"createdBlock" db:"createdBlock"`
	Network          string    `json:"network" db:"network"`
	CreatedTimestamp time.Time `json:"createdTimestamp" db:"createdTimestamp"`
	Deployer         *string   `json:"deployer,omitempty" db:"deployer"`
	Metadata         *string   `json:"metadata,omitempty" db:"metadata"`
}

//easyjson:json
type QuoteToken struct {
	Identifier string `json:"identifier" db:"identifier"`
	Name       string `json:"name" db:"name"`
	Symbol     string `json:"symbol" db:"symbol"`
	Address    string `json:"address" db:"address"`
	Decimals   uint8  `json:"decimals" db:"decimals"`
}

//easyjson:json
type QuoteTokenSimple struct {
	Identifier string `json:"identifier" db:"identifier"`
	Address    string `json:"address" db:"address"`
}

//easyjson:json
type Pair struct {
	Address          string           `json:"address" db:"address"`
	Network          string           `json:"network" db:"network"`
	Exchange         string           `json:"exchange" db:"exchange"`
	Token            string           `json:"token" db:"token"`
	QuoteToken       QuoteTokenSimple `json:"quoteToken" db:"quoteToken"`
	CreatedBlock     int64            `json:"createdblock" db:"createdblock"`
	CreatedTimestamp time.Time        `json:"createdtimestamp" db:"createdtimestamp"`
}

//easyjson:json
type TrackerResponse struct {
	Data struct {
		Amount   string `json:"amount"`
		Base     string `json:"base"`
		Currency string `json:"currency"`
	} `json:"data"`
}

type BalanceSheet struct {
	Wallet string
	Token  string
	Amount float64
}

type QueryAll struct {
	Source string         `json:"source" db:"source"`
	Wallet sql.NullString `json:"wallet,omitempty" db:"wallet"`
	Token  sql.NullString `json:"token,omitempty" db:"token"`
	Name   sql.NullString `json:"name,omitempty" db:"name"`
	Symbol sql.NullString `json:"symbol,omitempty" db:"symbol"`
	Pair   sql.NullString `json:"pair" db:"pair"`
}
