# 项目目录结构

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
├── docs/ --相关文档说明
│   ├── directory.md
│   └── main.json
├── etc/ -- 配置文件目录
│   └── tiktok-api.yaml
├── internal/ -- 项目内部代码目录
│   ├── config/ -- 配置文件相关
│   │   └── config.go
│   ├── dal/  --数据库操作相关
│   │   ├── repository/  --聚合层，也对相关数据操作进行一层封装，方便后续扩展，也是复杂操作的内聚
│   │   │   ├── comment_baseinfo.go
│   │   │   ├── repository.go
│   │   │   ├── user_baseinfo.go
│   │   │   ├── user_follow.go
│   │   │   ├── video_baseinfo.go
│   │   │   ├── video_liker.go
│   │   │   └── video_popular.go
│   │   ├── tables/  -- 数据库表相关，基础的 crud
│   │   │   ├── comment_baseinfo/
│   │   │   │   └── mysql.go
│   │   │   ├── user_baseinfo/
│   │   │   │   └── mysql.go
│   │   │   ├── user_follow/
│   │   │   │   └── mysql.go
│   │   │   ├── video_baseinfo/
│   │   │   │   └── mysql.go
│   │   │   ├── video_liker/
│   │   │   │   ├── mysql.go
│   │   │   │   └── redis.go
│   │   │   └── video_popular/
│   │   │       ├── mysql.go
│   │   │       └── redis.go
│   │   └── init.go
│   ├── handler/  --入口层，这个层次主要进行数据的解析和请求的处理，方便将纯净的数据传入 logic 层
│   │   ├── communication/
│   │   │   ├── getfanslisthandler.go
│   │   │   ├── getfriendlisthandler.go
│   │   │   ├── getsubscriberlisthandler.go
│   │   │   └── subscribehandler.go
│   │   ├── interaction/
│   │   │   ├── commentvideohandler.go
│   │   │   ├── deletecommenthandler.go
│   │   │   ├── getcommentlisthandler.go
│   │   │   ├── getlikelisthandler.go
│   │   │   └── likevideohandler.go
│   │   ├── user/
│   │   │   ├── getuserinfohandler.go
│   │   │   ├── loginhandler.go
│   │   │   ├── postuserphotohandler.go
│   │   │   ├── registerhandler.go
│   │   │   └── refreshtokenhandler.go
│   │   ├── video/
│   │   │   ├── getvideolisthandler.go
│   │   │   ├── publishvideohandler.go
│   │   │   ├── videopopularhandler.go
│   │   │   └── videosearchhandler.go
│   │   └── routes.go
│   ├── logic/ -- 业务逻辑层，这个层次主要进行相关的业务逻辑的处理，进行复杂的判断等操作
│   │   ├── communication/
│   │   │   ├── getfanslistlogic.go
│   │   │   ├── getfriendlistlogic.go
│   │   │   ├── getsubscriberlistlogic.go
│   │   │   └── subscribelogic.go
│   │   ├── interaction/
│   │   │   ├── commentvideologic.go
│   │   │   ├── deletecommentlogic.go
│   │   │   ├── getcommentlistlogic.go
│   │   │   ├── getlikelistlogic.go
│   │   │   └── likevideologic.go
│   │   ├── user/
│   │   │   ├── getuserinfologic.go
│   │   │   ├── loginlogic.go
│   │   │   ├── postuserphotologic.go
│   │   │   ├── registerlogic.go
│   │   │   └── refreshtokenlogic.go
│   │   ├── video/
│   │   │   ├── getvideolistlogic.go
│   │   │   ├── publishvideologic.go
│   │   │   ├── videopopularlogic.go
│   │   │   └── videosearchlogic.go
│   ├── mw/ --中间件层
│   │   ├── ali/
│   │   │   ├── aliconfig.go
│   │   │   └── aliconfig.yaml
│   │   └── token/
│   │       ├── auth.go
│   │       ├── claims.go
│   │       ├── jwt.go
│   │       └── store.go
│   ├── svc/
│   │   ├── servicecontext.go
│   │   └── xerr/
│   │       └── error.go
│   └── types/
│       └── types.go
├── testdata/
│   ├── images/
│   ├── user1/
│   ├── user2/
│   └── videos/
├── utils/
│   ├── decode.go
│   ├── encryption.go
│   ├── id.go
│   └── time.go
├── Dockerfile
├── compose.yml
├── go.mod
├── go.sum
├── tiktok.go --主函数入口
```
