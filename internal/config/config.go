// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
)

type Config struct {
	rest.RestConf
	DataSource string          `json:"DataSource"`
	Redis      redis.RedisConf `json:"Redis"`
	Auth       AuthConfig      `json:"Auth"`
}

type AuthConfig struct {
	AccessSecret string `json:"AccessSecret"`
	AccessExpire int64  `json:"AccessExpire"`
}
