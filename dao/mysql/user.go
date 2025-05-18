package mysql

import (
	"errors"
	"gorm.io/gorm"
	"gosocial/models"
)

// CheckEmail 检查邮箱是否唯一
func CheckEmail(email string) (err error) {
	var exists bool

	// 高性能查询：利用唯一索引加速
	err = db.Model(&models.User{}).
		Select("1").
		Where("email = ?", email).
		Limit(1).
		Find(&exists).
		Error

	// 明确处理「记录不存在」的情况
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil // 无记录代表邮箱可用
	}

	// 数据库异常处理
	if err != nil {
		return ErrorDBSelect
	}

	// 存在记录时返回业务错误
	if exists {
		return ErrEmailExists
	}
	return nil
}

// SaveUser 保存用户信息
func SaveUser(user *models.User) (err error) {
	return db.Create(user).Error
}

// GetUserByEmail 通过邮箱获取用户信息
func GetUserByEmail(email string) (*models.User, error) {
	var user models.User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrorUserNotExist
		}
		return nil, result.Error
	}
	return &user, nil
}

// GetUserByUID 通过UID查询用户信息
func GetUserByUID(uid int64) (*models.User, error) {
	var user models.User
	result := db.Where("user_id = ?", uid).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrorUserNotExist
		}
		return nil, result.Error
	}
	return &user, nil
}

// IsUserExist 判断用户是否存在
func IsUserExist(uid int64) error {
	_, err := GetUserByUID(uid)
	if errors.Is(err, ErrorUserNotExist) {
		return ErrorUserNotExist
	}
	return err
}

func UpdateUserLastLogin(uid int64) error {
	return db.Model(&models.User{}).
		Where("user_id = ?", uid).
		UpdateColumn("last_login", gorm.Expr("now()")).Error
}

// UpdateUserInfo 更新用户信息
func UpdateUserInfo(uid int64, updateData map[string]interface{}) error {
	if len(updateData) == 0 {
		return nil
	}
	return db.Model(&models.User{}).
		Where("user_id = ?", uid).
		Updates(updateData).Error
}

// UpdateUserPassword 更新用户密码
func UpdateUserPassword(uid int64, newPassword string) error {
	return db.Model(&models.User{}).
		Where("user_id = ?", uid).
		Update("password_hash", newPassword).Error
}

func UpdateAge(uid int64, age int) error {
	return db.Model(&models.User{}).
		Where("user_id = ?", uid).
		Update("age", age).Error
}
