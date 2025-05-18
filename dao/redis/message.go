package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"go.uber.org/zap"
	"gosocial/models"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

const (
	MessageTTL           = 7 * 24 * time.Hour // 消息过期时间
	UnreadKeyPrefix      = "unread:"          // 未读消息数key前缀
	ChatKeyPrefix        = "chat:"            // 聊天记录key前缀
	FileMetaPrefix       = "file:"            // 文件元信息key前缀
	CounterUpdateChannel = "counter_updates"  // 计数器更新频道
)

type MessageDao struct {
	rdb *redis.Client
}

func NewMessageDao(rdb *redis.Client) *MessageDao {
	return &MessageDao{rdb: rdb}
}

// SendMessage 发送消息
func (d *MessageDao) SendMessage(ctx context.Context, msg *models.Message) error {
	// 生成聊天记录key
	chatKey := GetChatKey(msg.From, msg.To)

	// 使用管道批量操作
	pipe := d.rdb.TxPipeline()

	// 存储消息到ZSET
	msgJSON, err := json.Marshal(msg)
	if err != nil {
		zap.L().Error("marshal message failed", zap.Error(err))
		return err
	}
	pipe.ZAdd(ctx, chatKey, &redis.Z{
		Score:  float64(msg.CreatedAt.Unix()),
		Member: msgJSON,
	})

	// 设置过期时间
	pipe.Expire(ctx, chatKey, MessageTTL)

	// 更新未读计数
	unreadKey := UnreadKeyPrefix + fmt.Sprintf("%d", msg.To)
	pipe.HIncrBy(ctx, unreadKey, fmt.Sprintf("%d", msg.From), 1)
	// 发布消息到接收者的频道
	channel := fmt.Sprintf("user:%d:messages", msg.To)
	if err = d.rdb.Publish(ctx, channel, msgJSON).Err(); err != nil {
		zap.L().Error("publish message failed", zap.Error(err))
	}

	_, err = pipe.Exec(ctx)
	return err
}

// GetUnreadCounts 获取未读消息数
func (d *MessageDao) GetUnreadCounts(ctx context.Context, userID int64) (map[int64]int64, error) {
	key := UnreadKeyPrefix + fmt.Sprintf("%d", userID)
	result, err := d.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	counts := make(map[int64]int64, len(result))
	for k, v := range result {
		fromID, err := strconv.ParseInt(k, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid fromID format: %v", err)
		}
		count, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("invalid count format: %v", err)
		}
		counts[fromID] = count
	}
	return counts, nil
}

// GetMessages 获取聊天记录
func (d *MessageDao) GetMessages(ctx context.Context, from, to int64, start, end int64) ([]models.Message, error) {
	key := GetChatKey(from, to)
	results, err := d.rdb.ZRangeByScore(ctx, key, &redis.ZRangeBy{
		Min: fmt.Sprintf("%d", start),
		Max: fmt.Sprintf("%d", end),
	}).Result()
	if err != nil {
		return nil, err
	}

	// 反序列化消息
	var messages []models.Message
	for _, result := range results {
		var msg models.Message
		if err = json.Unmarshal([]byte(result), &msg); err != nil {
			return nil, fmt.Errorf("failed to unmarshal message: %v", err)
		}
		messages = append(messages, msg)
	}

	return messages, nil
}

// StoreFileMeta 存储文件元信息
func (d *MessageDao) StoreFileMeta(ctx context.Context, file *models.FileMeta) error {
	key := FileMetaPrefix + file.URL
	fileJSON, err := json.Marshal(file)
	if err != nil {
		return fmt.Errorf("marshal file meta failed: %v", err)
	}
	return d.rdb.Set(ctx, key, fileJSON, MessageTTL).Err()
}

// GetFileMeta 获取文件元信息
func (d *MessageDao) GetFileMeta(ctx context.Context, url string) (*models.FileMeta, error) {
	key := FileMetaPrefix + url
	fileJSON, err := d.rdb.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	var file models.FileMeta
	if err = json.Unmarshal([]byte(fileJSON), &file); err != nil {
		return nil, fmt.Errorf("unmarshal file meta failed: %v", err)
	}
	return &file, nil
}

// HDel 删除哈希表中的字段
func (d *MessageDao) HDel(ctx context.Context, key string, fields ...string) (int64, error) {
	return d.rdb.HDel(ctx, key, fields...).Result()
}

// Publish 发布消息到指定频道
func (d *MessageDao) Publish(ctx context.Context, channel string, message interface{}) error {
	return d.rdb.Publish(ctx, channel, message).Err()
}

// GetChatKey 生成聊天键
func GetChatKey(user1, user2 int64) string {
	if user1 < user2 {
		return fmt.Sprintf("%s%d:%d", ChatKeyPrefix, user1, user2)
	}
	return fmt.Sprintf("%s%d:%d", ChatKeyPrefix, user2, user1)
}
