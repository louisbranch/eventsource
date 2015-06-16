package eventsource

import (
	"fmt"
	"time"
)

type EventStats struct {
	Start time.Time
	End   time.Time
	Event Event
}

type Stats interface {
	ClientsCount(int)
	EventSent(EventStats)
	EventEnd(EventStats)
}

type StatsJSONLogger struct {
	eventCounter int
}

func (s *StatsJSONLogger) ClientsCount(count int) {
	json := fmt.Sprintf(`{"type": "clients.count", "count": %d}`, count)
	fmt.Println(json)
}

func (s *StatsJSONLogger) EventSent(stats EventStats) {
	size := len(stats.Event.Message)
	duration := stats.End.Sub(stats.Start).Nanoseconds()
	json := fmt.Sprintf(`{"type": "events.sent", "size": %d, "duration": %d}`,
		size, duration)
	fmt.Println(json)
}

func (s *StatsJSONLogger) EventEnd(stats EventStats) {
	s.eventCounter++
	json := fmt.Sprintf(`{"type": "events.count", "count": %d}`, s.eventCounter)
	fmt.Println(json)
}
