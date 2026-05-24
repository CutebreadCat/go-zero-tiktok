package main

import (
	"flag"
	"fmt"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/handler"
	"go_zero-tiktok/internal/middleware/token"
	"go_zero-tiktok/internal/svc"

	"go_zero-tiktok/internal/shared/xerr"
	"net/http"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/rest/httpx"
)

var configFile = flag.String("f", "etc/tiktok-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf, token.WithAuth(c.Auth.AccessSecret))
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		return http.StatusOK, xerr.HandleError(err)
	})

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
