package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gosocial/dao/mysql"
	"gosocial/logic"
	"gosocial/models"
	"path/filepath"
	"strings"
	"time"
)

// RegisterHandler 用户注册业务接口
// @Summary 用户注册接口
// @Description 用户通过邮箱、用户名和密码进行注册，返回生成的唯一UID
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param   body  body   models.ParamRegister  true  "注册参数"
// @Security ApiKeyAuth
// @Success 200 {object}  int "用户uid"
// @Failure 400 {object}  models.Response "参数格式错误（具体错误字段提示）"
// @Failure 422 {object}  models.Response "该邮箱已被注册"
// @Failure 500 {object}  models.Response "服务器内部错误"
// @Router /api/v1/register [post]
func RegisterHandler(c *gin.Context) {
	// 1. 获取参数及校验
	user := new(models.ParamRegister)
	err := c.ShouldBind(user)
	if err != nil {
		zap.L().Error("SignUp with invalid param", zap.Error(err))
		//判断err是不是validator.ValidatorErrors 类型
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if !ok {
			zap.L().Error("SignUp with invalid param", zap.Error(err))
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans))) //翻译错误
		return
	}
	//2.业务处理
	var uid int64
	if uid, err = logic.RegisterLogic(user); err != nil {
		zap.L().Error("RegisterLogic failed", zap.Error(err))
		switch {
		case errors.Is(err, mysql.ErrEmailExists):
			ResponseError(c, CodeEmailExist)
			return
		default:
			ResponseError(c, CodeServerBusy)
			return
		}
	}
	//3.返回
	ResponseSuccess(c, fmt.Sprintf("用户注册成功，您的UID为：%d", uid))
}

// LoginHandler 处理用户登录
// @Summary 用户登录接口
// @Description 用户选择uid或者邮箱两种方式并输入自己的密码进行登录，返回用户信息及token
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param   body  body   models.ParamLogin  true  "登录参数"
// @Security ApiKeyAuth
// @Success 200 {object}  models.Response "登录成功，返回用户信息和令牌"
// @Failure 400 {object}  models.Response "参数格式错误（具体错误字段提示）"
// @Failure 404 {object}  models.Response "用户不存在"
// @Failure 401 {object}  models.Response "密码错误"
// @Failure 500 {object}  models.Response "服务器内部错误"
// @Router /api/v1/login [post]
func LoginHandler(c *gin.Context) {
	//1.获取参数及校验
	p := new(models.ParamLogin)
	err := c.ShouldBind(p)
	if err != nil {
		//请求参数有误，直接返回响应
		zap.L().Error("Login with invalid param", zap.Error(err))
		//判断err是不是validator.ValidatorErrors 类型
		var errs validator.ValidationErrors
		ok := errors.As(err, &errs)
		if !ok {
			zap.L().Error("SignUp with invalid param", zap.Error(err))
			ResponseError(c, CodeInvalidParam)
			return
		}
		ResponseErrorWithMsg(c, CodeInvalidParam, removeTopStruct(errs.Translate(trans)))
		return
	}
	//2.业务逻辑处理
	user, err := logic.LoginLogic(p)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExist) {
			ResponseError(c, CodeUserNotExist)
			return
		} else if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			ResponseError(c, CodeInvalidPassword)
			return
		} else if errors.Is(err, mysql.ErrorInvalidPassword) {
			ResponseError(c, CodeInvalidPassword)
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 确保 user 非空
	if user == nil {
		zap.L().Error("用户名不存在")
		ResponseError(c, CodeUserNotExist)
		return
	}

	ResponseSuccess(c, gin.H{
		"message":    fmt.Sprintf("欢迎用户 %s", user.Username),
		"user_id":    fmt.Sprintf("%d", user.UserID), //转化为string类型防止前端数据溢出(JSON的maxInt小于int64)
		"username":   user.Username,
		"token":      user.Token,
		"avatar_url": user.AvatarURL, // 添加头像URL返回
	})
}

