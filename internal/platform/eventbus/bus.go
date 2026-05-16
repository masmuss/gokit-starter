// Package eventbus provides an internal event bus for decoupled communication.
package eventbus

import (
	"context"
	"errors"
	"sync"

	"github.com/google/uuid"
)

// Event represents a domain event.
type Event struct {
	ID       uuid.UUID
	Name     string
	Payload  any
	Metadata map[string]string
}

// Handler defines a function that processes an event.
type Handler func(ctx context.Context, e Event) error

// Bus defines the interface for an event bus.
type Bus interface {
	Publish(ctx context.Context, e Event) error
	Subscribe(eventName string, handler Handler)
}

// InternalBus implements Bus using memory-based dispatching.
type InternalBus struct {
	mu          sync.RWMutex
	subscribers map[string][]Handler
}

// NewInternalBus creates a new in-memory event bus.
func NewInternalBus() *InternalBus {
	return &InternalBus{
		subscribers: make(map[string][]Handler),
	}
}

// Publish dispatches an event to all registered handlers for that event name.
func (b *InternalBus) Publish(ctx context.Context, e Event) error {
	b.mu.RLock()
	handlers, ok := b.subscribers[e.Name]
	b.mu.RUnlock()

	if !ok {
		return nil
	}

	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		errs []error
	)

	for _, h := range handlers {
		wg.Add(1)
		go func(handler Handler) {
			defer wg.Done()
			if err := handler(ctx, e); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
			}
		}(h)
	}

	wg.Wait()

	return errors.Join(errs...)
}

// Subscribe registers a handler for a specific event name.
func (b *InternalBus) Subscribe(eventName string, handler Handler) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.subscribers[eventName] = append(b.subscribers[eventName], handler)
}

// NewEvent creates a new event with a unique ID.
func NewEvent(name string, payload any) Event {
	return Event{
		ID:       uuid.New(),
		Name:     name,
		Payload:  payload,
		Metadata: make(map[string]string),
	}
}
