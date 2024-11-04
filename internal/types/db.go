package types

import "time"

type SwapLog struct {
	ID          string    `json:"id" db:"id"`
	Wallet      string    `json:"wallet" db:"wallet"`
	Network     string    `json:"network" db:"network"`
	Exchange    string    `json:"exchange" db:"exchange"`
	BlockNumber uint64    `json:"blockNumber" db:"blockNumber"`
	BlockHash   string    `json:"blockHash" db:"blockHash"`
	Timestamp   time.Time `json:"timestamp" db:"timestamp"`
	Type        string    `json:"type" db:"type"`
	AmountOut   float64   `json:"amountOut" db:"amountOut"`
	AmountIn    float64   `json:"amountIn" db:"amountIn"`
	Price       float64   `json:"price" db:"price"`
	Pair        string    `json:"pair" db:"pair"`
	LogIndex    string    `json:"logIndex" db:"logIndex"`
	Processed   bool      `json:"processed" db:"processed"`
}
