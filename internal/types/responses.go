package types

type WalletActivityResponse struct {
	Results []SwapLogModified `json:"results"`
}

type TopTradersResponse struct {
	Results []string `json:"results"`
}

type TokenLookupResponse struct {
	Token Token  `json:"token"`
	Pairs []Pair `json:"pairs"`
}

type PairLookupResponse struct {
	Pair       Pair       `json:"pair"`
	QuoteToken QuoteToken `json:"quoteToken"`
}

type TokenAndPnl struct {
	Token string   `json:"token"`
	PnL   TokenPnL `json:"pnl"`
}

type TokenPNLResponse struct {
	Tokens     []TokenAndPnl `json:"tokens"`
	Pagination Pagination    `json:"pagination"`
}

type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"pageSize"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

type AggregatedPnLResponse struct {
	Results AggregatedPnL `json:"results"`
}

type SearchQueryResponse struct {
	Results []QueryAll `json:"results"`
}

type HoldingsLookupResponse struct {
	Results float64 `json:"results"`
}
