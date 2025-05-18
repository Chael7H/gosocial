package controllers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gosocial/dao/mysql"
	"gosocial/logic"
	"strconv"
)

// GetFriendListHandler	获取好友列表
// @Summary 获取好友列表
// @Description 按最后互动时间排序的好友列表
// @Tags 好友管理
// @Produce json
// @Security ApiKeyAuth
// @Param request body models.ParamFriendItem true "好友列表请求参数"
// @Success 200 {object} models.Response{data=[]models.ParamFriendItem}
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /friends [get]
func GetFriendListHandler(c *gin.Context) {
	// 获取用户ID
	userID := c.MustGet("uid").(int64)
	// 获取好友列表
	friendList, err := logic.GetFriendList(userID)
	if err != nil {
		zap.L().Error("GetFriendListHandler failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	// 返回好友列表
	ResponseSuccess(c, friendList)
}

// AddFriendHandler 添加好友
// @Summary 添加好友
// @Description 添加好友关系
// @Tags 好友管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Success 200 {object} models.Response
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 401 {object} models.Response "不能添加自己为好友"
// @Failure 404 {object} models.Response "好友不存在"
// @Failure 409 {object} models.Response "已经是好友关系"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /friends/{friendID} [post]
func AddFriendHandler(c *gin.Context) {
	// 获取用户ID
	userID := c.MustGet("uid").(int64)

	// 获取参数
	FriendID, _ := strconv.ParseInt(c.Param("friendID"), 10, 64)
	// 调用logic添加好友
	if err := logic.AddFriend(userID, FriendID); err != nil {
		zap.L().Debug("AddFriendHandler() failed", zap.Error(err))
		switch {
		case errors.Is(err, mysql.ErrorCannotAddSelf):
			zap.L().Error("AddFriendHandler() cannot add self", zap.Error(err)) //不能添加自己为好友
			ResponseError(c, CodeCannotAddSelf)
		case errors.Is(err, mysql.ErrorUserNotExist):
			zap.L().Error("AddFriendHandler() user not exist", zap.Error(err)) //好友不存在
			ResponseError(c, CodeUserNotExist)
		case errors.Is(err, mysql.ErrorIsFriend):
			zap.L().Error("AddFriendHandler() user have been your friend", zap.Error(err)) //不能重复添加好友
			ResponseError(c, CodeIsFriend)
		default:
			zap.L().Error("AddFriendHandler() failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
		}
		return
	}

	// 返回成功
	ResponseSuccess(c, nil)
}

// GetFriendDetailHandler 获取好友信息
// @Summary 获取好友信息接口
// @Description 获取当前好友的个人信息
// @Tags 用户管理
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param friendID path int true "好友ID"
// @Success 200 {object} models.ParamFriendInfoResponse "成功获取用户信息"
// @Failure 400 {object} models.Response "该用户不是您的好友"
// @Failure 401 {object} models.Response "未授权"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /friends/{friendID} [get]
func GetFriendDetailHandler(c *gin.Context) {
	// 获取用户ID
	userID := c.MustGet("uid").(int64)

	// 获取好友ID
	FriendID, _ := strconv.ParseInt(c.Param("friendID"), 10, 64)

	// 检查是否为好友
	if err := logic.IsFriend(userID, FriendID); errors.Is(err, mysql.ErrorIsNotFriend) {
		zap.L().Error("GetUserInfoLogic failed", zap.Error(err))
		ResponseError(c, CodeIsNotFriend)
		return
	}

	// 调用业务逻辑获取用户信息
	user, err := logic.GetFriendInfoLogic(FriendID)
	if err != nil {
		zap.L().Error("GetUserInfoLogic failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 返回响应
	ResponseSuccess(c, user)
}

// SearchFriendHandler 搜索好友
// @Summary 搜索好友
// @Description 根据备注或用户名搜索好友，按匹配优先级排序
// @Tags 好友管理
// @Produce json
// @Security ApiKeyAuth
// @Param keyword query string true "搜索关键字"
// @Success 200 {object} models.Response{data=[]models.ParamFriendItem}
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /friends/search [get]
func SearchFriendHandler(c *gin.Context) {
	// 获取用户ID
	userID := c.MustGet("uid").(int64)

	// 获取搜索关键字
	keyword := c.Query("keyword")

	zap.L().Debug("SearchFriendHandler", zap.String("keyword", keyword))
	// 调用logic搜索好友
	friendList, err := logic.SearchFriend(userID, keyword)
	if err != nil {
		zap.L().Error("SearchFriendHandler failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 返回搜索结果
	ResponseSuccess(c, friendList)
}

// DeleteFriendHandler 删除好友
// @Summary 删除好友
// @Description 删除好友关系
// @Tags 好友管理
// @Produce json
// @Security ApiKeyAuth
// @Param uid query int true "好友ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 403 {object} models.Response "不是好友关系"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /friends [delete]
func DeleteFriendHandler(c *gin.Context) {
	// 获取用户ID
	userID := c.MustGet("uid").(int64)

	// 获取参数
	friendID, _ := strconv.ParseInt(c.Query("uid"), 10, 64)
	if friendID == 0 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 调用logic删除好友
	if err := logic.DeleteFriend(userID, friendID); err != nil {
		if errors.Is(err, mysql.ErrorIsNotFriend) {
			zap.L().Error("DeleteFriend() failed", zap.Error(err))
			ResponseError(c, CodeIsNotFriend)
			return
		}
		zap.L().Error("DeleteFriend() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, gin.H{
		"message": "您已成功删除该好友",
	})
}

// UpdateFriendRemarkHandler 更新好友备注
// @Summary 更新好友备注
// @Description 更新好友备注信息
// @Tags 好友管理
// @Produce json
// @Security ApiKeyAuth
// @Param uid query int true "好友ID"
// @Param remark query string true "新备注"
// @Success 200 {object} models.Response "备注更新成功"
// @Failure 400 {object} models.Response "参数格式错误"
// @Failure 403 {object} models.Response "不是好友关系"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /friends/remark [put]
func UpdateFriendRemarkHandler(c *gin.Context) {
	// 获取用户ID
	userID := c.MustGet("uid").(int64)

	// 获取参数
	friendID, _ := strconv.ParseInt(c.Query("uid"), 10, 64)
	remark := c.Query("remark")
	if friendID == 0 {
		ResponseError(c, CodeInvalidParam)
		return
	}
	// 调用logic更新好友备注
	if err := logic.UpdateFriendRemark(userID, friendID, remark); err != nil {
		if errors.Is(err, mysql.ErrorIsNotFriend) {
			zap.L().Error("UpdateFriendRemark() failed", zap.Error(err))
			ResponseError(c, CodeIsNotFriend)
			return
		}
		zap.L().Error("UpdateFriendRemark() failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	// 修改成功
	ResponseSuccess(c, gin.H{
		"message": "好友备注修改成功",
		"remark":  remark,
	})
}
