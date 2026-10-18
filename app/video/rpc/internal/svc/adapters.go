package svc

import (
	"io"

	"go_zero-tiktok/pkg/storage/aliyun"
)

type StorageAdapter struct{}

// UploadFile 上传文件到 OSS，返回 object key（不是完整 URL）。
func (a *StorageAdapter) UploadFile(reader io.Reader, objectKey string) (string, error) {
	if _, err := aliyun.UploadBytes(reader, objectKey); err != nil {
		return "", err
	}
	return objectKey, nil
}

// DeleteFile 删除 OSS 对象，用于清理并发重复发布时未被采用的上传文件。
func (a *StorageAdapter) DeleteFile(objectKey string) error {
	return aliyun.DeleteFile(objectKey)
}
