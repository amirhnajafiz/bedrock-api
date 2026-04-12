package xerrors

import "errors"

var (
	// ContainerManagerErrNotFound is returned when a container is not found in the container manager.
	ContainerManagerErrNotFound = errors.New("container not found")
)
