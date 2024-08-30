package lb

import (
	"net/url"

	"github.com/grussorusso/serverledge/internal/lbcommon"
	"github.com/grussorusso/serverledge/internal/mab"
	"github.com/grussorusso/serverledge/internal/registration"
)

const LB = "[LB]:"

type TargetsInfo struct {
	targets       []*url.URL
	targetsStatus []*registration.StatusInformation
}

type LBProxy struct {
	targetsInfo  *TargetsInfo
	lbPolicyName lbcommon.Policy
	lbPolicy     LBPolicy
	oldStats     *mab.Stats
	newStats     *mab.Stats
}
