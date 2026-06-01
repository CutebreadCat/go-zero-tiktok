package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string `json:"DataSource"`

	JwtAuth  AuthConfig      `json:"JwtAuth"`
	AppRedis redis.RedisConf `json:"AppRedis"`
}

type AuthConfig struct {
	AccessSecret string `json:"AccessSecret"`
	AccessExpire int64  `json:"AccessExpire"`
}
