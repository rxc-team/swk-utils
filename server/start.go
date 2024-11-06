package server

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"rxcsoft.cn/utils/config"
	"rxcsoft.cn/utils/logger"
)

var (
	log               = logger.New()
	configEnvFileName = "config.env"
)

// Start 加载环境变量和配置文件
func Start() {
	// 加载环境变量
	InitConfigEnv()
	// 加载配置文件
	config.InitConfig()
}

// InitConfigEnv 初始化env配置
func InitConfigEnv() {
	file := fmt.Sprintf("%v/%v", getCwd(), configEnvFileName)
	if err := godotenv.Load(file); err != nil {
		log.Fatalf("ERROR LOADING DOTENV %v at path %v", err, file)
	}

	log.Infof("config.env initialize!")
}

func getCwd() string {
	pwd, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// if path is root, return empty instead
	if pwd == "/" {
		pwd = ""
	}

	return pwd
}
