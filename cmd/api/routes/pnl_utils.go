package routes

import (
	"blocsy/internal/types"
	"context"
	"log"
	"math/big"
	"sort"
	"time"
)

func CalculateTokenPnL(
	ctx context.Context,
	swapLogs []types.SwapLog,
	usdPrice float64,
	findLatestSwapFn func(context.Context, string) ([]types.SwapLog, error),
) (types.TokenPnL, *big.Float, *big.Float, *big.Float, *big.Float, *big.Float, time.Duration, *big.Float) {
	totalBuyTokens := new(big.Float)
	totalSellTokens := new(big.Float)
	totalBuyValue := new(big.Float)
	totalSellValue := new(big.Float)

	var buyQueue []types.TokenLot
	var totalHeldTime time.Duration
	var totalSoldAmount = big.NewFloat(0)
	var totalValueRemaining = big.NewFloat(0)

	// Sort swap logs to ensure oldest first
	sort.Slice(swapLogs, func(i, j int) bool {
		return swapLogs[i].Timestamp.Before(swapLogs[j].Timestamp)
	})

	// Process all swaps
	for _, swap := range swapLogs {
		amountOutFloat := new(big.Float).SetFloat64(swap.AmountOut)
		amountInFloat := new(big.Float).SetFloat64(swap.AmountIn)

		if swap.Action == "BUY" || swap.Action == "RECEIVE" {
			totalBuyTokens.Add(totalBuyTokens, amountInFloat)
			if swap.Action == "BUY" {
				totalBuyValue.Add(totalBuyValue, new(big.Float).Mul(amountOutFloat, big.NewFloat(usdPrice)))
			}
			buyQueue = append(buyQueue, types.TokenLot{
				Amount:    new(big.Float).Set(amountInFloat),
				Timestamp: swap.Timestamp,
			})
		} else if swap.Action == "SELL" || swap.Action == "TRANSFER" {
			totalSellTokens.Add(totalSellTokens, amountOutFloat)
			if swap.Action == "SELL" {
				totalSellValue.Add(totalSellValue, new(big.Float).Mul(amountInFloat, big.NewFloat(usdPrice)))
			}

			toSell := new(big.Float).Set(amountOutFloat)
			for len(buyQueue) > 0 && toSell.Cmp(big.NewFloat(0)) > 0 {
				currentLot := &buyQueue[0]
				if currentLot.Amount.Cmp(toSell) <= 0 {
					heldDuration := swap.Timestamp.Sub(currentLot.Timestamp)
					amountFloat, _ := currentLot.Amount.Float64()
					totalHeldTime += time.Duration(float64(heldDuration.Nanoseconds()) * amountFloat)

					totalSoldAmount.Add(totalSoldAmount, currentLot.Amount)
					toSell.Sub(toSell, currentLot.Amount)
					buyQueue = buyQueue[1:]
				} else {
					heldDuration := swap.Timestamp.Sub(currentLot.Timestamp)
					toSellFloat, _ := toSell.Float64()
					totalHeldTime += time.Duration(float64(heldDuration.Nanoseconds()) * toSellFloat)
					totalSoldAmount.Add(totalSoldAmount, toSell)
					currentLot.Amount.Sub(currentLot.Amount, toSell)
					toSell = big.NewFloat(0)
				}
			}
		}
	}

	// Calculate realized PnL
	realizedPNL := new(big.Float)
	if totalSellTokens.Cmp(big.NewFloat(0)) != 0 {
		realizedPNL.Sub(totalSellValue, totalBuyValue)
	} else {
		realizedPNL.SetFloat64(0)
	}

	// Calculate unrealized PnL
	unrealizedPNL := new(big.Float)
	remainingAmount := new(big.Float).Sub(totalBuyTokens, totalSellTokens)

	if remainingAmount.Cmp(big.NewFloat(0)) > 0 {
		mostRecentPrice := new(big.Float)

		// Get the most recent price from the latest swap
		if len(swapLogs) > 0 {
			pair := swapLogs[0].Pair
			mostRecentSwap, err := findLatestSwapFn(ctx, pair)
			if err == nil && len(mostRecentSwap) > 0 {
				amountOutFloat := new(big.Float).SetFloat64(mostRecentSwap[0].AmountOut)
				amountInFloat := new(big.Float).SetFloat64(mostRecentSwap[0].AmountIn)
				if mostRecentSwap[0].Action == "BUY" {
					mostRecentPrice = new(big.Float).Quo(amountOutFloat, amountInFloat)
				} else if mostRecentSwap[0].Action == "SELL" {
					mostRecentPrice = new(big.Float).Quo(amountInFloat, amountOutFloat)
				}
				log.Printf("%s | Calculated most recent price: %s", *swapLogs[0].TokenSymbol, mostRecentPrice.Text('f', 18))
			}
		}

		currentValue := new(big.Float).Mul(remainingAmount, mostRecentPrice)
		totalValueRemaining = new(big.Float).Add(totalValueRemaining, currentValue)

		avgBuyPrice := new(big.Float)
		if totalBuyTokens.Cmp(big.NewFloat(0)) > 0 {
			avgBuyPrice = new(big.Float).Quo(totalBuyValue, totalBuyTokens)
		}
		costBasis := new(big.Float).Mul(remainingAmount, avgBuyPrice)

		unrealizedPNL = new(big.Float).Sub(currentValue, costBasis)
	}

	// Convert to TokenPnL structure
	var pnlResults types.TokenPnL

	realizedPNLFloatUSD, _ := realizedPNL.Float64()
	pnlResults.RealizedPnLUSD = realizedPNLFloatUSD

	unrealizedPNLFloatUSD, _ := unrealizedPNL.Float64()
	pnlResults.UnrealizedPnLUSD = unrealizedPNLFloatUSD

	// Calculate realized ROI
	if totalSellValue.Cmp(big.NewFloat(0)) > 0 {
		realizedROI := new(big.Float).Quo(realizedPNL, totalSellValue)
		realizedROIFloat, _ := realizedROI.Float64()
		pnlResults.RealizedROI = realizedROIFloat * 100
	}

	// Calculate unrealized ROI
	if remainingAmount.Cmp(big.NewFloat(0)) > 0 {
		avgBuyPrice := new(big.Float).Quo(totalBuyValue, totalBuyTokens)
		remainingCost := new(big.Float).Mul(avgBuyPrice, remainingAmount)

		if remainingCost.Cmp(big.NewFloat(0)) > 0 {
			unrealizedROI := new(big.Float).Quo(unrealizedPNL, remainingCost)
			unrealizedROIFloat, _ := unrealizedROI.Float64()
			pnlResults.UnrealizedROI = unrealizedROIFloat * 100
		}
	}

	// Calculate total ROI
	if totalBuyValue.Cmp(big.NewFloat(0)) > 0 {
		totalROI := new(big.Float).Quo(new(big.Float).Add(realizedPNL, unrealizedPNL), totalBuyValue)
		finalROI, _ := totalROI.Float64()
		pnlResults.ROI = finalROI * 100
	}

	// Calculate total PnL
	pnlResults.PnLUSD = pnlResults.RealizedPnLUSD + pnlResults.UnrealizedPnLUSD
	pnlResults.TotalTrades = len(swapLogs)

	boughtTokensFloat, _ := totalBuyTokens.Float64()
	pnlResults.BoughtTokens = boughtTokensFloat

	boughtUSDFloat, _ := totalBuyValue.Float64()
	pnlResults.BoughtUSD = boughtUSDFloat

	soldTokensFloat, _ := totalSellTokens.Float64()
	pnlResults.SoldTokens = soldTokensFloat

	soldUSDFloat, _ := totalSellValue.Float64()
	pnlResults.SoldUSD = soldUSDFloat

	remainingTokensFloat, _ := remainingAmount.Float64()
	pnlResults.RemainingTokens = remainingTokensFloat

	remainingUSDFloat, _ := totalValueRemaining.Float64()
	pnlResults.RemainingUSD = remainingUSDFloat

	return pnlResults, totalBuyValue, totalSellValue, totalBuyTokens, totalSellTokens, totalSoldAmount, totalHeldTime, totalValueRemaining
}
