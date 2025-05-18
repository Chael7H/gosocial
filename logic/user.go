package logic

import (
	"errors"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gosocial/dao/mysql"
	"gosocial/models"
	"gosocial/pkg/jwt"
	"gosocial/pkg/snowflake"
	"strconv"
	"strings"
	"time"
)

// RegisterLogic 注册逻辑函数
// func RegisterLogic(user *models.ParamRegister) (uid int64, err error) {
var RegisterLogic = func(user *models.ParamRegister) (uid int64, err error) {
	//检查邮箱是否唯一
	if err = mysql.CheckEmail(user.Email); err != nil {
		return 0, err
	}
	//使用雪花算法生成uid并保存到数据库
	//var uid int64
	if uid, err = snowflake.GenID(); err != nil {
		zap.L().Error("生成uid失败", zap.Error(err))
		return 0, mysql.ErrorSystem
	}
	//对密码进行BCrypt哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		zap.L().Error("密码加密失败", zap.Error(err))
		return 0, mysql.ErrorSystem
	}
	//保存用户信息
	userDetail := &models.User{
		UserID:       uid,
		Username:     user.Username,
		PasswordHash: string(hashedPassword),
		Email:        user.Email,
		AvatarURL:    "https://s3.bmp.ovh/imgs/2025/05/04/e272b0b155df44bd.png", // 设置默认头像URL
	}

	if err = mysql.SaveUser(userDetail); err != nil {
		zap.L().Error("注册失败", zap.Error(err))
		return 0, mysql.ErrorSystem
	}
	return
}

// LoginLogic 登录逻辑函数
func LoginLogic(p *models.ParamLogin) (*models.User, error) {
	// 根据输入类型选择查询方式
	var user *models.User
	var err error
	if isEmail(p.Identifier) {
		user, err = mysql.GetUserByEmail(p.Identifier)
	} else { // UID登录（需解析字符串为int64）
		uid, parseErr := strconv.ParseInt(p.Identifier, 10, 64)
		if parseErr != nil {
			// UID格式无效
			return nil, mysql.ErrorInvalidId
		}
		user, err = mysql.GetUserByUID(uid)
	}
	// 处理用户查询错误
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExist) {
			return nil, mysql.ErrorUserNotExist
		}
		return nil, mysql.ErrorSystem
	}
	// 验证密码
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(p.Password)); err != nil {
		return nil, mysql.ErrorInvalidPassword
	}
	// 生成JWT Token
	token, err := jwt.GenToken(uint64(user.UserID), user.Username)
	if err != nil {
		zap.L().Error("生成Token失败", zap.Error(err))
		return nil, mysql.ErrorGenerateToken
	}
	zap.L().Info("生成的Token", zap.String("token", token))
	user.Token = token
	zap.L().Info("赋值后的User", zap.Any("user", user))

	// 更新用户登录时间
	if err = mysql.UpdateUserLastLogin(user.UserID); err != nil {
		return nil, mysql.ErrorSystem
	}
	return user, nil
}

// GetUserInfoLogic 获取用户信息逻辑
func GetUserInfoLogic(userID int64) (*models.User, error) {
	// 从数据库获取用户信息
	user, err := mysql.GetUserByUID(userID)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExist) {
			return nil, mysql.ErrorUserNotExist
		}
		return nil, mysql.ErrorSystem
	}

	// 计算用户年龄
	if user.Birthday != nil {
		user.Age = calculateAge(*user.Birthday)
	}
	return user, nil
}

// UpdateUserInfoLogic 更新用户信息逻辑
func UpdateUserInfoLogic(userID int64, req models.ParamUpdateUserInfoRequest) error {
	// 准备更新数据
	updateData := make(map[string]interface{})
	if req.Username != "" {
		updateData["username"] = req.Username
	}
	if req.Gender != "" {
		updateData["gender"] = req.Gender
	}
	if req.Signature != "" {
		updateData["signature"] = req.Signature
	}
	if req.AvatarURL != "" {
		updateData["avatar_url"] = req.AvatarURL
	}
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
					return mysql.ErrorInvalidParam
				}
			}
		}
		updateData["birthday"] = birthday
	}

	// 调用DAO层更新用户信息
	if err := mysql.UpdateUserInfo(userID, updateData); err != nil {
		return mysql.ErrorSystem
	}

	return nil
}

// UpdatePasswordLogic 更新密码逻辑
func UpdatePasswordLogic(userID int64, oldPassword, newPassword string) error {
	// 获取用户当前密码
	user, err := mysql.GetUserByUID(userID)
	if err != nil {
		return err
	}

	// 验证旧密码
	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return mysql.ErrorInvalidPassword
	}

	// 对新密码进行哈希处理
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return mysql.ErrorSystem
	}

	// 更新密码
	if err = mysql.UpdateUserPassword(userID, string(hashedPassword)); err != nil {
		return mysql.ErrorSystem
	}

	return nil
}

// calculateAge 根据生日计算年龄
func calculateAge(birthday time.Time) int {
	now := time.Now()
	years := now.Year() - birthday.Year()
	if now.Month() < birthday.Month() || (now.Month() == birthday.Month() && now.Day() < birthday.Day()) {
		years--
	}
	return years
}

// 简单邮箱格式验证
func isEmail(input string) bool {
	return strings.Contains(input, "@")
}

func UpdateAge(userId int64, age int) error {
	return mysql.UpdateAge(userId, age)
}

// UpdateUserAvatar 更新用户头像URL
func UpdateUserAvatar(userID int64, avatarURL string) error {
	updateData := map[string]interface{}{
		"avatar_url": avatarURL,
	}
	return mysql.UpdateUserInfo(userID, updateData)
}
