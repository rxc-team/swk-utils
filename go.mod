module rxcsoft.cn/utils

go 1.13

require (
	cloud.google.com/go v0.68.0 // indirect
	cloud.google.com/go/storage v1.12.0
	github.com/antonfisher/nested-logrus-formatter v1.3.0
	github.com/dimchansky/utfbom v1.1.0
	github.com/garyburd/redigo v1.6.2
	github.com/globalsign/mgo v0.0.0-20181015135952-eeefdecb41b8
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.1.2
	github.com/joho/godotenv v1.3.0
	github.com/kr/pretty v0.2.1 // indirect
	github.com/micro/go-micro/v2 v2.9.1
	github.com/micro/go-plugins/broker/rabbitmq/v2 v2.9.1
	github.com/minio/minio-go/v7 v7.0.23
	github.com/olivere/elastic/v7 v7.0.21
	github.com/saintfish/chardet v0.0.0-20120816061221-3af4cd4741ca
	github.com/sirupsen/logrus v1.8.1
	go.mongodb.org/mongo-driver v1.4.2
	go.opencensus.io v0.22.5 // indirect
	golang.org/x/oauth2 v0.0.0-20200902213428-5d25da1a8d43
	golang.org/x/text v0.3.6
	google.golang.org/api v0.34.0
	google.golang.org/genproto v0.0.0-20201102152239-715cce707fb0 // indirect
)

replace (
	google.golang.org/api => google.golang.org/api v0.14.0
	google.golang.org/grpc => google.golang.org/grpc v1.26.0
)
