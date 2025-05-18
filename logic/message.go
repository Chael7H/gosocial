package logic

import (
	"context"
	"fmt"
	"gosocial/dao/mysql"
	"gosocial/dao/redis"
	"gosocial/models"
	"sort"
	"time"
)

var oneWeek = 7 * 24 * time.Hour

type MessageLogic struct {
	messageDao *redis.MessageDao
	mysqlDao   *mysql.MessageDao
}

func NewMessageLogic(messageDao *redis.MessageDao, mysqlDao *mysql.MessageDao) *MessageLogic {
	return &MessageLogic{
		messageDao: messageDao,
		mysqlDao:   mysqlDao,
	}
}

// SendTextMessage 发送文本消息
func (l *MessageLogic) SendTextMessage(ctx context.Context, from, to int64, content string) error {
	msg := &models.Message{
		From:      from,
		To:        to,
		Content:   content,
		Type:      1, // 文本消息
		CreatedAt: time.Now(),
	}

	// 为发送方和接收方都存储消息
	if err := l.messageDao.SendMessage(ctx, msg); err != nil {
		return err
	}

	// 异步持久化到MySQL(7天后)
	go func() {
		time.Sleep(oneWeek)
		_ = l.mysqlDao.SaveMessage(context.Background(), msg)
	}()

	return nil
}

// SendFileMessage 发送文件消息
func (l *MessageLogic) SendFileMessage(ctx context.Context, from, to int64, file *models.FileMeta) error {
	// 存储文件元信息
	if err := l.messageDao.StoreFileMeta(ctx, file); err != nil {
		return err
	}

	msg := &models.Message{
		From:      from,
		To:        to,
		Content:   file.URL,
		Type:      3, // 文件消息
		FileURL:   file.URL,
		CreatedAt: time.Now(),
	}

	// 为发送方和接收方都存储消息
	if err := l.messageDao.SendMessage(ctx, msg); err != nil {
		return err
	}

	return nil
}

// GetUnreadCounts 获取所有好友的未读消息数
func (l *MessageLogic) GetUnreadCounts(ctx context.Context, userID int64) (map[int64]int64, error) {
	return l.messageDao.GetUnreadCounts(ctx, userID)
}

// GetMessages 获取聊天记录
func (l *MessageLogic) GetMessages(ctx context.Context, from, to int64, startTime, endTime time.Time) ([]models.Message, error) {

	// 优先从Redis获取最新消息
	redisMsgs, err := l.messageDao.GetMessages(ctx, from, to,
		startTime.Unix(),
		endTime.Unix())
	if err != nil {
		return nil, fmt.Errorf("get redis messages failed: %v", err)
	}

	// 从MySQL获取历史消息(只查询7天前的数据)
	if startTime.Before(time.Now().Add(-oneWeek)) {
		mysqlMsgs, err := l.mysqlDao.GetMessages(ctx, from, to,
			startTime,
			endTime)
		if err != nil {
			return nil, fmt.Errorf("get mysql messages failed: %v", err)
		}
		redisMsgs = append(redisMsgs, mysqlMsgs...)
	}

	// 按时间排序
	sort.Slice(redisMsgs, func(i, j int) bool {
		return redisMsgs[i].CreatedAt.Before(redisMsgs[j].CreatedAt)
	})

	// 智能时间显示处理
	for i := 1; i < len(redisMsgs); i++ {
		if redisMsgs[i].CreatedAt.Sub(redisMsgs[i-1].CreatedAt) < 5*time.Minute {
			redisMsgs[i].HideTime = true
		}
	}

	return redisMsgs, nil
}

// MarkMessagesAsRead 将消息标记为已读
func (l *MessageLogic) MarkMessagesAsRead(ctx context.Context, userID, friendID int64) error {
	// 删除未读计数
	unreadKey := redis.UnreadKeyPrefix + fmt.Sprintf("%d", userID)
	_, err := l.messageDao.HDel(ctx, unreadKey, fmt.Sprintf("%d", friendID))
	if err != nil {
		return fmt.Errorf("failed to mark messages as read: %v", err)
	}

	// 更新前端计数器显示
	go func() {
		time.Sleep(500 * time.Millisecond) // 等待前端更新
		l.messageDao.Publish(ctx, redis.CounterUpdateChannel, fmt.Sprintf("%d:%d", userID, friendID))
	}()

	return nil
}
