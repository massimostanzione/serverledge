package lb

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/grussorusso/serverledge/internal/registration"
)

// WRRCostPolicy: è un load balancer che utilizza la politica wrr-cost
type WRRCostPolicy struct {
	mu            sync.Mutex
	lbProxy       *LBProxy
	servers       []Server
	weights       []int
	totalWeight   int
	index         int
	requestCounts []int
	totalReqs     int
}

// NewWRRCostPolicy: crea un nuovo load balancer wrr-cost
func NewWRRCostPolicy(lbProxy *LBProxy) *WRRCostPolicy {

	log.Println(LB, "WRRCostPolicy created")

	// Recupero il costo dei nodi cloud
	costs := make([]float64, len(lbProxy.targets))
	for i, target := range lbProxy.targets {
		costs[i] = getCost(target)
	}

	// Determino il costo massimo tra i server
	maxCost := costs[0]
	for _, value := range costs[1:] {
		if value > maxCost {
			maxCost = value
		}
	}

	// Calcolo i pesi di ciascun server
	servers := make([]Server, len(lbProxy.targets))
	totalWeight := 0
	weights := make([]int, len(lbProxy.targets))
	for i, target := range lbProxy.targets {
		weight := MULT_FACTOR * int(maxCost/costs[i])
		if weight < 1 {
			weight = 1
		}
		weights[i] = weight
		totalWeight += weight
		servers[i] = Server{target: target, weight: weight}
	}

	//log.Println(LB, "WRRCostPolicy end")

	// Ritorno della struttura inizializzata
	return &WRRCostPolicy{
		lbProxy:       lbProxy,
		servers:       servers,
		weights:       weights,
		totalWeight:   totalWeight,
		index:         0,
		requestCounts: make([]int, len(servers)),
		totalReqs:     0,
	}
}

// SelectTarget: seleziona il prossimo target utilizzando la politica wrr-cost
func (r *WRRCostPolicy) SelectTarget(funName string) *url.URL {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(r.servers) == 0 {
		return nil
	}

	for {
		if r.totalReqs == r.totalWeight {
			r.totalReqs = 0
			r.requestCounts = make([]int, len(r.servers))
			r.index = 0
		}
		server := &r.servers[r.index]
		if r.requestCounts[r.index] < server.weight {
			r.requestCounts[r.index]++
			r.totalReqs++
			r.index = (r.index + 1) % len(r.servers)
			return server.target
		}
		r.index = (r.index + 1) % len(r.servers)
	}
}

func getCost(target *url.URL) float64 {
	url := fmt.Sprintf("http://%s/status", target.Host)
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalf("%s Invocation to get status failed: %v", LB, err)
	}

	// Read the response body
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("%s Error reading response body: %v", LB, err)
	}

	// Check the status code
	if resp.StatusCode == http.StatusOK {
		var statusInfo registration.StatusInformation

		// Decode the JSON into a StatusInformation structure
		err = json.Unmarshal(body, &statusInfo)
		if err != nil {
			log.Fatalf("%s Error decoding JSON: %v", LB, err)
		}

		return statusInfo.CostCloud
	}

	return 1
}
