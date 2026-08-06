package user_baseinfo

import (
	"time"
)

// UserBaseinfo 用户基础信息数据库模型(含 MFA,重构后 user_mfa 合并入本表,见 migrations)
type UserBaseinfo struct {
	UserID           int64     `gorm:"primaryKey;type:bigint;column:user_id"`
	Username         string    `gorm:"type:varchar(64);column:username"`
	Password         string    `gorm:"not null;type:varchar(64);column:password"`
	PhotoURL         string    `gorm:"not null;type:varchar(255);column:photo_url"`
	MFASecret        string    `gorm:"type:varchar(64);column:mfa_secret"`
	MFAEnabled       bool      `gorm:"not null;default:false;type:tinyint(1);column:mfa_enabled"`
	MFAPendingSecret string    `gorm:"type:varchar(64);column:mfa_pending_secret"`
	Status           int8      `gorm:"not null;default:1;type:tinyint;column:status"`
	CreatedAt        time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt        time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (UserBaseinfo) TableName() string {
	return "user_baseinfo"
}