package handlers

import (
	"encoding/base64"
	"net/http"

	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/config"
	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/utils"
	"github.com/dingdinglz/vivo"
	"github.com/gin-gonic/gin"
)

// OCR请求结构体
type OCRRequest struct {
	Image  string `json:"image" binding:"required"` // base64编码的图片
	Mode   int    `json:"mode"`                     // OCR模式，默认为0（仅返回文字）
	AppID  string `json:"app_id,omitempty"`         // vivo AI AppID
	AppKey string `json:"app_key,omitempty"`        // vivo AI AppKey
}

// OCR响应结构体
type OCRResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// OCRHandler 处理OCR识别请求
func OCRHandler(app *vivo.Vivo, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		utils.Log.Info("OCR Handler called")

		var req OCRRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			utils.Log.WithError(err).Error("OCR请求参数错误")
			c.JSON(http.StatusBadRequest, OCRResponse{
				Success: false,
				Message: "请求参数错误: " + err.Error(),
			})
			return
		}

		// 验证模式参数
		if req.Mode < 0 || req.Mode > 2 {
			c.JSON(http.StatusBadRequest, OCRResponse{
				Success: false,
				Message: "不支持的OCR模式，请使用0-2之间的值",
			})
			return
		}

		// 获取AppID和AppKey，优先使用前端传来的参数
		var appID, appKey string
		var vivoApp *vivo.Vivo

		if req.AppID != "" && req.AppKey != "" {
			// 使用前端传来的AppID和AppKey创建新的vivo实例
			appID = req.AppID
			appKey = req.AppKey
			utils.Log.Info("使用前端传来的vivo AI凭据")
			vivoApp = vivo.NewVivoAIGC(vivo.Config{
				AppID:  appID,
				AppKey: appKey,
			})
		} else {
			// 使用默认的vivo实例
			appID = cfg.VivoAI.AppID
			appKey = cfg.VivoAI.AppKey
			utils.Log.Info("使用配置文件中的vivo AI凭据")
			vivoApp = app
		}

		if appID == "" || appKey == "" {
			utils.Log.WithFields(map[string]interface{}{
				"has_app_id":    appID != "",
				"has_app_key":   appKey != "",
				"from_frontend": req.AppID != "",
			}).Error("蓝心大模型配置缺失")
			c.JSON(http.StatusBadRequest, OCRResponse{
				Success: false,
				Message: "缺少vivo AI凭据，请在前端设置页面配置AppID和AppKey",
			})
			return
		}

		// 解码base64图片
		imageData, err := base64.StdEncoding.DecodeString(req.Image)
		if err != nil {
			utils.Log.WithError(err).Error("图片base64解码失败")
			c.JSON(http.StatusBadRequest, OCRResponse{
				Success: false,
				Message: "图片格式错误",
			})
			return
		}

		// 使用vivo包的OCR方法
		result, err := vivoApp.OCR(imageData, req.Mode)
		if err != nil {
			utils.Log.WithError(err).Error("OCR识别失败")
			c.JSON(http.StatusInternalServerError, OCRResponse{
				Success: false,
				Message: "OCR识别失败: " + err.Error(),
			})
			return
		}

		c.JSON(http.StatusOK, OCRResponse{
			Success: true,
			Message: "OCR识别成功",
			Data:    result,
		})
	}
}
