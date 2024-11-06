package gcs

import (
	"context"
	"fmt"
	"io"
	"path"
	"path/filepath"
	"time"

	cloud "cloud.google.com/go/storage"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	"rxcsoft.cn/utils/logger"
	"rxcsoft.cn/utils/storage"
)

type (
	// Service 文件服务器客户端
	Service struct {
		Endpoint       string
		ServiceAccount string
		Region         string
		BucketName     string
		PublicPath     string
		ProjectID      string

		client *cloud.Client
	}
)

var (
	// GCSJSONFileName key file name for api key access
	GCSJSONFileName = "GCS_JSON_FILE_NAME"
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
	ctx := context.Background()
	// 盘点是否存在客户端，存在则直接返回
	if svc.client != nil {
		return nil
	}
	// 创建新的客户端
	// 判断服务的必要字段是否为空，空则抛出错误
	if ok := (svc.ServiceAccount != "" &&
		svc.Region != "" &&
		svc.BucketName != "" &&
		svc.ProjectID != "" &&
		svc.PublicPath != ""); !ok {
		return fmt.Errorf("Invalid service struct: %v", svc)
	}

	client, err := cloud.NewClient(ctx, option.WithCredentialsJSON([]byte(svc.ServiceAccount)))
	if err != nil {
		return fmt.Errorf("Unable to create storage service: %v", err)
	}

	bucket := client.Bucket(svc.BucketName)
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()
	if err := bucket.Create(ctx, svc.ProjectID, &cloud.BucketAttrs{
		StorageClass: "STANDARD",
		Location:     "asia",
	}); err != nil {
		return fmt.Errorf("Failed to ensure public folder: %v", err)
	}

	svc.client = client
	return nil
}

// NewObject 基础的创建一个文件对象
func (svc *Service) NewObject(objectName string, file io.Reader, contentType string) (*storage.ObjectInfo, error) {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	wc := bucket.Object(objectName).NewWriter(ctx)
	defer wc.Close()
	if _, err := io.Copy(wc, file); err != nil {
		log.Errorf("gcs.WriterObject failed: %v", err)
		return nil, err
	}

	return &storage.ObjectInfo{
		Name:         wc.Name,
		MediaLink:    fmt.Sprintf("/storage/%s", wc.MediaLink),
		SelfLink:     wc.MediaLink,
		ContentType:  wc.ContentType,
		Size:         wc.Size,
		ETag:         wc.Etag,
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
	return svc.NewObject(
		objectName,
		file, contentType)
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

	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	src := bucket.Object(srcObjectName)
	dst := bucket.Object(dstObjectName)

	uploadInfo, err := dst.CopierFrom(src).Run(ctx)
	if err != nil {
		log.Errorf("client.CopyObject in minio.CopyObject failed: %v", err)
		return nil, err
	}

	return &storage.ObjectInfo{
		Name:         uploadInfo.Name,
		MediaLink:    fmt.Sprintf("/storage/%s", uploadInfo.MediaLink),
		SelfLink:     uploadInfo.MediaLink,
		Size:         uploadInfo.Size,
		ETag:         uploadInfo.Etag,
		LastModified: time.Now(),
	}, nil
}

// GetObject 获取文件对象
func (svc *Service) GetObject(objectName string) (io.ReadCloser, error) {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	object, err := bucket.Object(objectName).NewReader(ctx)
	if err != nil {
		log.Errorf("Error GetObject '%s/%s': %v", svc.BucketName, objectName, err)
		return nil, err
	}
	return object, nil
}

// DeleteObject 基础的删除文件对象
func (svc *Service) DeleteObject(objectName string) error {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	if err := bucket.Object(objectName).Delete(ctx); err != nil {
		log.Warnf("error DeleteObject: %v[%v/%v]", err, svc.BucketName, objectName)
		return err
	}
	return nil
}

// DeleteBucket 删除桶中的所有文件
func (svc *Service) DeleteBucket() error {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	if err := bucket.Delete(ctx); err != nil {
		log.Warnf("error DeleteBucket: %v[%v]", err, svc.BucketName)
		return err
	}
	return nil
}

// DeletePath 删除当前路径下的的所有文件
func (svc *Service) DeletePath(ph string) (int64, error) {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	var total int64 = 0

	publicFiles := bucket.Objects(ctx, &cloud.Query{
		Prefix: path.Join(svc.PublicPath, ph),
	})

	// 删除公共空间下 app 的对象
	for {
		obj, err := publicFiles.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}

		total += obj.Size

		if err := bucket.Object(obj.Name).Delete(ctx); err != nil {
			log.Warnf("error DeleteObject: %v[%v/%v]", err, svc.BucketName, obj.Name)
			return 0, err
		}
	}
	privateFiles := bucket.Objects(ctx, &cloud.Query{
		Prefix: path.Join(ph),
	})

	// 删除私有空间下 app 的对象
	for {
		obj, err := privateFiles.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}

		total += obj.Size

		if err := bucket.Object(obj.Name).Delete(ctx); err != nil {
			log.Warnf("error DeleteObject: %v[%v/%v]", err, svc.BucketName, obj.Name)
			return 0, err
		}
	}

	return total, nil
}

