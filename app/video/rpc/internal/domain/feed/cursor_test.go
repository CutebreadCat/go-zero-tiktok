package feed

import (
	"testing"
)

func TestEncodeDecodeRecommendCursor(t *testing.T) {
	tests := []struct {
		name    string
		cursor  *RecommendCursor
		wantNil bool
	}{
		{
			name: "正常游标",
			cursor: &RecommendCursor{
				Score:   12345,
				VideoID: 67890,
			},
			wantNil: false,
		},
		{
			name:    "nil 游标返回空字符串",
			cursor:  nil,
			wantNil: true,
		},
		{
			name: "无效 video_id 返回空字符串",
			cursor: &RecommendCursor{
				Score:   12345,
				VideoID: 0,
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded := EncodeRecommendCursor(tt.cursor)
			if tt.wantNil {
				if encoded != "" {
					t.Errorf("EncodeRecommendCursor() = %q, want empty", encoded)
				}
				return
			}

			decoded, err := DecodeRecommendCursor(encoded)
			if err != nil {
				t.Errorf("DecodeRecommendCursor() error = %v", err)
				return
			}
			if decoded.Score != tt.cursor.Score || decoded.VideoID != tt.cursor.VideoID {
				t.Errorf("DecodeRecommendCursor() = %+v, want %+v", decoded, tt.cursor)
			}
		})
	}
}

func TestDecodeRecommendCursor_Empty(t *testing.T) {
	decoded, err := DecodeRecommendCursor("")
	if err != nil {
		t.Errorf("DecodeRecommendCursor(\"\") error = %v", err)
	}
	if decoded != nil {
		t.Errorf("DecodeRecommendCursor(\"\") = %+v, want nil", decoded)
	}
}

func TestDecodeRecommendCursor_Invalid(t *testing.T) {
	_, err := DecodeRecommendCursor("invalid-cursor")
	if err == nil {
		t.Error("DecodeRecommendCursor(\"invalid-cursor\") expected error")
	}
}
