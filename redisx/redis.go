package redisx

import (
	"fmt"
	"sync"

	"github.com/go-redis/redis/v8"
	"rxcsoft.cn/utils/config"
)

var (
	opts    redis.Options
	redisdb *redis.Client
	hasInit bool = false
	once    sync.Once
)

type (
	// HashResponse 结构体返回
	HashResponse map[string]string
)

//使用单例模式创建redis client
func New() *redis.Client {

	if !hasInit {
		opts.Addr = "localhost:6379"
	}

	once.Do(func() {
		redisdb = redis.NewClient(&opts)
	})
	return redisdb
}

// StartRedis 开启redis的连接
func StartRedis(config config.DB) {
	opts.Addr = fmt.Sprintf("%s:%s", config.Host, config.Port)
	opts.Password = config.Password
	opts.PoolSize = 10
	hasInit = true
}