// GetObjectInfo 获取文件对象的详细情报
func (svc *Service) GetObjectInfo(objectName string) (*storage.ObjectInfo, error) {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	obj, err := bucket.Object(objectName).Attrs(ctx)
	if err != nil {
		return nil, err
	}
	return &storage.ObjectInfo{
		Name:         objectName,
		MediaLink:    fmt.Sprintf("/storage/%s", obj.MediaLink),
		SelfLink:     obj.MediaLink,
		ContentType:  obj.ContentType,
		Size:         obj.Size,
		ETag:         obj.Etag,
		LastModified: obj.Updated,
	}, nil
}

// GetSharedURL 获取文件的临时分享链接
func (svc *Service) GetSharedURL(objectName string) (string, error) {
	conf, err := google.JWTConfigFromJSON([]byte(svc.ServiceAccount))
	if err != nil {
		return "", fmt.Errorf("google.JWTConfigFromJSON: %v", err)
	}
	opts := &cloud.SignedURLOptions{
		Scheme:         cloud.SigningSchemeV4,
		Method:         "GET",
		GoogleAccessID: conf.Email,
		PrivateKey:     conf.PrivateKey,
		Expires:        time.Now().Add(24 * time.Hour),
	}
	u, err := cloud.SignedURL(svc.BucketName, objectName, opts)
	if err != nil {
		return "", fmt.Errorf("storage.SignedURL: %v", err)
	}

	return u, nil
}

// GetListObjects 获取所有文件
func (svc *Service) GetListObjects(prefix string, recursive bool) ([]string, error) {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	q := cloud.Query{
		Prefix: prefix,
	}

	it := bucket.Objects(ctx, &q)
	var objects []string
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		objects = append(objects, attrs.Name)
	}
	return objects, nil
}

// GetFolderSize 获取文件夹大小
func (svc *Service) GetFolderSize(prefix string, recursive bool) (int64, error) {
	ctx := context.Background()
	bucket := svc.client.Bucket(svc.BucketName)

	ctx, cancel := context.WithTimeout(ctx, time.Minute*5)
	defer cancel()

	q := cloud.Query{
		Prefix: prefix,
	}

	it := bucket.Objects(ctx, &q)
	var size int64 = 0
	for {
		attrs, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return 0, err
		}
		size += attrs.Size
	}
	return size, nil
}

// CopyPath 复制一个文件夹
func (svc *Service) CopyPath(src, dst string, recursive bool) (int64, error) {
	return 0, nil
}

// CopyPath 复制一个文件夹
func (svc *Service) RenameFolder(src, dst string) error {
	return nil
}
