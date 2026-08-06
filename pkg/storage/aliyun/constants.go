package aliyun

const (
	aliConfigName = "aliconfig"
	aliConfigType = "yaml"

	aliConfigRootPath = "."
	aliConfigPkgPath  = "pkg/storage/aliyun"

	ossAccessIDKey         = "oss_access.id"
	ossAccessSecretKey     = "oss_access.secret"
	ossAccessEndpointKey   = "oss_access.endpoint"
	ossAccessBucketNameKey = "oss_access.bucket_name"

	// ObjectType 系列：OSS 对象存储路径的首段类型前缀，用于区分不同资源。
	ObjectTypeVideo  = "videos"
	ObjectTypeCover  = "covers"
	ObjectTypeAvatar = "avatars"
)