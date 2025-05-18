package logic

import (
	"errors"
	"gosocial/dao/mysql"
	"gosocial/models"
	"time"
)

// GetFriendPosts 获取好友动态列表
func GetFriendPosts(userID int64, offset int) ([]models.ParamPostWithUserInfo, error) {
	// 1. 获取好友列表
	friends, err := GetFriendInfo(userID)
	if err != nil {
		return nil, err
	}
	// 2. 获取好友ID列表
	friendIDs := make([]int64, len(friends))
	for i, friend := range friends {
		friendIDs[i] = friend.FriendID
	}
	postIDs := append(friendIDs, userID)
	// 3. 获取好友和自己的动态并附上备注
	posts, err := mysql.GetPostsByUserIDs(postIDs, offset, 10)

	if err != nil {
		return nil, err
	}
	for k, _ := range posts {
		posts[k].Username, err = GetRemarkByPost(userID, posts[k])
		if err != nil {
			return nil, err
		}
	}

	// 4. 补充动态的简略个人信息
	var result []models.ParamPostWithUserInfo
	for _, post := range posts {
		result = append(result, ToParamPostWithUserInfo(post))
	}
	return result, nil
}

// GetUserPosts 获取指定用户动态列表
func GetUserPosts(currentUserID, targetUserID int64, offset int) ([]models.ParamPostWithUserInfo, error) {
	var result []models.ParamPostWithUserInfo
	//1.检查是否该用户为自己,若是，则直接获取动态
	if currentUserID == targetUserID {
		posts, err := mysql.GetPostsByUserID(targetUserID, offset, 10)
		if err != nil {
			return nil, err
		}
		//  补充用户信息
		for _, post := range posts {
			result = append(result, ToParamPostWithUserInfo(post))
		}
		return result, nil
	}

	// 2. 检查是否是好友关系，非好友返回错误
	if err := mysql.IsFriend(currentUserID, targetUserID); !errors.Is(err, mysql.ErrorIsFriend) {
		return nil, mysql.ErrorIsNotFriend
	}

	// 3. 获取用户动态
	posts, err := mysql.GetPostsByUserID(targetUserID, offset, 10)
	if err != nil {
		return nil, err
	}
	for k, _ := range posts {
		posts[k].Username, err = GetRemarkByPost(currentUserID, posts[k])
		if err != nil {
			return nil, err
		}
	}
	//  补充用户信息
	for _, post := range posts {
		result = append(result, ToParamPostWithUserInfo(post))
	}
	return result, nil
}

// ToParamPostWithUserInfo 封装动态信息与用户信息
func ToParamPostWithUserInfo(post models.Post) models.ParamPostWithUserInfo {
	return models.ParamPostWithUserInfo{
		ViewCount: post.ViewCount,
		Avatar:    post.AvatarURL,
		Nickname:  post.Username,
		Content:   post.Content,
		Images:    post.Images,
		CreatedAt: post.CreatedAt,
	}
}

// GetRemarkByPost 获取好友备注
func GetRemarkByPost(userID int64, post models.Post) (string, error) {
	return mysql.GetRemarkByPost(userID, post)
}

// IncrementPostViewCount 增加动态浏览量
func IncrementPostViewCount(postID int64) error {
	return mysql.IncrementPostViewCount(postID)
}

// CreatePost 创建用户动态
func CreatePost(userID int64, content, images string) (*models.Post, error) {
	// 1. 验证用户存在性
	user, err := mysql.GetUserByUID(userID)
	if err != nil {
		return nil, err
	}

	// 2. 创建动态
	post := models.Post{
		UserID:    userID,
		Username:  user.Username,
		AvatarURL: user.AvatarURL,
		Content:   content,
		Images:    images,
		ViewCount: 0,
		CreatedAt: time.Now(),
	}

	// 3. 保存到数据库
	if err = mysql.CreatePost(&post); err != nil {
		return nil, err
	}

	return &post, nil
}

func DeletePostByUser(userID int64) error {
	return mysql.DeletePostByUser(userID)
}
