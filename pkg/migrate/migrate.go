// Package migrate 提供基于 golang-migrate 的数据库迁移启动器。
//
// 注意：本项目数据库运行在 docker 容器内（compose.infrastructure.yml 中的 mysql:8.0），
// 本地开发通过 127.0.0.1:3309 映射访问的也是容器内的 MySQL。
// 本启动器直接复用服务配置中的 DataSource DSN，因此只会作用于容器数据库。
package migrate

import (
	"errors"
	"fmt"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/zeromicro/go-zero/core/logx"
)

// Run 对 DSN 指向的数据库执行 migrationsDir 目录下所有未应用的迁移。
// dsn 为 go-sql-driver/mysql 标准格式，如
//
//	root:yourpassword@tcp(mysql:3306)/gozero-tiktok?charset=utf8mb4&parseTime=True&loc=Local
//
// 内部会转换为 golang-migrate 需要的 URL 格式。
func Run(dsn, migrationsDir string) error {
	url := toMigrateURL(dsn)
	m, err := migrate.New("file://"+migrationsDir, url)
	if err != nil {
		return fmt.Errorf("migrate.New: %w", err)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			logx.Errorf("migrate close source error: %v", srcErr)
		}
		if dbErr != nil {
			logx.Errorf("migrate close database error: %v", dbErr)
		}
	}()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("migrate.Up: %w", err)
	}
	return nil
}

// toMigrateURL 将 go-sql-driver DSN 转为 golang-migrate 的 mysql:// URL。
// 若已带 mysql:// 前缀则原样返回。
func toMigrateURL(dsn string) string {
	if strings.HasPrefix(dsn, "mysql://") {
		return dsn
	}
	return "mysql://" + dsn
}
