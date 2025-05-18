package logic

import (
	"errors"
	"gosocial/dao/mysql"
	"gosocial/models"
	"sort"
	"strings"
)

// GetFriendList 获取好友列表
func GetFriendList(userID int64) (friendList []models.ParamFriendItem, err error) {
	// 查询数据库
	friendships, err := mysql.GetFriendList(userID)
	if err != nil {
		return nil, err
	}
	// 转换数据格式
	for _, f := range friendships {
		friendList = append(friendList, ToFriendItem(f))
	}
	return
}

// ToFriendItem 转换 Friendship 到前端需要的 FriendItem
func ToFriendItem(f models.Friendship) models.ParamFriendItem {
	return models.ParamFriendItem{
		FriendID:       f.FriendID,
		DisplayName:    ToNickname(f),
		AvatarURL:      f.Friend.AvatarURL,
		LastInteractAt: f.LastInteractAt,
	}
}

// AddFriend 添加好友关系
func AddFriend(userID, friendID int64) error {
	// 不能添加自己为好友
	if userID == friendID {
		return mysql.ErrorCannotAddSelf
	}
	// 判断userID是否存在
	if err := mysql.IsUserExist(friendID); err != nil {
		return err
	}
	// 判断是否已经是好友
	if err := mysql.IsFriend(userID, friendID); err != nil {
		return err
	}
	// 调用dao层添加好友
	return mysql.AddFriend(userID, friendID)
}

// SearchFriend 搜索好友
func SearchFriend(userID int64, keyword string) ([]models.ParamFriendItem, error) {
	// 获取好友列表
	friendships, err := mysql.GetFriendList(userID)
	if err != nil {
		return nil, err
	}

	// 定义匹配优先级
	type matchResult struct {
		friend models.ParamFriendItem
		score  int // 0:不匹配 1:包含 2:前缀 3:完全匹配
	}

	var results []matchResult
	for _, f := range friendships {
		displayName := f.Remark
		if displayName == "" {
			displayName = f.Friend.Username
		}

		// 检查匹配情况
		var score int
		switch {
		case displayName == keyword:
			score = 3 // 完全匹配
		case len(displayName) >= len(keyword) && displayName[:len(keyword)] == keyword:
			score = 2 // 前缀匹配
		case strings.Contains(displayName, keyword):
			score = 1 // 包含匹配
		default:
			continue // 不匹配则跳过
		}

		results = append(results, matchResult{
			friend: ToFriendItem(f),
			score:  score,
		})
	}

	// 按匹配优先级排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].score > results[j].score
	})

	// 提取排序后的好友列表
	var friendList []models.ParamFriendItem
	for _, r := range results {
		friendList = append(friendList, r.friend)
	}

	return friendList, nil
}

// GetFriendInfo 获取好友列表
func GetFriendInfo(userID int64) (friendList []models.Friendship, err error) {
	// 查询数据库
	friendships, err := mysql.GetFriendList(userID)
	if err != nil {
		return nil, err
	}
	for i := 0; i < len(friendships); i++ {
		friendships[i].Remark = ToNickname(friendships[i])
	}
	return friendships, nil
}

// GetFriendInfoLogic 获取好友信息
func GetFriendInfoLogic(userID int64) (*models.ParamFriendInfoResponse, error) {
	// 从数据库获取用户信息
	user, err := mysql.GetUserByUID(userID)
	if err != nil {
		if errors.Is(err, mysql.ErrorUserNotExist) {
			return nil, mysql.ErrorUserNotExist
		}
		return nil, mysql.ErrorSystem
	}
	// 格式化生日为xx-xx格式
	birthday := ""
	if user.Birthday != nil {
		birthday = user.Birthday.Format("01-02")
	}

	// 计算用户年龄
	if user.Birthday != nil {
		user.Age = calculateAge(*user.Birthday)
	}

	return &models.ParamFriendInfoResponse{
		AvatarURL: user.AvatarURL,
		Username:  user.Username,
		UserID:    user.UserID,
		Gender:    user.Gender,
		Birthday:  birthday,
		Age:       user.Age,
		Signature: user.Signature,
		LastLogin: user.LastLogin,
	}, nil
}

// ToNickname 获取好友的昵称
func ToNickname(friend models.Friendship) string {
	displayName := friend.Remark
	if displayName == "" {
		displayName = friend.Friend.Username
	}
	return displayName
}

// DeleteFriend 删除好友
func DeleteFriend(userID int64, friendID int64) error {
	if userID == friendID {
		return mysql.ErrorCannotDeleteSelf
	}
	//检查该用户是否是您的好友
	if err := mysql.IsFriend(userID, friendID); !errors.Is(err, mysql.ErrorIsFriend) {
		return mysql.ErrorIsNotFriend
	}
	return mysql.DeleteFriend(userID, friendID)
}

// UpdateFriendRemark 更新好友备注
func UpdateFriendRemark(userID int64, friendID int64, remark string) error {
	//检查该用户是否是您的好友
	if err := mysql.IsFriend(userID, friendID); !errors.Is(err, mysql.ErrorIsFriend) {
		return mysql.ErrorIsNotFriend
	}
	//更新备注
	return mysql.UpdateFriendRemark(userID, friendID, remark)
}

// IsFriend 判断是否为好友
func IsFriend(userID int64, friendID int64) error {
	err := mysql.IsFriend(userID, friendID)
	if !errors.Is(err, mysql.ErrorIsFriend) {
		return mysql.ErrorIsNotFriend
	}
	return nil
}
