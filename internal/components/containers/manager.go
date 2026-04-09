package containers

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
)

// ContainerManager manages Docker container lifecycles.
// Implementations must be safe for concurrent use.
type ContainerManager interface {
	// Start starts a new container. Returns the container ID.
	Start(ctx context.Context, cfg *ContainerConfig) (string, error)
	// StoreLogs writes the container's stdout and stderr to the given file path.
	StoreLogs(ctx context.Context, containerID string, filePath string) error
	// List returns all containers managed by this instance.
	List(ctx context.Context) ([]*ContainerInfo, error)
	// Get returns information about a specific container.
	Get(ctx context.Context, containerID string) (*ContainerInfo, error)
	// Stop stops a running container.
	Stop(ctx context.Context, containerID string) error
	// Remove removes a container.
	Remove(ctx context.Context, containerID string) error
	// Get client returns the underlying container runtime client.
	GetClient() ContainerClient
}

// NewContainerManager returns a ContainerManager backed by the runtime client.
func NewContainerManager(rc string) (ContainerManager, error) {
	var (
		err error
		cli ContainerClient
	)

	switch rc {
	case "simulator":
		cli = newSimulatorClient()
	case "docker":
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("unsupported runtime client: %s", rc)
	}

	return &dockerManager{client: cli}, nil
}
