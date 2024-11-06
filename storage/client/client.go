package client

import (
	"errors"
	"fmt"

	"rxcsoft.cn/utils/config"
	"rxcsoft.cn/utils/logger"
	"rxcsoft.cn/utils/storage"
	"rxcsoft.cn/utils/storage/gcs"
	"rxcsoft.cn/utils/storage/minio"
)

var (
	log           = logger.New()
	storageConfig config.Storage
)

// InitStorageClient 初始化配置文件
func InitStorageClient() {
	var err error
	storageConfig, err = config.GetStorageConf()
	if err != nil {
		panic(errors.New("storage config has error"))
	}

	switch storageConfig.Platform {
	case "gcs":
		isEmpty := (len(storageConfig.Endpoint) > 0 &&
			len(storageConfig.ServiceAccount) > 0 &&
			len(storageConfig.ProjectID) > 0 &&
			len(storageConfig.Region) > 0 &&
			len(storageConfig.Bucket) > 0 &&
			len(storageConfig.PublicPath) > 0)

		// 判断服务的必要字段是否为空，空则抛出错误
		if !isEmpty {
			panic(errors.New("storage config has error"))
		}
	case "minio":
		isEmpty := (len(storageConfig.Endpoint) > 0 &&
			len(storageConfig.AccessID) > 0 &&
			len(storageConfig.SecretKey) > 0 &&
			len(storageConfig.Region) > 0 &&
			len(storageConfig.Bucket) > 0 &&
			len(storageConfig.PublicPath) > 0)

		// 判断服务的必要字段是否为空，空则抛出错误
		if !isEmpty {
			panic(errors.New("storage config has error"))
		}
	default:
		isEmpty := (len(storageConfig.Endpoint) > 0 &&
			len(storageConfig.AccessID) > 0 &&
			len(storageConfig.SecretKey) > 0 &&
			len(storageConfig.Region) > 0 &&
			len(storageConfig.Bucket) > 0 &&
			len(storageConfig.PublicPath) > 0)

		// 判断服务的必要字段是否为空，空则抛出错误
		if !isEmpty {
			panic(errors.New("storage config has error"))
		}
	}
}

// NewClient 获取一个新的客户端
func NewClient(bName string) (cli storage.Service, err error) {
	if storageConfig.Platform == "gcs" {
		var client *gcs.Service

		if len(bName) > 0 {
			bn := fmt.Sprintf("%s-%s", storageConfig.Bucket, bName)
			client = &gcs.Service{
				Endpoint:       storageConfig.Endpoint,
				ServiceAccount: storageConfig.ServiceAccount,
				ProjectID:      storageConfig.ProjectID,
				Region:         storageConfig.Region,
				BucketName:     bn,
				PublicPath:     storageConfig.PublicPath,
			}
		} else {
			client = &gcs.Service{
				Endpoint:       storageConfig.Endpoint,
				ServiceAccount: storageConfig.ServiceAccount,
				ProjectID:      storageConfig.ProjectID,
				Region:         storageConfig.Region,
				BucketName:     storageConfig.Bucket,
				PublicPath:     storageConfig.PublicPath,
			}
		}

		log.Infof("InitStorageClient %v", client)
		if err := client.Initialize(); err != nil {
			log.Infof("InitStorageClient has error: %v", err)
			return nil, err
		}
		return client, nil
	}

	var client *minio.Service

	if len(bName) > 0 {
		bn := fmt.Sprintf("%s-%s", storageConfig.Bucket, bName)
		client = &minio.Service{
			Endpoint:   storageConfig.Endpoint,
			AccessID:   storageConfig.AccessID,
			SecretKey:  storageConfig.SecretKey,
			UseSSL:     false,
			Region:     storageConfig.Region,
			BucketName: bn,
			PublicPath: storageConfig.PublicPath,
		}
	} else {
		client = &minio.Service{
			Endpoint:   storageConfig.Endpoint,
			AccessID:   storageConfig.AccessID,
			SecretKey:  storageConfig.SecretKey,
			UseSSL:     false,
			Region:     storageConfig.Region,
			BucketName: storageConfig.Bucket,
			PublicPath: storageConfig.PublicPath,
		}
	}

	log.Infof("InitStorageClient %v", client)
	if err := client.Initialize(); err != nil {
		log.Infof("InitStorageClient has error: %v", err)
		return nil, err
	}
	return client, nil
}
