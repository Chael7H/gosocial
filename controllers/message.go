package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gosocial/dao/mysql"
	r "gosocial/dao/redis"
	"gosocial/logic"
	"gosocial/models"
	"strconv"
	"time"
)

type MessageController struct {
	logic *logic.MessageLogic
}

// NewMessageController 构造函数，接收MessageDao
func NewMessageController(messageDao *r.MessageDao, mysqlDao *mysql.MessageDao) *MessageController {
	return &MessageController{
		logic: logic.NewMessageLogic(messageDao, mysqlDao),
	}
}

// SendMessageHandler 发送消息
// @Summary 发送文本消息
// @Description 向指定好友发送文本消息(图片和文件消息请使用/messages/image和/messages/file路由)
// @Tags 消息
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param object body models.ParamTextReq true "文本消息参数"
// @Success 200 {object} models.Response "{"timestamp":时间戳,"receiver_id":接收者ID,"status":"sent"}"
// @Failure 400 {object} models.Response "错误信息"
// @Router /api/v1/messages [post]
func (c *MessageController) SendMessageHandler(ctx *gin.Context) {
	// 解析请求体中的消息类型和内容
	var req models.ParamTextReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zap.L().Error("parse request body failed",
			zap.Error(err),
			zap.Any("request", ctx.Request.Body),
		)
		// 设置默认值
		if req.To == 0 {
			req.To, _ = strconv.ParseInt(ctx.PostForm("to"), 10, 64)
		}
		if req.Content == "" {
			req.Content = ctx.PostForm("content")
		}
		// 再次验证必要字段
		if req.To == 0 || req.Content == "" {
			ResponseErrorWithMsg(ctx, CodeInvalidParam, "收件人和消息内容不能为空")
			return
		}
	}

	// 获取当前用户ID,并判断是否为好友
	from, _ := ctx.Get("uid")
	if err := mysql.IsFriend(from.(int64), req.To); !errors.Is(err, mysql.ErrorIsFriend) {
		zap.L().Error("is friend failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeIsNotFriend)
		return
	}

	err := c.logic.SendTextMessage(ctx, from.(int64), req.To, req.Content)
	if err != nil {
		zap.L().Error("send message failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeMessageSendFail)
		return
	}
	// 返回完整响应，包含前端需要的所有字段
	ResponseSuccess(ctx, gin.H{
		"from":       from.(int64),
		"to":         req.To,
		"direct":     1,
		"content":    req.Content,
		"type":       1,
		"created_at": time.Now().Unix(),
		"status":     "sent",
		"avatar_url": "", // 需要从用户信息获取
		"hide_time":  false,
	})
}

// GetMessagesHandler 获取聊天记录
// @Summary 获取聊天记录
// @Description 获取与指定好友的聊天记录
// @Tags 消息
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param friend_id query string true "好友ID"
// @Param start_time query int false "开始时间戳(默认7天前)"
// @Param end_time query int false "结束时间戳(默认当前时间)"
// @Param mark_read query bool false "是否标记为已读"
// @Param history query bool false "是否获取全部历史消息"
// @Success 200 {object} models.Response "{"messages":[{"from":发送者ID,"to":接收者ID,"direct":消息发送方向 ,"created_at":时间,"content":"内容","avatar_url":"头像URL","hide_time":是否隐藏时间}]}"
// @Failure 400 {object} models.Response "错误信息"
// @Router /api/v1/messages [get]
func (c *MessageController) GetMessagesHandler(ctx *gin.Context) {
	userID, _ := ctx.Get("uid")
	friendID, _ := strconv.ParseInt(ctx.Query("friend_id"), 10, 64)
	//判断是否为好友
	if err := mysql.IsFriend(userID.(int64), friendID); errors.Is(err, mysql.ErrorIsNotFriend) {
		zap.L().Error("is friend failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeIsNotFriend)
		return
	}

	// 处理时间范围
	var startTime, endTime time.Time

	if ctx.Query("history") == "true" {
		// 获取全部历史消息
		startTime = time.Unix(0, 0)
		endTime = time.Now()
	} else {
		// 默认获取7天内消息
		startTime = time.Now().Add(-7 * 24 * time.Hour)
		endTime = time.Now()

		// 处理自定义时间范围
		if startStr := ctx.Query("start_time"); startStr != "" {
			if startUnix, err := strconv.ParseInt(startStr, 10, 64); err == nil {
				startTime = time.Unix(startUnix, 0)
			}
		}
		if endStr := ctx.Query("end_time"); endStr != "" {
			if endUnix, err := strconv.ParseInt(endStr, 10, 64); err == nil {
				endTime = time.Unix(endUnix, 0)
			}
		}
	}

	// 获取消息列表
	messages, err := c.logic.GetMessages(ctx, userID.(int64), friendID, startTime, endTime)
	if err != nil {
		zap.L().Error("get messages failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}

	// 标记为已读
	if ctx.Query("mark_read") == "true" {
		if err = c.logic.MarkMessagesAsRead(ctx, userID.(int64), friendID); err != nil {
			zap.L().Error("mark messages as read failed, err: ", zap.Error(err))
			ResponseError(ctx, CodeServerBusy)
			return
		}
	}

	// 转换为前端需要的格式
	responseMsgs := make([]gin.H, 0, len(messages))
	//确定消息方向
	var direct = 0
	for _, msg := range messages {
		if userID == msg.From {
			direct = 1
		} else {
			direct = 2
		}
		responseMsgs = append(responseMsgs, gin.H{
			"from":       msg.From,
			"to":         msg.To,
			"direct":     direct, //direct=1 方向则为用户发给好友，direct=2 方向则为好友发给用户
			"created_at": msg.CreatedAt,
			"content":    msg.Content,
			"avatar_url": msg.My.AvatarURL,
			"hide_time":  msg.HideTime,
		})
	}

	ResponseSuccess(ctx, gin.H{
		"messages": responseMsgs,
	})
}

// SendImageMessageHandler 发送图片消息
// @Summary 发送图片消息
// @Description 向指定好友发送图片消息
// @Tags 消息
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param object body models.ParamImageReq true "图片消息参数"
// @Success 200 {object} models.Response "{"timestamp":时间戳,"receiver_id":接收者ID,"status":"sent"}"
// @Failure 400 {object} models.Response "错误信息"
// @Router /api/v1/messages/image [post]
func (c *MessageController) SendImageMessageHandler(ctx *gin.Context) {
	// 解析请求体
	var req models.ParamImageReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zap.L().Error("parse image request body failed", zap.Error(err))
		ResponseErrorWithMsg(ctx, CodeInvalidParam, "收件人和图片内容不能为空")
		return
	}

	// 获取当前用户ID,并判断是否为好友
	from, _ := ctx.Get("uid")
	if err := mysql.IsFriend(from.(int64), req.To); !errors.Is(err, mysql.ErrorIsFriend) {
		zap.L().Error("is friend failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeIsNotFriend)
		return
	}

	// 处理图片消息
	file := &models.FileMeta{
		URL:    req.Content,
		Name:   "image", // 图片消息特殊标记
		Size:   0,       // 图片大小不强制要求
		Type:   "image",
		Width:  req.Width,
		Height: req.Height,
	}
	err := c.logic.SendFileMessage(ctx, from.(int64), req.To, file)
	if err != nil {
		zap.L().Error("send image message failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeMessageSendFail)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"from":       from.(int64),
		"to":         req.To,
		"direct":     1,
		"content":    req.Content,
		"type":       2, // 图片消息类型
		"created_at": time.Now().Unix(),
		"status":     "sent",
		"avatar_url": "",
		"hide_time":  false,
	})
}

// SendFileMessageHandler 发送文件消息
// @Summary 发送文件消息
// @Description 向指定好友发送文件消息
// @Tags 消息
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Param object body models.ParamFileReq true "文件消息参数"
// @Success 200 {object} models.Response "{"timestamp":时间戳,"receiver_id":接收者ID,"status":"sent"}"
// @Failure 400 {object} models.Response "错误信息"
// @Router /api/v1/messages/file [post]
func (c *MessageController) SendFileMessageHandler(ctx *gin.Context) {
	// 解析请求体
	var req models.ParamFileReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		zap.L().Error("parse file request body failed", zap.Error(err))
		ResponseErrorWithMsg(ctx, CodeInvalidParam, "收件人和文件内容不能为空")
		return
	}

	// 获取当前用户ID,并判断是否为好友
	from, _ := ctx.Get("uid")
	if err := mysql.IsFriend(from.(int64), req.To); !errors.Is(err, mysql.ErrorIsFriend) {
		zap.L().Error("is friend failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeIsNotFriend)
		return
	}

	// 处理文件消息
	file := &models.FileMeta{
		URL:  req.Content,
		Name: req.Name,
		Size: req.Size,
		Type: req.Type,
	}
	err := c.logic.SendFileMessage(ctx, from.(int64), req.To, file)
	if err != nil {
		zap.L().Error("send file message failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeMessageSendFail)
		return
	}

	ResponseSuccess(ctx, gin.H{
		"from":       from.(int64),
		"to":         req.To,
		"direct":     1,
		"content":    req.Content,
		"type":       3, // 文件消息类型
		"created_at": time.Now().Unix(),
		"status":     "sent",
		"avatar_url": "",
		"hide_time":  false,
	})
}

// GetUnreadCountsHandler 获取未读消息数
// @Summary 获取未读消息数
// @Description 获取所有好友的未读消息数
// @Tags 消息
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer 用户令牌"
// @Success 200 {object} models.Response "{"counts":{"好友ID":未读数量}}"
// @Failure 400 {object} models.Response "错误信息"
// @Router /api/v1/messages/unread [get]
func (c *MessageController) GetUnreadCountsHandler(ctx *gin.Context) {
	// 获取当前用户ID
	from, _ := ctx.Get("uid")
	// 获取未读消息数
	counts, err := c.logic.GetUnreadCounts(ctx, from.(int64))
	if err != nil {
		zap.L().Error("get unread count failed, err: ", zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}
	// 返回成功响应，包含未读消息数
	ResponseSuccess(ctx, gin.H{
		"counts": counts,
	})
}
