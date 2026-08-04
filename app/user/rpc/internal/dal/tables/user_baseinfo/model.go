package user_baseinfo

import (
	"time"
)

// UserBaseinfo 用户基础信息数据库模型
type UserBaseinfo struct {
	UserID    int64     `gorm:"primaryKey;type:bigint;column:user_id"`
	Username  string    `gorm:"type:varchar(64);column:username"`
	Password  string    `gorm:"not null;type:varchar(64);column:password"`
	PhotoURL  string    `gorm:"not null;type:varchar(255);column:photo_url"`
	Status    int8      `gorm:"not null;default:1;type:tinyint;column:status"`
	CreatedAt time.Time `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime;column:updated_at"`
}

func (UserBaseinfo) TableName() string {
	return "user_baseinfo"
}

// UserMFA 用户MFA信息数据库模型
type UserMFA struct {
	UserID           int64  `gorm:"primaryKey;type:bigint;column:user_id"`
	MFASecret        string `gorm:"not null;type:varchar(64);column:mfa_secret"`
	MFAEnabled       bool   `gorm:"default:false;type:boolean;column:mfa_enabled"`
	PasswordHash     string `gorm:"not null;type:varchar(64);column:password_hash"`
	MFAPendingSecret string `gorm:"type:varchar(64);column:mfa_pending_secret"`
}

func (UserMFA) TableName() string {
	return "user_mfa"
}
