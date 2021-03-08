// Simple Event System
package events

import (
	"sync"
	"time"
)

type Event interface {
	New(args ...interface{}) Event
	Name() string
	Time() time.Time
}

type EventManager struct {
	mu           sync.Mutex
	EventChannel chan Event
}

func NewManager() EventManager {
	return EventManager{
		mu:           sync.Mutex{},
		EventChannel: make(chan Event, 100),
	}
}

func (em *EventManager) Emit(event Event) {
	if len(em.EventChannel)+1 >= cap(em.EventChannel) {
		em.mu.Lock()
		// discord 20 oldest events
		for i := 0; i < 20; i++ {
			_ = <-em.EventChannel
		}
		em.mu.Unlock()
	}
	em.EventChannel <- event
}