// GetUserInfoHandler 获取用户信息
// @Summary 获取用户信息接口
// @Description 获取当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.ParamUserInfoResponse "成功获取用户信息"
// @Failure 401 {object} models.Response "未授权"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/user/info [get]
func GetUserInfoHandler(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		zap.L().Error("GetUserInfoHandler: user_id not found in context")
		ResponseError(c, CodeInvalidToken)
		return
	}

	// 调用业务逻辑获取用户信息
	user, err := logic.GetUserInfoLogic(userID.(int64))
	if err != nil {
		zap.L().Error("GetUserInfoLogic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 格式化生日为xx-xx格式
	birthday := ""
	if user.Birthday != nil {
		birthday = user.Birthday.Format("01-02")
	}

	// 格式化最后登录时间
	lastLogin := ""
	if user.LastLogin != nil {
		lastLogin = user.LastLogin.Format("2006-01-02 15:04:05")
	}

	// 返回响应
	ResponseSuccess(c, models.ParamUserInfoResponse{
		AvatarURL: user.AvatarURL,
		Username:  user.Username,
		UserID:    user.UserID,
		Gender:    user.Gender,
		Birthday:  birthday,
		Age:       user.Age,
		Signature: user.Signature,
		LastLogin: lastLogin,
	})
}

// UpdateUserInfoHandler 更新用户信息
// @Summary 更新用户信息接口
// @Description 更新当前登录用户的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param   body  body   models.ParamUpdateUserInfoRequest  true  "更新参数"
// @Security ApiKeyAuth
// @Success 200 {object} models.Response "更新成功"
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 401 {object} models.Response "未授权"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/user/update_info [put]
func UpdateUserInfoHandler(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		zap.L().Error("UpdateUserInfoHandler: user_id not found in context")
		ResponseError(c, CodeInvalidToken)
		return
	}

	// 解析请求参数
	var req models.ParamUpdateUserInfoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("UpdateUserInfo with invalid param", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 处理生日字段并计算年龄
	if req.Birthday != "" {
		// 尝试多种日期格式解析
		var birthday time.Time
		var err error

		// 尝试ISO 8601格式 (2006-01-02T15:04:05Z)
		birthday, err = time.Parse(time.RFC3339, req.Birthday)
		if err != nil {
			// 尝试短横线分隔格式 (2006-01-02)
			birthday, err = time.Parse("2006-01-02", req.Birthday)
			if err != nil {
				// 尝试斜杠分隔格式 (2006/01/02)
				birthday, err = time.Parse("2006/01/02", req.Birthday)
				if err != nil {
					ResponseError(c, CodeInvalidParam)
					return
				}
			}
		}
		// 计算年龄（必须满1年才能+1岁）
		now := time.Now()
		years := now.Year() - birthday.Year()
		if now.Month() < birthday.Month() || (now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
			years--
		}

		_ = logic.UpdateAge(userID.(int64), years)
	}

	// 调用业务逻辑更新用户信息
	if err := logic.UpdateUserInfoLogic(userID.(int64), req); err != nil {
		zap.L().Error("UpdateUserInfoLogic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, "更新成功!")
}

// UploadAvatarHandler 上传头像
// @Summary 上传头像接口
// @Description 上传用户头像文件并更新用户信息
// @Tags 用户管理
// @Accept multipart/form-data
// @Produce json
// @Param avatar formData file true "头像文件"
// @Security ApiKeyAuth
// @Success 200 {object} models.Response "上传成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/user/upload_avatar [post]
func UploadAvatarHandler(c *gin.Context) {
	// 获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		zap.L().Error("UploadAvatarHandler: user_id not found in context")
		ResponseError(c, CodeInvalidToken)
		return
	}

	// 获取上传文件
	file, err := c.FormFile("avatar")
	if err != nil {
		zap.L().Error("UploadAvatarHandler: get file failed", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 验证文件类型
	if !strings.HasSuffix(file.Filename, ".jpg") && !strings.HasSuffix(file.Filename, ".png") {
		zap.L().Error("UploadAvatarHandler: invalid file type", zap.String("filename", file.Filename))
		ResponseError(c, CodeInvalidImageFormat)
		return
	}

	// 生成唯一文件名
	fileName := fmt.Sprintf("%d_%d%s", userID.(int64), time.Now().Unix(), filepath.Ext(file.Filename))
	savePath := filepath.Join("static", "images", fileName)

	// 保存文件
	if err = c.SaveUploadedFile(file, savePath); err != nil {
		zap.L().Error("UploadAvatarHandler: save file failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 更新用户头像URL
	avatarURL := fmt.Sprintf("/static/images/%s", fileName)
	if err = logic.UpdateUserAvatar(userID.(int64), avatarURL); err != nil {
		zap.L().Error("UpdateUserAvatar failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, gin.H{
		"message":    "上传成功",
		"avatar_url": avatarURL,
	})
}

// UpdatePasswordHandler 更新密码
// @Summary 更新密码接口
// @Description 更新当前登录用户的密码
// @Tags 用户管理
// @Accept json
// @Produce json
// @Param   body  body   models.ParamUpdatePasswordRequest  true  "更新密码参数"
// @Security ApiKeyAuth
// @Success 200 {object} models.Response "密码更新成功"
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 401 {object} models.Response "未授权或旧密码错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/user/update_password [put]
func UpdatePasswordHandler(c *gin.Context) {
	// 从上下文中获取用户ID
	userID, exists := c.Get("uid")
	if !exists {
		zap.L().Error("UpdatePasswordHandler: user_id not found in context")
		ResponseError(c, CodeInvalidToken)
		return
	}

	// 解析请求参数
	var req models.ParamUpdatePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("UpdatePasswordWith invalid param", zap.Error(err))
		flag := len(req.NewPassword)
		switch {
		case flag < 6:
			ResponseError(c, CodePasswordTooShort)
			return
		case flag > 16:
			ResponseError(c, CodePasswordTooLong)
			return
		}
		ResponseError(c, CodeInvalidParam)
		return
	}
	//两次密码不能一致
	if req.NewPassword == req.OldPassword {
		ResponseError(c, CodePasswordSame)
	}

	// 调用业务逻辑更新密码
	if err := logic.UpdatePasswordLogic(userID.(int64), req.OldPassword, req.NewPassword); err != nil {
		if errors.Is(err, mysql.ErrorInvalidPassword) {
			zap.L().Error("UpdatePasswordLogic failed", zap.Error(err))
			ResponseError(c, CodeInvalidPassword)
		} else {
			zap.L().Error("UpdatePasswordLogic failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
		}
		return
	}
	ResponseSuccess(c, "密码修改成功")
}
