package video_baseinfo

import (
	"context"
	"database/sql"
	"testing"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newDryRunDB(t *testing.T) *gorm.DB {
	t.Helper()

	sqlDB, err := sql.Open("mysql", "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}
	return db
}

func TestGetVideoByLastTimeParsesDefaultLayout(t *testing.T) {
	db := newDryRunDB(t)

	if _, _, err := GetVideoByLastTime(context.Background(), db, "2006-01-02 15:04:05 ", 1, 10); err != nil {
		t.Fatalf("GetVideoByLastTime unexpected error: %v", err)
	}
}

