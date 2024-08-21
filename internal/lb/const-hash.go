package lb

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"sync"
	"time"

	"github.com/grussorusso/serverledge/internal/function"
	"github.com/grussorusso/serverledge/internal/registration"
	"github.com/grussorusso/serverledge/utils"
)

// ConstHashPolicy: è un load balancer che utilizza la politica Consistent Hashing
type ConstHashPolicy struct {
	mu      sync.Mutex
	lbProxy *LBProxy
	ring    *Ring
}

// Ring is a structure that maintains nodes in sorted order
type Ring struct {
	nodes []*url.URL
	ring  []struct {
		key  uint64
		node *url.URL
	}
}

// NewConstHashPolicy: crea un nuovo load balancer Consistent Hashing
func NewConstHashPolicy(lbProxy *LBProxy) *ConstHashPolicy {
	log.Println(LB, "ConstHashPolicy created")
	ring := &Ring{
		nodes: lbProxy.targets,
	}
	for _, node := range lbProxy.targets {
		ring.AddNode(node)
	}
	return &ConstHashPolicy{
		lbProxy: lbProxy,
		ring:    ring,
	}
}

// SelectTarget: seleziona il prossimo target utilizzando la politica Consistent Hashing
func (r *ConstHashPolicy) SelectTarget(funName string) *url.URL {
	r.mu.Lock()
	defer r.mu.Unlock()
	fun, ok := getFromEtcd(funName)
	if !ok {
		log.Fatalf("%s Dropping request for unknown fun '%s'", LB, funName)
	}
	node := r.ring.GetNode(fun)
	return node
}

func getEtcdKey(funcName string) string {
	return fmt.Sprintf("/function/%s", funcName)
}

func getFromEtcd(name string) (*function.Function, bool) {
	cli, err := utils.GetEtcdClient()
	if err != nil {
		return nil, false
	}
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	getResponse, err := cli.Get(ctx, getEtcdKey(name))
	if err != nil || len(getResponse.Kvs) < 1 {
		return nil, false
	}

	var f function.Function
	err = json.Unmarshal(getResponse.Kvs[0].Value, &f)
	if err != nil {
		return nil, false
	}

	return &f, true
}

// Helper function to generate a key
func hash(key string) uint64 {
	hash := sha256.New()
	hash.Write([]byte(key))
	hashBytes := hash.Sum(nil)
	var hashValue uint64
	for _, b := range hashBytes[:8] { // Use only the first 8 bytes to get a uint64
		hashValue = hashValue<<8 + uint64(b)
	}
	return hashValue
}

// Adds a node to the ring
func (r *Ring) AddNode(node *url.URL) {
	hostport := node.Hostname() + ":" + node.Port()
	key := hash(hostport)
	r.ring = append(r.ring, struct {
		key  uint64
		node *url.URL
	}{key, node})
	sort.Slice(r.ring, func(i, j int) bool {
		return r.ring[i].key < r.ring[j].key
	})
}

// Finds the closest node based on the function
func (r *Ring) GetNode(fun *function.Function) *url.URL {
	key := hash(fun.Name)

	// Find the index where the key should be inserted
	startIndex := sort.Search(len(r.ring), func(i int) bool {
		return r.ring[i].key > key
	})

	if startIndex >= len(r.ring) {
		return r.ring[0].node
	}

	// Retrieve status information for all nodes
	nodesStatusInfo := make(map[*url.URL]*registration.StatusInformation)
	for _, node := range r.nodes {
		statusInfo := getNodesStatusInfo(node)
		if statusInfo != nil {
			nodesStatusInfo[node] = statusInfo
		} else {
			return nil
		}
	}

	// Store the starting node
	startNode := r.ring[startIndex].node

	// Check nodes starting from the startIndex and wrap around if necessary
	for _, entry := range append(r.ring[startIndex:], r.ring[:startIndex]...) {
		node := entry.node
		count, ok := nodesStatusInfo[node].AvailableWarmContainers[fun.Name]
		if (ok && count > 0) || (nodesStatusInfo[node].AvailableMemMB > fun.MemoryMB) {
			return node
		}
	}
	// If no suitable node is found, return the starting node
	return startNode
}

// Helper function to retrieve node status information via HTTP
func getNodesStatusInfo(node *url.URL) *registration.StatusInformation {
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
