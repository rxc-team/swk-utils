package database

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/globalsign/mgo"
	"rxcsoft.cn/utils/config"
)

var (
	// MongodbSession 会话
	MongodbSession *mgo.Session

	// Db 当前的db名称
	Db string

	// ensureMaxWrite 默认最大写入条数
	ensureMaxWrite = 1

	// maxSyncTimeout 默认最大同步超时次数
	maxSyncTimeout time.Duration = 1
)

const ()

// StartMongodb 启动mongodb的连接
func StartMongodb(env config.DB) {
	cfe, host := getMongodbHost(env)

	mongoDBDialInfo := buildMongodBconn(cfe, host)
	session, err := mgo.DialWithInfo(&mongoDBDialInfo)
	if err != nil {
		panic(fmt.Sprintf("failed to connect mongodb:%v", err))
	}
	MongodbSession = session

	// 开启调试
	if os.Getenv("MONGODB_DEBUG") == "1" {
		mgo.SetDebug(true)
	}

	setDatabaseName(cfe)
	log.Infof(fmt.Sprintf("connected to mongodb... %v(db:%s)", host, Db))

	printLiveServers(session)
}

// 构建db连接信息
func buildMongodBconn(cfe config.DB, hosts string) mgo.DialInfo {
	mongodbHosts := splitMongodbInstances(hosts)
	ensureMaxWrite = len(mongodbHosts)
	for i, host := range mongodbHosts {
		if !strings.Contains(host, ":") {
			mongodbHosts[i] = fmt.Sprintf("%s:%s", host, cfe.Port)
		}
	}
	log.Infof("connecting to mongodb.. %s", mongodbHosts)
	conn := mgo.DialInfo{
		Addrs:     mongodbHosts,
		Timeout:   10 * time.Second,
		Source:    "admin",
		Username:  cfe.Username,
		PoolLimit: 1000,
		Password:  cfe.Password,

		Direct: false,
	}

	if cfe.ReplicaSetName != "" {
		conn.ReplicaSetName = cfe.ReplicaSetName
	}

	if cfe.Source != "" {
		conn.Source = cfe.Source
	}

	return conn
}

// splitMongodbInstances 将mongodb实例字符串拆分为切片
func splitMongodbInstances(instances string) []string {
	var hosts []string
	for _, host := range strings.Split(instances, ",") {
		hosts = append(hosts, host)
	}

	return hosts
}

func getMongodbHost(env config.DB) (config.DB, string) {
	return env, fmt.Sprintf("%v", env.Host)
}

// IsConnected 判断是否连接
func IsConnected() bool {
	connected := true

	if MongodbSession == nil {
		connected = false
	}

	return connected
}

// SessionCopy 制作一个mongodb会话的副本
func SessionCopy() *mgo.Session {
	MongodbSession.Refresh()

	sc := MongodbSession.Copy()

	sc.SetSyncTimeout(maxSyncTimeout * time.Second)
	sc.SetSocketTimeout(maxSyncTimeout * time.Hour)
	return sc
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
func printLiveServers(session *mgo.Session) {
	log.Infof("mongodb liveServers=%v", session.LiveServers())
}

// BeginMongo 开启mongodb会话
func BeginMongo() (time.Time, *mgo.Session) {
	return time.Now(), SessionCopy()
}

// BeginMongoConn 开启mongoBD连接信息
type BeginMongoConn struct {
	Time time.Time
	Conn *mgo.Session
	Col  *mgo.Collection
}

// BeginMongoWCol 开启一个关联到集合的连接
func BeginMongoWCol() func(string) BeginMongoConn {
	return func(cname string) BeginMongoConn {
		conn := SessionCopy()
		return BeginMongoConn{time.Now(), conn, conn.DB(Db).C(cname)}
	}
}

// setDatabaseName 设置DB名称
func setDatabaseName(env config.DB) error {
	if env.Database != "" {
		Db = env.Database
		return nil
	}

	panic("mongodb database name is not set")
}
