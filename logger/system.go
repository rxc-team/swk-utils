package logger

import (
	"os"

	nested "github.com/antonfisher/nested-logrus-formatter"
	"github.com/sirupsen/logrus"
)

// New 获取一个系统log实例
func New() *logrus.Logger {
	log := logrus.New()
	// 设置控制台输出
	log.Out = os.Stdout
	// 设置等级到debug模式
	log.Level = logrus.InfoLevel
	// 设置格式为文本格式，时间戳格式为2006-01-02 15:04:05.000000
	formatter := &nested.Formatter{
		HideKeys:        true,
		NoFieldsColors:  false,
		NoColors:        false,
		TimestampFormat: "2006-01-02 15:04:05",
	}

	log.SetFormatter(formatter)

	return log
}
