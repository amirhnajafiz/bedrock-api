package workers

import (
	"context"
	"time"

	"github.com/amirhnajafiz/bedrock-api/internal/scheduler"
)

// WorkerDockerDHealthCheck continuously checks the health status of Docker daemons by listening to an input channel
// for updates and using a ticker to periodically remove stale entries from the health map.
func WorkerDockerDHealthCheck(ctx context.Context, input chan string, interval time.Duration) {
	// get a reference to the scheduler instance
	scheduler := scheduler.NewRoundRobin()

	// healthMap keeps track of the last time a health update was received for each Docker daemon
	healthMap := make(map[string]time.Time)

	// ticker is used to periodically check for stale entries in the healthMap
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case dockerd := <-input:
			// update the healthMap with the current time for the received Docker daemon
			healthMap[dockerd] = time.Now()
			scheduler.Append(dockerd)
		case <-ticker.C:
			timeSnapshot := time.Now()

			// loop through the healthMap and remove any entries that haven't been updated within the interval
			for dockerd, lastUpdated := range healthMap {
				if timeSnapshot.Sub(lastUpdated) > interval {
					delete(healthMap, dockerd)
					scheduler.Drop(dockerd)
				}
			}
		}
	}
}
