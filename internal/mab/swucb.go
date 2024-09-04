package mab

import (
	"container/list"
	"log"
	"math"
	"time"

	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// SlidingWindowUCB struct
type SlidingWindowUCB struct {
	windowSize        int                         // Size of the sliding window
	explorationFactor float64                     // Exploration factor of UCB
	policies          []lbcommon.Policy           // List of available policies (actions)
	rewards           map[lbcommon.Policy]float64 // Average rewards for each policy within the window
	plays             map[lbcommon.Policy]int     // Count of times each policy was played within the window
	totalPlays        int                         // Total number of plays across all policies within the window
	history           *list.List                  // List to maintain policy history and their rewards
	rewardConfig      RewardConfig                // Reward calculation parameters
	resetCounter      int
	influxDBWriter    *influxwriter.InfluxDBWriter // Writer for InfluxDB
}

// Entry struct to store policy-reward pairs in the history
type Entry struct {
	Policy lbcommon.Policy
	Reward float64
}

// NewSlidingWindowUCB creates a new instance of the SlidingWindowUCB strategy
func NewSlidingWindowUCB(windowSize int, explorationFactor float64, policies []lbcommon.Policy, rewardConfig RewardConfig, influxDBWriter *influxwriter.InfluxDBWriter) *SlidingWindowUCB {
	return &SlidingWindowUCB{
		windowSize:        windowSize,
		explorationFactor: explorationFactor,
		policies:          policies,
		rewards:           make(map[lbcommon.Policy]float64),
		plays:             make(map[lbcommon.Policy]int),
		totalPlays:        0,
		history:           list.New(),
		rewardConfig:      rewardConfig,
		resetCounter:      0,
		influxDBWriter:    influxDBWriter,
	}
}

// Update method updates the average reward and plays count for a given policy
func (swucb *SlidingWindowUCB) Update(newStats, oldStats Stats) {
	policy := newStats.LBPolicy
	reward := CalculateReward(swucb.rewardConfig, newStats, oldStats)

	swucb.resetCounter++

	// If the window is full, remove the oldest entry and update counts
	if swucb.history.Len() == swucb.windowSize {
		// Remove the oldest entry if the window is full
		oldest := swucb.history.Remove(swucb.history.Front()).(Entry)
		swucb.decrementCountsAndRewards(oldest.Policy, oldest.Reward)
	}

	// Add the new entry to the window and update counts
	swucb.history.PushBack(Entry{Policy: policy, Reward: reward})
	swucb.incrementCountsAndRewards(policy, reward)

	log.Println(lbcommon.MAB, "plays updated", swucb.plays)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	influxMABStats := populateInfluxMABStats(newStats, oldStats, timestamp, reward)
	swucb.influxDBWriter.WriteJSON(influxMABStats)

	// Check if it's time to reset
	if swucb.resetCounter == 15 {
		swucb.reset()
	}

}

// reset method resets the rewards and plays maps
func (swucb *SlidingWindowUCB) reset() {
	swucb.explorationFactor = 0.03
	swucb.rewardConfig.Beta = 1
	swucb.rewardConfig.Gamma = 0
}

// Increments counts and rewards
func (swucb *SlidingWindowUCB) incrementCountsAndRewards(policy lbcommon.Policy, reward float64) {
	if _, exists := swucb.plays[policy]; !exists {
		swucb.plays[policy] = 0
		swucb.rewards[policy] = 0
	}
	swucb.plays[policy]++
	swucb.totalPlays++

	// Incremental update of the average reward
	avgReward := swucb.rewards[policy]
	count := float64(swucb.plays[policy])
	swucb.rewards[policy] = avgReward + (1/count)*(reward-avgReward)
}

// Decrements counts and rewards
func (swucb *SlidingWindowUCB) decrementCountsAndRewards(policy lbcommon.Policy, reward float64) {
	if swucb.plays[policy] > 0 {
		swucb.plays[policy]--
		swucb.totalPlays--

		if swucb.plays[policy] == 0 {
			swucb.rewards[policy] = 0
		} else {
			// Adjust the average reward after removing an old reward
			avgReward := swucb.rewards[policy]
			count := float64(swucb.plays[policy])
			swucb.rewards[policy] = (avgReward*count - reward) / count
		}
	}
}

// SelectPolicy selects a policy based on the SlidingWindowUCB algorithm
func (swucb *SlidingWindowUCB) SelectPolicy() lbcommon.Policy {

	log.Println(lbcommon.MAB, "Rewards", swucb.rewards)

	var bestPolicy lbcommon.Policy
	bestUCBValue := -math.MaxFloat64

	for _, policy := range swucb.policies {
		avgReward, exists := swucb.rewards[policy]
		if !exists {
			// If the policy has not been played, give it a high UCB value (epxlore it)
			return policy
		}

		// Compute the UCB value for this policy
		n := float64(swucb.plays[policy])
		bonus := swucb.explorationFactor * math.Sqrt((2 * math.Log(float64(swucb.totalPlays)) / n))
		ucbValue := avgReward + bonus

		// Select the policy with the highest UCB value
		if ucbValue > bestUCBValue {
			bestPolicy = policy
			bestUCBValue = ucbValue
		}
	}

	return bestPolicy
}
