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
