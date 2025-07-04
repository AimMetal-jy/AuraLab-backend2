package handlers

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/config"
	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/utils"
	"github.com/dingdinglz/vivo"
	"github.com/gin-gonic/gin"
)

// MultimodalChatHandler handles the AI chat requests with image support.
func MultimodalChatHandler(app *vivo.Vivo, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取表单数据
		message := ctx.PostForm("message")
		sessionID := ctx.PostForm("session_id")
		historyMessagesStr := ctx.PostForm("history_messages")
		appID := ctx.PostForm("app_id")
		appKey := ctx.PostForm("app_key")

		// 检查消息是否为空
		if message == "" {
			utils.AbortWithBadRequest(ctx, nil, "Message cannot be empty")
			return
		}

		// 处理图片上传
		var imageBase64 string
		file, header, err := ctx.Request.FormFile("image")
		if err == nil {
			defer file.Close()

			// 读取图片数据
			imageData, err := io.ReadAll(file)
			if err != nil {
				utils.AbortWithBadRequest(ctx, err, "Failed to read image file")
				return
			}

			// 将图片转换为base64
			imageBase64 = base64.StdEncoding.EncodeToString(imageData)

			// 添加MIME类型前缀
			contentType := header.Header.Get("Content-Type")
			if contentType == "" {
				// 根据文件扩展名推断MIME类型
				ext := strings.ToLower(header.Filename[strings.LastIndex(header.Filename, ".")+1:])
				switch ext {
				case "jpg", "jpeg":
					contentType = "image/jpeg"
				case "png":
					contentType = "image/png"
				case "gif":
					contentType = "image/gif"
				case "webp":
					contentType = "image/webp"
				default:
					contentType = "image/jpeg"
				}
			}
			imageBase64 = fmt.Sprintf("data:%s;base64,%s", contentType, imageBase64)
		}

		// 创建蓝心大模型应用实例
		chatApp := createBlueLMApp(appID, appKey, cfg)

		// 生成或使用现有的会话ID
		if sessionID == "" {
			sessionID = vivo.GenerateSessionID()
		}

		// 解析历史消息
		var historyMessages []vivo.ChatMessage
		if historyMessagesStr != "" {
			// TODO: 解析历史消息JSON字符串
			// 这里暂时跳过历史消息的解析
		}

		// 构建当前消息内容
		var messageContent string
		if imageBase64 != "" {
			// 如果有图片，使用多模态格式
			messageContent = fmt.Sprintf("%s\n[图片:%s]", message, imageBase64)
		} else {
			messageContent = message
		}

		// 添加当前用户消息
		historyMessages = append(historyMessages, vivo.ChatMessage{
			Role:    vivo.CHAT_ROLE_USER,
			Content: messageContent,
		})

		// 调用蓝心大模型的多模态接口
		res, err := chatApp.Chat(vivo.GenerateSessionID(), sessionID, historyMessages, nil)
		if err != nil {
			utils.AbortWithInternalServerError(ctx, err)
			return
		}

		// 将AI回复添加到消息历史
		historyMessages = append(historyMessages, res)

		// 返回响应
		ctx.JSON(http.StatusOK, gin.H{
			"success":    true,
			"message":    "Chat completed successfully",
			"timestamp":  time.Now().Format("2006-01-02 15:04:05"),
			"session_id": sessionID,
			"data": gin.H{
				"reply":    res.Content,
				"role":     res.Role,
				"messages": historyMessages,
			},
			"app_info": gin.H{
				"app_id":  appID,
				"app_key": appKey,
			},
		})
	}
}
