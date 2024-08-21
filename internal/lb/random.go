package lb

import (
	"log"
	"math/rand"
	"net/url"
)

// RandomPolicy: è un load balancer che utilizza la politica Random
type RandomPolicy struct {
	lbProxy *LBProxy
}

// NewRandomPolicy: crea un nuovo load balancer Random
func NewRandomPolicy(lbProxy *LBProxy) *RandomPolicy {
	log.Println(LB, "RandomPolicy created")
	return &RandomPolicy{
		lbProxy: lbProxy,
	}
}

// SelectTarget: seleziona
func (r *RandomPolicy) SelectTarget(funName string) *url.URL {
	nodes := r.lbProxy.targets
	if len(nodes) == 0 {
		return nil
	}
	return nodes[rand.Intn(len(nodes))]
}
