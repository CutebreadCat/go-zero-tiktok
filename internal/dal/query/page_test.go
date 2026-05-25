package query

import (
	"database/sql"
	"testing"

	"go_zero-tiktok/internal/dal/page"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type paginateTestModel struct {
	ID string
}

func TestPaginate(t *testing.T) {
	sqlDB, err := sql.Open("mysql", "gorm:gorm@tcp(localhost:9910)/gorm?charset=utf8&parseTime=True&loc=Local")
	if err != nil {
		t.Fatalf("open sql db: %v", err)
	}
	defer sqlDB.Close()

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DryRun: true, DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("open dry-run db: %v", err)
	}

	tests := []struct {
		name       string
		pageNum    int
		pageSize   int
		wantOffset int
		wantLimit  int
	}{
		{name: "valid page", pageNum: 2, pageSize: 20, wantOffset: 20, wantLimit: 20},
		{name: "zero page number", pageNum: 0, pageSize: 20, wantOffset: 0, wantLimit: 20},
		{name: "negative page number", pageNum: -1, pageSize: 20, wantOffset: 0, wantLimit: 20},
		{name: "zero page size", pageNum: 2, pageSize: 0, wantOffset: page.DefaultPageSize, wantLimit: page.DefaultPageSize},
		{name: "negative page size", pageNum: 2, pageSize: -5, wantOffset: page.DefaultPageSize, wantLimit: page.DefaultPageSize},
		{name: "zero page and size", pageNum: 0, pageSize: 0, wantOffset: 0, wantLimit: page.DefaultPageSize},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tx := db.Session(&gorm.Session{DryRun: true}).
				Model(&paginateTestModel{}).
				Scopes(Paginate(tt.pageNum, tt.pageSize)).
				Find(&[]paginateTestModel{})

			limitClause, ok := tx.Statement.Clauses["LIMIT"]
			if !ok {
				t.Fatal("expected LIMIT clause")
			}
			limit, ok := limitClause.Expression.(clause.Limit)
			if !ok {
				t.Fatalf("LIMIT clause type = %T, want clause.Limit", limitClause.Expression)
			}
			if limit.Limit == nil {
				t.Fatal("expected limit value")
			}
			if got := *limit.Limit; got != tt.wantLimit {
				t.Errorf("limit = %d, want %d", got, tt.wantLimit)
			}
			if limit.Offset != tt.wantOffset {
				t.Errorf("offset = %d, want %d", limit.Offset, tt.wantOffset)
			}
		})
	}
}
