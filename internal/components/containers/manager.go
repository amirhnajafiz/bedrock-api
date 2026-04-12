package containers

import (
	"context"

	"github.com/amirhnajafiz/bedrock-api/internal/components/containers/docker"
	"github.com/amirhnajafiz/bedrock-api/internal/components/containers/simulator"
	"github.com/amirhnajafiz/bedrock-api/pkg/models"
)

// ContainerManager defines the interface for container management operations.
// This interface abstracts the underlying container runtime.
type ContainerManager interface {
	// Create sets up a new container and returns the container ID.
	Create(ctx context.Context, cfg *models.ContainerConfig) (string, error)
	// Start starts a created container.
	Start(ctx context.Context, containerID string) error
	// Stop stops a running container.
	Stop(ctx context.Context, containerID string) error
	// Remove removes a container.
	Remove(ctx context.Context, containerID string) error
	// StoreLogs writes the container's stdout and stderr to the given file path.
	StoreLogs(ctx context.Context, containerID string, filePath string) error
	// List returns all containers managed by this instance.
	List(ctx context.Context, labels map[string]string) ([]*models.ContainerInfo, error)
	// Get returns information about a specific container.
	Get(ctx context.Context, containerID string) (*models.ContainerInfo, error)
}

// NewDockerManager creates a new ContainerManager that uses Docker as the container runtime.
func NewDockerManager() (ContainerManager, error) {
	return docker.NewDockerContainerManager()
}

// NewSimulatorManager creates a new ContainerManager that uses a simulator for testing purposes.
func NewSimulatorManager() (ContainerManager, error) {
	return simulator.NewSimulatorContainerManager()
}
