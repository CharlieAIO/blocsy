package types

import "time"

type SwapLog struct {
	ID          string    `json:"id" db:"id"`
	Wallet      string    `json:"wallet" db:"wallet"`
	Source      string    `json:"source" db:"source"`
	BlockNumber uint64    `json:"blockNumber" db:"blockNumber"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	AmountOut   float64   `json:"amountOut" db:"amountOut"`
	AmountIn    float64   `json:"amountIn" db:"amountIn"`
	Action      string    `json:"action" db:"action"`
	Pair        string    `json:"pair" db:"pair"`
	Token       string    `json:"token" db:"token"`
	Processed   bool      `json:"processed" db:"processed"`
}
