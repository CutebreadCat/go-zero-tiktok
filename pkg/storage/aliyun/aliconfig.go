package aliyun

import (
	"fmt"
	"io"
	"path/filepath"
	"strings"

	appLogger "go_zero-tiktok/pkg/logger"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
)

// Config 阿里云 OSS 配置
type Config struct {
	OSSAccess struct {
		ID         string `yaml:"id"`
		Secret     string `yaml:"secret"`
		Endpoint   string `yaml:"endpoint"`
		BucketName string `yaml:"bucket_name"`
	} `yaml:"oss_access"`
}

var (
	conf   Config
	client *oss.Client
)

// LoadConfig 从本地 yaml 加载 OSS 配置
func LoadConfig() {
	viper.SetConfigName(aliConfigName)
	viper.SetConfigType(aliConfigType)
	viper.AddConfigPath(aliConfigRootPath)
	viper.AddConfigPath(aliConfigPkgPath)

	if err := viper.ReadInConfig(); err != nil {
		appLogger.Errorf("读取阿里云配置失败: %v", err)
		return
	}

	conf.OSSAccess.ID = viper.GetString(ossAccessIDKey)
	conf.OSSAccess.Secret = viper.GetString(ossAccessSecretKey)
	conf.OSSAccess.Endpoint = viper.GetString(ossAccessEndpointKey)
	conf.OSSAccess.BucketName = viper.GetString(ossAccessBucketNameKey)
	appLogger.Info("阿里云配置已加载")
}

// InitClient 初始化 OSS 客户端
func InitClient() {
	endpoint := conf.OSSAccess.Endpoint
	accessKeyID := conf.OSSAccess.ID
	accessKeySecret := conf.OSSAccess.Secret

	if endpoint == "" || accessKeyID == "" || accessKeySecret == "" {
		appLogger.Warn("阿里云配置不完整，OSS 上传已禁用")
		return
	}

	c, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		appLogger.Errorf("初始化阿里云 OSS 客户端失败: %v", err)
		return
	}

	client = c
	appLogger.Info("阿里云 OSS 客户端已初始化")
}

// UploadFile 上传本地文件到 OSS
func UploadFile(localFilePath, objectKey string) (string, error) {
	bucket, err := getBucket()
	if err != nil {
		return "", err
	}

	if err := bucket.PutObjectFromFile(objectKey, localFilePath); err != nil {
		return "", fmt.Errorf("upload file to oss failed: %w", err)
	}

	return buildObjectURL(objectKey), nil
}

// DeleteFile 删除 OSS 对象
func DeleteFile(objectKey string) error {
	bucket, err := getBucket()
	if err != nil {
		return err
	}

	if err := bucket.DeleteObject(objectKey); err != nil {
		return fmt.Errorf("delete oss object failed: %w", err)
	}

	return nil
}

// UploadBytes 上传字节流到 OSS
func UploadBytes(reader io.Reader, objectKey string) (string, error) {
	bucket, err := getBucket()
	if err != nil {
		return "", err
	}

	if err := bucket.PutObject(objectKey, reader); err != nil {
		return "", fmt.Errorf("upload bytes to oss failed: %w", err)
	}

	return buildObjectURL(objectKey), nil
}

// BuildObjectKey 构造 OSS 对象存储路径：{type}/{userId}/{entityId}/{filename}
// objectType 用 ObjectType 系列常量；filename 经 filepath.Base 过滤，防止路径注入。
func BuildObjectKey(objectType string, userID, entityID int64, filename string) string {
	return fmt.Sprintf("%s/%d/%d/%s", objectType, userID, entityID, filepath.Base(filename))
}

func getBucket() (*oss.Bucket, error) {
	if client == nil {
		return nil, fmt.Errorf("aliyun oss client is not initialized")
	}

	bucket, err := client.Bucket(conf.OSSAccess.BucketName)
	if err != nil {
		return nil, fmt.Errorf("get oss bucket failed: %w", err)
	}

	return bucket, nil
}

func buildObjectURL(objectKey string) string {
	endpoint := strings.TrimPrefix(conf.OSSAccess.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return fmt.Sprintf("https://%s.%s/%s", conf.OSSAccess.BucketName, endpoint, objectKey)
}