package models

import "time"

type Message struct {
	ID          int64     `gorm:"primaryKey" json:"id"` // 消息ID
	From        int64     `json:"from,string"`          // 发送者ID
	To          int64     `json:"to,string"`            // 接收者ID
	Content     string    `json:"content"`              // 消息内容
	Type        int       `json:"type"`                 // 消息类型(1:文本 2:图片 3:文件)
	FileURL     string    `json:"file_url"`             // 文件URL
	CreatedAt   time.Time `json:"created_at"`           // 创建时间
	Status      string    `json:"status"`               // 消息状态（sent/delivered/read）
	IsPersisted bool      `json:"is_persisted"`         // 是否已持久化到数据库
	HideTime    bool      `json:"hide_time"`            // 是否隐藏时间显示

	// 关联发送者的用户信息（非数据库字段）
	My User `gorm:"foreignKey:From;references:UserID"`
}

type FileMeta struct {
	Name   string `json:"name"`   // 文件名
	Size   int64  `json:"size"`   // 文件大小
	Width  int    `json:"width"`  //图片宽度
	Height int    `json:"height"` //图片高度
	Type   string `json:"type"`   // 文件类型
	URL    string `json:"url"`    // 文件URL
}
