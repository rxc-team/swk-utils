package minio

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"rxcsoft.cn/utils/logger"
	"rxcsoft.cn/utils/storage"
)

type (
	// Service 文件服务器客户端
	Service struct {
		Endpoint   string
		AccessID   string
		SecretKey  string
		UseSSL     bool
		Region     string
		BucketName string
		PublicPath string

		client *minio.Client
	}
)

var log = logger.New()

// GetBucketName 获取bucket名
func (svc *Service) GetBucketName() string {
	return svc.BucketName
}

// GetPublicPath 获取公共路径
func (svc *Service) GetPublicPath() string {
	return svc.PublicPath
}

// GetRegion 获取区域
func (svc *Service) GetRegion() string {
	return svc.Region
}

// GetEndpoint 获取端点
func (svc *Service) GetEndpoint() string {
	return svc.Endpoint
}

// Initialize 初始化客户端
func (svc *Service) Initialize() error {
	// 盘点是否存在客户端，存在则直接返回
	if svc.client != nil {
		return nil
	}
	// 创建新的客户端
	// 判断服务的必要字段是否为空，空则抛出错误
	if ok := (svc.Endpoint != "" &&
		svc.AccessID != "" &&
		svc.SecretKey != "" &&
		svc.Region != "" &&
		svc.BucketName != "" &&
		svc.PublicPath != ""); !ok {
		return fmt.Errorf("Invalid service struct: %v", svc)
	}
	client, err := minio.New(
		svc.Endpoint,
		&minio.Options{
			Creds:  credentials.NewStaticV4(svc.AccessID, svc.SecretKey, ""),
			Secure: svc.UseSSL,
			Region: svc.Region,
		},
	)
	if err != nil {
		return fmt.Errorf("Unable to create storage service: %v", err)
	}
	var found bool
	found, err = createBucket(client, svc.BucketName, svc.Region)
	// 设置权限为可读
	if !found && err == nil {
		err = client.SetBucketPolicy(
			context.Background(),
			svc.BucketName,
			svc.publicPolicy(),
		)
	}
	if err != nil {
		return fmt.Errorf("Failed to ensure public folder: %v", err)
	}
	svc.client = client
	return nil
}

// NewObject 基础的创建一个文件对象
func (svc *Service) NewObject(objectName string, file io.Reader, contentType string) (*storage.ObjectInfo, error) {
	info, err := svc.client.PutObject(context.Background(), svc.BucketName, objectName, file, -1, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Errorf("minio.PutObject failed: %v", err)
		return nil, err
	}
	return &storage.ObjectInfo{
		Name:         info.Key,
		MediaLink:    fmt.Sprintf("/storage/%s/%s", svc.BucketName, objectName),
		SelfLink:     fmt.Sprintf("%s/%s", svc.BucketName, objectName),
		ContentType:  contentType,
		Size:         info.Size,
		ETag:         info.ETag,
		LastModified: time.Now(),
	}, nil
}

// SaveObject 保存为随机名称的文件对象
func (svc *Service) SaveObject(file io.Reader, path, contentType string) (*storage.ObjectInfo, error) {
	objectName := generateObjectName(path)
	return svc.NewObject(objectName, file, contentType)
}

// 生成带路径的文件名
func generateObjectName(filePath string) string {
	paths, fileName := filepath.Split(filePath)
	name := time.Now().Format("20060102030405") + "_" + fileName
	return path.Join(paths, name)
}

// createPublicObject 创建公共路径下的文件对象
func (svc *Service) createPublicObject(objectName string, file io.Reader, contentType string) (*storage.ObjectInfo, error) {
	objectName = fmt.Sprintf("%s/%s", svc.PublicPath, objectName)
	return svc.NewObject(objectName, file, contentType)
}

// SavePublicObject 保存文件对象到公共路径下
func (svc *Service) SavePublicObject(file io.Reader, path, contentType string) (*storage.ObjectInfo, error) {
	fileName := generateObjectName(path)
	return svc.createPublicObject(
		fileName,
		file,
		contentType,
	)
}

