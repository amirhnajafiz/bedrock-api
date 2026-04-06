//go:build e2e
// +build e2e

package e2e

import (
	"testing"

	"github.com/amirhnajafiz/bedrock-api/internal/components/containers"
)

// TestContainers tests the container management functionality of the DockerD.
// It creates a container, checks if it's running, stops it, checks if it's stopped, stores the logs, and finally removes the container.
// Note: This test requires Docker to be installed and running on the host machine.
func TestContainers(t *testing.T) {
	// create a context
	ctx := t.Context()

	// create a container manager
	cm, err := containers.NewDockerManager()
	if err != nil {
		t.Fatalf("failed to create container manager: %v", err)
	}

	// create an nginx container
	containerID, err := cm.Start(ctx, &containers.ContainerConfig{
		Name:  "nginx-container",
		Image: "nginx:latest",
	})
	if err != nil {
		t.Fatalf("failed to create container: %v", err)
	}

	// list containers and check if the nginx container is running
	containersList, err := cm.List(ctx)
	if err != nil {
		t.Fatalf("failed to list containers: %v", err)
	}

	found := false
	for _, c := range containersList {
		if c.ID == containerID {
			t.Logf("nginx container is running with ID: %s", containerID)
			found = true
			break
		}
	}
	if !found {
		t.Errorf("nginx container not found")
	}

	// stop the nginx container
	err = cm.Stop(ctx, containerID)
	if err != nil {
		t.Fatalf("failed to stop container: %v", err)
	}

	// list containers again and check if the nginx container is stopped
	containersList, err = cm.List(ctx)
	if err != nil {
		t.Fatalf("failed to list containers: %v", err)
	}

	for _, c := range containersList {
		if c.ID == containerID {
			t.Errorf("nginx container is still running with ID: %s", containerID)
		}
	}
	t.Logf("nginx container stopped successfully")

	// store the container logs
	err = cm.StoreLogs(ctx, containerID, "nginx-logs.txt")
	if err != nil {
		t.Fatalf("failed to get container logs: %v", err)
	}

	// remove the nginx container
	err = cm.Remove(ctx, containerID)
	if err != nil {
		t.Fatalf("failed to remove container: %v", err)
	}

	// list containers again and check if the nginx container is removed
	containersList, err = cm.List(ctx)
	if err != nil {
		t.Fatalf("failed to list containers: %v", err)
	}

	for _, c := range containersList {
		if c.ID == containerID {
			t.Errorf("nginx container is still present with ID: %s", containerID)
		}
	}
	t.Logf("nginx container removed successfully")
}
