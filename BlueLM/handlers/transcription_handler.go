package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"time"

	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/config"
	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/utils"
	"github.com/dingdinglz/vivo"
	"github.com/gin-gonic/gin"
)

// createBlueLMApp 创建蓝心大模型应用实例，考虑配置优先级
// 优先级：前端传递的配置 > 系统环境变量 > config.yaml文件
func createBlueLMApp(frontendAppID, frontendAppKey string, cfg *config.Config) *vivo.Vivo {
	// 1. 优先使用前端传递的配置
	if frontendAppID != "" && frontendAppKey != "" {
		utils.Log.Infof("Using frontend provided BlueLM config")
		return vivo.NewVivoAIGC(vivo.Config{
			AppID:  frontendAppID,
			AppKey: frontendAppKey,
		})
	}

	// 2. 使用系统环境变量
	envAppID := os.Getenv("BLUELM_APP_ID")
	envAppKey := os.Getenv("BLUELM_APP_KEY")
	if envAppID != "" && envAppKey != "" {
		utils.Log.Infof("Using environment variables for BlueLM config")
		return vivo.NewVivoAIGC(vivo.Config{
			AppID:  envAppID,
			AppKey: envAppKey,
		})
	}

	// 3. 使用config.yaml文件中的配置
	if cfg.VivoAI.AppID != "" && cfg.VivoAI.AppKey != "" {
		utils.Log.Infof("Using config.yaml for BlueLM config")
		return vivo.NewVivoAIGC(vivo.Config{
			AppID:  cfg.VivoAI.AppID,
			AppKey: cfg.VivoAI.AppKey,
		})
	}

	// 如果都没有配置，返回空配置的应用实例（可能会失败）
	utils.Log.Warnf("No BlueLM configuration found, using empty config")
	return vivo.NewVivoAIGC(vivo.Config{
		AppID:  "",
		AppKey: "",
	})
}

// TranscriptionHandler 处理长语音转写请求
func TranscriptionHandler(app *vivo.Vivo, cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 保存上传的文件
		uploadFilePath, err := utils.SaveUploadedFile(c, cfg.FilePaths.UploadDir)
		if err != nil {
			utils.AbortWithBadRequest(c, err, "Failed to save uploaded file")
			return
		}

		// 获取蓝心大模型配置（前端传递的优先级最高）
		appID := c.PostForm("app_id")
		appKey := c.PostForm("app_key")

		// 创建蓝心大模型应用实例，考虑配置优先级
		transcriptionApp := createBlueLMApp(appID, appKey, cfg)

		//调用蓝心大模型长语音转写
		trans := transcriptionApp.NewTranscription(uploadFilePath)
		e := trans.Upload()
		if e != nil {
			utils.AbortWithInternalServerError(c, e)
			return
		}

		e = trans.Start()
		if e != nil {
			utils.AbortWithInternalServerError(c, e)
			return
		}

		taskID := getTaskID(trans)
		// 获取原始文件名
		filename := "unknown"
		if file, err := c.FormFile("file"); err == nil {
			filename = file.Filename
		}

		// 创建任务记录
		GlobalTaskManager.CreateTask(taskID, filename)
		GlobalTaskManager.UpdateTaskStatus(taskID, TaskStatusProcessing, "Processing started")

		go pollTranscriptionStatus(trans, cfg, taskID)

		c.JSON(http.StatusOK, gin.H{"task_id": taskID})
	}
}

// getTaskID uses reflection to access the private taskID field
func getTaskID(trans *vivo.Transcription) string {
	v := reflect.ValueOf(trans).Elem()
	taskIDField := v.FieldByName("taskID")
	if !taskIDField.IsValid() {
		return "unknown"
	}
	return taskIDField.String()
}

func pollTranscriptionStatus(trans *vivo.Transcription, cfg *config.Config, taskID string) {
	process := 0
	var e error
	for process != 100 {
		time.Sleep(1 * time.Second)
		// 查询任务进度
		process, e = trans.GetTaskInfo()
		if e != nil {
			utils.Log.Warnf("Failed to get task info for task %s: %v", taskID, e)
			GlobalTaskManager.UpdateTaskStatus(taskID, TaskStatusFailed, fmt.Sprintf("Error getting progress: %v", e))
			// 不中断轮询，继续尝试
			continue
		}
		utils.Log.Infof("Task %s progress: %d%%", taskID, process)
		GlobalTaskManager.UpdateTaskStatus(taskID, TaskStatusProcessing, fmt.Sprintf("Progress: %d%%", process))
	}

	result, e := trans.GetResult()
	if e != nil {
		utils.Log.Errorf("Failed to get result for task %s: %v", taskID, e)
		GlobalTaskManager.UpdateTaskStatus(taskID, TaskStatusFailed, fmt.Sprintf("Error getting result: %v", e))
		return
	}

	jsonData, err := json.Marshal(result)
	if err != nil {
		utils.Log.Errorf("Failed to marshal result for task %s: %v", taskID, err)
		GlobalTaskManager.UpdateTaskStatus(taskID, TaskStatusFailed, fmt.Sprintf("Error serializing result: %v", err))
		return
	}

	filename := fmt.Sprintf("transcription_%s.json", taskID)
	downloadFilePath := filepath.Join(cfg.FilePaths.DownloadDir, filename)
	//将json数据写入文件
	err = os.WriteFile(downloadFilePath, jsonData, 0644)
	if err != nil {
		utils.Log.Errorf("Failed to write result to file for task %s: %v", taskID, err)
		GlobalTaskManager.UpdateTaskStatus(taskID, TaskStatusFailed, fmt.Sprintf("Error writing file: %v", err))
		return
	}

	// 更新任务状态为完成
	GlobalTaskManager.SetTaskFilePath(taskID, downloadFilePath)
	GlobalTaskManager.UpdateTaskStatus(taskID, TaskStatusCompleted, "Transcription completed successfully")
	utils.Log.Infof("Transcription task %s completed successfully. Result saved to %s", taskID, downloadFilePath)
}
