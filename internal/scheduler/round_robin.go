package scheduler

import (
	"errors"
	"slices"
	"sync"
)

var ErrEmpty = errors.New("scheduler is empty")

type RoundRobinScheduler struct {
	mu    sync.Mutex
	queue []string
}

func NewRoundRobin() Scheduler {
	return &RoundRobinScheduler{
		queue: make([]string, 0),
	}
}

func (r *RoundRobinScheduler) Append(item string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if slices.Contains(r.queue, item) {
		return
	}

	r.queue = append(r.queue, item)
}

func (r *RoundRobinScheduler) Pick() (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.queue) == 0 {
		return "", ErrEmpty
	}

	item := r.queue[0]

	// rotate queue
	r.queue = append(r.queue[1:], item)

	return item, nil
}
