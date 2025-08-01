package rdis

import (
	"context"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/mangohow/cloud-ide/cmd/webserver/internal/conf"
)

var client *redis.Client

func RedisInstance() *redis.Client {
	return client
}

func InitRedis() error {
	client = redis.NewClient(&redis.Options{
		Addr:         conf.RedisConfig.Addr,
		Password:     conf.RedisConfig.Password,
		DB:           int(conf.RedisConfig.DB),
		PoolSize:     int(conf.RedisConfig.PoolSize),     // 连接池最大socket连接数
		MinIdleConns: int(conf.RedisConfig.MinIdleConns), // 最少连接维持数
	})

	timeoutCtx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	_, err := client.Ping(timeoutCtx).Result()
	if err != nil {
		return err
	}

	return nil
}

func CloseRedisConn() {
	client.Close()
}
