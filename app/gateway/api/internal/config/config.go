// Code scaffolded by goctl. Safe to edit.
// goctl 1.10.1

package config

import (
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	rest.RestConf
	UserRpc          zrpc.RpcClientConf
	VideoRpc         zrpc.RpcClientConf
	InteractionRpc   zrpc.RpcClientConf
	CommunicationRpc zrpc.RpcClientConf
	Auth             AuthConfig
	// Kafka 配置，用于异步发送热度分重算事件（解耦 Gateway 与 video.rpc）。
	Kafka KafkaConfig
}

type AuthConfig struct {
	AccessSecret string
	AccessExpire int64
}

// KafkaConfig Kafka 配置。
type KafkaConfig struct {
	Brokers []string `json:"Brokers"`
	Topic   string   `json:"Topic"`
	// Enable 是否启用 Kafka 热度分重算链路；false 时回退到同步 RPC 调用。
	Enable bool `json:"Enable"`
}
