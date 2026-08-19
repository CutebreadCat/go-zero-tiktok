package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string `json:"DataSource"`

	// AppRedis 应用 Redis，用于 Feed 候选池（feed:global ZSet）。
	AppRedis redis.RedisConf `json:"AppRedis"`
}
