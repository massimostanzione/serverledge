package container

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/containers/podman/v4/libpod/define"
	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/bindings/images"
	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/grussorusso/serverledge/internal/config"
	"github.com/opencontainers/runtime-spec/specs-go"
)

type PodmanFactory struct {
	ctx context.Context
}

// Initialize the container factory, acquiring the context
func InitPodmanContainerFactory() *PodmanFactory {
	ctx, err := bindings.NewConnection(context.Background(), config.PODMAN_SOCKET)
	if err != nil {
		panic(err)
	}

	podmanFact := &PodmanFactory{ctx}
	cf = podmanFact
	return podmanFact
}

// Create a container
func (cf *PodmanFactory) Create(image string, opts *ContainerOptions) (ContainerID, error) {
	if !cf.HasImage(image) {
		log.Printf("Pulling image: %s", image)
		_, err := images.Pull(cf.ctx, image, new(images.PullOptions))
		if err != nil {
			log.Printf("Could not pull image: %s", image)
			// we do not return here, as a stale copy of the image
			// could still be available locally
		}
	}

	s := specgen.NewSpecGenerator(image, false)
	s.Image = image
	s.Command = opts.Cmd
	s.EnvMerge = opts.Env
	s.Terminal = false
	memory_limit := (opts.MemoryMB * 1048576)
	s.ResourceLimits = new(specs.LinuxResources)
	s.ResourceLimits.Memory = new(specs.LinuxMemory)
	s.ResourceLimits.Memory.Limit = &memory_limit
	s.SeccompProfilePath = "unconfined"
	r, err := containers.CreateWithSpec(cf.ctx, s, new(containers.CreateOptions))
	return r.ID, err
}

/* Copy a file to the container. Podman API doesnt support container
files copy, so this function does the copy by shell.*/
func (cf *PodmanFactory) CopyToContainer(contID ContainerID, content io.Reader, destPath string) error {
	b, _ := io.ReadAll(content)                                             // Get the function bytes
	b = bytes.Trim(b, "\x00")                                               // Remove null bytes
	functionBody := strings.Split(string(b), "\n")                          // Get the function body
	functionInfo := strings.Split(functionBody[0], "0")                     // The first line contains the file name and the function signature
	fileName := strings.Trim(functionInfo[0], "\x00")                       // Isolate the function file name
	functionName := strings.Trim(functionInfo[len(functionInfo)-1], "\x00") // Isolate the function signature
	functionBody = functionBody[1:]                                         // Now consider the function body as the remaining part of the content

	tmpFile, err := os.Create("/tmp/" + fileName) // Create a temporary copy file to transfer it to the container
	if err != nil {
		return err
	}
	defer os.Remove("/tmp/" + fileName)

	fmt.Fprintln(tmpFile, functionName) // Append the function signaure and then all the remaining lines
	for _, line := range functionBody {
		if line != "" {
			fmt.Fprintln(tmpFile, line)
		}
	}
	tmpFile.Close()

	// Now the temporary function file copy is ready: do the copy
	err = exec.Command("podman", "cp", "/tmp/"+fileName, contID+":"+destPath).Run()

	return err
}

// Start an existing container
func (cf *PodmanFactory) Start(contID ContainerID) error {
	err := containers.Start(cf.ctx, contID, nil)
	if err != nil {
		log.Printf("The container %s could not be started: %v", contID, err)
		return err
	}
	running := define.ContainerStateRunning
	_, err = containers.Wait(cf.ctx, contID, new(containers.WaitOptions).WithCondition([]define.ContainerStatus{running}))
	return err
}

// Destroy an existing container
func (cf *PodmanFactory) Destroy(contID ContainerID) error {
	// force set to true causes running container to be killed (and then removed)
	err := containers.Stop(cf.ctx, contID, new(containers.StopOptions).WithTimeout(0))
	if err != nil {
		log.Printf("The container %s could not be stopped: %v", contID, err)
		return err
	}
	_, err = containers.Remove(cf.ctx, contID, new(containers.RemoveOptions))
	return err
}

// Check if the image exists locally
func (cf *PodmanFactory) HasImage(image string) bool {
	// TODO: we should try using cf.cli.ImageList(...)
	cmd := fmt.Sprintf("podman images %s | grep -vF REPOSITORY", image)
	_, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return false
	}

	// We have the image, but we may need to refresh it
	if config.GetBool(config.FACTORY_REFRESH_IMAGES, false) {
		if refreshed, ok := refreshedImages[image]; !ok || !refreshed {
			return false
		}
	}
	return true
}

// Get the IP address of the container
func (cf *PodmanFactory) GetIPAddress(contID ContainerID) (string, error) {
	contJson, err := containers.Inspect(cf.ctx, contID, new(containers.InspectOptions))
	if err != nil {
		return "", err
	}
	return contJson.NetworkSettings.IPAddress, nil
}

// Get the container memory in MB
func (cf *PodmanFactory) GetMemoryMB(contID ContainerID) (int64, error) {
	contJson, err := containers.Inspect(cf.ctx, contID, new(containers.InspectOptions))
	if err != nil {
		return -1, err
	}
	return contJson.HostConfig.Memory / 1048576, nil
}

// Checkpoint the container memory/execution state into a local tar archive
func (cf *PodmanFactory) CheckpointContainer(contID ContainerID, archiveName string) error {
	// Container checkpoint
	options := new(containers.CheckpointOptions).WithExport(archiveName).WithTCPEstablished(true)
	_, err := containers.Checkpoint(cf.ctx, contID, options)
	if err != nil {
		log.Printf("The container %s could not be checkpointed: %v", contID, err)
	}
	Destroy(contID)
	return err
}

// Restore a container execution from a local tar checkpoint archive
func (cf *PodmanFactory) RestoreContainer(contID ContainerID, archiveName string) (string, error) {
	// Container restore
	options := new(containers.RestoreOptions).WithImportArchive(archiveName).WithTCPEstablished(true)
	restoreReport, err := containers.Restore(cf.ctx, contID, options)
	if err != nil {
		log.Printf("The container %s could not be restored: %v", contID, err)
	}
	return restoreReport.Id, err
}
