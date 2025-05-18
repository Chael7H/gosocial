package models

type Response struct {
	Code int         `json:"code" example:"200"`    // 状态码
	Msg  string      `json:"msg" example:"success"` // 消息描述
	Data interface{} `json:"data"`                  // 业务数据（如UID字符串）
}
