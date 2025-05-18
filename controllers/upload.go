package controllers

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"gosocial/pkg/snowflake"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type UploadController struct{}

func NewUploadController() *UploadController {
	return &UploadController{}
}

// UploadFileHandler 处理文件上传
// @Summary 上传文件
// @Description 上传图片或文件
// @Tags 上传
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "上传的文件"
// @Success 200 {object} models.Response "{"url":"文件访问URL"}"
// @Failure 400 {object} models.Response "错误信息"
// @Router /api/v1/upload [post]
func (c *UploadController) UploadFileHandler(ctx *gin.Context) {
	// 1. 获取上传文件
	file, err := ctx.FormFile("file")
	if err != nil {
		zap.L().Error("获取上传文件失败", zap.Error(err))
		ResponseError(ctx, CodeInvalidParam)
		return
	}

	// 2. 验证文件大小 (限制10MB)
	const maxFileSize = 10 << 20 // 10MB
	if file.Size > maxFileSize {
		zap.L().Error("文件大小超过限制",
			zap.String("filename", file.Filename),
			zap.Int64("size", file.Size))
		ResponseErrorWithMsg(ctx, CodeInvalidParam, "文件大小不能超过10MB")
		return
	}

	// 3. 验证文件类型
	ext := strings.ToLower(filepath.Ext(file.Filename))
	if !isAllowedFileType(ext) {
		zap.L().Error("不支持的文件类型",
			zap.String("filename", file.Filename),
			zap.String("ext", ext))
		ResponseErrorWithMsg(ctx, CodeInvalidParam, "不支持的文件类型")
		return
	}

	// 4. 创建上传目录 (按日期组织)
	uploadDir := filepath.Join("static", "upload", time.Now().Format("2006/01/02"))
	if err = os.MkdirAll(uploadDir, 0755); err != nil {
		zap.L().Error("创建上传目录失败",
			zap.String("path", uploadDir),
			zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}

	// 5. 生成唯一文件名
	fileUid, _ := snowflake.GenID()
	filename := fmt.Sprintf("%d_%v%s",
		time.Now().Unix(),
		fileUid,
		ext)
	filePath := filepath.Join(uploadDir, filename)

	// 6. 保存文件
	if err = ctx.SaveUploadedFile(file, filePath); err != nil {
		zap.L().Error("保存文件失败",
			zap.String("path", filePath),
			zap.Error(err))
		ResponseError(ctx, CodeServerBusy)
		return
	}

	// 7. 返回文件URL (相对路径)
	ResponseSuccess(ctx, gin.H{
		"url": "/" + filePath,
	})
}

// isAllowedFileType 检查文件类型是否允许
func isAllowedFileType(ext string) bool {
	allowedTypes := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
		".gif":  true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".zip":  true,
		".rar":  true,
	}
	return allowedTypes[ext]
}
