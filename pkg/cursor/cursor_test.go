package cursor

import (
	"encoding/base64"
	"encoding/json"
	"testing"
)

type testCursor struct {
	PublishedAt int64 `json:"published_at"`
	VideoID     int64 `json:"video_id"`
}

func TestEncodeDecode(t *testing.T) {
	original := &testCursor{PublishedAt: 1723891200000, VideoID: 42}

	encoded, err := Encode(original)
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if encoded == "" {
		t.Fatal("encoded cursor is empty")
	}

	var decoded testCursor
	if err := Decode(encoded, &decoded); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if decoded.PublishedAt != original.PublishedAt || decoded.VideoID != original.VideoID {
		t.Fatalf("decoded mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestDecode_EmptyCursor(t *testing.T) {
	var decoded testCursor
	err := Decode("", &decoded)
	if err == nil {
		t.Fatal("expected error for empty cursor")
	}
}

func TestDecode_InvalidCursor(t *testing.T) {
	var decoded testCursor
	err := Decode("not-a-valid-cursor", &decoded)
	if err == nil {
		t.Fatal("expected error for invalid cursor")
	}
}

func TestDecode_StdEncodingCompatible(t *testing.T) {
	// 模拟用 StdEncoding 编码的旧 cursor，Decode 应能兼容解析。
	original := &testCursor{PublishedAt: 1723891200000, VideoID: 42}
	content, _ := json.Marshal(original)
	encoded := base64.StdEncoding.EncodeToString(content)

	var decoded testCursor
	if err := Decode(encoded, &decoded); err != nil {
		t.Fatalf("decode std encoding failed: %v", err)
	}

	if decoded.PublishedAt != original.PublishedAt || decoded.VideoID != original.VideoID {
		t.Fatalf("decoded mismatch: got %+v, want %+v", decoded, original)
	}
}

func TestEncode_NilPayload(t *testing.T) {
	_, err := Encode(nil)
	if err == nil {
		t.Fatal("expected error for nil payload")
	}
}
