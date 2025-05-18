package models

import "time"

// ParamRegister  用户注册参数结构体
type ParamRegister struct {
	Username   string `json:"username" binding:"required"`
	Password   string `json:"password" binding:"required,min=6,max=20"`
	RePassword string `json:"re_password" binding:"required,eqfield=Password" `
	Email      string `json:"email" binding:"required,email"`
}

// ParamLogin  用户登录参数结构体
type ParamLogin struct {
	Identifier string `json:"identifier" binding:"required"` // 登录标识（邮箱或用户ID）
	Password   string `json:"password" binding:"required"`
}

// ParamFriendItem	好友列表参数结构体
type ParamFriendItem struct {
	FriendID       int64     `json:"friend_id,string"`
	DisplayName    string    `json:"display_name"` // 优先显示备注，否则显示昵称
	AvatarURL      string    `json:"avatar_url"`
	LastInteractAt time.Time `json:"last_interact_at"`
}

// ParamTextReq  发送文本消息模型结构体
type ParamTextReq struct {
	To      int64  `json:"to,string" binding:"required"`
	Content string `json:"content" binding:"required"`
}

// ParamImageReq  发送图片消息模型结构体
type ParamImageReq struct {
	To      int64  `json:"to,string" binding:"required"`
	Content string `json:"content" binding:"required"` // 图片URL
	Width   int    `json:"width"`                      // 图片宽度(像素)
	Height  int    `json:"height"`                     // 图片高度(像素)
}

// ParamFileReq  发送文件消息模型结构体
type ParamFileReq struct {
	To      int64  `json:"to,string" binding:"required"`
	Content string `json:"content" binding:"required"` // 文件URL
	Name    string `json:"name"`                       // 文件名
	Size    int64  `json:"size"`                       // 文件大小(字节)
	Type    string `json:"type"`                       // 文件类型
}

// ParamFriendAdd  添加好友模型
type ParamFriendAdd struct {
	FriendID int64 `json:"friend_id,string" binding:"required"`
}

// ParamPostWithUserInfo 包含用户信息的动态
type ParamPostWithUserInfo struct {
	ID        int64     `json:"id,string"`
	ViewCount uint64    `json:"view_count"` // 浏览量
	Avatar    string    `json:"avatar"`     // 用户头像
	Nickname  string    `json:"nickname"`   // 用户备注或用户名
	Content   string    `json:"content"`    // 文字内容
	Images    string    `json:"images"`     // 图片URL，多个用逗号分隔
	CreatedAt time.Time `json:"created_at"` // 发布时间
}

// ParamUserInfoResponse 用户信息响应结构
type ParamUserInfoResponse struct {
	AvatarURL string `json:"avatar_url"`
	Username  string `json:"username"`
	UserID    int64  `json:"user_id,string"`
	Gender    string `json:"gender"`
	Birthday  string `json:"birthday"` // 格式: xx-xx
	Age       int    `json:"age"`
	Signature string `json:"signature"`
	LastLogin string `json:"last_login"`
}

// ParamFriendInfoResponse 好友信息响应结构
type ParamFriendInfoResponse struct {
	AvatarURL string     `json:"avatar_url"`
	Username  string     `json:"username"`
	Remark    string     `json:"remark"`
	UserID    int64      `json:"user_id,string"`
	Gender    string     `json:"gender"`
	Birthday  string     `json:"birthday"` // 格式: xx-xx
	Age       int        `json:"age"`
	Signature string     `json:"signature"`
	LastLogin *time.Time `json:"last_login"`
}

// ParamUpdateUserInfoRequest 更新用户信息请求结构
type ParamUpdateUserInfoRequest struct {
	Username  string `json:"username" binding:"omitempty,min=2,max=20"`
	Gender    string `json:"gender" binding:"omitempty,oneof=male female"`
	Signature string `json:"signature" binding:"omitempty,max=255"`
	AvatarURL string `json:"avatar_url" binding:"omitempty,url"`
	Birthday  string `json:"birthday" binding:"omitempty"`
}

// ParamUpdatePasswordRequest 更新密码请求结构
type ParamUpdatePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required,min=6,max=20"`
	NewPassword string `json:"new_password" binding:"required,min=6,max=20"`
}
