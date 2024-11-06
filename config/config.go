package config

import (
	"os"

	"github.com/micro/go-micro/v2/config"
)

const (
	// MysqlKey mysql key
	MysqlKey = "mysql"
	// RedisKey redis key
	RedisKey = "redis"
	// MongoKey mongo key
	MongoKey = "mongo"
	// Neo4jKey neo4j key
	Neo4jKey = "neo4j"
)

type (
	// DB env struct for config
	DB struct {
		Host           string `json:"host"`
		Port           string `json:"port"`
		Username       string `json:"username"`
		Database       string `json:"database"`
		Password       string `json:"password"`
		ReplicaSetName string `json:"replicasetname"`
		Source         string `json:"source"`
	}
	// Storage env struct for config
	Storage struct {
		Platform       string `json:"platform"`
		Bucket         string `json:"bucket"`
		Region         string `json:"region"`
		PublicPath     string `json:"public_path"`
		Endpoint       string `json:"endpoint"`
		AccessID       string `json:"access_id"`
		SecretKey      string `json:"secret_key"`
		ServiceAccount string `json:"service_account"`
		ProjectID      string `json:"project_id"`
	}
)

// GetConf 从go-micro的config中获取mongoDB配置文件,错误情况下返回默认配置
func GetConf(key string) DB {

	// 默认从go-micro中获取信息
	env, err := GetConfFromMicro(key)
	if err != nil {
		// 如果没有获取到，则设置默认的DB配置
		return getDefaultConf(key)
	}

	return env
}

// GetConfFromMicro 从micro中读取配置信息
func GetConfFromMicro(key string) (DB, error) {
	// 获取当前运行环境变量
	env := os.Getenv("ENV")
	var db DB
	// 从micro配置中读取
	if err := config.Get(key, env).Scan(&db); err != nil {
		log.Errorf("get %v db config from micro error: %v", key, err)
		return db, err
	}

	return db, nil
}

// GetStorageConf 从go-micro的config中获取mongoDB配置文件,错误情况下返回默认配置
func GetStorageConf() (Storage, error) {

	// 默认从go-micro中获取信息
	env, err := GetStorageConfFromMicro()
	if err != nil {
		// 如果没有获取到，返回错误
		return env, err
	}

	return env, nil
}

// GetStorageConfFromMicro 从micro中读取配置信息
func GetStorageConfFromMicro() (Storage, error) {
	// 获取当前运行环境变量
	env := os.Getenv("ENV")
	var sg Storage
	// 从micro配置中读取
	if err := config.Get("storage", env).Scan(&sg); err != nil {
		log.Errorf("get storage config from micro error: %v", err)
		return sg, err
	}

	return sg, nil
}

// 设置默认配置
func getDefaultConf(key string) DB {
	var env DB

	switch key {
	case RedisKey:
		env = DB{
			Host: "localhost",
			Port: "6379",
		}
	case MongoKey:
		env = DB{
			Host:     "localhost",
			Port:     "27017",
			Database: "pit3-dev",
		}
	case Neo4jKey:
		env = DB{
			Host: "localhost",
			Port: "7474",
		}
	}

	return env
}
