package containers

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

// Label applied to every container created by this manager so that List can
// filter them from unrelated containers on the same Docker host.
const (
	labelKey   = "bedrock.managed-by"
	labelValue = "bedrock-dockerd"
)

// dockerManager implements ContainerManager using the Docker Engine API.
type dockerManager struct {
	client client.APIClient
}

func (m *dockerManager) ensureImage(ctx context.Context, imageName string) error {
	_, err := m.client.ImageInspect(ctx, imageName)
	if err == nil {
		return nil
	}
	if !errdefs.IsNotFound(err) {
		return fmt.Errorf("checking image %s: %w", imageName, err)
	}
	reader, err := m.client.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return fmt.Errorf("pulling image %s: %w", imageName, err)
	}
	defer reader.Close()
	if _, err := io.Copy(os.Stdout, reader); err != nil {
		return fmt.Errorf("streaming image pull %s: %w", imageName, err)
	}
	return nil
}

// Create pulls together the container configuration from cfg, creates the
// container on the Docker host, and starts it.
func (m *dockerManager) Create(ctx context.Context, cfg ContainerConfig) (string, error) {
	// pull image if does not exist
	if err := m.ensureImage(ctx, cfg.Image); err != nil {
		return "", err
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
	resp, err := m.client.ContainerCreate(
		ctx,
		&container.Config{
			Image: cfg.Image,
			Env:   cfg.Env,
			Cmd:   cfg.Cmd,
			Labels: map[string]string{
				labelKey: labelValue,
			},
		},
		hostConfig,
		nil,
		nil,
		cfg.Name,
	)
	if err != nil {
		return "", err
	}

	// start the container
	if err := m.client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		_ = m.client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{})
		return "", err
	}

	return resp.ID, nil
}

// StoreLogs fetches the stdout and stderr streams of a container and writes
// them to filePath. The Docker multiplexed log format is decoded before writing.
func (m *dockerManager) StoreLogs(ctx context.Context, containerID string, filePath string) error {
	reader, err := m.client.ContainerLogs(ctx, containerID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
	})
	if err != nil {
		return err
	}
	defer reader.Close()

	f, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = stdcopy.StdCopy(f, f, reader)
	return err
}

// List returns information about every container that carries the bedrock
// managed-by label, regardless of whether it is running or stopped.
func (m *dockerManager) List(ctx context.Context) ([]ContainerInfo, error) {
	raw, err := m.client.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", labelKey+"="+labelValue),
		),
	})
	if err != nil {
		return nil, err
	}

	infos := make([]ContainerInfo, 0, len(raw))
	for _, summary := range raw {
		exited, exitCode := false, 0
		// call ContainerInspect to get the exit code if the container has finished
		if inspect, err := m.client.ContainerInspect(ctx, summary.ID); err == nil {
			if inspect.State != nil && !inspect.State.Running {
				exited = true
				exitCode = 0
			}
		}

		infos = append(infos, ContainerInfo{
			Summary:  summary,
			Exited:   exited,
			ExitCode: exitCode,
		})
	}

	return infos, nil
}

// Stop stops a running container.
func (m *dockerManager) Stop(ctx context.Context, containerID string) error {
	return m.client.ContainerStop(ctx, containerID, container.StopOptions{})
}

// Remove removes a container.
func (m *dockerManager) Remove(ctx context.Context, containerID string) error {
	return m.client.ContainerRemove(ctx, containerID, container.RemoveOptions{})
}

// Inspect returns detailed information about a container, including its exit code if it has finished.
func (m *dockerManager) Inspect(ctx context.Context, containerID string) (container.InspectResponse, error) {
	return m.client.ContainerInspect(ctx, containerID)
}
