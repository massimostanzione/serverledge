package mab

import (
	"log"

	"github.com/grussorusso/serverledge/internal/config"
	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// NewMABAgent initializes the agent with a specific strategy
func NewMABAgent() *MABAgent {

	// Initialize InfluxDB
	url := config.GetString(config.STORAGE_DB_ADDRESS, "http://localhost:8086")
	token := config.GetString(config.STORAGE_DB_TOKEN, "serverledge")
	org := config.GetString(config.STORAGE_DB_ORGNAME, "serverledge")
	bucket := "stats"
	influxDBWriter := influxwriter.NewInfluxDBWriter(url, token, org, bucket)

	// Configure reward
	alpha := config.GetFloat(config.MAB_AGENT_REWARD_ALPHA, 0)
	beta := config.GetFloat(config.MAB_AGENT_REWARD_BETA, 0)
	gamma := config.GetFloat(config.MAB_AGENT_REWARD_GAMMA, 0)
	delta := config.GetFloat(config.MAB_AGENT_REWARD_DELTA, 0)
	rewardConfig := RewardConfig{Alpha: alpha, Beta: beta, Gamma: gamma, Delta: delta}

	// Get available load balancing policies
	policies := lbcommon.GetAllPolicies()
	log.Println(lbcommon.MAB, "Policies found by the MAB agent:", policies)

	// Configure MAB agent's exploration strategy and create it
	mabStrategy := config.GetString(config.MAB_AGENT_STRATEGY, string(EpsilonGreedyStrategy))
	var selectionStrategy SelectionStrategy
	switch Strategy(mabStrategy) {
	case UCBStrategy:
		explorationFactor := config.GetFloat(config.MAB_AGENT_EXPLORATIONFACTOR, 0.05)
		selectionStrategy = NewUCB(explorationFactor, policies, rewardConfig, influxDBWriter)
		log.Println(lbcommon.MAB, "Using UCB Strategy (exploration factor =", explorationFactor, ")")
	case ResetUCBStrategy:
		explorationFactor := config.GetFloat(config.MAB_AGENT_EXPLORATIONFACTOR, 0.05)
		resetInterval := config.GetInt(config.MAB_AGENT_RUCB_RESETINTERVAL, 0)
		selectionStrategy = NewResetUCB(resetInterval, explorationFactor, policies, rewardConfig, influxDBWriter)
		log.Println(lbcommon.MAB, "Using ResetUCB Strategy (exploration factor =", explorationFactor, ", reset interval = ", resetInterval, ")")
	case SWUCBStrategy:
		explorationFactor := config.GetFloat(config.MAB_AGENT_EXPLORATIONFACTOR, 0.05)
		windowSize := config.GetInt(config.MAB_AGENT_SWUCB_WINDOWSIZE, 10)
		selectionStrategy = NewSlidingWindowUCB(windowSize, explorationFactor, policies, rewardConfig, influxDBWriter)
		log.Println(lbcommon.MAB, "Using SWUCB Strategy (exploration factor =", explorationFactor, ", window size =", windowSize, ")")
	default: // EpsilonGreedyStrategy default
		epsilon := config.GetFloat(config.MAB_AGENT_EPSILON, 0.1)
		selectionStrategy = NewEpsilonGreedy(epsilon, policies, rewardConfig, influxDBWriter)
		log.Println(lbcommon.MAB, "Using Epsilon-Greedy Strategy (epsilon =", epsilon, ")")
	}

	return createMABAgent(selectionStrategy, influxDBWriter)
}

// Update method delegates to the strategy's Update method
func (mab *MABAgent) Update(newStats, oldStats Stats) {
	mab.strategy.Update(newStats, oldStats)
}

// SelectPolicy method delegates to the strategy's SelectPolicy method
func (mab *MABAgent) SelectPolicy() lbcommon.Policy {
	return mab.strategy.SelectPolicy()
}

// createMABAgent initializes and returns a Multi-Armed Bandit (MAB) agent
// based on the configuration settings. The agent's exploration strategy
// is selected from available options like Epsilon-Greedy, UCB, ResetUCB,
// or SWUCB, and is configured with the appropriate reward parameters.
func createMABAgent(strategy SelectionStrategy, influxDBWriter *influxwriter.InfluxDBWriter) *MABAgent {
	return &MABAgent{
		strategy:       strategy,
		influxDBWriter: influxDBWriter,
	}
}
