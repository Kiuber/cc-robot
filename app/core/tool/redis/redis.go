package redis

import (
	cboot "cc-robot/core/boot"
	"fmt"
	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func Setup() {
	c := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cboot.GV.Config.Infra.RedisConfig.Host, cboot.GV.Config.Infra.RedisConfig.Port),
		Password: cboot.GV.Config.Infra.RedisConfig.Password,
		DB:       0,
	})
	client = c
}

func RdbClient() *redis.Client {
	if client == nil {
		Setup()
	}
	return client
}
