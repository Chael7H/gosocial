package mysql

import (
	"context"
	"gosocial/models"
	"time"
)

type MessageDao struct{}

func NewMessageDao() *MessageDao {
	return &MessageDao{}
}

// SaveMessage 保存消息到MySQL
func (d *MessageDao) SaveMessage(ctx context.Context, msg *models.Message) error {
	return GetDB().WithContext(ctx).Create(msg).Error
}

// GetMessages 从MySQL获取历史消息
func (d *MessageDao) GetMessages(ctx context.Context, from, to int64, start, end time.Time) ([]models.Message, error) {
	var messages []models.Message

	err := GetDB().WithContext(ctx).
		Where("((`from` = ? AND `to` = ?) OR (`from` = ? AND `to` = ?))",
			from, to, to, from).
		Where("created_at BETWEEN ? AND ?", start, end).
		Order("created_at ASC").
		Find(&messages).Error

	if err != nil {
		return nil, err
	}
	return messages, nil
}
