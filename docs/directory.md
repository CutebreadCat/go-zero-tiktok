# 项目目录结构（仅文件夹，最多三层；已省略 .gocache）

```
go_zero-tiktok/
├── .github/ -- GitHub 配置
│   └── workflows/ -- CI 工作流
├── api/ -- gozero的api定义
│   ├── chat/
│   ├── communication/
│   ├── interaction/
│   ├── user/
│   └── video/
├── docs/ -- 文档与说明
├── etc/ -- 配置文件
├── internal/ -- 内部业务代码
│   ├── config/ -- 配置加载
│   ├── dal/ -- 数据访问层
│   │   ├── repository/ -- 聚合仓储与封装
│   │   └── tables/ -- 表结构与基础 CRUD
│   ├── domain/ -- 领域模型
│   │   ├── chat/ -- 聊天领域
│   │   ├── comment/ -- 评论领域
│   │   ├── user/ -- 用户与关注领域
│   │   ├── video/ -- 视频领域
│   │   └── websocket/ -- WebSocket 领域
│   ├── handler/  参数校验层
│   │   ├── chat/
│   │   ├── communication/
│   │   ├── interaction/
│   │   ├── user/
│   │   └── video/
│   ├── infra/ -- 外部依赖与基础设施
│   │   ├── ai/ -- AI 相关能力
│   │   ├── cache/ -- 缓存实现
│   │   ├── cronjob/ -- 定时任务(未来想做)
│   │   ├── mq/ -- 消息队列
│   │   └── storage/ -- 对象存储
│   ├── logic/  逻辑层处理函数
│   │   ├── chat/
│   │   ├── communication/
│   │   ├── interaction/
│   │   ├── user/
│   │   └── video/
│   ├── middleware/ -- 中间件
│   │   ├── government/ -- 治理相关
│   │   ├── mfa/ -- 多因素认证
│   │   └── token/ -- Token 认证
│   ├── shared/ -- 契约层
│   │   ├── ctxkey/ -- 上下文键
│   │   ├── mq/ -- MQ 公共定义
│   │   └── xerr/ -- 错误码
│   ├── svc/ -- 依赖注入与上下文
│   │   └── mock/ -- Mock 实现
│   ├── types/ -- 通用类型
│   └── utils/ -- 工具函数
├── sql/ -- 数据库sql
├── .dockerignoer
├── .editorconfig
├── .env
├── .gitattributes
├── .gitignore
├── .golangci.yaml
├── compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── tiktok.go  主程序入口

```
