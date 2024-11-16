package mab

import (
	"log"
	"math"
	"os"
	"time"

	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// UCB2 struct
type UCB2 struct {
	explorationFactor float64                     // Exploration factor of UCB2
	alpha             float64                     // \alpha parameter of UCB2
	policies          []lbcommon.Policy           // List of available policies (actions)
	currentPolicy     lbcommon.Policy             // Current selected policy (to be locked by UCB2 when required)
	rewards           map[lbcommon.Policy]float64 // Average rewards for each policy
	plays             map[lbcommon.Policy]int     // Count of times each policy was played
	epochs            map[lbcommon.Policy]int     // Count of epochs each policy was played (i.e., "R")
	totalPlays        int                         // Total number of plays across all policies
	remLockedPlays    int                         // number of arm selection locked by the most promising arm selected in previous epochs
	rewardConfig      RewardConfig                // Reward calculation parameters
	resetCounter      int
	influxDBWriter    *influxwriter.InfluxDBWriter // Writer for InfluxDB
}

// NewUCB2 creates a new instance of the UCB2 strategy
func NewUCB2(explorationFactor float64, alpha float64, policies []lbcommon.Policy, rewardConfig RewardConfig, influxDBWriter *influxwriter.InfluxDBWriter) *UCB2 {
	return &UCB2{
		explorationFactor: explorationFactor,
		alpha:             alpha,
		policies:          policies,
		currentPolicy:     "",
		rewards:           make(map[lbcommon.Policy]float64),
		plays:             make(map[lbcommon.Policy]int),
		epochs:            make(map[lbcommon.Policy]int),
		totalPlays:        0,
		remLockedPlays:    0,
		rewardConfig:      rewardConfig,
		resetCounter:      0,
		influxDBWriter:    influxDBWriter,
	}
}

// Update method updates the average reward and plays count for a given policy
func (ucb *UCB2) Update(newStats, oldStats Stats) {
	policy := newStats.LBPolicy
	ucb.currentPolicy = policy
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
func (ucb *UCB2) reset() {
	ucb.explorationFactor = 0.03
	ucb.rewardConfig.Beta = 1
	ucb.rewardConfig.Gamma = 0
}

func (ucb *UCB2) _tau(r int) float64 {
	return math.Ceil(math.Pow(1+ucb.alpha, float64(r)))
}

// SelectPolicy selects a policy based on the UCB algorithm
func (ucb *UCB2) SelectPolicy() lbcommon.Policy {

	log.Println(lbcommon.MAB, "Rewards", ucb.rewards)

	var bestPolicy lbcommon.Policy
	bestUCBValue := -math.MaxFloat64

	// if a specific arm was previously selected,
	// continue executing it for the remaining f(\tau(R)) times (see below)
	if ucb.remLockedPlays > 0 {
		ucb.remLockedPlays -= 1
		log.Println(lbcommon.MAB, "Policy selection locked by UCB2 on", ucb.currentPolicy, ",", ucb.remLockedPlays,
			"plays remaining.")
		return ucb.currentPolicy
	}
	for _, policy := range ucb.policies {
		avgReward, exists := ucb.rewards[policy]
		if !exists {
			// If the policy has not been played, give it a high UCB2 value (explore it)
			// This is also the init UCB2 step ("play every arm once")
			return policy
		}

		// Compute the UCB2 value for this policy
		n := float64(ucb.plays[policy])
		tau_r := ucb._tau(ucb.epochs[policy])
		bonus := ucb.explorationFactor * math.Sqrt(((1.0+ucb.alpha)*(math.E*n/tau_r))/(2*tau_r))

		ucbValue := avgReward + bonus

		// Select the policy with the highest UCB2 value
		if ucbValue > bestUCBValue {
			bestPolicy = policy
			bestUCBValue = ucbValue
		}

	}

	// once the arm is selected, "lock" it for f(\tau(R)) subsequent plays:
	tau_r_selected := ucb._tau(ucb.epochs[bestPolicy])
	tau_r1_selected := ucb._tau(ucb.epochs[bestPolicy] + 1)

	if tau_r1_selected-tau_r_selected < 0 {
		log.Println(lbcommon.MAB, "ERROR: negative remLockedPlays =", ucb.remLockedPlays)
		os.Exit(-1)
	}

	// \tau differences could be zero, set a minimum of 1 execution
	ucb.remLockedPlays = int(math.Max(1, tau_r1_selected-tau_r_selected))
	ucb.remLockedPlays -= 1 // because decrementing is done while selecting the policy, i.e. here

	ucb.epochs[bestPolicy] += 1 // increment the epoch counter
	log.Println(lbcommon.MAB, "Starting epoch no.", ucb.epochs[bestPolicy], "for policy", bestPolicy, " - it will last for", ucb.remLockedPlays+1, "subsequent plays.")

	return bestPolicy
}
