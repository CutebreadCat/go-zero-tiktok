package ali

import (
	"fmt"
	"io"
	"log"

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
	viper.SetConfigName("aliconfig")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("internal/mw/ali")
	err := viper.ReadInConfig()
	if err != nil {
		log.Printf("读取阿里云配置文件失败: %v", err)
		return
	} else {
		log.Printf("成功找到阿里云配置文件")
		// 打印所有配置键值，检查是否正确读取
		keys := viper.AllKeys()
		log.Printf("配置文件中的键: %v", keys)

		// 尝试直接获取配置值
		id := viper.GetString("oss_access.id")
		secret := viper.GetString("oss_access.secret")
		endpoint := viper.GetString("oss_access.endpoint")
		bucket := viper.GetString("oss_access.bucket_name")

		// 无论 Unmarshal 是否成功，都手动设置配置值
		AliConf.OSSAccess.ID = id
		AliConf.OSSAccess.Secret = secret
		AliConf.OSSAccess.Endpoint = endpoint
		AliConf.OSSAccess.BucketName = bucket
	}
	log.Printf("✅ 成功读取阿里云配置文件")

}

func AliInit() {
	// 从环境变量读取密钥
	endpoint := AliConf.OSSAccess.Endpoint
	accessKeyID := AliConf.OSSAccess.ID
	accessKeySecret := AliConf.OSSAccess.Secret

	// 检查配置是否完整
	if endpoint == "" || accessKeyID == "" || accessKeySecret == "" {
		log.Printf("阿里云配置不完整，文件上传功能将不可用")
		return
	}

	// 初始化客户端
	client, err := oss.New(endpoint, accessKeyID, accessKeySecret)
	if err != nil {
		log.Printf("初始化阿里云客户端失败: %v，文件上传功能将不可用", err)
		return
	}
	AliClient = client
	log.Println("✅ 成功初始化阿里云 OSS 客户端！")

}

func UploadFileToOSS(localFilePath, objectKey string) (string, error) {
	bucket, err := AliClient.Bucket(AliConf.OSSAccess.BucketName)
	if err != nil {
		return "", fmt.Errorf("获取Bucket失败: %v", err)
	}
	err = bucket.PutObjectFromFile(objectKey, localFilePath)
	if err != nil {
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	fileURL := fmt.Sprintf("https://%s.%s/%s", AliConf.OSSAccess.BucketName, AliConf.OSSAccess.Endpoint, objectKey)
	return fileURL, nil
}

func DeleteFileFromOSS(objectKey string) error {
	bucket, err := AliClient.Bucket(AliConf.OSSAccess.BucketName)
	if err != nil {
		return fmt.Errorf("获取Bucket失败: %v", err)
	}
	err = bucket.DeleteObject(objectKey)
	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}
	return nil
}

func UploadBytesToOSS(reader io.Reader, objectKey string) (string, error) {
	bucket, err := AliClient.Bucket(AliConf.OSSAccess.BucketName)
	if err != nil {
		return "", fmt.Errorf("获取Bucket失败: %v", err)
	}
	err = bucket.PutObject(objectKey, reader)
	if err != nil {
		return "", fmt.Errorf("上传字节数据失败: %v", err)
	}
	fileURL := fmt.Sprintf("https://%s.%s/%s", AliConf.OSSAccess.BucketName, AliConf.OSSAccess.Endpoint, objectKey)
	return fileURL, nil
}
