package docker

import (
	"context"
	"io"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
)

// DockerContainerClient abstracts the Docker Engine Client methods needed by the Docker manager.
type DockerContainerClient interface {
	// Container management
	ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkingConfig *network.NetworkingConfig, platform *ocispec.Platform, containerName string) (container.CreateResponse, error)
	ContainerStart(ctx context.Context, containerID string, options container.StartOptions) error
	ContainerStop(ctx context.Context, containerID string, options container.StopOptions) error
	ContainerRemove(ctx context.Context, containerID string, options container.RemoveOptions) error
	ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error)
	ContainerLogs(ctx context.Context, container string, options container.LogsOptions) (io.ReadCloser, error)
	ContainerInspect(ctx context.Context, containerID string) (containerJSON container.InspectResponse, err error)
	// Image management
	ImageInspectWithRaw(ctx context.Context, imageName string) (image.InspectResponse, []byte, error)
	ImagePull(ctx context.Context, refStr string, options image.PullOptions) (io.ReadCloser, error)
	ImageRemove(ctx context.Context, imageName string, options image.RemoveOptions) ([]image.DeleteResponse, error)
}

// NewDockerContainerClient creates a new DockerContainerClient using environment variables for configuration.
func NewDockerContainerClient() (DockerContainerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return cli, nil
}
