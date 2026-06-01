// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	DataSource string          `json:"DataSource"`
	Redis      redis.RedisConf `json:"Redis"`

	// RPC Clients
	UserRpc          zrpc.RpcClientConf
	VideoRpc         zrpc.RpcClientConf
	InteractionRpc   zrpc.RpcClientConf
	CommunicationRpc zrpc.RpcClientConf
	ChatRpc          zrpc.RpcClientConf

	Auth  AuthConfig  `json:"Auth"`
	Kafka KafkaConfig `json:"Kafka"`
}

type KafkaConfig struct {
	Brokers []string `json:"Brokers"`
	Topic   string   `json:"Topic"`
	GroupID string   `json:"GroupId"`
}

type AuthConfig struct {
	AccessSecret string `json:"AccessSecret"`
	AccessExpire int64  `json:"AccessExpire"`
}
