package database

import (
	"fmt"
	"time"

	"github.com/garyburd/redigo/redis"
	"rxcsoft.cn/utils/config"
	"rxcsoft.cn/utils/logger"
)

var (
	log = logger.New()
	// RedisPool redis连接池
	RedisPool *redis.Pool
)

type (
	// HashResponse 结构体返回
	HashResponse map[string]string
)

// StartRedis 开启redis的连接
func StartRedis(config config.DB) error {
	host := fmt.Sprintf("%v:%v", config.Host, config.Port)
	RedisPool = newPool(host, config.Password)
	log.Infof(fmt.Sprintf("connected to redis! %v", host))
	return nil
}

// 创建一个新的连接池
func newPool(serverConnStr, password string) *redis.Pool {
	log.Infof(fmt.Sprintf("connecting new redis pool! -->> %v", serverConnStr))
	return &redis.Pool{
		MaxIdle:     1024,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", serverConnStr)
			if err != nil {
				log.Errorf(fmt.Sprintf("error dialing to redis.. %v", err))
				return nil, err
			}
			if password != "" {
				if _, err := c.Do("AUTH", password); err != nil {
					c.Close()
					return nil, err
				}

			}
			return c, err
		},
	}
}

// GetRedisCon 从连接池中获取一个连接
func GetRedisCon() redis.Conn {
	return RedisPool.Get()
}

// HMSet 设置map类型的值
func HMSet(key string, values []interface{}) error {
	s := time.Now()

	if values == nil {
		return nil
	}

	args := []interface{}{key}
	i := 0
	var k, v interface{}
	for _, arg := range values {
		args = append(args, arg)

		// for logging
		if i%2 == 0 {
			k = arg
		} else {
			v = arg
			log.Infof(fmt.Sprintf("confing add [%s][key:%v value:%v]", key, k, v))
		}
		i++
	}

	c := GetRedisCon()
	defer c.Close()

	if _, err := c.Do("HMSET", args...); err != nil {
		log.Errorf(fmt.Sprintf("error HMSet [%s] %v", key, err))
		return err
	}

	log.Infof(fmt.Sprintf("HMSet [%s] took: %v", key, time.Since(s)))
	return nil
}

// HGetALL 获取map类型的值
func HGetALL(key string) (HashResponse, error) {
	c := GetRedisCon()
	defer c.Close()

	res, err := redis.StringMap(c.Do("HGETALL", key))
	if err != nil {
		log.Errorf(fmt.Sprintf("error HGetALL %v", err))
		return res, err
	}
	return res, nil
}
