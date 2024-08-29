package mab

import (
	"log"
	"math/rand"
	"time"

	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// EpsilonGreedy struct
type EpsilonGreedy struct {
	epsilon        float64                      // Probability of exploring
	policies       []lbcommon.Policy            // List of available policies (actions)
	rewards        map[lbcommon.Policy]float64  // Average rewards for each policy
	plays          map[lbcommon.Policy]int      // Count of times each action was played
	rng            *rand.Rand                   // Random number generator
	rewardConfig   RewardConfig                 // Reward calculation parameters
	influxDBWriter *influxwriter.InfluxDBWriter // Writer for InfluxDB
}

// NewEpsilonGreedy creates a new instance of the EpsilonGreedy strategy
func NewEpsilonGreedy(epsilon float64, policies []lbcommon.Policy, rewardConfig RewardConfig, influxDBWriter *influxwriter.InfluxDBWriter) *EpsilonGreedy {
	return &EpsilonGreedy{
		epsilon:        epsilon,
		policies:       policies,
		rewards:        make(map[lbcommon.Policy]float64),
		plays:          make(map[lbcommon.Policy]int),
		rng:            rand.New(rand.NewSource(time.Now().UnixNano())),
		rewardConfig:   rewardConfig,
		influxDBWriter: influxDBWriter,
	}
}

// Update method updates the average reward and plays count for a given policy
func (eg *EpsilonGreedy) Update(newStats, oldStats Stats) {
	policy := newStats.LBPolicy
	reward := CalculateReward(eg.rewardConfig, newStats, oldStats)

	//log.Println(lbcommon.MAB, "New stats", newStats)
	//log.Println(lbcommon.MAB, "Old stats", oldStats)

	// Update the plays count
	if _, exists := eg.plays[policy]; !exists {
		eg.plays[policy] = 0
	}
	// Increment the count of plays for the policy
	eg.plays[policy]++

	// Apply the incremental update formula
	if eg.plays[policy] == 1 {
		// If this is the first time the policy is played, set the reward directly
		eg.rewards[policy] = reward
	} else {
		// Update the average reward using the incremental formula
		avgReward := eg.rewards[policy]
		count := float64(eg.plays[policy])
		eg.rewards[policy] = avgReward + (1/count)*(reward-avgReward)
	}

	log.Println(lbcommon.MAB, "Plays updated", eg.plays)

	timestamp := time.Now().UTC().Format(time.RFC3339)
	influxMABStats := populateInfluxMABStats(newStats, oldStats, timestamp, reward)
	eg.influxDBWriter.WriteJSON(influxMABStats)
}

// SelectPolicy selects a policy based on the espilon-greedy algorithm
func (eg *EpsilonGreedy) SelectPolicy() lbcommon.Policy {

	log.Println(lbcommon.MAB, "Rewards", eg.rewards)

	// With probability epsilon, explore (choose a random action)
	if eg.rng.Float64() < eg.epsilon {
		return eg.policies[eg.rng.Intn(len(eg.policies))]
	}

	// Otherwise, exploit (choose the action with the highest average reward)
	var bestPolicy lbcommon.Policy
	var bestAvgReward float64
	for _, policy := range eg.policies {
		avgReward, exists := eg.rewards[policy]
		if !exists || bestPolicy == "" || avgReward > bestAvgReward {
			bestPolicy = policy
			bestAvgReward = avgReward
		}
	}

	return bestPolicy
}
