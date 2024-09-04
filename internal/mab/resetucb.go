package mab

import (
	"log"
	"math"
	"time"

	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// ResetUCB struct
type ResetUCB struct {
	resetInterval     int                          // Number of actions before resetting
	explorationFactor float64                      // Exploration factor of UCB
	policies          []lbcommon.Policy            // List of available policies (actions)
	rewards           map[lbcommon.Policy]float64  // Average rewards for each policy
	plays             map[lbcommon.Policy]int      // Count of times each policy was played
	totalPlays        int                          // Total number of plays across all policies
	resetCounter      int                          // Counter to track when to reset
	rewardConfig      RewardConfig                 // Reward calculation parameters
	influxDBWriter    *influxwriter.InfluxDBWriter // Writer for InfluxDB
}

// NewResetUCB creates a new instance of the ResetUCB strategy
func NewResetUCB(resetInterval int, explorationFactor float64, policies []lbcommon.Policy, rewardConfig RewardConfig, influxDBWriter *influxwriter.InfluxDBWriter) *ResetUCB {
	return &ResetUCB{
		resetInterval:     resetInterval,
		explorationFactor: explorationFactor,
		policies:          policies,
		rewards:           make(map[lbcommon.Policy]float64),
		plays:             make(map[lbcommon.Policy]int),
		totalPlays:        0,
		resetCounter:      0,
		rewardConfig:      rewardConfig,
		influxDBWriter:    influxDBWriter,
	}
}

// Update method updates the average reward and plays count for a given policy
func (rucb *ResetUCB) Update(newStats, oldStats Stats) {
	policy := newStats.LBPolicy
	reward := CalculateReward(rucb.rewardConfig, newStats, oldStats)

	// Update the count of plays for the policy
	if _, exists := rucb.plays[policy]; !exists {
		rucb.plays[policy] = 0
	}
	rucb.plays[policy]++
	rucb.totalPlays++
	rucb.resetCounter++

	// Apply the incremental update formula
	if rucb.plays[policy] == 1 {
		// If this is the first time the policy is played, set the reward directly
		rucb.rewards[policy] = reward
	} else {
		// Update the average reward using the incremental formula
		avgReward := rucb.rewards[policy]
		count := float64(rucb.plays[policy])
		rucb.rewards[policy] = avgReward + (1/count)*(reward-avgReward)
	}

	log.Println(lbcommon.MAB, "plays updated", rucb.plays)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	influxMABStats := populateInfluxMABStats(newStats, oldStats, timestamp, reward)
	rucb.influxDBWriter.WriteJSON(influxMABStats)

	// Check if it's time to reset
	if rucb.resetCounter == rucb.resetInterval {
		rucb.reset()
	}
}

// reset method resets the rewards and plays maps
func (rucb *ResetUCB) reset() {
	rucb.rewards = make(map[lbcommon.Policy]float64)
	rucb.plays = make(map[lbcommon.Policy]int)
	rucb.totalPlays = 0
	//rucb.resetCounter = 0
	rucb.explorationFactor = 0.03
	rucb.rewardConfig.Beta = 1
	rucb.rewardConfig.Gamma = 0
}

// SelectPolicy selects a policy based on the ResetUCB algorithm
func (rucb *ResetUCB) SelectPolicy() lbcommon.Policy {

	log.Println(lbcommon.MAB, "Rewards", rucb.rewards)

	var bestPolicy lbcommon.Policy
	bestUCBValue := -math.MaxFloat64

	for _, policy := range rucb.policies {
		avgReward, exists := rucb.rewards[policy]
		if !exists {
			// If the policy has not been played, give it a high UCB value (explore it)
			return policy
		}

		// Compute the UCB value for this policy
		n := float64(rucb.plays[policy])
		bonus := rucb.explorationFactor * math.Sqrt((2 * math.Log(float64(rucb.totalPlays)) / n))
		ucbValue := avgReward + bonus

		// Select the policy with the highest UCB value
		if ucbValue > bestUCBValue {
			bestPolicy = policy
			bestUCBValue = ucbValue
		}
	}

	return bestPolicy
}
