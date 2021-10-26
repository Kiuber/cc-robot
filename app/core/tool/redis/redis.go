package redis

import (
	"cc-robot/model"
	"fmt"
	"github.com/go-redis/redis/v8"
)

var client *redis.Client

func Setup(ctx model.Context) {
	c := redis.NewClient(&redis.Options {
		Addr:	  fmt.Sprintf("%s:%s", ctx.Config.Infra.RedisConfig.Host, ctx.Config.Infra.RedisConfig.Port),
		Password: ctx.Config.Infra.RedisConfig.Password,
		DB:		  0,
	})
	client = c
}

func RdbClient(ctx model.Context) *redis.Client {
	if client == nil {
		Setup(ctx)
	}
	return client
}
