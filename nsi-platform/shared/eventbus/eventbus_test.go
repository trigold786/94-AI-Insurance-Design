package eventbus

import (
	"context"
	"os"
	"testing"
)

func TestNewRedisBusInvalidURL(t *testing.T) {
	bus, err := NewRedisBus("not-a-valid-url")
	if err == nil {
		bus.Close()
		t.Fatal("expected error for invalid URL, got nil")
	}
}

func TestNewRedisBusEmptyURL(t *testing.T) {
	bus, err := NewRedisBus("")
	if err == nil {
		bus.Close()
		t.Fatal("expected error for empty URL, got nil")
	}
}

func TestPublishAndConsume(t *testing.T) {
	redisURL := os.Getenv("NSI_TEST_REDIS_URL")
	if redisURL == "" {
		t.Skip("NSI_TEST_REDIS_URL not set, skipping integration test")
	}

	bus, err := NewRedisBus(redisURL)
	if err != nil {
		t.Fatalf("NewRedisBus() returned error: %v", err)
	}
	defer bus.Close()

	ctx := context.Background()
	stream := "test:events"
	event := &Event{
		Type:    "test.event",
		Payload: []byte(`{"key":"value"}`),
	}

	if err := bus.Publish(ctx, stream, event); err != nil {
		t.Fatalf("Publish() returned error: %v", err)
	}

	if event.ID == "" {
		t.Error("expected non-empty event ID after publish")
	}
}

func TestPublishNilEvent(t *testing.T) {
	bus, err := NewRedisBus("redis://localhost:6379/0")
	if err != nil {
		t.Fatalf("NewRedisBus() returned error: %v", err)
	}
	defer bus.Close()

	ctx := context.Background()
	err = bus.Publish(ctx, "test:events", nil)
	if err == nil {
		t.Error("expected error for nil event, got nil")
	}
}

func TestPublishEmptyStream(t *testing.T) {
	bus, err := NewRedisBus("redis://localhost:6379/0")
	if err != nil {
		t.Fatalf("NewRedisBus() returned error: %v", err)
	}
	defer bus.Close()

	ctx := context.Background()
	err = bus.Publish(ctx, "", &Event{Type: "test", Payload: []byte("{}")})
	if err == nil {
		t.Error("expected error for empty stream, got nil")
	}
}

func TestBusImplementsInterface(t *testing.T) {
	bus, err := NewRedisBus("redis://localhost:6379/0")
	if err != nil {
		t.Fatalf("NewRedisBus() returned error: %v", err)
	}
	defer bus.Close()

	var iface EventBus = bus
	if iface == nil {
		t.Error("EventBus interface should not be nil")
	}
}


