package lb

import (
	"net/url"

	"github.com/grussorusso/serverledge/internal/lbcommon"
	"github.com/grussorusso/serverledge/internal/mab"
)

const LB = "[LB]:"

type LBProxy struct {
	//mu           sync.RWMutex
	targets      []*url.URL
	lbPolicyName lbcommon.Policy
	lbPolicy     LBPolicy
	oldStats     *mab.Stats
	newStats     *mab.Stats
}
