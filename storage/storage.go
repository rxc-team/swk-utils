package storage

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"rxcsoft.cn/utils/logger"
)

type (
	// Service 客户端接口
	Service interface {
		// Initialize 初始化
		Initialize() error
		// SaveObject 创建随机名称的文件对象
		SaveObject(file io.Reader, path, contentType string) (*ObjectInfo, error)
		// SavePublicObject 保存文件对象到公开的路径
		SavePublicObject(file io.Reader, path, contentType string) (*ObjectInfo, error)
		// CopyObject 复制文件对象
		CopyObject(srcObjectName, dstObjectName string) (*ObjectInfo, error)
		// DeleteObject 删除文件对象
		DeleteObject(objectName string) error
		// DeleteBucket 删除桶中的所有文件
		DeleteBucket() error
		// DeletePath 删除当前路径下的的所有文件
		DeletePath(path string) (int64, error)
		// GetObject 获取文件对象
		GetObject(objectName string) (io.ReadCloser, error)
		// GetObjectInfo 获取文件对象的信息
		GetObjectInfo(objectName string) (*ObjectInfo, error)
		// GetObjectInfo 获取文件对象的信息
		GetSharedURL(objectName string) (string, error)
		// GetListObjects 获取所有文件
		GetListObjects(prefix string, recursive bool) ([]string, error)
		// GetFolderSize 获取文件夹的占用大小
		GetFolderSize(prefix string, recursive bool) (int64, error)
		// CopyPath 复制一个文件夹
		CopyPath(src, dst string, recursive bool) (int64, error)
		// RenameFolder 将文件夹名改为另一个
		RenameFolder(src, dst string) error

		// 获取公共信息
		// GetBucketName 获取bucket名
		GetBucketName() string
		// GetPublicPath 获取公共路径
		GetPublicPath() string
		// GetRegion 获取域
		GetRegion() string
		// GetEndpoint 获取端点信息
		GetEndpoint() string
	}

	// FileObject 文件对象
	FileObject struct {
		File       io.ReadCloser // the object reader
		ObjectInfo               // object info
	}

	// ObjectInfo 文件详细情报
	ObjectInfo struct {
		Name         string    // 对象名
		SelfLink     string    // 客户端使用路径
		MediaLink    string    // mini中的路径
		ContentType  string    // 文件类型
		Size         int64     // 文件大小
		ETag         string    // metadata
		LastModified time.Time // 最后更新时间
	}
)

var (
	// ErrNotImplemented 未实现的接口api的错误
	ErrNotImplemented = fmt.Errorf("The API is not yet implemented")
	// NameLength 文件名称长度
	NameLength  = 10
	tmpPathName = "_pit"
	log         = logger.New()
)

// NewTempFile 为临时文件准备一个字符串
func NewTempFile(name string, prefixDir ...string) string {
	folder := fmt.Sprintf("%s%s", os.TempDir(), tmpPathName)
	for _, prefix := range prefixDir {
		folder = fmt.Sprintf("%s/%s", folder, prefix)
	}
	if err := os.MkdirAll(folder, os.ModePerm); err != nil {
		log.Errorf("cant mkdir to %v", err)
		return ""
	}
	return filepath.FromSlash(fmt.Sprintf("%s/%s", folder, name))
}
