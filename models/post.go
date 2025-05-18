package models

import (
	"time"
)

// Post 用户动态模型
type Post struct {
	ID        int64     `gorm:"primaryKey" json:"-"`             // 动态ID
	UserID    int64     `json:"user_id,string"`                  // 发布用户ID
	Username  string    `json:"username"`                        // 发布用户昵称
	AvatarURL string    `json:"avatar_url"`                      //发布用户头像
	Content   string    `gorm:"type:text" json:"content"`        // 文字内容
	Images    string    `gorm:"type:varchar(255)" json:"images"` // 图片URL，多个用逗号分隔
	ViewCount uint64    `json:"view_count"`                      // 浏览量
	CreatedAt time.Time `json:"created_at"`                      // 发布时间
}
