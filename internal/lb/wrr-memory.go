package lb

import (
	"log"
	"net/url"
	"sync"
)

var servers []Server
var weights []int
var totalWeight int

// WRRMemoryPolicy: è un load balancer che utilizza la politica wrr-memory
type WRRMemoryPolicy struct {
	mu            sync.Mutex
	lbProxy       *LBProxy
	servers       []Server
	weights       []int
	totalWeight   int
	index         int
	requestCounts []int
	totalReqs     int
}

// NewWRRMemoryPolicy: crea un nuovo load balancer wrr-memory
func NewWRRMemoryPolicy(lbProxy *LBProxy) *WRRMemoryPolicy {

	log.Println(LB, "WRRMemoryPolicy created")

	// Recupero le memorie dei nodi cloud
	memories := make([]int, len(lbProxy.targetsInfo.targets))
	for i, target := range lbProxy.targetsInfo.targets {
		memories[i] = getMemory(lbProxy, target.String())
	}

	// Determino la memoria minima tra i server
	minMem := memories[0]
	for _, value := range memories[1:] {
		if value < minMem {
			minMem = value
		}
	}

	// Calcolo i pesi di ciascun server
	servers := make([]Server, len(lbProxy.targetsInfo.targets))
	totalWeight := 0
	weights := make([]int, len(lbProxy.targetsInfo.targets))
	for i, target := range lbProxy.targetsInfo.targets {
		weight := MULT_FACTOR * int(memories[i]/minMem)
		if weight < 1 {
			weight = 1
		}
		weights[i] = weight
		totalWeight += weight
		servers[i] = Server{target: target, weight: weight}
	}

	log.Println(LB, "Servers:", servers)

	// Ritorno della struttura inizializzata
	return &WRRMemoryPolicy{
		lbProxy:       lbProxy,
		servers:       servers,
		weights:       weights,
		totalWeight:   totalWeight,
		index:         0,
		requestCounts: make([]int, len(servers)),
		totalReqs:     0,
	}
}

// SelectTarget: seleziona il prossimo target utilizzando la politica wrr-memory
func (p *WRRMemoryPolicy) SelectTarget(funName string) *url.URL {
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
func getMemory(target *url.URL) int {
	url := fmt.Sprintf("http://%s/status", target.Host)
	resp, err := http.Get(url)
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

		// Decode the JSON into a StatusInformation structure
		err = json.Unmarshal(body, &statusInfo)
		if err != nil {
			log.Fatalf("%s Error decoding JSON: %v", LB, err)
		}

		return int(statusInfo.MaxMemMB)
	}

	return 1
}
*/

func getMemory(lbP *LBProxy, target string) int {
	memory := 1
	for _, targetStatus := range lbP.targetsInfo.targetsStatus {
		if targetStatus.Addresses.NodeAddress == target {
			memory = int(targetStatus.MaxMemMB)
			break
		}
	}
	return memory
}
