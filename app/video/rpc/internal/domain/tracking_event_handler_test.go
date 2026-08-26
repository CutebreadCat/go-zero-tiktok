package domain

import (
	"context"
	"testing"

	"go_zero-tiktok/pkg/event"
	"go_zero-tiktok/pkg/kafka"
)

type mockFeedSeenRepo struct {
	seen map[int64][]int64
}

func (m *mockFeedSeenRepo) MarkSeen(ctx context.Context, userID int64, videoIDs []int64) error {
	m.seen[userID] = append(m.seen[userID], videoIDs...)
	return nil
}

type mockVideoViewEventRepo struct {
	events []*eventRecord
}

type eventRecord struct {
	userID    int64
	videoID   int64
	scene     string
	requestID string
	eventType string
	watchMs   int64
	completed int8
}

func (m *mockVideoViewEventRepo) CreateEvent(ctx context.Context, userID, videoID int64, scene, requestID, eventType string, watchMs int64, completed int8) error {
	m.events = append(m.events, &eventRecord{
		userID:    userID,
		videoID:   videoID,
		scene:     scene,
		requestID: requestID,
		eventType: eventType,
		watchMs:   watchMs,
		completed: completed,
	})
	return nil
}

func TestTrackingEventHandler_Impression(t *testing.T) {
	seenRepo := &mockFeedSeenRepo{seen: make(map[int64][]int64)}
	viewRepo := &mockVideoViewEventRepo{}
	handler := NewTrackingEventHandler(seenRepo, viewRepo, nil)

	ev := event.ImpressionEvent{
		TrackingEventBase: event.TrackingEventBase{UserID: 1},
		VideoID:           100,
		Scene:             "recommend",
		RequestID:         "req-1",
		Position:          2,
	}
	if err := handler.Consume(context.Background(), ev.ToKafkaEvent("tracking-events")); err != nil {
		t.Fatalf("consume failed: %v", err)
	}

	if len(seenRepo.seen[1]) != 1 || seenRepo.seen[1][0] != 100 {
		t.Errorf("expected video 100 marked seen for user 1, got %v", seenRepo.seen[1])
	}
	if len(viewRepo.events) != 1 || viewRepo.events[0].eventType != "exposed" {
		t.Errorf("expected one exposed event, got %+v", viewRepo.events)
	}
}

func TestTrackingEventHandler_PlayAndComplete(t *testing.T) {
	viewRepo := &mockVideoViewEventRepo{}
	handler := NewTrackingEventHandler(nil, viewRepo, nil)

	playEv := event.PlayEvent{
		TrackingEventBase: event.TrackingEventBase{UserID: 1},
		VideoID:           200,
		Scene:             "recommend",
		DurationMs:        15000,
	}
	if err := handler.Consume(context.Background(), playEv.ToKafkaEvent("tracking-events")); err != nil {
		t.Fatalf("consume play failed: %v", err)
	}

	completeEv := event.CompleteEvent{
		TrackingEventBase: event.TrackingEventBase{UserID: 1},
		VideoID:           200,
		WatchMs:           15000,
		DurationMs:        15000,
	}
	if err := handler.Consume(context.Background(), completeEv.ToKafkaEvent("tracking-events")); err != nil {
		t.Fatalf("consume complete failed: %v", err)
	}

	if len(viewRepo.events) != 2 {
		t.Fatalf("expected 2 events, got %d", len(viewRepo.events))
	}
	if viewRepo.events[0].eventType != "play" {
		t.Errorf("first event type = %s, want play", viewRepo.events[0].eventType)
	}
	if viewRepo.events[1].eventType != "complete" || viewRepo.events[1].completed != 1 {
		t.Errorf("complete event mismatch: %+v", viewRepo.events[1])
	}
}

func TestTrackingEventHandler_UnknownType(t *testing.T) {
	handler := NewTrackingEventHandler(nil, nil, nil)
	if err := handler.Consume(context.Background(), &kafka.Event{Type: "Unknown"}); err != nil {
		t.Fatalf("unknown type should be ignored, got err: %v", err)
	}
}
