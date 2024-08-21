package lb

import "net/url"

// Interfaccia per le politiche di load balancing
type LBPolicy interface {
	SelectTarget(funName string) *url.URL
}
