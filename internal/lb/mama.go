package lb

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sync"

	"github.com/grussorusso/serverledge/internal/registration"
)

// MAMAPolicy: è un load balancer che utilizza la politica MAMA
type MAMAPolicy struct {
	mu      sync.Mutex
	lbProxy *LBProxy
}

// NewMAMAPolicy: crea un nuovo load balancer MAMA
func NewMAMAPolicy(lbProxy *LBProxy) *MAMAPolicy {
	log.Println(LB, "MAMAPolicy created")
	return &MAMAPolicy{
		lbProxy: lbProxy,
	}
}

// SelectTarget: seleziona il prossimo target utilizzando la politica MAMA
func (r *MAMAPolicy) SelectTarget(funName string) *url.URL {
	r.mu.Lock()
	defer r.mu.Unlock()
	nodes := r.lbProxy.targetsInfo.targets
	if len(nodes) == 0 {
		return nil
	}

	/*
		// Retrieve status information for all nodes
		var nodesStatus []*registration.StatusInformation
		for _, node := range nodes {
			statusInfo := getNodeStatus(node)
			if statusInfo != nil {
				nodesStatus = append(nodesStatus, statusInfo)
			} else {
				return nil
			}
		}
	*/

	// Filter nodes with warm containers
	var nodesWarm []*registration.StatusInformation
	for _, nodeStatus := range r.lbProxy.targetsInfo.targetsStatus {
		if count, ok := nodeStatus.AvailableWarmContainers[funName]; ok && count > 0 {
			nodesWarm = append(nodesWarm, nodeStatus)
		}
	}

	var selectedNode *registration.StatusInformation
	if len(nodesWarm) == 0 { // No warm containers available (cold start)
		selectedNode = getNodeWithMaxAvailableMem(r.lbProxy.targetsInfo.targetsStatus)
	} else { // Warm containers available (warm start)
		selectedNode = getNodeWithMaxAvailableMem(nodesWarm)
	}

	// Parse the selected node's URL and return it
	selectedNodeURL, err := url.Parse(selectedNode.Addresses.NodeAddress)
	if err != nil {
		log.Fatalf("%s Error parsing url: %v", LB, err)
	}

	/*
		for _, node := range nodes {
			if node.String() == selectedNodeURL.String() {
				return node
			}
		}
	*/

	return selectedNodeURL
}

// Helper function to retrieve node status information via HTTP
func getNodeStatus(node *url.URL) *registration.StatusInformation {
	resp, err := http.Get(node.String() + "/status")
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
		if err := json.Unmarshal(body, &statusInfo); err != nil {
			log.Fatalf("%s Error decoding JSON: %v", LB, err)
		}
		return &statusInfo
	}

	return nil
}

// Helper function to select the ndde with the maximum available memory
func getNodeWithMaxAvailableMem(nodesStatus []*registration.StatusInformation) *registration.StatusInformation {
	if len(nodesStatus) == 0 {
		return nil
	}

	// Start with the first node
	maxNode := nodesStatus[0]

	// Compare available memory of the other nodes
	for _, node := range nodesStatus[1:] {
		if node.AvailableMemMB > maxNode.AvailableMemMB {
			maxNode = node
		}
	}

	return maxNode
}
