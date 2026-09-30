package service

import "sync"

// Event is a single SSE message: Name becomes the SSE "event:" field,
// Data becomes the "data:" field (already serialized, e.g. JSON or plain text).
type Event struct {
	Name string
	Data string
}

// EventBroadcaster fans out Events to any number of subscribers (SSE clients).
// Safe for concurrent use.
type EventBroadcaster struct {
	mu   sync.Mutex
	subs map[chan Event]struct{}
}

func NewEventBroadcaster() *EventBroadcaster {
	return &EventBroadcaster{subs: make(map[chan Event]struct{})}
}

// Subscribe registers a new subscriber and returns a channel of Events for
// it, plus an unsubscribe function that must be called when the client
// disconnects.
func (b *EventBroadcaster) Subscribe() (<-chan Event, func()) {
	ch := make(chan Event, 8)
	b.mu.Lock()
	b.subs[ch] = struct{}{}
	b.mu.Unlock()

	unsubscribe := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		if _, ok := b.subs[ch]; ok {
			delete(b.subs, ch)
			close(ch)
		}
	}
	return ch, unsubscribe
}

// Publish sends ev to every current subscriber. Slow/full subscribers are
// skipped rather than blocking the publisher.
func (b *EventBroadcaster) Publish(ev Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ch := range b.subs {
		select {
		case ch <- ev:
		default:
		}
	}
}
