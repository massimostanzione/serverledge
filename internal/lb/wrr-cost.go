package lb

import (
	"log"
	"net/url"
	"sync"
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
	costs := make([]float64, len(lbProxy.targetsInfo.targets))
	for i, target := range lbProxy.targetsInfo.targets {
		costs[i] = getCost(lbProxy, target.String())
	}

	// Determino il costo massimo tra i server
	maxCost := costs[0]
	for _, value := range costs[1:] {
		if value > maxCost {
			maxCost = value
		}
	}

	// Calcolo i pesi di ciascun server
	servers := make([]Server, len(lbProxy.targetsInfo.targets))
	totalWeight := 0
	weights := make([]int, len(lbProxy.targetsInfo.targets))
	for i, target := range lbProxy.targetsInfo.targets {
		weight := MULT_FACTOR * int(maxCost/costs[i])
		if weight < 1 {
			weight = 1
		}
		weights[i] = weight
		totalWeight += weight
		servers[i] = Server{target: target, weight: weight}
	}

	log.Println(LB, "Servers:", servers)

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
func (p *WRRCostPolicy) SelectTarget(funName string) *url.URL {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.servers) == 0 {
		return nil
	}

	for {
		if p.totalReqs == p.totalWeight {
			p.totalReqs = 0
			p.requestCounts = make([]int, len(p.servers))
			p.index = 0
		}
		server := &p.servers[p.index]
		if p.requestCounts[p.index] < server.weight {
			p.requestCounts[p.index]++
			p.totalReqs++
			p.index = (p.index + 1) % len(p.servers)
			return server.target
		}
		p.index = (p.index + 1) % len(p.servers)
	}
}

/*
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
*/

func getCost(lbP *LBProxy, target string) float64 {
	cost := 1.0
	for _, targetStatus := range lbP.targetsInfo.targetsStatus {
		if targetStatus.Addresses.NodeAddress == target {
			cost = targetStatus.CostCloud
			break
		}
	}
	return cost
}
