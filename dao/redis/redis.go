package redis

import (
	"context"
	"fmt"
	"gosocial/settings"

	"github.com/go-redis/redis/v8"
)

// 声明一个全局变量
var rdb *redis.Client

// Init 初始化redis连接
func Init(cfg *settings.RedisConfig) (err error) {
	rdb = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password, //  "" :no password set
		DB:       cfg.DB,       //use default DB
		PoolSize: cfg.PoolSize,
	})
	_, err = rdb.Ping(context.Background()).Result()
	return
}

func Close() {
	_ = rdb.Close()
}
