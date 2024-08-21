package influxwriter

const INFLUXDB = "[INFLUXDB]:"

// InfluxMABStats holds statistics related to the performance of a Multi-Armed Bandit (MAB) load balancing algorithm.
// Each field in the structure captures different aspects of the load balancing process, with values either representing
// incremental changes since the previous timestamp or cumulative totals up to the current timestamp.
// Fields with "Cum" in their names represent cumulative totals up to the current timestamp.
// All other fields represent values that have changed between the previous timestamp and the current one.
type InfluxMABStats struct {
	Time           string  // The timestamp at which the statistics were recorded
	Policy         string  // The name of the load balancing policy in use
	ServerLoads    []int   // The number of requests received by each server between the previous timestamp and this one
	ServerLoadsCum []int   // The cumulative number of requests received by each server up to this timestamp
	DroppedReqs    []int   // The number of requests dropped by each server between the previous timestamp and this one
	DroppedReqsCum []int   // The cumulative number of requests dropped by each server up to this timestamp
	AvgRespTime    float64 // The average response time between the previous timestamp and this one
	Cost           float64 // The total cost incurred between the previous timestamp and this one
	Utility        float64 // The total utility calculated between the previous timestamp and this one
	Reward         float64 // The reward value calculated based on metrics between the previous timestamp and this one
}
