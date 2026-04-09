package containers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"testing"
	"time"

	cerrdefs "github.com/containerd/errdefs"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/pkg/stdcopy"
)

func TestSimulatorClient_ImageLifecycle(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newSimulatorClient("alpine:latest")

	_, _, err := client.ImageInspectWithRaw(ctx, "alpine:latest")
	if err != nil {
		t.Fatalf("ImageInspectWithRaw existing image error: %v", err)
	}

	_, _, err = client.ImageInspectWithRaw(ctx, "busybox:latest")
	if !cerrdefs.IsNotFound(err) {
		t.Fatalf("ImageInspectWithRaw missing image error = %v, want not found", err)
	}

	pullReader, err := client.ImagePull(ctx, "busybox:latest", image.PullOptions{})
	if err != nil {
		t.Fatalf("ImagePull error: %v", err)
	}
	defer pullReader.Close()

	pullData, err := io.ReadAll(pullReader)
	if err != nil {
		t.Fatalf("reading pull stream failed: %v", err)
	}

	var msg map[string]string
	if err := json.Unmarshal(bytes.TrimSpace(pullData), &msg); err != nil {
		t.Fatalf("pull stream JSON decode failed: %v", err)
	}
	if msg["status"] == "" {
		t.Fatalf("pull stream status missing, got: %s", string(pullData))
	}

	_, _, err = client.ImageInspectWithRaw(ctx, "busybox:latest")
	if err != nil {
		t.Fatalf("ImageInspectWithRaw pulled image error: %v", err)
	}

	removed, err := client.ImageRemove(ctx, "busybox:latest", image.RemoveOptions{})
	if err != nil {
		t.Fatalf("ImageRemove error: %v", err)
	}
	if len(removed) != 1 || removed[0].Deleted != "busybox:latest" {
		t.Fatalf("ImageRemove response = %+v, want one deleted record for busybox:latest", removed)
	}

	_, _, err = client.ImageInspectWithRaw(ctx, "busybox:latest")
	if !cerrdefs.IsNotFound(err) {
		t.Fatalf("ImageInspectWithRaw removed image error = %v, want not found", err)
	}
}

func TestSimulatorClient_ContainerLifecycleAndInspect(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newSimulatorClient()

	resp, err := client.ContainerCreate(ctx, &container.Config{
		Image: "alpine:latest",
		Labels: map[string]string{
			"bedrock.managed-by": "bedrock-dockerd",
		},
	}, nil, nil, nil, "tracer-1")
	if err != nil {
		t.Fatalf("ContainerCreate error: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("ContainerCreate returned empty ID")
	}

	if err := client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		t.Fatalf("ContainerStart error: %v", err)
	}

	inspectRunning, err := client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		t.Fatalf("ContainerInspect running error: %v", err)
	}
	if inspectRunning.ID != resp.ID {
		t.Fatalf("inspect ID = %q, want %q", inspectRunning.ID, resp.ID)
	}
	if inspectRunning.Name != "/tracer-1" {
		t.Fatalf("inspect Name = %q, want %q", inspectRunning.Name, "/tracer-1")
	}
	if inspectRunning.Config == nil || inspectRunning.Config.Image != "alpine:latest" {
		t.Fatalf("inspect Config.Image = %+v, want alpine:latest", inspectRunning.Config)
	}
	if inspectRunning.State == nil || !inspectRunning.State.Running || inspectRunning.State.Status != "running" {
		t.Fatalf("inspect running state = %+v, want running", inspectRunning.State)
	}
	if _, err := time.Parse(time.RFC3339, inspectRunning.Created); err != nil {
		t.Fatalf("inspect Created parse error: %v", err)
	}

	if err := client.ContainerStop(ctx, resp.ID, container.StopOptions{}); err != nil {
		t.Fatalf("ContainerStop error: %v", err)
	}

	inspectStopped, err := client.ContainerInspect(ctx, resp.ID)
	if err != nil {
		t.Fatalf("ContainerInspect stopped error: %v", err)
	}
	if inspectStopped.State == nil || inspectStopped.State.Running || inspectStopped.State.Status != "exited" {
		t.Fatalf("inspect stopped state = %+v, want exited", inspectStopped.State)
	}

	if err := client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
		t.Fatalf("ContainerRemove error: %v", err)
	}

	_, err = client.ContainerInspect(ctx, resp.ID)
	if !cerrdefs.IsNotFound(err) {
		t.Fatalf("ContainerInspect removed container error = %v, want not found", err)
	}
}