// CopyObject 复制文件对象
func (svc *Service) CopyObject(srcObjectName, dstObjectName string) (*storage.ObjectInfo, error) {
	srcOpts := minio.CopySrcOptions{
		Bucket: svc.BucketName,
		Object: srcObjectName,
	}

	// Destination object
	dstOpts := minio.CopyDestOptions{
		Bucket: svc.BucketName,
		Object: dstObjectName,
	}

	uploadInfo, err := svc.client.CopyObject(context.Background(), dstOpts, srcOpts)
	if err != nil {
		log.Errorf("client.CopyObject in minio.CopyObject failed: %v", err)
		return nil, err
	}

	return &storage.ObjectInfo{
		Name:         dstObjectName,
		MediaLink:    fmt.Sprintf("/storage/%s/%s", svc.BucketName, dstObjectName),
		SelfLink:     fmt.Sprintf("%s/%s", svc.BucketName, dstObjectName),
		Size:         uploadInfo.Size,
		ETag:         uploadInfo.ETag,
		LastModified: time.Now(),
		// Size:         uint64(n),
	}, nil
}

// GetObject 获取文件对象
func (svc *Service) GetObject(objectName string) (io.ReadCloser, error) {
	object, err := svc.client.GetObject(context.Background(), svc.BucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("Error GetObject '%s/%s': %v", svc.BucketName, objectName, err)
		return nil, err
	}
	return object, nil
}

// DeleteObject 基础的删除文件对象
func (svc *Service) DeleteObject(objectName string) error {
	if err := svc.client.RemoveObject(context.Background(), svc.BucketName, objectName, minio.RemoveObjectOptions{}); err != nil {
		log.Warnf("error DeleteObject: %v[%v/%v]", err, svc.BucketName, objectName)
		return err
	}
	return nil
}

// DeleteBucket 删除桶中的所有文件
func (svc *Service) DeleteBucket() error {
	objectsCh := make(chan minio.ObjectInfo)

	go func() {
		defer close(objectsCh)
		for object := range svc.client.ListObjects(context.Background(), svc.BucketName, minio.ListObjectsOptions{Recursive: true, Prefix: "/"}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object
		}
	}()

	for err := range svc.client.RemoveObjects(context.Background(), svc.BucketName, objectsCh, minio.RemoveObjectsOptions{}) {
		if err.Err != nil {
			return err.Err
		}
		log.Warnf("error DeleteBucket: %v[%v]", err.Err, svc.BucketName)
		return err.Err
	}

	if err := svc.client.RemoveBucket(context.Background(), svc.BucketName); err != nil {
		log.Warnf("error DeleteBucket: %v[%v]", err, svc.BucketName)
		return err
	}
	return nil
}

// DeletePath 删除当前路径下的的所有文件
func (svc *Service) DeletePath(ph string) (int64, error) {
	objectsCh := make(chan minio.ObjectInfo)
	objectsSize := make(chan int64)

	go func() {
		defer close(objectsCh)
		defer close(objectsSize)

		for object := range svc.client.ListObjects(context.Background(), svc.BucketName, minio.ListObjectsOptions{Recursive: true, Prefix: path.Join(svc.PublicPath, ph)}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object
			objectsSize <- object.Size
		}

		for object := range svc.client.ListObjects(context.Background(), svc.BucketName, minio.ListObjectsOptions{Recursive: true, Prefix: path.Join(ph)}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}
			objectsCh <- object
			objectsSize <- object.Size
		}
	}()

	var total int64 = 0
	var wg sync.WaitGroup

	wg.Add(1)

	go func() {
		defer wg.Done()
		for size := range objectsSize {
			atomic.AddInt64(&total, size)
		}
	}()

	wg.Add(1)

	go func() {
		defer wg.Done()
		for err := range svc.client.RemoveObjects(context.Background(), svc.BucketName, objectsCh, minio.RemoveObjectsOptions{
			GovernanceBypass: true,
		}) {
			if err.Err != nil {
				log.Warnf("error DeleteBucket: %v[%v]", err.Err, svc.BucketName)
				return
			}
		}
	}()

	wg.Wait()

	return total, nil
}

