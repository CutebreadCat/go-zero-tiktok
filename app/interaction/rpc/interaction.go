package main

import (
	"flag"
	"fmt"

	appLogger "go_zero-tiktok/pkg/logger"
	"go_zero-tiktok/app/interaction/rpc/interaction_pb"
	"go_zero-tiktok/app/interaction/rpc/internal/config"
	"go_zero-tiktok/app/interaction/rpc/internal/server"
	"go_zero-tiktok/app/interaction/rpc/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/interaction.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c, conf.UseEnv())
	level := "info"
	if c.Mode == service.DevMode || c.Mode == service.TestMode {
		level = "debug"
	}
	appLogger.Init("interaction-rpc", level)
	appLogger.RegisterOTelTraceExtractor()
	defer appLogger.Close()
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		interaction_pb.RegisterInteractionServiceServer(grpcServer, server.NewInteractionServiceServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	appLogger.RegisterLogxBridge()
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
