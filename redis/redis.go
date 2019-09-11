package infra

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-redis/redis"
	"github.com/nocai/infra/consul"
	"os"
)

func NewGoRedisClient(l log.Logger) *redis.Client {
	masterName := consul.GetString("redis.sentinel.masterName")
	sentinelAddres := consul.GetStringSlice("redis.sentinel.addrs")
	password := consul.GetString("redis.sentinel.password")
	db := consul.GetInt("redis.sentinel.db")
	poolSize := consul.GetInt("redis.sentinel.poolSize")

	failoverOptions := redis.FailoverOptions{
		MasterName:    masterName,
		SentinelAddrs: sentinelAddres,
		Password:      password,
		DB:            db,
		PoolSize:      poolSize,
	}
	redisClient := redis.NewFailoverClient(&failoverOptions)

	p, err := redisClient.Ping().Result()
	if err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(1)
	}
	_ = level.Info(l).Log("msg", fmt.Sprintf("starting sentinel, Ping: [%v]", p))
	return redisClient
}
