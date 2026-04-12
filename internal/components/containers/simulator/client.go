package simulator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"maps"
	"sort"
	"strings"
	"sync"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/pkg/stdcopy"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// SimulatorContainerClient is an in-memory ContainerClient implementation intended for
// tests and local simulation where containers are assumed to run successfully.
type SimulatorContainerClient struct {
	mu sync.RWMutex

	nextID int64

	images     map[string]struct{}
	containers map[string]*inMemoryContainer
}

// inMemoryContainer is a simplified in-memory representation of a container, used by SimulatorContainerClient.
type inMemoryContainer struct {
	id      string
	name    string
	image   string
	labels  map[string]string
	created time.Time

	running  bool
	exitCode int

	stdout string
	stderr string
}

// NewSimulatorContainerClient creates a new SimulatorContainerClient with the given initial images available.
func NewSimulatorContainerClient(initialImages ...string) *SimulatorContainerClient {
	images := make(map[string]struct{}, len(initialImages))
	for _, img := range initialImages {
		images[img] = struct{}{}
	}

	return &SimulatorContainerClient{
		nextID:     1,
		images:     images,
		containers: make(map[string]*inMemoryContainer),
	}
}

// ContainerCreate creates a new container with the given configuration and returns its ID.
func (s *SimulatorContainerClient) ContainerCreate(
	_ context.Context,
	config *container.Config,
	_ *container.HostConfig,
	_ *network.NetworkingConfig,
	_ *ocispec.Platform,
	containerName string,
) (container.CreateResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := fmt.Sprintf("sim-%d", s.nextID)
	s.nextID++

	labels := make(map[string]string)
	if config != nil && config.Labels != nil {
		maps.Copy(labels, config.Labels)
	}

	imageName := ""
	if config != nil {
		imageName = config.Image
	}

	s.containers[id] = &inMemoryContainer{
		id:       id,
		name:     containerName,
		image:    imageName,
		labels:   labels,
		created:  time.Now().UTC(),
		running:  false,
		exitCode: 0,
		stdout:   "",
		stderr:   "",
	}

	return container.CreateResponse{ID: id}, nil
}

// ContainerStart starts the container with the given ID.
// In this simulator, we use the tag to determine if the container
// should fail to start. If the tag "fail" is present, the container will fail to start.
func (s *SimulatorContainerClient) ContainerStart(
	_ context.Context,
	containerID string,
	_ container.StartOptions,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.containers[containerID]
	if !ok {
		return cerrdefs.ErrNotFound
	}

	if _, fail := c.labels["fail"]; fail {
		c.running = false
		c.exitCode = 1
		c.stderr = "Simulated container start failure"

		return nil
	}

	c.running = true
	c.exitCode = 0

	return nil
}

// ContainerStop stops the container with the given ID.
func (s *SimulatorContainerClient) ContainerStop(
	_ context.Context,
	containerID string,
	_ container.StopOptions,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	c, ok := s.containers[containerID]
	if !ok {
		return cerrdefs.ErrNotFound
	}

	c.running = false

	return nil
}

// ContainerRemove removes the container with the given ID.
func (s *SimulatorContainerClient) ContainerRemove(
	_ context.Context,
	containerID string,
	_ container.RemoveOptions,
) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.containers[containerID]; !ok {
		return cerrdefs.ErrNotFound
	}

	delete(s.containers, containerID)

	return nil
}

// ContainerList returns a list of containers, filtered according to the provided options.
func (s *SimulatorContainerClient) ContainerList(
	_ context.Context,
	options container.ListOptions,
) ([]container.Summary, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	ids := make([]string, 0, len(s.containers))
	for id := range s.containers {
		ids = append(ids, id)
	}
	sort.Strings(ids)

	out := make([]container.Summary, 0, len(ids))
	for _, id := range ids {
		c := s.containers[id]

		if !options.All && !c.running {
			continue
		}

		if options.Filters.Len() > 0 && !options.Filters.MatchKVList("label", c.labels) {
			continue
		}

		state := "exited"
		status := fmt.Sprintf("Exited (%d)", c.exitCode)
		if c.running {
			state = "running"
			status = "Up"
		}

		name := c.name
		if name != "" && !strings.HasPrefix(name, "/") {
			name = "/" + name
		}

		out = append(out, container.Summary{
			ID:      c.id,
			Names:   []string{name},
			Image:   c.image,
			State:   state,
			Status:  status,
			Created: c.created.Unix(),
			Labels:  c.labels,
		})
	}

	return out, nil
}

// ContainerLogs returns the logs for the container with the given ID.
func (s *SimulatorContainerClient) ContainerLogs(
	_ context.Context,
	containerID string,
	_ container.LogsOptions,
) (io.ReadCloser, error) {
	s.mu.RLock()
	c, ok := s.containers[containerID]
	if !ok {
		s.mu.RUnlock()
		return nil, cerrdefs.ErrNotFound
	}

	stdout := c.stdout
	stderr := c.stderr
	s.mu.RUnlock()

	var buf bytes.Buffer
	if stdout != "" {
		_, _ = stdcopy.NewStdWriter(&buf, stdcopy.Stdout).Write([]byte(stdout))
	}
	if stderr != "" {
		_, _ = stdcopy.NewStdWriter(&buf, stdcopy.Stderr).Write([]byte(stderr))
	}

	return io.NopCloser(bytes.NewReader(buf.Bytes())), nil
}

// ContainerInspect returns detailed information about the container with the given ID.
func (s *SimulatorContainerClient) ContainerInspect(
	_ context.Context,
	containerID string,
) (container.InspectResponse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.containers[containerID]
	if !ok {
		return container.InspectResponse{}, cerrdefs.ErrNotFound
	}

	state := &container.State{
		Running:  c.running,
		ExitCode: c.exitCode,
	}
	if c.running {
		state.Status = "running"
	} else {
		state.Status = "exited"
	}

	return container.InspectResponse{
		ContainerJSONBase: &container.ContainerJSONBase{
			ID:      c.id,
			Name:    "/" + c.name,
			Created: c.created.Format(time.RFC3339),
			State:   state,
			Image:   c.image,
		},
		Config: &container.Config{
			Image:  c.image,
			Labels: c.labels,
		},
	}, nil
}

// ImageInspectWithRaw returns detailed information about the image with the given name.
func (s *SimulatorContainerClient) ImageInspectWithRaw(
	_ context.Context,
	imageName string,
) (image.InspectResponse, []byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, ok := s.images[imageName]; !ok {
		return image.InspectResponse{}, nil, cerrdefs.ErrNotFound
	}

	return image.InspectResponse{}, nil, nil
}

// ImagePull simulates pulling an image by adding it to the in-memory set of available images.
func (s *SimulatorContainerClient) ImagePull(
	_ context.Context,
	refStr string,
	_ image.PullOptions,
) (io.ReadCloser, error) {
	s.mu.Lock()
	s.images[refStr] = struct{}{}
	s.mu.Unlock()

	payload := map[string]string{"status": "Downloaded newer image"}
	b, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return io.NopCloser(strings.NewReader(string(b) + "\n")), nil
}

// ImageRemove simulates removing an image by deleting it from the in-memory set of available images.
func (s *SimulatorContainerClient) ImageRemove(
	_ context.Context,
	imageName string,
	_ image.RemoveOptions,
) ([]image.DeleteResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.images[imageName]; !ok {
		return nil, cerrdefs.ErrNotFound
	}

	delete(s.images, imageName)

	return []image.DeleteResponse{{Deleted: imageName}}, nil
}
