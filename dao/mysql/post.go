package mysql

import (
	"gorm.io/gorm"
	"gosocial/models"
)

// GetPostsByUserIDs 获取多个用户的动态,按动态发布时间降序排序
func GetPostsByUserIDs(userIDs []int64, offset, limit int) ([]models.Post, error) {
	var posts []models.Post
	err := db.Where("user_id IN ?", userIDs).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// GetPostsByUserID 获取单个用户的动态,按动态发布时间降序排序
func GetPostsByUserID(userID int64, offset, limit int) ([]models.Post, error) {
	var posts []models.Post
	err := db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&posts).Error
	if err != nil {
		return nil, err
	}
	return posts, nil
}

// IncrementPostViewCount 增加动态浏览量
func IncrementPostViewCount(postID int64) error {
	return db.Model(&models.Post{}).
		Where("id = ?", postID).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

// GetRemarkByPost 获取用户对好友动态的备注
func GetRemarkByPost(userID int64, post models.Post) (remark string, err error) {
	//若是自己的动态，则直接返回自己的Username
	if userID == post.UserID {
		return post.Username, nil
	}
	var friendship models.Friendship
	err = db.Select("remark").Where("user_id = ? AND friend_id = ?", userID, post.UserID).
		First(&friendship).Error
	if err != nil {
		return "", err
	}
	if friendship.Remark == "" {
		return post.Username, nil
	}
	return friendship.Remark, nil
}

// CreatePost 创建用户动态
func CreatePost(post *models.Post) error {
	return db.Create(post).Error
}

func DeletePostByUser(userID int64) error {
	var post models.Post
	err := db.Where("user_id = ?", userID).Delete(&post).Error
	if err != nil {
		return err
	}
	return nil
}
