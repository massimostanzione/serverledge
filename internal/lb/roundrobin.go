package lb

import (
	"log"
	"net/url"
	"sync"
)

// RoundRobinPolicy: è un load balancer che utilizza la politica Round Robin
type RoundRobinPolicy struct {
	mu      sync.Mutex
	index   int
	lbProxy *LBProxy
}

// NewRoundRobinPolicy: crea un nuovo load balancer Round Robin
func NewRoundRobinPolicy(lbProxy *LBProxy) *RoundRobinPolicy {
	log.Println(LB, "RoundRobinPolicy created")
	return &RoundRobinPolicy{
		index:   0,
		lbProxy: lbProxy,
	}
}

// SelectTarget: seleziona il prossimo target utilizzando la politica Round-Robin
func (r *RoundRobinPolicy) SelectTarget(funName string) *url.URL {
	r.mu.Lock()
	defer r.mu.Unlock()
	nodes := r.lbProxy.targets
	if len(nodes) == 0 {
		return nil
	}
	targetIndex := r.index % len(nodes)
	r.index = targetIndex
	r.index = r.index + 1
	return nodes[targetIndex]
}
