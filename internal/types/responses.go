package types

type WalletActivityResponse struct {
	Results []SwapLog `json:"results"`
}

type TopTradersResponse struct {
	Results []string `json:"results"`
}

type TopRecentTokensResponse struct {
	Results []TopRecentToken `json:"results"`
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
	Token       string   `json:"token"`
	TokenSymbol string   `json:"tokenSymbol"`
	PnL         TokenPnL `json:"pnl"`
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
	Results   AggregatedPnL `json:"results"`
	TotalBuy  int64         `json:"totalBuy"`
	TotalSell int64         `json:"totalSell"`
}

// QueryAllResponse is a wrapper for QueryAll that handles sql.NullString fields properly
type QueryAllResponse struct {
	Source string `json:"source"`
	Wallet string `json:"wallet,omitempty"`
	Token  string `json:"token,omitempty"`
	Name   string `json:"name,omitempty"`
	Symbol string `json:"symbol,omitempty"`
	Pair   string `json:"pair"`
}

// ConvertToResponse converts a QueryAll to a QueryAllResponse
func ConvertToResponse(q QueryAll) QueryAllResponse {
	resp := QueryAllResponse{
		Source: q.Source,
	}

	if q.Wallet.Valid {
		resp.Wallet = q.Wallet.String
	}

	if q.Token.Valid {
		resp.Token = q.Token.String
	}

	if q.Name.Valid {
		resp.Name = q.Name.String
	}

	if q.Symbol.Valid {
		resp.Symbol = q.Symbol.String
	}

	if q.Pair.Valid {
		resp.Pair = q.Pair.String
	}

	return resp
}

type SearchQueryResponse struct {
	Results []QueryAllResponse `json:"results"`
}

type HoldingsLookupResponse struct {
	Results float64 `json:"results"`
}

type PriceLookupResponse struct {
	Price    float64 `json:"price"`
	Symbol   string  `json:"symbol"`
	Currency string  `json:"currency"`
}
