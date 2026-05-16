package user_baseinfo

import (
	"time"
)

// UserBaseinfo 用户基础信息数据库模型
type UserBaseinfo struct {
	UserID       string     `gorm:"primaryKey;type:varchar(64);column:user_id"`
	Username     string     `gorm:"unique;type:varchar(64);column:username"`
	Password     string     `gorm:"not null;type:varchar(255);column:password"`
	PhotoURL     string     `gorm:"not null;type:varchar(255);column:photo_url"`
	JwchID       string     `gorm:"default:null;type:varchar(10);column:jwch_id"`
	JwchPassword string     `gorm:"default:null;type:varchar(15);column:jwch_password"`
	CreatedAt    time.Time  `gorm:"autoCreateTime;column:created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime;column:updated_at"`
	DeletedAt    *time.Time `gorm:"column:deleted_at"`
}

func (UserBaseinfo) TableName() string {
	return "user_baseinfo"
}

// UserMFA 用户MFA信息数据库模型
type UserMFA struct {
	UserID           string `gorm:"primaryKey;type:varchar(64);column:user_id"`
	MFASecret        string `gorm:"not null;type:varchar(255);column:mfa_secret"`
	MFAEnabled       bool   `gorm:"default:false;type:boolean;column:mfa_enabled"`
	PasswordHash     string `gorm:"not null;type:varchar(255);column:password_hash"`
	MFAPendingSecret string `gorm:"type:varchar(255);column:mfa_pending_secret"`
}

func (UserMFA) TableName() string {
	return "user_mfa"
}
