package events

import (
	"sync"
)

// EventType represents an event type.
type EventType string

const (
	EventWebsiteCreated EventType = "website.created"
	EventWebsiteDeleted EventType = "website.deleted"
	EventAlertFired     EventType = "alert.fired"
	EventMetricsUpdate  EventType = "metrics.update"
	EventTaskProgress   EventType = "task.progress"
	EventTaskComplete   EventType = "task.complete"
)

// Event represents a broadcast event.
type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data"`
}

// Bus is a simple pub-sub using Go channels.
type Bus struct {
	subscribers map[EventType][]chan Event
	mu          sync.RWMutex
}

// NewBus creates a new event bus.
func NewBus() *Bus {
	return &Bus{subscribers: make(map[EventType][]chan Event)}
}

// Subscribe returns a channel that receives events of the given type.
func (b *Bus) Subscribe(et EventType) chan Event {
	ch := make(chan Event, 10)
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[et] = append(b.subscribers[et], ch)
	return ch
}

// Unsubscribe removes a channel from the subscribers.
func (b *Bus) Unsubscribe(et EventType, ch chan Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	subs := b.subscribers[et]
	for i, s := range subs {
		if s == ch {
			b.subscribers[et] = append(subs[:i], subs[i+1:]...)
			close(ch)
			break
		}
	}
}

// Publish sends an event to all subscribers.
func (b *Bus) Publish(ev Event) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	for _, ch := range b.subscribers[ev.Type] {
		select {
		case ch <- ev:
		default:
		}
	}
}