// GetObjectInfo 获取文件对象的详细情报
func (svc *Service) GetObjectInfo(objectName string) (*storage.ObjectInfo, error) {
	obj, err := svc.client.StatObject(context.Background(), svc.BucketName, objectName, minio.StatObjectOptions{})
	if err != nil {
		return nil, err
	}
	return &storage.ObjectInfo{
		Name:         objectName,
		MediaLink:    fmt.Sprintf("/storage/%s/%s", svc.BucketName, objectName),
		SelfLink:     fmt.Sprintf("%s/%s", svc.BucketName, objectName),
		ContentType:  obj.ContentType,
		Size:         int64(obj.Size),
		ETag:         obj.ETag,
		LastModified: obj.LastModified,
	}, nil
}

// GetSharedURL 获取文件的临时分享链接
func (svc *Service) GetSharedURL(objectName string) (string, error) {
	// Set request parameters for content-disposition.
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "attachment; filename=\"objectName\"")

	// Generates a presigned url which expires in a day.
	presignedURL, err := svc.client.PresignedGetObject(context.Background(), svc.BucketName, objectName, time.Second*24*60*60, reqParams)
	if err != nil {
		return "", err
	}
	return presignedURL.Path, nil
}

// GetListObjects 获取所有文件
func (svc *Service) GetListObjects(prefix string, recursive bool) ([]string, error) {
	objectCh := svc.client.ListObjects(context.Background(), svc.BucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	var objects []string
	for object := range objectCh {
		if object.Err != nil {
			return nil, object.Err
		}
		fmt.Println(object)
		objects = append(objects, object.Key)
	}
	return objects, nil
}

// GetFolderSize 获取文件夹大小
func (svc *Service) GetFolderSize(prefix string, recursive bool) (int64, error) {
	objectCh := svc.client.ListObjects(context.Background(), svc.BucketName, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: recursive,
	})

	var size int64 = 0
	for object := range objectCh {
		if object.Err != nil {
			return 0, object.Err
		}
		fmt.Println(object)
		size += object.Size
	}
	return size, nil
}

// CopyPath 复制一个文件夹
func (svc *Service) CopyPath(src, dst string, recursive bool) (int64, error) {
	ctx := context.Background()
	var wg sync.WaitGroup
	var total int64 = 0
	wg.Add(1)
	go func() {
		defer wg.Done()

		for object := range svc.client.ListObjects(ctx, svc.BucketName, minio.ListObjectsOptions{Recursive: recursive, Prefix: path.Join(src)}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}

			srcOpts := minio.CopySrcOptions{
				Bucket: svc.BucketName,
				Object: object.Key,
			}

			// Destination object
			dstOpts := minio.CopyDestOptions{
				Bucket: svc.BucketName,
				Object: strings.Replace(object.Key, src, path.Join(dst), 1),
			}

			_, err := svc.client.CopyObject(ctx, dstOpts, srcOpts)
			if err != nil {
				log.Errorf("client.CopyObject in minio.CopyObject failed: %v", err)
				return
			}

			atomic.AddInt64(&total, object.Size)
		}
	}()

	wg.Wait()

	return total, nil
}

// CopyPath 复制一个文件夹
func (svc *Service) RenameFolder(src, dst string) error {
	ctx := context.Background()
	objectsCh := make(chan minio.ObjectInfo)
	go func() {
		defer close(objectsCh)

		for object := range svc.client.ListObjects(ctx, svc.BucketName, minio.ListObjectsOptions{Recursive: true, Prefix: path.Join(src)}) {
			if object.Err != nil {
				log.Fatalln(object.Err)
			}

			srcOpts := minio.CopySrcOptions{
				Bucket: svc.BucketName,
				Object: object.Key,
			}

			// Destination object
			dstOpts := minio.CopyDestOptions{
				Bucket: svc.BucketName,
				Object: strings.Replace(object.Key, src, path.Join(dst), 1),
			}

			_, err := svc.client.CopyObject(ctx, dstOpts, srcOpts)
			if err != nil {
				log.Errorf("client.CopyObject in minio.CopyObject failed: %v", err)
				return
			}

			objectsCh <- object
		}
	}()

	for err := range svc.client.RemoveObjects(context.Background(), svc.BucketName, objectsCh, minio.RemoveObjectsOptions{
		GovernanceBypass: true,
	}) {
		if err.Err != nil {
			log.Warnf("error DeleteBucket: %v[%v]", err.Err, svc.BucketName)
			return err.Err
		}
	}

	return nil
}
