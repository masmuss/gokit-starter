package eventbus

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestInternalBus_Publish_NoSubscribers(t *testing.T) {
	bus := NewInternalBus()
	err := bus.Publish(context.Background(), NewEvent("test.event", "data"))
	require.NoError(t, err)
}

func TestInternalBus_SubscribeAndPublish(t *testing.T) {
	bus := NewInternalBus()
	ctx := context.Background()

	var received atomic.Value
	received.Store(Event{})

	bus.Subscribe("test.event", func(_ context.Context, e Event) error {
		received.Store(e)
		return nil
	})

	event := NewEvent("test.event", "hello")
	err := bus.Publish(ctx, event)
	require.NoError(t, err)

	got, ok := received.Load().(Event)
	require.True(t, ok)
	require.Equal(t, event.Name, got.Name)
	require.Equal(t, event.Payload, got.Payload)
	require.NotEqual(t, uuid.Nil, got.ID)
}

func TestInternalBus_MultipleSubscribers(t *testing.T) {
	bus := NewInternalBus()
	ctx := context.Background()

	var counter atomic.Int32
	handler := func(_ context.Context, _ Event) error {
		counter.Add(1)
		return nil
	}

	bus.Subscribe("test.event", handler)
	bus.Subscribe("test.event", handler)
	bus.Subscribe("test.event", handler)

	err := bus.Publish(ctx, NewEvent("test.event", nil))
	require.NoError(t, err)
	require.Equal(t, int32(3), counter.Load())
}

func TestInternalBus_EventNameFiltering(t *testing.T) {
	bus := NewInternalBus()
	ctx := context.Background()

	var called atomic.Bool
	bus.Subscribe("event.a", func(_ context.Context, _ Event) error {
		called.Store(true)
		return nil
	})

	err := bus.Publish(ctx, NewEvent("event.b", nil))
	require.NoError(t, err)
	require.False(t, called.Load())
}

func TestInternalBus_HandlerError(t *testing.T) {
	bus := NewInternalBus()
	ctx := context.Background()

	bus.Subscribe("test.event", func(_ context.Context, _ Event) error {
		return errors.New("handler failed")
	})

	err := bus.Publish(ctx, NewEvent("test.event", nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), "handler failed")
}

func TestInternalBus_MultipleHandlerErrors(t *testing.T) {
	bus := NewInternalBus()
	ctx := context.Background()

	bus.Subscribe("test.event", func(_ context.Context, _ Event) error {
		return errors.New("error 1")
	})
	bus.Subscribe("test.event", func(_ context.Context, _ Event) error {
		return errors.New("error 2")
	})

	err := bus.Publish(ctx, NewEvent("test.event", nil))
	require.Error(t, err)
	require.Contains(t, err.Error(), "error 1")
	require.Contains(t, err.Error(), "error 2")
}

func TestInternalBus_NewEventID(t *testing.T) {
	e1 := NewEvent("test", nil)
	e2 := NewEvent("test", nil)
	require.NotEqual(t, e1.ID, e2.ID)
}

func TestInternalBus_ConcurrentSafe(t *testing.T) {
	bus := NewInternalBus()
	ctx := context.Background()

	var counter atomic.Int32
	bus.Subscribe("test.event", func(_ context.Context, _ Event) error {
		counter.Add(1)
		return nil
	})

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = bus.Publish(ctx, NewEvent("test.event", nil))
		}()
	}
	wg.Wait()

	require.Equal(t, int32(10), counter.Load())
}
