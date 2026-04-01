package scheduler

// Scheduler components helps API to pick DockerD instances for sessions.
type Scheduler interface {
	// Append a new DockerD instance to the scheduler.
	Append(string)
	// Exists checks if a DockerD instance is registered in the scheduler.
	Exists(string) bool
	// Pick returns a DockerD instance, it returns error if fails.
	Pick() (string, error)
}
