package mab

import (
	"log"
	"math"
	"time"

	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// KLUCB struct
type KLUCB struct {
	explorationFactor float64                     // Exploration factor of UCB
	c                 float64                     // \c factor of KL-UCB
	policies          []lbcommon.Policy           // List of available policies (actions)
	rewards           map[lbcommon.Policy]float64 // Average rewards for each policy
	plays             map[lbcommon.Policy]int     // Count of times each policy was played
	totalPlays        int                         // Total number of plays across all policies
	rewardConfig      RewardConfig                // Reward calculation parameters
	resetCounter      int
	influxDBWriter    *influxwriter.InfluxDBWriter // Writer for InfluxDB
}

// NewKLUCB creates a new instance of the KLUCB strategy
func NewKLUCB(explorationFactor float64, c float64, policies []lbcommon.Policy, rewardConfig RewardConfig, influxDBWriter *influxwriter.InfluxDBWriter) *KLUCB {
	return &KLUCB{
		explorationFactor: explorationFactor,
		c:                 c,
		policies:          policies,
		rewards:           make(map[lbcommon.Policy]float64),
		plays:             make(map[lbcommon.Policy]int),
		totalPlays:        0,
		rewardConfig:      rewardConfig,
		resetCounter:      0,
		influxDBWriter:    influxDBWriter,
	}
}

// Update method updates the average reward and plays count for a given policy
func (ucb *KLUCB) Update(newStats, oldStats Stats) {
	policy := newStats.LBPolicy
	reward := CalculateReward(ucb.rewardConfig, newStats, oldStats)

	ucb.resetCounter++

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

	// Check if it's time to reset
	//if ucb.resetCounter == 15 {
	//	ucb.reset()
	//}
}

// reset method resets the rewards and plays maps
func (ucb *KLUCB) reset() {
	ucb.explorationFactor = 0.03
	ucb.rewardConfig.Beta = 1
	ucb.rewardConfig.Gamma = 0
}

func (ucb *KLUCB) _kl(p float64, q float64) float64 {
	if p == q {
		return 0.0
	} else if q == 0 || q == 1 {
		return math.MaxFloat64
	}
	return (p * math.Log(p/q)) + ((1 - p) * math.Log((1-p)/(1-q)))
}

func (ucb *KLUCB) _q(policy lbcommon.Policy) float64 {
	t := ucb.totalPlays
	upper_limit := 1.0
	lower_limit := ucb.rewards[policy] + 1
	epsilon := 1e-6 // tolerance
	target := (math.Log(float64(t)) + ucb.c*math.Log(math.Log(float64(t)))) / float64(ucb.plays[policy])

	// find the q value via binary searhc
	for upper_limit-lower_limit > epsilon {
		q := (upper_limit + lower_limit) / 2
		if ucb._kl(ucb.rewards[policy]+1, q) <= target {
			lower_limit = q
		} else {
			upper_limit = q
		}
	}
	return (upper_limit + lower_limit) / 2
}

// SelectPolicy selects a policy based on the UCB algorithm
func (ucb *KLUCB) SelectPolicy() lbcommon.Policy {

	log.Println(lbcommon.MAB, "Rewards", ucb.rewards)

	var bestPolicy lbcommon.Policy
	bestUCBValue := -math.MaxFloat64

	for _, policy := range ucb.policies {
		_, exists := ucb.rewards[policy]
		if !exists {
			// If the policy has not been played, give it a high UCB value (explore it)
			return policy
		}

		// Compute the UCB value for this policy
		ucbValue := ucb._q(policy)

		// Select the policy with the highest UCB value
		if ucbValue > bestUCBValue {
			bestPolicy = policy
			bestUCBValue = ucbValue
		}
	}

	return bestPolicy
}
