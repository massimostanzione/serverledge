package mab

import (
	"github.com/grussorusso/serverledge/internal/influxwriter"
	"github.com/grussorusso/serverledge/internal/lbcommon"
)

// Stats holds various metrics related to the performance of a load balancing policy.
// This structure captures both per-server statistics and overall metrics for the system.
// The data includes the number of requests, completions, and drops, as well as performance
// metrics such as response time, cost, and utility.
type Stats struct {
	LBPolicy    lbcommon.Policy // The load balancing policy currently in use
	ServerLoads map[string]int  // The total number of requests received by each server
	DroppedReqs map[string]int  // The total number of requests dropped by each server
	Arrivals    int             // The total number of incoming requests
	Completions int             // The total number of completed requests
	Violations  int             // The total number of response time violations
	Drops       int             // The total number of dropped requests
	RespTime    float64         // The cumulative response time
	Cost        float64         // The cumulative cost
	RawUtility  float64         // The cumulative utility
}

type Strategy string

const (
	EpsilonGreedyStrategy Strategy = "Epsilon-Greedy"
	UCBStrategy           Strategy = "UCB"
	ResetUCBStrategy      Strategy = "ResetUCB"
	SWUCBStrategy         Strategy = "SWUCB"
	UCB2Strategy          Strategy = "UCB2"
	UCBTunedStrategy      Strategy = "UCBTuned"
	KLUCBStrategy         Strategy = "KL-UCB"
)

// SelectionStrategy defines the methods that any strategy should implement
type SelectionStrategy interface {
	Update(newStats, oldStats Stats)
	SelectPolicy() lbcommon.Policy
}

// MABAgent represents an agent that utilizes a Multi-Armed Bandit (MAB) strategy for load balancing or decision-making.
// The agent selects actions or policies based on the specified strategy and can log performance data to an InfluxDB database.
type MABAgent struct {
	strategy       SelectionStrategy
	influxDBWriter *influxwriter.InfluxDBWriter
}

// RewardConfig holds the parameters for reward calculation
type RewardConfig struct {
	Alpha float64 // Coefficient for load imbalance
	Beta  float64 // Coefficient for response time
	Gamma float64 // Coefficient for cost
	Delta float64 // Coefficient for utility
	Zeta  float64 // Coefficient for violations count
}
