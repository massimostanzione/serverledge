package mab

import (
	"sort"

	"github.com/grussorusso/serverledge/internal/influxwriter"
)

// Function to get the sorted keys of a map
func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Function to calculate differences between two maps
func calculateDifferences(newStats, oldStats map[string]int) map[string]int {
	// Create a new map to store the differences
	diffs := make(map[string]int)

	// Iterate over the keys in the newStats.ServerLoads map
	for key, newValue := range newStats {
		// Get the old value if it exists; otherwise, use zero
		oldValue := 0
		if oldVal, ok := oldStats[key]; ok {
			oldValue = oldVal
		}

		// Calculate the difference and store it in the map
		diffs[key] = newValue - oldValue
	}

	return diffs
}

func populateInfluxMABStats(newStats, oldStats Stats, timestamp string, reward float64) influxwriter.InfluxMABStats {
	// Get the sorted keys of the maps
	sortedServerLoadsKeys := sortedKeys(newStats.ServerLoads)
	sortedDroppedReqsKeys := sortedKeys(newStats.DroppedReqs)

	// Calculate differences
	serverLoadsDiff := calculateDifferences(newStats.ServerLoads, oldStats.ServerLoads)
	droppedReqsDiff := calculateDifferences(newStats.DroppedReqs, oldStats.DroppedReqs)

	// Convert the map values to sorted slices
	serverLoads := make([]int, 0, len(sortedServerLoadsKeys))
	serverLoadsCum := make([]int, 0, len(sortedServerLoadsKeys))
	droppedReqs := make([]int, 0, len(sortedDroppedReqsKeys))
	droppedReqsCum := make([]int, 0, len(sortedDroppedReqsKeys))

	// Fill server loads and cumulative loads
	for _, key := range sortedServerLoadsKeys {
		loadDiff := serverLoadsDiff[key]
		serverLoads = append(serverLoads, loadDiff)
		serverLoadsCum = append(serverLoadsCum, newStats.ServerLoads[key])
	}

	// Fill dropped requests and cumulative dropped requests
	for _, key := range sortedDroppedReqsKeys {
		dropDiff := droppedReqsDiff[key]
		droppedReqs = append(droppedReqs, dropDiff)
		droppedReqsCum = append(droppedReqsCum, newStats.DroppedReqs[key])
	}

	// Calculate the average response time
	var avgRespTime float64
	totalResponseTime := newStats.RespTime - oldStats.RespTime
	totalCompletions := newStats.Completions - oldStats.Completions
	if totalCompletions == 0 {
		avgRespTime = 0
	} else {
		avgRespTime = totalResponseTime / float64(totalCompletions)
	}

	return influxwriter.InfluxMABStats{
		Time:           timestamp,
		Policy:         string(newStats.LBPolicy),
		ServerLoads:    serverLoads,
		ServerLoadsCum: serverLoadsCum,
		DroppedReqs:    droppedReqs,
		DroppedReqsCum: droppedReqsCum,
		Arrivals:       newStats.Arrivals - oldStats.Arrivals,
		Completions:    newStats.Completions - oldStats.Completions,
		AvgRespTime:    avgRespTime,
		Cost:           newStats.Cost - oldStats.Cost,
		Utility:        newStats.RawUtility - oldStats.RawUtility,
		Reward:         reward,
	}
}
