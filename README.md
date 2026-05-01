```
go_zero-tiktok/
├── api/ -- 接口定义目录
│   ├── communication.api
│   ├── interaction.api
│   ├── main.api
│   ├── model.api
│   ├── user.api
│   ├── user_auth.api
│   ├── video.api
│   └── video_auth.api
├── docs/ -- 相关文档说明
│   ├── directory.md
│   └── main.json
├── etc/ -- 配置文件目录
│   └── tiktok-api.yaml
├── internal/ -- 项目内部代码目录
│   ├── config/ -- 配置文件相关
│   │   └── config.go
│   ├── dal/ -- 数据库操作相关
│   │   ├── repository/ -- 聚合层，对相关数据操作进行封装，方便后续扩展
│   │   └── tables/ -- 数据库表相关，基础 CRUD
│   ├── handler/ -- 入口层，负责解析请求并将纯净数据传入 logic 层
│   │   ├── communication/
│   │   ├── interaction/
│   │   ├── user/
│   │   └── video/
│   ├── infra/ -- 外部基础设施层
│   │   └── storage/
│   │       └── aliyun/ -- 阿里云 OSS
│   ├── logic/ -- 业务逻辑层
│   │   ├── communication/
│   │   ├── interaction/
│   │   ├── user/
│   │   └── video/
│   ├── middleware/ -- 中间件层
│   │   ├── mfa/ -- MFA 校验
│   │   ├── token/ -- 鉴权认证
│   │   └── useragentmiddleware.go
│   ├── svc/ -- 依赖注入层
│   ├── types/
│   └── utils/ -- 常用工具
├── testdata/
│   ├── images/
│   ├── user1/
│   ├── user2/
│   └── videos/
├── Dockerfile
├── compose.yml
├── go.mod
├── go.sum
└── tiktok.go -- 主函数入口
```
