package scheduler

import (
	"slices"
	"sync"
)

var (
	// roundRobinSchedulerInstance is a singleton instance of RoundRobinScheduler.
	roundRobinSchedulerInstance Scheduler
)

// RoundRobinScheduler selects an instance using RoundRobin algorithm.
type RoundRobinScheduler struct {
	mu    sync.Mutex
	queue []string
}

// NewRoundRobin returns a singleton instance of RoundRobinScheduler.
func NewRoundRobin() Scheduler {
	if roundRobinSchedulerInstance == nil {
		roundRobinSchedulerInstance = &RoundRobinScheduler{
			queue: make([]string, 0),
		}
	}

	return roundRobinSchedulerInstance
}

func (r *RoundRobinScheduler) Append(item string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// avoid duplicate append
	if slices.Contains(r.queue, item) {
		return
	}

	r.queue = append(r.queue, item)
}

func (r *RoundRobinScheduler) Drop(item string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// remove item from queue
	for i, v := range r.queue {
		if v == item {
			r.queue = append(r.queue[:i], r.queue[i+1:]...)
			return
		}
	}
}

func (r *RoundRobinScheduler) Exists(item string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return slices.Contains(r.queue, item)
}

func (r *RoundRobinScheduler) Pick() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// error if the queue is empty
	if len(r.queue) == 0 {
		return "", ErrEmpty
	}

	item := r.queue[0]

	// rotate queue
	r.queue = append(r.queue[1:], item)

	return item, nil
}
