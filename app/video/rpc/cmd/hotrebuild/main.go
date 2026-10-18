// Command hotrebuild 从 MySQL 计数离线重建 hot:videos 热度榜。
//
// 复用 video RPC 的领域服务与打分参数：临时键 + Lua 原子 RENAME 发布，
// 重建期间线上读请求始终读旧榜，失败不会替换正式榜。
//
// 示例：
//
//	go run ./app/video/rpc/cmd/hotrebuild -f app/video/rpc/etc/video.yaml -dry-run
//	go run ./app/video/rpc/cmd/hotrebuild -f app/video/rpc/etc/video.yaml -since 168h -keep 10000
package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go_zero-tiktok/app/video/rpc/internal/config"
	videodomain "go_zero-tiktok/app/video/rpc/internal/domain"
	feedpkg "go_zero-tiktok/app/video/rpc/internal/domain/feed"
	"go_zero-tiktok/app/video/rpc/internal/svc"
	appLogger "go_zero-tiktok/pkg/logger"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	configFile = flag.String("f", "app/video/rpc/etc/video.yaml", "配置文件路径（复用 video RPC 的 DataSource/AppRedis/Hot）")
	since      = flag.Duration("since", 0, "候选窗口：只重建最近该时长内发布的视频，0 表示全量重建")
	batchSize  = flag.Int("batch", 500, "单批扫描行数")
	keepTopN   = flag.Int("keep", -1, "榜单保留条数，-1 表示沿用配置 Hot.KeepTopN")
	dryRun     = flag.Bool("dry-run", false, "只统计候选数量，不抢锁、不写 Redis")
	timeout    = flag.Duration("timeout", 0, "整体超时时间，0 表示不限制")
)

// rebuildConfig 重建命令所需的最小配置集。
// 复用 video.yaml 中的 DataSource/AppRedis/Hot，其余字段（RPC/Kafka/QoS 等）被忽略。
type rebuildConfig struct {
	DataSource string           `json:"DataSource"`
	AppRedis   redis.RedisConf  `json:"AppRedis"`
	Hot        config.HotConfig `json:"Hot"`
}

func main() {
	flag.Parse()

	var c rebuildConfig
	conf.MustLoad(*configFile, &c, conf.UseEnv())

	appLogger.Init("video-hot-rebuild", "info")
	defer appLogger.Close()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	if *timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, *timeout)
		defer cancel()
	}

	videoService, err := newVideoService(c)
	if err != nil {
		logx.Errorf("init video service failed: %v", err)
		os.Exit(1)
	}

	opts := videodomain.HotRebuildOptions{
		BatchSize: int32(*batchSize),
		KeepTopN:  resolveKeepTopN(*keepTopN, c.Hot.KeepTopN),
		DryRun:    *dryRun,
	}
	if *since > 0 {
		opts.Since = time.Now().Add(-*since)
	}

	logx.Infof("hot ranking rebuild start: since=%s batch=%d keep_top_n=%d dry_run=%t",
		formatSince(opts.Since), opts.BatchSize, opts.KeepTopN, opts.DryRun)

	scanned, members, err := videoService.RebuildHotRanking(ctx, opts)
	if err != nil {
		logx.Errorf("hot ranking rebuild failed: %v", err)
		os.Exit(1)
	}

	if opts.DryRun {
		fmt.Printf("dry-run: %d candidate videos in window %s\n", scanned, formatSince(opts.Since))
		return
	}
	fmt.Printf("rebuild done: scanned=%d members=%d keep_top_n=%d\n", scanned, members, opts.KeepTopN)
}

// newVideoService 构造仅用于离线重建的领域服务。
// 不启动 Kafka/worker，也不依赖对象存储，避免命令进程常驻后台任务。
func newVideoService(c rebuildConfig) (*videodomain.VideoService, error) {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("open mysql: %w", err)
	}
	rdb := redis.MustNewRedis(c.AppRedis)

	repos, err := svc.NewRepositories(db, rdb)
	if err != nil {
		return nil, fmt.Errorf("init repositories: %w", err)
	}

	return videodomain.NewVideoService(
		repos.Video,
		repos.VideoStat,
		repos.VideoQoS,
		nil,
		repos.Feed,
		nil,
		svc.NewHotScoreCalculator(c.Hot),
		nil,
		feedpkg.DefaultRecommendConfig(),
	), nil
}

// resolveKeepTopN 合并命令行与配置：命令行 -1 表示跟随配置。
func resolveKeepTopN(flagValue, configValue int) int {
	if flagValue >= 0 {
		return flagValue
	}
	return configValue
}

func formatSince(t time.Time) string {
	if t.IsZero() {
		return "all"
	}
	return t.Format(time.RFC3339)
}
