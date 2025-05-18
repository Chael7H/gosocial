package models

import "time"

// Friendship 好友关系模型
type Friendship struct {
	ID             int64     `gorm:"primaryKey;autoIncrement"`
	UserID         int64     `gorm:"index;not null;comment:用户ID" json:"user_id,string"`
	FriendID       int64     `gorm:"index;not null;comment:好友ID" json:"friend_id,string"`
	Remark         string    `gorm:"type:varchar(50);default:'';comment:好友备注"`
	LastInteractAt time.Time `gorm:"index;comment:最后互动时间"`
	CreatedAt      time.Time `gorm:"autoCreateTime;comment:添加时间"`

	// 关联好友的用户信息（非数据库字段）
	Friend User `gorm:"foreignKey:FriendID;references:UserID"`
}
