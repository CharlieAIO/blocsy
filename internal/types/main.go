package types

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
	Name             string  `json:"name"`
	Symbol           string  `json:"symbol"`
	Decimals         uint8   `json:"decimals"`
	Address          string  `json:"address"`
	Supply           string  `json:"supply"`
	CreatedBlock     int64   `json:"createdBlock"`
	Network          string  `json:"network"`
	CreatedTimestamp uint64  `json:"createdTimestamp"`
	Deployer         *string `json:"deployer,omitempty"`
	Metadata         *string `json:"metadata,omitempty"`
}

//easyjson:json
type QuoteToken struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	Symbol     string `json:"symbol"`
	Address    string `json:"address"`
	Decimals   uint8  `json:"decimals"`
}

//easyjson:json
type QuoteTokenSimple struct {
	Identifier string `json:"identifier"`
	Address    string `json:"address"`
}

//easyjson:json
type Pair struct {
	Address          string           `json:"pair"`
	Network          string           `json:"network"`
	Exchange         string           `json:"exchange"`
	Token            string           `json:"token"`
	QuoteToken       QuoteTokenSimple `json:"quoteToken"`
	CreatedBlock     int64            `json:"createdBlock"`
	CreatedTimestamp uint64           `json:"createdTimestamp"`
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
