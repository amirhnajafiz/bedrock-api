package bdtracer

import (
	"fmt"
	"os"
)

// BdTracer represents the configuration for the Bedrock Tracer.
type BdTracer struct {
	BaseDir string
	Image   string
}

// CreateTracerOutputDir creates the output directory for the tracer if it doesn't exist.
func (b *BdTracer) CreateTracerOutputDir(sessionId string) error {
	outputDir := fmt.Sprintf("%s/%s", b.BaseDir, sessionId)
	return os.MkdirAll(outputDir, 0755)
}

// DefaultContainerFlags returns the default flags for the tracer container.
func (b *BdTracer) DefaultContainerFlags() map[string]any {
	return map[string]any{
		"pid":        "host",
		"privileged": true,
	}
}

// DefaultTracerVolumes returns the default volume mappings for the tracer container.
func (b *BdTracer) DefaultTracerVolumes(sessionId string) map[string]string {
	return map[string]string{
		"/sys":                      "/sys:rw",
		"/lib/modules":              "/lib/modules:ro",
		"/var/run/docker.sock":      "/var/run/docker.sock",
		b.BaseDir + "/" + sessionId: "/logs",
	}
}

// DefaultTracerCommand returns the default command to run the tracer container.
func (b *BdTracer) DefaultTracerCommand(targetContainerName string) []string {
	return []string{
		"bdtrace",
		"--container",
		targetContainerName,
		"-o",
		"/logs",
	}
}
