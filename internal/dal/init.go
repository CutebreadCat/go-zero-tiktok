// internal/dal/dal.go
package dal

import (
	"log"
	"time"

	"context"

	"go_zero-tiktok/internal/types"

	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var Db *gorm.DB
var Rdb *redis.Redis

// InitMysql 初始化 GORM 数据库连接
// dsn: 数据库连接串，从 config 层获取
func InitMysql(dsn string) error {
	var err error
	// 1. 建立 GORM 连接
	Db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		return err
	}

	// 2. 获取底层的 *sql.DB 对象，用于配置连接池
	// 这一步非常重要，可以优化数据库连接性能
	sqlDB, err := Db.DB()
	if err != nil {
		return err
	}

	// 3. 配置连接池参数
	sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
	sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
	sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大存活时间

	// 4. 测试连接是否成功
	if err = sqlDB.Ping(); err != nil {
		log.Printf("数据库连接测试失败: %v", err)
		return err
	}

	log.Println("数据库连接成功！")
	return nil
}

// conf 是你从 config.yaml 读取出来的配置结构
func InitRedis(conf redis.RedisConf) {
	// 1. 直接使用传入的 conf，不要重新定义
	// go-zero 的 MustNewRedis 内部会自动处理 Host, Pass, Type 等配置
	rdb := redis.MustNewRedis(conf)

	// 2. 赋值给全局变量
	Rdb = rdb

	// 3. 使用 PingCtx 进行健康检查
	// 注意：go-zero 的 PingCtx 需要传入 context
	ctx := context.Background()
	if err := rdb.PingCtx(ctx); err != true {
		log.Fatalf("❌ 无法连接到 Redis (%s): %v", conf.Host, err)
	}

	log.Println("✅ 成功连接到 Redis！")
}

func InitTables() error {
	// 这里可以调用 AutoMigrate 来自动创建表
	// 注意：AutoMigrate 只能创建表，不能修改已存在的表结构
	err := Db.AutoMigrate(
		&types.UserBaseinfo{},
		&types.VideoBaseinfo{},
		&types.VideoPopular{},
		&types.CommentBaseinfo{},
		&types.VideoLiker{},
		&types.UserFollow{},
	)
	if err != nil {
		log.Printf("数据表初始化失败: %v", err)
		return err
	}

	log.Println("✅ 数据表初始化成功！")
	return nil
}
