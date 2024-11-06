package config

import (
	"github.com/micro/go-micro/v2/config"
	"github.com/micro/go-micro/v2/config/source/file"
	"rxcsoft.cn/utils/logger"
)

var (
	log = logger.New()
)

// InitConfig 配置文件初始化
func InitConfig() {
	// 加载配置文件
	if err := loadDBConfig(); err != nil {
		log.Fatalf("db config file not found, error: %v", err)
	}
}

// loadConfig 加载db配置文件
func loadDBConfig() error {
	filePath := "./db-config.json"

	// 通过micro工具类加载db配置文件
	if err := config.Load(file.NewSource(file.WithPath(filePath))); err != nil {
		log.Fatalf("db config file not found, error: %v", err)
		return err
	}
	return nil
}
