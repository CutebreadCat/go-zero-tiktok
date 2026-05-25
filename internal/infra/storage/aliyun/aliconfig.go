package aliyun

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/spf13/viper"
)

type AliConfig struct {
	OSSAccess struct {
		ID         string `yaml:"id"`
		Secret     string `yaml:"secret"`
		Endpoint   string `yaml:"endpoint"`
		BucketName string `yaml:"bucket_name"`
	} `yaml:"oss_access"`
}

var AliConf AliConfig
var AliClient *oss.Client

func GetAliConfig() {
	viper.SetConfigName(aliConfigName)
	viper.SetConfigType(aliConfigType)
	viper.AddConfigPath(aliConfigRootPath)
	viper.AddConfigPath(aliConfigPkgPath)

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("读取阿里云配置失败: %v", err)
		return
	}

	AliConf.OSSAccess.ID = viper.GetString(ossAccessIDKey)
	AliConf.OSSAccess.Secret = viper.GetString(ossAccessSecretKey)
	AliConf.OSSAccess.Endpoint = viper.GetString(ossAccessEndpointKey)
	AliConf.OSSAccess.BucketName = viper.GetString(ossAccessBucketNameKey)
	log.Println("阿里云配置已加载")
}

func AliInit() {
	endpoint := AliConf.OSSAccess.Endpoint
	accessKeyID := AliConf.OSSAccess.ID
	accessKeySecret := AliConf.OSSAccess.Secret

	if endpoint == "" || accessKeyID == "" || accessKeySecret == "" {
		log.Println("阿里云配置不完整，OSS 上传已禁用")
		return
	}

	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		log.Printf("初始化阿里云 OSS 客户端失败: %v", err)
		return
	}

	AliClient = client
	log.Println("阿里云 OSS 客户端已初始化")
}

func UploadFileToOSS(localFilePath, objectKey string) (string, error) {
	bucket, err := bucket()
	if err != nil {
		return "", err
	}

	if err := bucket.PutObjectFromFile(objectKey, localFilePath); err != nil {
		return "", fmt.Errorf("upload file to oss failed: %w", err)
	}

	return objectURL(objectKey), nil
}

func DeleteFileFromOSS(objectKey string) error {
	bucket, err := bucket()
	if err != nil {
		return err
	}

	if err := bucket.DeleteObject(objectKey); err != nil {
		return fmt.Errorf("delete oss object failed: %w", err)
	}

	return nil
}

func UploadBytesToOSS(reader io.Reader, objectKey string) (string, error) {
	bucket, err := bucket()
	if err != nil {
		return "", err
	}

	if err := bucket.PutObject(objectKey, reader); err != nil {
		return "", fmt.Errorf("upload bytes to oss failed: %w", err)
	}

	return objectURL(objectKey), nil
}

func bucket() (*oss.Bucket, error) {
	if AliClient == nil {
		return nil, fmt.Errorf("aliyun oss client is not initialized")
	}

	bucket, err := AliClient.Bucket(AliConf.OSSAccess.BucketName)
	if err != nil {
		return nil, fmt.Errorf("get oss bucket failed: %w", err)
	}

	return bucket, nil
}

func objectURL(objectKey string) string {
	endpoint := strings.TrimPrefix(AliConf.OSSAccess.Endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return fmt.Sprintf("https://%s.%s/%s", AliConf.OSSAccess.BucketName, endpoint, objectKey)
}
