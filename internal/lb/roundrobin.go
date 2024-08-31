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
func (p *RoundRobinPolicy) SelectTarget(funName string) *url.URL {
	p.mu.Lock()
	defer p.mu.Unlock()
	nodes := p.lbProxy.targetsInfo.targets
	if len(nodes) == 0 {
		return nil
	}
	targetIndex := p.index % len(nodes)
	p.index = targetIndex
	p.index = p.index + 1
	return nodes[targetIndex]
}
