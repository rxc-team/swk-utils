package mongo

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"rxcsoft.cn/utils/config"
	"rxcsoft.cn/utils/logger"
)

var (
	log = logger.New()
	// mongo连接
	client *mongo.Client
	// Db 当前的db名称
	Db string
)

const (
	// MaxPoolSize 连接池大小
	MaxPoolSize uint64 = 1000
)

// StartMongodb 启动mongodb的连接
func StartMongodb(env config.DB) {
	cfe, host := getMongodbHost(env)
	hosts := buildHost(host, cfe.Port)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	option := options.Client()
	option.SetHosts(hosts)                   // 设置连接host
	option.SetReplicaSet(env.ReplicaSetName) // 设置replica name
	option.SetMaxPoolSize(MaxPoolSize)       // 设置最大连接池的数量
	option.SetMinPoolSize(100)               // 设定最小连接池大小
	option.SetRetryReads(true)               // 增加读的重试
	option.SetRetryWrites(true)              // 增加写的重试
	// 设置读的偏好
	r := readpref.SecondaryPreferred()
	option.SetReadPreference(r)
	// 设置读隔离
	rn := readconcern.Local()
	option.SetReadConcern(rn)
	// 设置写隔离
	w := writeconcern.WriteConcern{}
	w.WithOptions(writeconcern.WMajority())
	option.SetWriteConcern(&w)
	if len(env.Username) > 0 && len(env.Password) > 0 {
		option.SetAuth(
			options.Credential{ // 设置认证信息
				AuthSource: env.Source,
				Username:   env.Username,
				Password:   env.Password,
			})
	}
	cli, err := mongo.Connect(ctx, option)
	if err != nil {
		log.Panic(err)
	}

	client = cli

	setDatabaseName(cfe)
	log.Infof("connected to mongodb... %v(db:%s)", host, Db)
	pingServers(client)
}

// 构建db连接信息
func buildHost(hosts, port string) []string {
	mongodbHosts := splitMongodbInstances(hosts)
	// 编辑host主机信息
	for i, host := range mongodbHosts {
		if !strings.Contains(host, ":") {
			mongodbHosts[i] = fmt.Sprintf("%s:%s", host, port)
		}
	}

	return mongodbHosts
}

// splitMongodbInstances 将mongodb实例字符串拆分为切片
func splitMongodbInstances(instances string) []string {
	var hosts []string
	hosts = append(hosts, strings.Split(instances, ",")...)

	return hosts
}

func getMongodbHost(env config.DB) (config.DB, string) {
	return env, fmt.Sprintf("%v", env.Host)
}

// New 返回一个连接
func New() *mongo.Client {
	return client
}

// GetDBName 获取带后缀的DB名称
func GetDBName(suffix string) string {
	name := strings.Builder{}
	name.WriteString(Db)
	name.WriteString("_")
	name.WriteString(suffix)
	return name.String()
}

// 打印live情报
func pingServers(client *mongo.Client) {
	client.Ping(context.TODO(), &readpref.ReadPref{})
}

// setDatabaseName 设置DB名称
func setDatabaseName(env config.DB) error {
	if env.Database != "" {
		Db = env.Database
		return nil
	}

	panic("mongodb database name is not set")
}
