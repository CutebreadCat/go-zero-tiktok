package query

import (
	"go_zero-tiktok/internal/dal/page"

	"gorm.io/gorm"
)

// Paginate applies offset/limit based on page number and size.
func Paginate(pageNum, pageSize int) func(*gorm.DB) *gorm.DB {
	if pageNum <= 0 {
		pageNum = page.DefaultPageNum
	}
	if pageSize <= 0 {
		pageSize = page.DefaultPageSize
	}

	return func(db *gorm.DB) *gorm.DB {
		offset := (pageNum - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