func TestSimulatorClient_ContainerListAndFilters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newSimulatorClient()

	create := func(name, managedLabel string) string {
		t.Helper()

		resp, err := client.ContainerCreate(ctx, &container.Config{
			Image: "img:latest",
			Labels: map[string]string{
				"bedrock.managed-by": managedLabel,
			},
		}, nil, nil, nil, name)
		if err != nil {
			t.Fatalf("ContainerCreate(%s) error: %v", name, err)
		}
		return resp.ID
	}

	idA := create("a", "bedrock-dockerd")
	idB := create("b", "bedrock-dockerd")
	idC := create("c", "other")

	if err := client.ContainerStart(ctx, idA, container.StartOptions{}); err != nil {
		t.Fatalf("ContainerStart(%s) error: %v", idA, err)
	}
	if err := client.ContainerStart(ctx, idC, container.StartOptions{}); err != nil {
		t.Fatalf("ContainerStart(%s) error: %v", idC, err)
	}

	runningOnly, err := client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		t.Fatalf("ContainerList running-only error: %v", err)
	}
	if len(runningOnly) != 2 {
		t.Fatalf("running-only count = %d, want 2", len(runningOnly))
	}
	for _, s := range runningOnly {
		if s.State != "running" {
			t.Fatalf("running-only state = %q, want running", s.State)
		}
	}

	managedAll, err := client.ContainerList(ctx, container.ListOptions{
		All: true,
		Filters: filters.NewArgs(
			filters.Arg("label", "bedrock.managed-by=bedrock-dockerd"),
		),
	})
	if err != nil {
		t.Fatalf("ContainerList filtered error: %v", err)
	}
	if len(managedAll) != 2 {
		t.Fatalf("filtered count = %d, want 2", len(managedAll))
	}

	ids := []string{managedAll[0].ID, managedAll[1].ID}
	sort.Strings(ids)
	wantIDs := []string{idA, idB}
	sort.Strings(wantIDs)
	if ids[0] != wantIDs[0] || ids[1] != wantIDs[1] {
		t.Fatalf("filtered IDs = %v, want %v", ids, wantIDs)
	}
}

func TestSimulatorClient_ContainerLogs(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newSimulatorClient()

	resp, err := client.ContainerCreate(ctx, &container.Config{Image: "alpine:latest"}, nil, nil, nil, "loggy")
	if err != nil {
		t.Fatalf("ContainerCreate error: %v", err)
	}

	impl, ok := client.(*simulatorClient)
	if !ok {
		t.Fatal("expected *simulatorClient implementation")
	}

	impl.mu.Lock()
	impl.containers[resp.ID].stdout = "hello stdout\n"
	impl.containers[resp.ID].stderr = "hello stderr\n"
	impl.mu.Unlock()

	reader, err := client.ContainerLogs(ctx, resp.ID, container.LogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		t.Fatalf("ContainerLogs error: %v", err)
	}
	defer reader.Close()

	var out bytes.Buffer
	_, err = stdcopy.StdCopy(&out, &out, reader)
	if err != nil {
		t.Fatalf("StdCopy demux failed: %v", err)
	}

	logs := out.String()
	if !strings.Contains(logs, "hello stdout") {
		t.Fatalf("logs missing stdout, got %q", logs)
	}
	if !strings.Contains(logs, "hello stderr") {
		t.Fatalf("logs missing stderr, got %q", logs)
	}
}

func TestSimulatorClient_NotFoundPaths(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newSimulatorClient()
	missingID := "sim-404"

	if err := client.ContainerStart(ctx, missingID, container.StartOptions{}); !cerrdefs.IsNotFound(err) {
		t.Fatalf("ContainerStart missing error = %v, want not found", err)
	}

	if err := client.ContainerStop(ctx, missingID, container.StopOptions{}); !cerrdefs.IsNotFound(err) {
		t.Fatalf("ContainerStop missing error = %v, want not found", err)
	}

	if err := client.ContainerRemove(ctx, missingID, container.RemoveOptions{}); !cerrdefs.IsNotFound(err) {
		t.Fatalf("ContainerRemove missing error = %v, want not found", err)
	}

	if _, err := client.ContainerLogs(ctx, missingID, container.LogsOptions{}); !cerrdefs.IsNotFound(err) {
		t.Fatalf("ContainerLogs missing error = %v, want not found", err)
	}

	if _, err := client.ContainerInspect(ctx, missingID); !cerrdefs.IsNotFound(err) {
		t.Fatalf("ContainerInspect missing error = %v, want not found", err)
	}

	if _, err := client.ImageRemove(ctx, "missing:image", image.RemoveOptions{}); !cerrdefs.IsNotFound(err) {
		t.Fatalf("ImageRemove missing error = %v, want not found", err)
	}
}

func TestSimulatorClient_ImplementsContainerClient(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	client := newSimulatorClient()

	resp, err := client.ContainerCreate(ctx, &container.Config{Image: "img"}, nil, nil, nil, "ctr")
	if err != nil {
		t.Fatalf("ContainerCreate error: %v", err)
	}
	if resp.ID == "" {
		t.Fatal("ContainerCreate returned empty ID")
	}

	if err := client.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		t.Fatalf("ContainerStart error: %v", err)
	}

	if err := client.ContainerStop(ctx, resp.ID, container.StopOptions{}); err != nil {
		t.Fatalf("ContainerStop error: %v", err)
	}

	if err := client.ContainerRemove(ctx, resp.ID, container.RemoveOptions{}); err != nil {
		t.Fatalf("ContainerRemove error: %v", err)
	}

	_, _, err = client.ImageInspectWithRaw(ctx, "img")
	if err != nil && !errors.Is(err, cerrdefs.ErrNotFound) {
		t.Fatalf("ImageInspectWithRaw unexpected error: %v", err)
	}
}
