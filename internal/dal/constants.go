package dal

import "time"

const (
	mysqlMaxOpenConns    = 100
	mysqlMaxIdleConns    = 10
	mysqlConnMaxLifetime = time.Hour
)
