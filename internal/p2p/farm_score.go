package p2p

import (
	"math"
)

// FarmScoreCalculator calculates farm scores based on performance metrics
type FarmScoreCalculator struct{}

// NewFarmScoreCalculator creates a new farm score calculator
func NewFarmScoreCalculator() *FarmScoreCalculator {
	return &FarmScoreCalculator{}
}

// CalculateFarmScore calculates the farm score based on the Dexponent protocol formula:
// FarmScore = 0.4(Sortino Ratio) + 0.4(Sharpe ratio) + 0.2(Maximum DrawDown) + 2(Returns)
func (f *FarmScoreCalculator) CalculateFarmScore(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	// Calculate metrics
	sharpeRatio := f.calculateSharpeRatio(returns)
	sortinoRatio := f.calculateSortinoRatio(returns)
	maxDrawdown := f.calculateMaximumDrawdown(returns)
	averageReturn := f.calculateAverageReturn(returns)

	// Apply the formula
	farmScore := 0.4*sortinoRatio + 0.4*sharpeRatio + 0.2*maxDrawdown + 2*averageReturn

	// Round to 6 decimal places to ensure consistent results across validators
	return math.Round(farmScore*1000000) / 1000000
}

// calculateSharpeRatio calculates the Sharpe ratio
// Sharpe Ratio = (Average Return - Risk Free Rate) / Standard Deviation
func (f *FarmScoreCalculator) calculateSharpeRatio(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// For simplicity, assume risk-free rate is 0
	riskFreeRate := 0.0

	// Calculate average return
	averageReturn := f.calculateAverageReturn(returns)

	// Calculate standard deviation
	variance := 0.0
	for _, r := range returns {
		variance += math.Pow(r-averageReturn, 2)
	}
	variance /= float64(len(returns) - 1)
	stdDev := math.Sqrt(variance)

	// Avoid division by zero
	if stdDev == 0 {
		return 0
	}

	return (averageReturn - riskFreeRate) / stdDev
}

// calculateSortinoRatio calculates the Sortino ratio
// Sortino Ratio = (Average Return - Risk Free Rate) / Downside Deviation
func (f *FarmScoreCalculator) calculateSortinoRatio(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// For simplicity, assume risk-free rate is 0
	riskFreeRate := 0.0

	// Calculate average return
	averageReturn := f.calculateAverageReturn(returns)

	// Calculate downside deviation (only negative returns)
	downsideSum := 0.0
	downsideCount := 0
	for _, r := range returns {
		if r < 0 {
			downsideSum += math.Pow(r, 2)
			downsideCount++
		}
	}

	// Avoid division by zero
	if downsideCount == 0 {
		return averageReturn * 10 // If no downside, return a high ratio
	}

	downsideDeviation := math.Sqrt(downsideSum / float64(downsideCount))

	// Avoid division by zero
	if downsideDeviation == 0 {
		return 0
	}

	return (averageReturn - riskFreeRate) / downsideDeviation
}

// calculateMaximumDrawdown calculates the maximum drawdown
// Maximum Drawdown = (Peak Value - Trough Value) / Peak Value
func (f *FarmScoreCalculator) calculateMaximumDrawdown(returns []float64) float64 {
	if len(returns) < 2 {
		return 0
	}

	// Convert returns to cumulative returns
	cumulativeReturns := make([]float64, len(returns))
	cumulativeReturns[0] = 1 + returns[0]
	for i := 1; i < len(returns); i++ {
		cumulativeReturns[i] = cumulativeReturns[i-1] * (1 + returns[i])
	}

	// Calculate maximum drawdown
	maxDrawdown := 0.0
	peak := cumulativeReturns[0]

	for _, value := range cumulativeReturns {
		if value > peak {
			peak = value
		} else {
			drawdown := (peak - value) / peak
			if drawdown > maxDrawdown {
				maxDrawdown = drawdown
			}
		}
	}

	return maxDrawdown
}

// calculateAverageReturn calculates the average return
func (f *FarmScoreCalculator) calculateAverageReturn(returns []float64) float64 {
	if len(returns) == 0 {
		return 0
	}

	sum := 0.0
	for _, r := range returns {
		sum += r
	}

	return sum / float64(len(returns))
}


