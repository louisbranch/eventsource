package eventsource

import (
	"log"
	"time"
)

type Metrics interface {
	ClientCount(int)
	EventDone(Event, []time.Duration)
}

// A DefaultMetrics implements the Metrics interface and does nothing. Useful
// for disable metrics.
type NoopMetrics struct{}

// The ClientCount function does nothing.
func (NoopMetrics) ClientCount(int) {}

// The EventDone function does nothing.
func (NoopMetrics) EventDone(Event, []time.Duration) {}

// A DefaultMetrics implements the Metrics interface and logs events to the
// stdout.
type DefaultMetrics struct{}

// The ClientCount function does nothing.
func (DefaultMetrics) ClientCount(int) {}

// The EventDone function logs to stdout the avg time an event to be sent to
// clients. Clients with error are ignored.
func (m DefaultMetrics) EventDone(e Event, durations []time.Duration) {
	var sum float64
	var count float64
	var avg float64
	for _, d := range durations {
		if d > 0 {
			sum += float64(d)
			count++
		}
	}
	if count > 0 {
		avg = sum / count
	}
	log.Printf("Event completed - clients %.f, avg time %.2f\n", count, avg)
}
