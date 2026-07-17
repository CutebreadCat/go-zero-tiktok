package svc

import (
	"io"

	"go_zero-tiktok/pkg/storage/aliyun"
)

type StorageAdapter struct{}

func (a *StorageAdapter) UploadFile(reader io.Reader, objectKey string) (string, error) {
	return aliyun.UploadBytesToOSS(reader, objectKey)
}
