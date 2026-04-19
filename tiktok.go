// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package main

import (
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
	// 【核心代码】全局注册错误处理器
	httpx.SetErrorHandler(func(err error) (int, interface{}) {
		// 判断错误类型
		if codeErr, ok := err.(*xerr.CodeError); ok {
			// 如果是我们的自定义业务错误，返回 HTTP 200，但 body 里带业务码
			// 你也可以选择返回 codeErr.Code 作为 HTTP 状态码，看你的需求
			return codeErr.Code, codeErr
		}
		// 如果是其他未知错误（如系统 panic），返回 500
		return http.StatusInternalServerError, nil
	})
	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
