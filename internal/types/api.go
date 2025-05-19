package types

type PNLInfo struct {
	Swaps              int    `json:"swaps"`
	TotalBuyVolume     string `json:"totalBuyVolume"`
	TotalBuyVolumeUSD  string `json:"totalBuyVolumeUSD"`
	TotalSellVolume    string `json:"totalSellVolume"`
	TotalSellVolumeUSD string `json:"totalSellVolumeUSD"`

	TotalPnL      string `json:"totalPnL"`
	TotalPnLUSD   string `json:"totalPnLUSD"`
	RoiPercentage string `json:"roiPercentage"`

	UnrealizedPnL           string `json:"unrealizedPnL"`
	UnrealizedPnLUSD        string `json:"unrealizedPnLUSD"`
	UnrealizedRoiPercentage string `json:"unrealizedRoiPercentage"`

	Pair             string `json:"pair"`
	QuoteTokenSymbol string `json:"quoteTokenSymbol"`
	Token            string `json:"token"`
	Network          string `json:"chain"`
	Exchange         string `json:"exchange"`
}

type AggregatedPnL struct {
	PnLUSD float64 `json:"pnlUSD"`
	ROI    float64 `json:"roi"`

	RealizedPnLUSD   float64 `json:"realizedPnLUSD"`
	RealizedROI      float64 `json:"realizedROI"`
	UnrealizedPnLUSD float64 `json:"unrealizedPnLUSD"`
	UnrealizedROI    float64 `json:"unrealizedROI"`
	TokensTraded     int     `json:"tokensTraded"`
	WinRate          float64 `json:"winRate"`
	AverageHoldTime  string  `json:"averageHoldTime"`

	TotalBuy  int64 `json:"totalBuy"`
	TotalSell int64 `json:"totalSell"`

	TotalBuyVolumeUSD       float64 `json:"totalBuyVolumeUSD"`
	TotalSellVolumeUSD      float64 `json:"totalSellVolumeUSD"`
	TotalVolumeUSD          float64 `json:"totalVolumeUSD"`
	TotalActivePositionsUSD float64 `json:"totalActivePositionsUSD"`
}

type TokenPnL struct {
	PnLUSD float64 `json:"pnlUSD"`
	ROI    float64 `json:"roi"`

	RealizedPnLUSD   float64 `json:"realizedPnLUSD"`
	RealizedROI      float64 `json:"realizedROI"`
	UnrealizedPnLUSD float64 `json:"unrealizedPnLUSD"`
	UnrealizedROI    float64 `json:"unrealizedROI"`
	TotalTrades      int     `json:"totalTrades"`
	HoldTime         string  `json:"holdTime"`

	BoughtTokens    float64 `json:"boughtTokens"`
	BoughtUSD       float64 `json:"boughtUSD"`
	SoldTokens      float64 `json:"soldTokens"`
	SoldUSD         float64 `json:"soldUSD"`
	RemainingTokens float64 `json:"remainingTokens"`
	RemainingUSD    float64 `json:"remainingUSD"`
}
