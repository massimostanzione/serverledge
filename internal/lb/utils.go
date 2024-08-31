package lb

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/grussorusso/serverledge/internal/function"
	"github.com/grussorusso/serverledge/internal/lbcommon"
	"github.com/grussorusso/serverledge/internal/mab"
	"github.com/grussorusso/serverledge/internal/registration"
)

// compareURLTargets checks if two slices of URL pointers contain the same URLs, regardless of order.
// It returns true if both slices have the same URLs with the same frequency; otherwise, it returns false.
// The function first compares the lengths of the slices, then uses a map to count URL occurrences and
// ensures that both slices have identical sets of URLs.
func compareURLTargets(oldTargets, newTargets []*url.URL) bool {
	// Se la lunghezza delle slice è diversa, sono sicuramente diverse
	if len(oldTargets) != len(newTargets) {
		return false
	}

	// Crea una mappa per contare le occorrenze delle URL in oldTargets
	urlMap := make(map[string]int)
	for _, u := range oldTargets {
		if u != nil {
			urlMap[u.String()]++
		}
	}

	// Sottrai le occorrenze per gli elementi in newTargets
	for _, u := range newTargets {
		if u != nil {
			urlMap[u.String()]--
			if urlMap[u.String()] == 0 {
				delete(urlMap, u.String())
			}
		}
	}

	// Se la mappa è vuota, allora le due slice contengono gli stessi elementi
	return len(urlMap) == 0
}

// getTargets retrieves a list of backend targets from a cloud registry based on the specified region.
// It parses each cloud node's JSON representation to extract the node address, converts it into a URL,
// and collects these URLs into a slice. The function returns the slice of URL pointers and any error encountered
// during retrieval or parsing.
func getTargets(region string) ([]*url.URL, error) {
	cloudNodes, err := registration.GetCloudNodes(region)
	if err != nil {
		return nil, err
	}

	targets := make([]*url.URL, 0, len(cloudNodes))
	i := 0
	for _, node := range cloudNodes {
		//log.Printf("Found target: %v", node)

		// Isola la parte JSON della stringa
		start := strings.Index(node, "{")
		end := strings.LastIndex(node, "}")
		jsonString := node[start : end+1]

		// Definisce una struttura per il JSON
		var result map[string]string
		err := json.Unmarshal([]byte(jsonString), &result)
		if err != nil {
			log.Println(LB, "Error parsing JSON:", err)
			return nil, err
		}

		// Estrai il valore di nodeAddress
		i++
		// log.Println("Target", i, ":", result["nodeAddress"])
		url, err := url.Parse(result["nodeAddress"])
		if err != nil {
			return nil, err
		}
		targets = append(targets, url)
	}

	// log.Printf("[LB]: found %d targets", len(targets))

	return targets, nil
}

// Helper function to retrieve node status information via HTTP
func getTargetStatus(target *url.URL) *registration.StatusInformation {
	resp, err := http.Get(target.String() + "/status")
	if err != nil {
		log.Fatalf("%s Invocation to get status failed: %v", LB, err)
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%s Error reading response body: %v", LB, err)
	}

	// Check the status code
	if resp.StatusCode == http.StatusOK {
		var statusInfo registration.StatusInformation
		if err := json.Unmarshal(body, &statusInfo); err != nil {
			log.Fatalf("%s Error decoding JSON: %v", LB, err)
		}
		return &statusInfo
	}

	return nil
}

