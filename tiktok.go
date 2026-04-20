// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
	"errors"
	"flag"
	"fmt"

	"go_zero-tiktok/internal/config"
	"go_zero-tiktok/internal/handler"
	"go_zero-tiktok/internal/svc"

	"go_zero-tiktok/internal/svc/xerr"
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

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	// 全局错误处理器：业务错误统一返回 200，错误码和消息放在 body 中
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		var codeErr *xerr.CodeError
		if errors.As(err, &codeErr) {
			return http.StatusOK, codeErr
		}

		// 兜底转换为统一业务错误，保证客户端总能收到可读错误信息
		return http.StatusOK, xerr.New(1004, "服务繁忙，请稍后重试")
	})
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
