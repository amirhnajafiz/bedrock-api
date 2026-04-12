package docker

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/amirhnajafiz/bedrock-api/pkg/models"
	"github.com/amirhnajafiz/bedrock-api/pkg/xerrors"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/pkg/stdcopy"
)

// DockerContainerManager implements ContainerManager using the Docker Engine API.
type DockerContainerManager struct {
	client DockerContainerClient
}

// NewDockerContainerManager creates a new DockerContainerManager with the given Docker client.
func NewDockerContainerManager() (*DockerContainerManager, error) {
	cli, err := NewDockerContainerClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %v", err)
	}

	return &DockerContainerManager{
		client: cli,
	}, nil
}

// ensureImage checks if the specified image is available locally, and attempts to pull it if not.
func (m *DockerContainerManager) ensureImage(ctx context.Context, imageName string) error {
	// pull only when the image is not available locally.
	if _, _, err := m.client.ImageInspectWithRaw(ctx, imageName); err != nil {
		if !cerrdefs.IsNotFound(err) {
			return fmt.Errorf("failed to inspect image: %w", err)
		}

		// image not found locally, attempt to pull it
		pullReader, err := m.client.ImagePull(ctx, imageName, image.PullOptions{})
		if err != nil {
			return fmt.Errorf("failed to pull image: %w", err)
		}
		defer pullReader.Close()

		// read the pull output to completion to ensure the image is fully pulled
		decoder := json.NewDecoder(pullReader)
		for {
			var msg map[string]any
			if err := decoder.Decode(&msg); err != nil {
				if err == io.EOF {
					break
				}
				return fmt.Errorf("failed reading pull stream: %w", err)
			}

			if e, ok := msg["error"]; ok {
				return fmt.Errorf("daemon pull error: %v", e)
			}
		}
	}

	return nil
}

// Create sets up a new container and returns the container ID.
func (m *DockerContainerManager) Create(ctx context.Context, cfg *models.ContainerConfig) (string, error) {
	// ensure the image is available locally before creating the container
	if err := m.ensureImage(ctx, cfg.Image); err != nil {
		return "", err
	}

	// check the labels
	if cfg.Labels == nil {
		cfg.Labels = make(map[string]string)
	}

	// set up volume mounts
	var mounts []mount.Mount
	for hostPath, containerPath := range cfg.Volumes {
		mounts = append(mounts, mount.Mount{
			Type:   mount.TypeBind,
			Source: hostPath,
			Target: containerPath,
		})
	}

	// set up host config with mounts and flags
	hostConfig := &container.HostConfig{
		AutoRemove:    false,
		Mounts:        mounts,
		RestartPolicy: container.RestartPolicy{Name: "no"},
	}
	if privileged, ok := cfg.Flags["privileged"].(bool); ok && privileged {
		hostConfig.Privileged = true
	}
	if pidMode, ok := cfg.Flags["pid"].(string); ok {
		hostConfig.PidMode = container.PidMode(pidMode)
	}

	// create and start the container
	dockerId, err := m.client.ContainerCreate(
		ctx,
		&container.Config{
			Image:  cfg.Image,
			Env:    cfg.Env,
			Cmd:    cfg.Cmd,
			Labels: cfg.Labels,
		},
		hostConfig,
		nil,
		nil,
		cfg.Name,
	)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}

	return dockerId.ID, nil
}

// Start starts a created container.
func (m *DockerContainerManager) Start(ctx context.Context, containerID string) error {
	// start the container
	if err := m.client.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		if cerrdefs.IsNotFound(err) {
			return xerrors.ContainerManagerErrNotFound
		}

		return fmt.Errorf("failed to start container: %w", err)
	}

	return nil
}

// StoreLogs fetches the stdout and stderr streams of a container and writes
// them to filePath. The Docker multiplexed log format is decoded before writing.
func (m *DockerContainerManager) StoreLogs(ctx context.Context, containerID string, filePath string) error {
	// fetch the logs as a multiplexed stream
	reader, err := m.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return xerrors.ContainerManagerErrNotFound
		}

		return fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	// create the log file and decode the multiplexed stream into it
	f, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create log file: %w", err)
	}
	defer f.Close()

	// redirect the decoded stdout and stderr to the same file
	_, err = stdcopy.StdCopy(f, f, reader)
	if err != nil {
		return fmt.Errorf("failed to write logs: %w", err)
	}

	return nil
}

// List returns information about every container regardless of whether it is running or stopped.
func (m *DockerContainerManager) List(ctx context.Context, labels map[string]string) ([]*models.ContainerInfo, error) {
	// create a filter to list all containers, including stopped ones
	args := filters.NewArgs()
	for k, v := range labels {
		args.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	raw, err := m.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: args,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	// convert the raw container data to ContainerInfo instances
	records := make([]*models.ContainerInfo, 0, len(raw))
	for _, c := range raw {
		name := ""
		if len(c.Names) > 0 {
			name = strings.TrimPrefix(c.Names[0], "/")
		}

		records = append(records, &models.ContainerInfo{
			ID:     c.ID,
			Name:   name,
			Image:  c.Image,
			Status: c.Status,
			Labels: c.Labels,
		})
	}

	return records, nil
}

// Get returns information about a specific container.
func (m *DockerContainerManager) Get(ctx context.Context, containerID string) (*models.ContainerInfo, error) {
	// call ContainerInspect to get detailed information about the container
	inspect, err := m.client.ContainerInspect(ctx, containerID)
	if err != nil {
		if cerrdefs.IsNotFound(err) {
			return nil, xerrors.ContainerManagerErrNotFound
		}

		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	// check if state is nil
	if inspect.State == nil {
		return nil, fmt.Errorf("container state is nil")
	}

	// extract the container name from the inspect data, trimming the leading slash if present
	name := ""
	if len(inspect.Name) > 0 {
		name = strings.TrimPrefix(inspect.Name, "/")
	}

	// convert the inspect created time string to a timestamp
	createdAt, err := time.Parse(time.RFC3339, inspect.Created)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created time: %w", err)
	}

	// create a container info instance with the inspect data
	record := &models.ContainerInfo{
		ID:        inspect.ID,
		Name:      name,
		Image:     inspect.Config.Image,
		Status:    inspect.State.Status,
		Exited:    false,
		ExitCode:  0,
		CreatedAt: createdAt,
	}

	// if the container is not running, set the Exited and ExitCode fields
	if !inspect.State.Running {
		record.Exited = true
		record.ExitCode = int(inspect.State.ExitCode)
	}

	return record, nil
}

// Stop stops a running container.
func (m *DockerContainerManager) Stop(ctx context.Context, containerID string) error {
	if err := m.client.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		if cerrdefs.IsNotFound(err) {
			return xerrors.ContainerManagerErrNotFound
		}

		return fmt.Errorf("failed to stop container: %w", err)
	}

	return nil
}

// Remove removes a container.
func (m *DockerContainerManager) Remove(ctx context.Context, containerID string) error {
	if err := m.client.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		if cerrdefs.IsNotFound(err) {
			return xerrors.ContainerManagerErrNotFound
		}

		return fmt.Errorf("failed to remove container: %w", err)
	}

	return nil
}
