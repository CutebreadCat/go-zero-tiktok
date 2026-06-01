package user_baseinfo

import "time"

type UserBaseinfo struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	Username  string    `gorm:"column:username"`
	Password  string    `gorm:"column:password"`
	PhotoURL  string    `gorm:"column:photo_url"`
	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (UserBaseinfo) TableName() string {
	return "user_baseinfo"
}
