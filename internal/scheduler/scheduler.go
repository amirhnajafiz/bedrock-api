package scheduler

type Scheduler interface {
	Append(string)
	Pick() (string, error)
}
