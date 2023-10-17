package function

import (
	"fmt"
	"time"
)

// Request represents a single function invocation, with a ReqId, reference to the Function, parameters and metrics data
type Request struct {
	ReqId      string
	Fun        *Function
	Params     map[string]interface{}
	Arrival    time.Time
	ExecReport ExecutionReport
	RequestQoS
	CanDoOffloading bool
	Async           bool
	IsInComposition bool // not currently used
}

type RequestQoS struct {
	Class    ServiceClass
	MaxRespT float64
}

type ExecutionReport struct {
	Result         string
	ResponseTime   float64 // time waited by the user to get the output: completion time - arrival time (offload + cold start + execution time)
	IsWarmStart    bool
	InitTime       float64 // time spent sleeping before initializing container
	OffloadLatency float64 // time spent offloading the request
	Duration       float64 // execution (service) time
	SchedAction    string
}

type Response struct {
	Success bool
	ExecutionReport
}

type AsyncResponse struct {
	ReqId string
}

func (r *Request) String() string {
	return fmt.Sprintf("[%s] Rq-%s", r.Fun.Name, r.ReqId)
}

type ServiceClass int64

const (
	LOW               ServiceClass = 0
	HIGH_PERFORMANCE               = 1
	HIGH_AVAILABILITY              = 2
)
