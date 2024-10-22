package types

//easyjson:json
type Token struct {
	Name             string `json:"name"`
	Symbol           string `json:"symbol"`
	Decimals         uint8  `json:"decimals"`
	Address          string `json:"address"`
	Supply           string `json:"supply"`
	CreatedBlock     int64  `json:"createdBlock"`
	Network          string `json:"network"`
	CreatedTimestamp uint64 `json:"createdTimestamp"`
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
