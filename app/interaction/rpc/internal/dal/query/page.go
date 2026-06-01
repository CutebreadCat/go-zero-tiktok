package query

import "gorm.io/gorm"

const (
	DefaultPageNum  = 1
	DefaultPageSize = 10
)

// Paginate applies offset/limit based on page number and size.
func Paginate(pageNum, pageSize int) func(*gorm.DB) *gorm.DB {
	if pageNum <= 0 {
		pageNum = DefaultPageNum
	}
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}

	return func(db *gorm.DB) *gorm.DB {
		offset := (pageNum - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}
