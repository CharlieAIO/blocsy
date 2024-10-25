package trackers

import (
	"context"
	"defi-intel/internal/types"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

var validSymbols = map[string]bool{
	"BTC":  true,
	"ETH":  true,
	"WETH": true,
	"SOL":  true,
	"USDT": true,
	"USDC": true,
}

type PriceTracker struct {
	priceTrackCache map[string]float64
	mx              sync.RWMutex
}

func NewPriceTracker() *PriceTracker {
	return &PriceTracker{
		priceTrackCache: make(map[string]float64),
		mx:              sync.RWMutex{},
	}
}

func (pt *PriceTracker) GetUSDPrice(symbol string) float64 {
	pt.mx.RLock()
	if price, exists := pt.priceTrackCache[symbol]; exists {
		pt.mx.RUnlock()
		return price
	}

	pt.mx.RUnlock()

	amount, err := lookupPrice(symbol)
	if err != nil {
		log.Println("Error looking up price:", err)
		return 0
	}

	pt.mx.Lock()
	pt.priceTrackCache[symbol] = amount
	pt.mx.Unlock()

	return amount
}

func (pt *PriceTracker) Run(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for symbol := range pt.priceTrackCache {
				amount, err := lookupPrice(symbol)
				if err != nil {
					log.Println("Error looking up price:", err)
					continue
				}

				pt.priceTrackCache[symbol] = amount
			}
		}
	}

}

func lookupPrice(symbol string) (float64, error) {
	if !validSymbols[symbol] {
		return 0, fmt.Errorf("Invalid symbol: %s", symbol)
	}

	symbol = strings.ToUpper(symbol)
	resp, err := http.Get("https://api.coinbase.com/v2/prices/" + symbol + "-USD/buy")
	if err != nil {
		return 0, fmt.Errorf("Error making request: %v", err)
	}

	defer resp.Body.Close()

	var response types.TrackerResponse
	err = json.NewDecoder(resp.Body).Decode(&response)
	if err != nil {
		return 0, fmt.Errorf("Error decoding response: %v", err)
	}

	amount, err := strconv.ParseFloat(response.Data.Amount, 64)
	if err != nil {
		return 0, fmt.Errorf("(%s) Error converting amount to float: %v", symbol, err)
	}

	if response.Data.Base != symbol {
		return 0, fmt.Errorf("Symbol mismatch: %s != %s", response.Data.Base, symbol)
	}

	return amount, nil
}
