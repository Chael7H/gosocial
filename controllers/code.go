package controllers

type ResCode int64

// 自定义错误码
const (
	CodeSuccess ResCode = 1000 + iota
	CodeInvalidParam
	CodeUserExist
	CodeUserNotExist
	CodeEmailExist
	CodeInvalidPassword
	CodeServerBusy
	CodeIsFriend
	CodeCannotAddSelf
	CodeIsNotFriend
	CodePasswordTooShort
	CodePasswordTooLong

	CodePasswordSame
	CodeNeedLogin
	CodeInvalidToken

	CodeMessageSendFail
	CodeTooManyImages
	CodeInvalidImageFormat
	CodeCannotDeleteOthersPost
)

var CodeMsg = map[ResCode]string{
	CodeSuccess:          "success",
	CodeInvalidParam:     "请求参数错误",
	CodeUserExist:        "用户名已存在",
	CodeUserNotExist:     "用户不存在",
	CodeEmailExist:       "该邮箱已被注册",
	CodeInvalidPassword:  "密码错误",
	CodeServerBusy:       "服务繁忙",
	CodeIsFriend:         "该用户已是您好友",
	CodeCannotAddSelf:    "不能添加自己为好友",
	CodeIsNotFriend:      "该用户不是您的好友",
	CodePasswordTooShort: "密码长度不能少于6位",
	CodePasswordTooLong:  "密码长度不能超过20位",

	CodePasswordSame: "新密码与旧密码相同",
	CodeNeedLogin:    "请先登录",
	CodeInvalidToken: "登录认证失效，请重新登录",

	CodeMessageSendFail:        "消息发送失败",
	CodeTooManyImages:          "最多上传9张图片",
	CodeInvalidImageFormat:     "图片格式不正确",
	CodeCannotDeleteOthersPost: "只能删除自己的动态",
}

func (c ResCode) Msg() string {
	msg, ok := CodeMsg[c]
	if !ok {
		msg = CodeMsg[CodeServerBusy]
	}
	return msg
}
