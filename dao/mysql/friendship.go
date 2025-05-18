package mysql

import (
	"gorm.io/gorm"
	"gosocial/models"
	"time"
)

// AddFriend 添加好友时插入双向记录
func AddFriend(userID, friendID int64) error {
	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&models.Friendship{
			UserID:         userID,
			FriendID:       friendID,
			LastInteractAt: time.Now(),
		}).Error; err != nil {
			return err
		}
		return tx.Create(&models.Friendship{
			UserID:         friendID,
			FriendID:       userID,
			LastInteractAt: time.Now(),
		}).Error
	})
}

// GetFriendList 获取好友列表，按最后互动时间降序排序
func GetFriendList(userID int64) ([]models.Friendship, error) {
	var friends []models.Friendship
	err := db.Preload("Friend").
		Where("user_id = ?", userID).
		Order("last_interact_at DESC").
		Find(&friends).Error
	return friends, err
}

// IsFriend 判断是否为好友
func IsFriend(userID, friendID int64) error {
	err := db.Where("user_id = ? AND friend_id = ?", userID, friendID).
		First(&models.Friendship{}).Error
	if err != nil {
		return nil
	}
	return ErrorIsFriend
}

// DeleteFriend 双向删除好友
func DeleteFriend(userID, friendID int64) error {
	err := db.Where("user_id = ? AND friend_id = ?", userID, friendID).
		Delete(&models.Friendship{}).Error
	if err != nil {
		return err
	}
	err = db.Where("user_id = ? AND friend_id = ?", friendID, userID).
		Delete(&models.Friendship{}).Error
	return err
}

func UpdateFriendRemark(userID, friendID int64, remark string) error {
	//如果备注为空，则删除备注
	if remark == "" {
		return db.Model(&models.Friendship{}).
			Where("user_id = ? AND friend_id = ?", userID, friendID).
			Update("remark", nil).Error
	}
	//备注不为空，更新备注
	return db.Model(&models.Friendship{}).
		Where("user_id = ? AND friend_id = ?", userID, friendID).
		Update("remark", remark).Error
}
