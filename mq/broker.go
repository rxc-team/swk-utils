package mq

import (
	"fmt"
	"os"
	"strings"

	"github.com/micro/go-micro/v2/broker"
	"github.com/micro/go-plugins/broker/rabbitmq/v2"
	"rxcsoft.cn/utils/logger"
)

var log = logger.New()

var rabbitmqBroker broker.Broker

func NewBroker() broker.Broker {

	if rabbitmqBroker != nil {
		return rabbitmqBroker
	}

	bk := rabbitmq.NewBroker(
		broker.Addrs(getMqAddr()...),
	)

	// 创建broker，并设置连接
	if err := bk.Init(); err != nil {
		log.Fatalf("broker.Init()", fmt.Sprintf("broker.Init() has error: %v", err))
	}
	if err := bk.Connect(); err != nil {
		log.Fatalf("broker.Connect()", fmt.Sprintf("broker.Connect() has error: %v", err))
	}

	rabbitmqBroker = bk

	return rabbitmqBroker
}

func getMqAddr() []string {

	var addList []string

	addrss := os.Getenv("RABBITMQ")

	if len(addrss) == 0 {
		addList = append(addList, "amqp://guest:guest@127.0.0.1:5672")
		return addList
	}

	addList = strings.Split(addrss, ",")

	return addList

}
