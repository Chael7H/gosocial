package mysql

import "errors"

var (
	ErrEmailExists        = errors.New("该邮箱已被注册")
	ErrorUserNotExist     = errors.New("该用户名不存在")
	ErrorInvalidPassword  = errors.New("密码错误")
	ErrorInvalidId        = errors.New("无效的ID")
	ErrorSystem           = errors.New("系统错误")
	ErrorDBSelect         = errors.New("数据库查询错误")
	ErrorGenerateToken    = errors.New("生成token失败")
	ErrorIsFriend         = errors.New("已经是好友")
	ErrorCannotAddSelf    = errors.New("不能添加自己为好友")
	ErrorCannotDeleteSelf = errors.New("不能删除自己")
	ErrorIsNotFriend      = errors.New("该用户不是您的好友")
	ErrorInvalidParam     = errors.New("无效的参数")
)
