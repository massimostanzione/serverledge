package mab

import (
	"log"
	"math"

	"github.com/grussorusso/serverledge/internal/lbcommon"
)

const MAX_COST = 5.0
const MAX_UTILITY = 10000.0

// CalculateReward computes the overall reward based on the given reward configuration and
// the differences between new and old statistics. It combines the contributions of load imbalance,
// response time, cost, and utility, each weighted by their respective coefficients (Alpha, Beta, Gamma, Delta)
// from the RewardConfig. The result represents the aggregated reward score based on the specified metrics.
func CalculateReward(rewardConfig RewardConfig, newStats, oldStats Stats) float64 {
	return rewardConfig.Alpha*calculateLoadImbalance(newStats, oldStats) + rewardConfig.Beta*calculateResponseTime(newStats, oldStats) + rewardConfig.Gamma*calculateCost(newStats, oldStats) + rewardConfig.Delta*calculateUtility(newStats, oldStats)
}

// calculateLoadImbalance computes the load imbalance based on the differences
// between the new and old statistics of server loads. It calculates the difference
// for each server, computes the mean and standard deviation of these differences,
// and returns the negative imbalance percentage. If there are no server load values,
// it returns 0.
func calculateLoadImbalance(newStats, oldStats Stats) float64 {
	// Create a new map to store the differences
	diffs := make(map[string]int)

	// Iterate over the keys in the newStats.ServerLoads map
	for key, newValue := range newStats.ServerLoads {
		// Get the old value if it exists; otherwise, use zero
		oldValue := 0
		if oldVal, ok := oldStats.ServerLoads[key]; ok {
			oldValue = oldVal
		}

		// Calculate the difference and store it in the map
		diffs[key] = newValue - oldValue
	}

	var serverLoads []int
	for _, value := range diffs {
		serverLoads = append(serverLoads, value)
	}

	// Handle empty slice case
	if len(serverLoads) == 0 {
		return 0
	}

	//log.Println(lbcommon.MAB, "ServerLoads:", diffs)

	// Calculate mean
	meanLoad := calculateMean(serverLoads)
	if meanLoad == 0 {
		return 0
	}

	// Calculate standard deviation
	stdDev := calculateStandardDeviation(serverLoads, meanLoad)

	// Calculate imbalance percentage
	imbalancePercentage := stdDev / meanLoad

	// Return negative imbalance percentage
	return -imbalancePercentage
}

// Helper function to calculate the mean of a slice
func calculateMean(values []int) float64 {
	var sum float64
	for _, value := range values {
		sum += float64(value)
	}
	return sum / float64(len(values))
}

// Helper function to calculate the standard deviation of a slice
func calculateStandardDeviation(values []int, mean float64) float64 {
	var sumOfSquares float64
	for _, value := range values {
		diff := float64(value) - mean
		sumOfSquares += diff * diff
	}
	variance := sumOfSquares / float64(len(values))
	return math.Sqrt(variance)
}

// calculateResponseTime computes the average response time based on the difference
// between the new and old statistics. It returns the negative average response time
// calculated as the total response time divided by the total number of completions.
// If there are no completions, it returns 0.
func calculateResponseTime(newStats, oldStats Stats) float64 {
	totalResponseTime := newStats.RespTime - oldStats.RespTime
	totalCompletions := newStats.Completions - oldStats.Completions
	if totalCompletions == 0 {
		return 0
	}
	avgRespTime := totalResponseTime / float64(totalCompletions)
	return -avgRespTime
}

// calculateCost computes the cost difference between new and old statistics.
// It returns the negative value of the cost difference divided by a constant MAX_COST.
// The result represents the normalized cost, scaled by MAX_COST.
func calculateCost(newStats, oldStats Stats) float64 {
	currentCost := newStats.Cost - oldStats.Cost
	log.Println(lbcommon.MAB, "CurrentCost", currentCost)
	return -(currentCost / MAX_COST)
}

// calculateUtility computes the utility difference between new and old statistics.
// It returns the negative value of one minus the utility difference divided by a constant MAX_UTILITY.
// The result represents the normalized utility, scaled by MAX_UTILITY.
func calculateUtility(newStats, oldStats Stats) float64 {
	currentUtility := newStats.RawUtility - oldStats.RawUtility
	return -(1 - (currentUtility / MAX_UTILITY))
}
