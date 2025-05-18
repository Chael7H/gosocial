package controllers

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gosocial/dao/mysql"
	"gosocial/logic"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// GetFriendPostsHandler 获取所有好友动态列表(类QQ个人空间)
// @Summary 获取好友动态列表
// @Description 按时间倒序获取所有好友的动态(每次10条)
// @Tags 动态
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} models.Response{data=[]models.ParamPostWithUserInfo} "成功获取动态列表"
// @Failure 401 {object} models.Response "未授权"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/posts [get]
func GetFriendPostsHandler(c *gin.Context) {
	// 获取当前用户ID
	userID := c.MustGet("uid").(int64)

	// 获取偏移量
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 调用逻辑层获取数据
	posts, err := logic.GetFriendPosts(userID, offset)
	if err != nil {
		zap.L().Error("logic.GetFriendPosts failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	//// 更新动态浏览量
	//for _, post := range posts {
	//	if err = logic.IncrementPostViewCount(post.ID); err != nil {
	//		zap.L().Error("logic.IncrementPostViewCount failed",
	//			zap.Int64("post_id", post.ID),
	//			zap.Error(err))
	//	}
	//}

	ResponseSuccess(c, posts)
}

// GetUserPostsHandler 获取指定用户动态列表
// @Summary 获取指定用户动态列表
// @Description 按时间倒序获取指定用户的动态(每次10条)
// @Tags 动态
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param user_id path int true "用户ID"
// @Param offset query int false "偏移量(默认0)"
// @Success 200 {object} models.Response{data=[]models.ParamPostWithUserInfo} "成功获取用户动态"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未授权"
// @Failure 403 {object} models.Response "非好友关系"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/posts/{user_id} [get]
func GetUserPostsHandler(c *gin.Context) {
	// 获取当前用户ID
	currentUserID := c.MustGet("uid").(int64)

	// 获取目标用户ID
	targetUserID, _ := strconv.ParseInt(c.Param("uid"), 10, 64)

	// 获取偏移量
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	// 调用逻辑层获取数据
	posts, err := logic.GetUserPosts(currentUserID, targetUserID, offset)
	if err != nil {
		if errors.Is(err, mysql.ErrorIsNotFriend) {
			zap.L().Error("The user is not current user's friend", zap.Error(err))
			ResponseError(c, CodeIsNotFriend)
			return
		}
		zap.L().Error("logic.GetUserPosts failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	//// 更新动态浏览量
	//for _, post := range posts {
	//	if err = logic.IncrementPostViewCount(post.ID); err != nil {
	//		zap.L().Error("logic.IncrementPostViewCount failed",
	//			zap.Int64("post_id", post.ID),
	//			zap.Error(err))
	//	}
	//}

	ResponseSuccess(c, posts)
}

// IncrementPostViewHandler 增加动态浏览量
// @Summary 增加动态浏览量
// @Description 每次浏览动态时增加1次浏览量
// @Tags 动态
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "动态ID"
// @Success 200 {object} models.Response "操作成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未授权"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/posts/{id}/view [put]
func IncrementPostViewHandler(c *gin.Context) {
	// 获取动态ID
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 调用逻辑层更新浏览量
	if err = logic.IncrementPostViewCount(postID); err != nil {
		zap.L().Error("logic.IncrementPostViewCount failed",
			zap.Int64("post_id", postID),
			zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, "浏览量更新成功")
}

// CreatePostHandler 创建用户动态
// @Summary 创建用户动态
// @Description 创建用户动态(支持纯文本、纯图片、图文混合)
// @Tags 动态
// @Accept multipart/form-data
// @Produce json
// @Security ApiKeyAuth
// @Param content formData string false "文字内容(不超过500字)"
// @Param images formData []file false "图片文件(最多9张,支持jpg/jpeg/png)"
// @Success 200 {object} models.Response{data=models.Post} "动态创建成功"
// @Failure 400 {object} models.Response "参数错误/图片格式错误/图片过多"
// @Failure 401 {object} models.Response "未授权"
// @Failure 413 {object} models.Response "文件过大"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/posts [post]
func CreatePostHandler(c *gin.Context) {
	// 获取当前用户ID
	userID := c.MustGet("uid").(int64)

	// 获取表单数据
	content := c.PostForm("content")
	form, err := c.MultipartForm()
	if err != nil {
		zap.L().Error("c.MultipartForm failed", zap.Error(err))
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 允许纯图片或纯文本
	if form == nil && content == "" {
		ResponseError(c, CodeInvalidParam)
		return
	}

	// 处理图片上传
	var imageURLs []string
	files := form.File["images"]
	if len(files) > 9 {
		ResponseError(c, CodeTooManyImages)
		return
	}

	for _, file := range files {
		// 检查图片格式
		ext := filepath.Ext(file.Filename)
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			ResponseError(c, CodeInvalidImageFormat)
			return
		}

		// 生成唯一文件名
		filename := fmt.Sprintf("%d_%s", time.Now().UnixNano(), file.Filename)
		dst := filepath.Join("static", "images", filename)

		// 保存文件
		if err = c.SaveUploadedFile(file, dst); err != nil {
			zap.L().Error("c.SaveUploadedFile failed", zap.Error(err))
			ResponseError(c, CodeServerBusy)
			return
		}

		imageURLs = append(imageURLs, "/static/images/"+filename)
	}

	// 调用逻辑层创建动态
	post, err := logic.CreatePost(userID, content, strings.Join(imageURLs, ","))
	if err != nil {
		zap.L().Error("logic.CreatePost failed", zap.Error(err))
		ResponseError(c, CodeServerBusy)
		return
	}
	ResponseSuccess(c, post)
}

// DeletePostHandler 删除动态
// @Summary 删除动态
// @Description 删除用户自己的动态
// @Tags 动态
// @Accept json
// @Produce json
// @Security ApiKeyAuth
// @Param id path int true "动态ID"
// @Success 200 {object} models.Response "删除成功"
// @Failure 400 {object} models.Response "参数错误"
// @Failure 401 {object} models.Response "未授权"
// @Failure 500 {object} models.Response "服务器内部错误"
// @Router /api/v1/posts/{id} [delete]
func DeletePostHandler(c *gin.Context) {
	// 获取当前用户ID
	userID := c.MustGet("uid").(int64)

	// 获取动态ID
	postID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		ResponseError(c, CodeInvalidParam)
		return
	}
	if userID != postID {
		zap.L().Error("cannot delete others post", zap.Int64("user_id", userID))
		ResponseError(c, CodeCannotDeleteOthersPost)
	}
	// 调用逻辑层删除动态
	if err = logic.DeletePostByUser(userID); err != nil {
		ResponseError(c, CodeServerBusy)
		return
	}

	ResponseSuccess(c, "动态删除成功")
}
