// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"flag"
	"fmt"

	appLogger "go_zero-tiktok/Prometheus/logger"
	"go_zero-tiktok/app/gateway/api/internal/config"
	"go_zero-tiktok/app/gateway/api/internal/handler"
	token "go_zero-tiktok/app/gateway/api/internal/middleware/token"
	"go_zero-tiktok/app/gateway/api/internal/svc"
	"go_zero-tiktok/pkg/xerr"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/tiktok-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	level := "info"
	if c.Mode == "dev" || c.Mode == "test" {
		level = "debug"
	}
	appLogger.Init("gateway-api", level)
	appLogger.RegisterOTelTraceExtractor()
	defer appLogger.Close()

	server := rest.MustNewServer(c.RestConf, token.WithAuth(c.Auth.AccessSecret))
	appLogger.RegisterLogxBridge()
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		return 200, xerr.HandleError(err)
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
