package events

import (
	"testing"
	"time"
)

func TestBus_PublishSubscribe(t *testing.T) {
	bus := NewBus()
	ch := bus.Subscribe(EventMetricsUpdate)
	defer bus.Unsubscribe(EventMetricsUpdate, ch)

	ev := Event{Type: EventMetricsUpdate, Data: map[string]int{"cpu": 50}}
	bus.Publish(ev)

	select {
	case received := <-ch:
		if received.Type != EventMetricsUpdate {
			t.Errorf("expected type %s, got %s", EventMetricsUpdate, received.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for event")
	}
}

func TestBus_Unsubscribe(t *testing.T) {
	bus := NewBus()
	ch := bus.Subscribe(EventMetricsUpdate)
	bus.Unsubscribe(EventMetricsUpdate, ch)

	// After unsubscribe, channel should be closed
	_, ok := <-ch
	if ok {
		t.Fatal("expected channel to be closed after unsubscribe")
	}
}
