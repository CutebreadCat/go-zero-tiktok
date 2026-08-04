package svc

import (
	"go_zero-tiktok/app/user/rpc/internal/config"
	userrepo "go_zero-tiktok/app/user/rpc/internal/dal/reposity"
	userdomain "go_zero-tiktok/app/user/rpc/internal/domain"
	"go_zero-tiktok/pkg/storage/aliyun"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type ServiceContext struct {
	Config             config.Config
	DB                 *gorm.DB
	Rdb                *redis.Redis
	Dal                *Repositories
	UserAuthService    *userdomain.AuthService
	UserMfaService     *userdomain.MfaService
	UserProfileService *userdomain.ProfileService
}

func NewServiceContext(c config.Config) *ServiceContext {
	db, err := gorm.Open(mysql.Open(c.DataSource), &gorm.Config{})
	logx.Must(err)

	rdb := redis.MustNewRedis(c.AppRedis)
	dalRepo := NewRepositories(db)

	aliyun.GetAliConfig()
	aliyun.AliInit()

	tokenAdapter := &TokenAdapter{}
	mfaAdapter := &MfaAdapter{}
	storageAdapter := &StorageAdapter{}

	return &ServiceContext{
		Config:             c,
		DB:                 db,
		Rdb:                rdb,
		Dal:                dalRepo,
		UserAuthService:    userdomain.NewAuthService(dalRepo.User, tokenAdapter, mfaAdapter, c.JwtAuth.AccessSecret, rdb),
		UserMfaService:     userdomain.NewMfaService(dalRepo.User, mfaAdapter, mfaAdapter),
		UserProfileService: userdomain.NewProfileService(dalRepo.User, storageAdapter),
	}
}

type Repositories struct {
	User *userrepo.UserBaseinfoRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		User: userrepo.NewUserBaseinfoRepo(db),
	}
}
