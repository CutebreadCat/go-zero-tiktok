package event

import (
	"encoding/json"
	"testing"

	"go_zero-tiktok/pkg/kafka"
)

func TestImpressionEvent_ToKafkaEvent(t *testing.T) {
	e := ImpressionEvent{
		TrackingEventBase: TrackingEventBase{
			Timestamp: 1234567890000,
			UserID:    100,
			DeviceID:  "device-1",
			ClientIP:  "127.0.0.1",
		},
		VideoID:   200,
		Scene:     "recommend",
		RequestID: "req-1",
		Position:  3,
	}

	ev := e.ToKafkaEvent("tracking-events")
	if ev.Type != ImpressionType {
		t.Fatalf("type = %s, want %s", ev.Type, ImpressionType)
	}
	if ev.Msg.Topic != "tracking-events" {
		t.Fatalf("topic = %s, want tracking-events", ev.Msg.Topic)
	}

	// 验证反序列化
	payload, _ := json.Marshal(ev)
	var parsed kafka.Event
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if parsed.Type != ImpressionType {
		t.Fatalf("parsed type = %s, want %s", parsed.Type, ImpressionType)
	}
	parsedImpression, ok := parsed.Data.(*ImpressionEvent)
	if !ok {
		t.Fatalf("parsed data type = %T, want *ImpressionEvent", parsed.Data)
	}
	if parsedImpression.UserID != 100 || parsedImpression.VideoID != 200 {
		t.Fatalf("parsed data mismatch: %+v", parsedImpression)
	}
}

func TestFollowEvent_ToKafkaEvent(t *testing.T) {
	e := FollowEvent{
		TrackingEventBase: TrackingEventBase{
			Timestamp: 1234567890000,
			UserID:    100,
		},
		TargetUserID: 200,
		Action:       "follow",
	}

	ev := e.ToKafkaEvent("tracking-events")
	if ev.Type != FollowType {
		t.Fatalf("type = %s, want %s", ev.Type, FollowType)
	}

	payload, _ := json.Marshal(ev)
	var parsed kafka.Event
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	parsedFollow, ok := parsed.Data.(*FollowEvent)
	if !ok {
		t.Fatalf("parsed data type = %T, want *FollowEvent", parsed.Data)
	}
	if parsedFollow.TargetUserID != 200 || parsedFollow.Action != "follow" {
		t.Fatalf("parsed data mismatch: %+v", parsedFollow)
	}
}
