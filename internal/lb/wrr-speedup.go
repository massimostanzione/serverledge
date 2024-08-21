package lb

import (
	"net/url"
	"sync"
)

// WRRSpeedupPolicy: è un load balancer che utilizza la politica wrr-speedup
type WRRSpeedupPolicy struct {
	mu            sync.Mutex
	servers       []Server
	weights       []int
	totalWeight   int
	index         int
	requestCounts []int
	totalReqs     int
}

// NewWRRSpeedupPolicy: crea un nuovo load balancer wrr-speedup
func NewWRRSpeedupPolicy(targets []*url.URL) *WRRSpeedupPolicy {

	// Recupero lo speedup dei nodi cloud
	speedups := make([]float64, len(targets))
	for i, target := range targets {
		speedups[i] = getSpeedup(target)
	}

	// Determino lo speedup minimo tra i server
	minSpeedup := speedups[0]
	for _, value := range speedups[1:] {
		if value < minSpeedup {
			minSpeedup = value
		}
	}

	// Calcolo i pesi di ciascun server
	servers := make([]Server, len(targets))
	totalWeight := 0
	weights := make([]int, len(targets))
	for i, target := range targets {
		weight := MULT_FACTOR * int(speedups[i]/minSpeedup)
		if weight < 1 {
			weight = 1
		}
		weights[i] = weight
		totalWeight += weight
		servers[i] = Server{target: target, weight: weight}
	}

	// Ritorno della struttura inizializzata
	return &WRRSpeedupPolicy{
		servers:       servers,
		weights:       weights,
		totalWeight:   totalWeight,
		index:         0,
		requestCounts: make([]int, len(servers)),
		totalReqs:     0,
	}
}

// SelectTarget: seleziona il prossimo target utilizzando la politica wrr-speedup
func (r *WRRSpeedupPolicy) SelectTarget(funName string) *url.URL {
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

func getSpeedup(target *url.URL) float64 {
	// TODO: Invocare la getStatus() del nodo Serverledge per recuperare lo speedup
	return 0
}
