package main

import (
	"flag"
	"fmt"

	"go_zero-tiktok/app/video/rpc/internal/config"
	"go_zero-tiktok/app/video/rpc/internal/server"
	"go_zero-tiktok/app/video/rpc/internal/svc"
	"go_zero-tiktok/app/video/rpc/video_pb"
	appLogger "go_zero-tiktok/pkg/logger"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/video.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	level := "info"
	if c.Mode == service.DevMode || c.Mode == service.TestMode {
		level = "debug"
	}
	appLogger.Init("video-rpc", level)
	appLogger.RegisterOTelTraceExtractor()
	defer appLogger.Close()
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		video_pb.RegisterVideoServiceServer(grpcServer, server.NewVideoServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	appLogger.RegisterLogxBridge()
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
