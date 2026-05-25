package myutils

import (
	"database/sql"
	"testing"
	"time"
)

func TestTimeHelpers(t *testing.T) {
	ts := int64(1704067200)
	if got := TsToStr(ts, "2006-01-02"); got != "2024-01-01" {
		t.Fatalf("TsToStr = %q, want %q", got, "2024-01-01")
	}

	tm := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)
	if got := TimeToStr(tm, ""); got != "2024-01-02 03:04:05" {
		t.Fatalf("TimeToStr = %q, want default layout value", got)
	}
	if got := TimeToStr(time.Time{}, ""); got != "" {
		t.Fatalf("zero TimeToStr = %q, want empty", got)
	}

	if got := NullTimeToStr(sql.NullTime{}, ""); got != "" {
		t.Fatalf("invalid NullTimeToStr = %q, want empty", got)
	}
	if got := NullTimeToStr(sql.NullTime{Time: tm, Valid: true}, "2006-01-02"); got != "2024-01-02" {
		t.Fatalf("valid NullTimeToStr = %q, want date", got)
	}

	parsed, err := StrToTime("2024-01-02 03:04:05", "")
	if err != nil {
		t.Fatalf("StrToTime unexpected error: %v", err)
	}
	if !parsed.Equal(tm) {
		t.Fatalf("StrToTime = %v, want %v", parsed, tm)
	}
	if _, err := StrToTime("bad-time", ""); err == nil {
		t.Fatal("expected StrToTime error")
	}

	if got := TimeToNullTime(nil); got.Valid {
		t.Fatal("nil TimeToNullTime should be invalid")
	}
	if got := TimeToNullTime(&tm); !got.Valid || !got.Time.Equal(tm) {
		t.Fatalf("TimeToNullTime = %+v, want valid time", got)
	}

	nt, err := StrToNullTime("2024-01-02", "2006-01-02")
	if err != nil {
		t.Fatalf("StrToNullTime unexpected error: %v", err)
	}
	if !nt.Valid {
		t.Fatal("StrToNullTime should be valid")
	}
	if nt, err := StrToNullTime("", ""); err != nil || nt.Valid {
		t.Fatalf("empty StrToNullTime = %+v, %v; want invalid nil-error", nt, err)
	}
	if _, err := StrToNullTime("bad-time", ""); err == nil {
		t.Fatal("expected StrToNullTime error")
	}

	if NowPtr() == nil {
		t.Fatal("NowPtr returned nil")
	}
}
