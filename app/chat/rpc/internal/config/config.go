package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
)

type Config struct {
	zrpc.RpcServerConf
	DataSource string          `json:"DataSource"`
	AppRedis   redis.RedisConf `json:"AppRedis"`
	Kafka      KafkaConfig     `json:"Kafka"`
}

type KafkaConfig struct {
	Brokers []string `json:"Brokers"`
	Topic   string   `json:"Topic"`
	GroupID string   `json:"GroupId"`
}
