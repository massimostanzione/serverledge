package scheduling

import (
	"fmt"
	"time"

	"github.com/grussorusso/serverledge/internal/config"
	"github.com/grussorusso/serverledge/internal/container"
	"github.com/grussorusso/serverledge/internal/executor"
)

const HANDLER_DIR = "/app"

// Execute serves a request on the specified container.
func Execute(contID container.ContainerID, r *scheduledRequest) error {
	//log.Printf("[%s] Executing on container: %v", r, contID)

	var req executor.InvocationRequest
	if r.Fun.Runtime == container.CUSTOM_RUNTIME {
		req = executor.InvocationRequest{
			Params: r.Params,
		}
	} else {
		cmd := container.RuntimeToInfo[r.Fun.Runtime].InvocationCmd
		req = executor.InvocationRequest{
			cmd,
			r.Params,
			r.Fun.Handler,
			HANDLER_DIR,
		}
	}

	t0 := time.Now()

	response, invocationWait, err := container.Execute(contID, &req)
	if err != nil {
		// notify scheduler
		completions <- &completion{scheduledRequest: r, contID: contID}
		return fmt.Errorf("[%s] Execution failed: %v", r, err)
	}

	if !response.Success {
		// notify scheduler
		completions <- &completion{scheduledRequest: r, contID: contID}
		return fmt.Errorf("Function execution failed")
	}

	r.ExecReport.Result = response.Result
	r.ExecReport.Duration = time.Now().Sub(t0).Seconds() - invocationWait.Seconds()
	r.ExecReport.ResponseTime = time.Now().Sub(r.Arrival).Seconds() + config.GetFloat(config.CLOUD_DELAY, 0.0)
	r.ExecReport.Cost = 0

	// Add the cost of the cloud in the report
	r.ExecReport.CostCloud = config.GetFloat(config.CLOUD_NODE_COST, 0.0) * r.ExecReport.Duration * (float64(r.Fun.MemoryMB) / 1024)

	// Add utility to the report
	r.ExecReport.Utility = r.ClassService.Utility

	// initializing containers may require invocation retries, adding
	// latency
	r.ExecReport.InitTime += invocationWait.Seconds()

	if r.ExecReport.ResponseTime > r.RequestQoS.MaxRespT {
		r.ExecReport.Violations += 1
	}

	// notify scheduler
	completions <- &completion{scheduledRequest: r, contID: contID}

	return nil
}