// getLBPolicy selects and returns a load balancing policy based on the provided
// policy type. It updates the lbProxy's policy name accordingly.
// func getLBPolicy(p lbcommon.Policy, targets []*url.URL) LBPolicy {
func getLBPolicy(p lbcommon.Policy, lbProxy *LBProxy) LBPolicy {
	switch p {
	case lbcommon.Random:
		lbProxy.lbPolicyName = lbcommon.Random
		return NewRandomPolicy(lbProxy)
	case lbcommon.RoundRobin:
		lbProxy.lbPolicyName = lbcommon.RoundRobin
		return NewRoundRobinPolicy(lbProxy)
	case lbcommon.MAMA:
		lbProxy.lbPolicyName = lbcommon.MAMA
		return NewMAMAPolicy(lbProxy)
	case lbcommon.WRRMemory:
		lbProxy.lbPolicyName = lbcommon.WRRMemory
		return NewWRRMemoryPolicy(lbProxy)
	case lbcommon.WRRCost:
		lbProxy.lbPolicyName = lbcommon.WRRCost
		return NewWRRCostPolicy(lbProxy)
	case lbcommon.CONST_HASH:
		lbProxy.lbPolicyName = lbcommon.CONST_HASH
		return NewConstHashPolicy(lbProxy)
	/*
		case lbcommon.WRRSpeedup:
			lbProxy.lbPolicyName = lbcommon.WRRSpeedup
			return NewWRRSpeedupPolicy(targets)
	*/
	default:
		lbProxy.lbPolicyName = lbcommon.Random
		return NewRandomPolicy(lbProxy)
	}
}

// newStats creates and returns a new instance of mab.Stats initialized with the provided load balancing policy.
// It sets up maps for server loads and dropped requests, and initializes other metrics (response time, completions, cost, and utility) to zero.
func newStats(lbPolicy lbcommon.Policy, targets []*url.URL) *mab.Stats {
	serverLoads := make(map[string]int)
	droppedReqs := make(map[string]int)
	for _, target := range targets {
		if target != nil {
			serverLoads[target.String()] = 0
			droppedReqs[target.String()] = 0
		}
	}
	return &mab.Stats{
		LBPolicy:    lbPolicy,
		ServerLoads: serverLoads,
		DroppedReqs: droppedReqs,
		Arrivals:    0,
		Completions: 0,
		Drops:       0,
		RespTime:    0,
		Cost:        0,
		RawUtility:  0,
	}
}

// updateStats updates the load balancer's statistics based on the execution report of a request.
func updateStats(lbP *LBProxy, executionReport function.ExecutionReport, backend string, dropped bool) {
	lbP.newStats.Arrivals += 1
	lbP.newStats.ServerLoads[backend] += 1
	if dropped { // Request dropped
		lbP.newStats.Drops += 1
		lbP.newStats.DroppedReqs[backend] += 1
	} else { // Request completed
		lbP.newStats.Completions += 1
		lbP.newStats.RespTime += executionReport.ResponseTime
		lbP.newStats.Cost += executionReport.CostCloud
		lbP.newStats.RawUtility += executionReport.Utility
	}
}

// copyStats copies the contents of newStats into oldStats.
// This includes both the shallow fields and a deep copy of the map fields.
// After execution, oldStats will contain a full copy of the data from newStats.
func copyStats(newStats *mab.Stats, oldStats *mab.Stats) {
	// Copy simple fields
	oldStats.LBPolicy = newStats.LBPolicy
	oldStats.Arrivals = newStats.Arrivals
	oldStats.Completions = newStats.Completions
	oldStats.Drops = newStats.Drops
	oldStats.RespTime = newStats.RespTime
	oldStats.Cost = newStats.Cost
	oldStats.RawUtility = newStats.RawUtility

	// Deep copy of the ServerLoads map
	if oldStats.ServerLoads == nil {
		oldStats.ServerLoads = make(map[string]int, len(newStats.ServerLoads))
	} else {
		for key := range oldStats.ServerLoads {
			delete(oldStats.ServerLoads, key)
		}
	}
	for key, value := range newStats.ServerLoads {
		oldStats.ServerLoads[key] = value
	}

	// Deep copy of the DroppedReqs map
	if oldStats.DroppedReqs == nil {
		oldStats.DroppedReqs = make(map[string]int, len(newStats.DroppedReqs))
	} else {
		for key := range oldStats.DroppedReqs {
			delete(oldStats.DroppedReqs, key)
		}
	}
	for key, value := range newStats.DroppedReqs {
		oldStats.DroppedReqs[key] = value
	}
}
