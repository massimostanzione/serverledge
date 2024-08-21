package mab

import (
	"log"
	"math"
	"time"

	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// UCB struct
type UCB struct {
	explorationFactor float64                      // Exploration factor of UCB
	policies          []lbcommon.Policy            // List of available policies (actions)
	rewards           map[lbcommon.Policy]float64  // Average rewards for each policy
	plays             map[lbcommon.Policy]int      // Count of times each policy was played
	totalPlays        int                          // Total number of plays across all policies
	rewardConfig      RewardConfig                 // Reward calculation parameters
	influxDBWriter    *influxwriter.InfluxDBWriter // Writer for InfluxDB
}

// NewUCB creates a new instance of the UCB strategy
func NewUCB(explorationFactor float64, policies []lbcommon.Policy, rewardConfig RewardConfig, influxDBWriter *influxwriter.InfluxDBWriter) *UCB {
	return &UCB{
		explorationFactor: explorationFactor,
		policies:          policies,
		rewards:           make(map[lbcommon.Policy]float64),
		plays:             make(map[lbcommon.Policy]int),
		totalPlays:        0,
		rewardConfig:      rewardConfig,
		influxDBWriter:    influxDBWriter,
	}
}

// Update method updates the average reward and plays count for a given policy
func (ucb *UCB) Update(newStats, oldStats Stats) {
	policy := newStats.LBPolicy
	reward := CalculateReward(ucb.rewardConfig, newStats, oldStats)

	// Update the count of plays for the policy
	if _, exists := ucb.plays[policy]; !exists {
		ucb.plays[policy] = 0
	}
	ucb.plays[policy]++
	ucb.totalPlays++

	// Apply the incremental update formula
	if ucb.plays[policy] == 1 {
		// If this is the first time the policy is played, set the reward directly
		ucb.rewards[policy] = reward
	} else {
		// Update the average reward using the incremental formula
		avgReward := ucb.rewards[policy]
		count := float64(ucb.plays[policy])
		ucb.rewards[policy] = avgReward + (1/count)*(reward-avgReward)
	}

	log.Println(lbcommon.MAB, "plays updated", ucb.plays)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	influxMABStats := populateInfluxMABStats(newStats, oldStats, timestamp, reward)
	ucb.influxDBWriter.WriteJSON(influxMABStats)
}

// SelectPolicy selects a policy based on the UCB algorithm
func (ucb *UCB) SelectPolicy() lbcommon.Policy {

	log.Println(lbcommon.MAB, "Rewards", ucb.rewards)

	var bestPolicy lbcommon.Policy
	bestUCBValue := -math.MaxFloat64

	for _, policy := range ucb.policies {
		avgReward, exists := ucb.rewards[policy]
		if !exists {
			// If the policy has not been played, give it a high UCB value (explore it)
			return policy
		}

		// Compute the UCB value for this policy
		n := float64(ucb.plays[policy])
		bonus := ucb.explorationFactor * math.Sqrt((2 * math.Log(float64(ucb.totalPlays)) / n))
		ucbValue := avgReward + bonus

		// Select the policy with the highest UCB value
		if ucbValue > bestUCBValue {
			bestPolicy = policy
			bestUCBValue = ucbValue
		}
	}

	return bestPolicy
}
