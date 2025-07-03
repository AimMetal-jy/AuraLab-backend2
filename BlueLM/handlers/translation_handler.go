package handlers

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// 翻译语言常量
const (
	TRANSLATE_LANGUAGE_AUTO     = "auto"
	TRANSLATE_LANGUAGE_CHINESE  = "zh-CHS"
	TRANSLATE_LANGUAGE_ENGLISH  = "en"
	TRANSLATE_LANGUAGE_JAPANESE = "ja"
	TRANSLATE_LANGUAGE_KOREAN   = "ko"
	TRANSLATE_LANGUAGE_FRENCH   = "fr"
	TRANSLATE_LANGUAGE_GERMAN   = "de"
	TRANSLATE_LANGUAGE_SPANISH  = "es"
	TRANSLATE_LANGUAGE_ITALIAN  = "it"
	TRANSLATE_LANGUAGE_RUSSIAN  = "ru"
	TRANSLATE_LANGUAGE_ARABIC   = "ar"
)

// 翻译请求结构
type TranslationRequest struct {
	From string `json:"from" binding:"required"`
	To   string `json:"to" binding:"required"`
	Text string `json:"text" binding:"required"`
}

// 翻译响应结构
type TranslationResponse struct {
	Code int `json:"code"`
	Data struct {
		Translation string `json:"translation"`
	} `json:"data"`
	Msg string `json:"msg"`
}

// 翻译结果
type TranslationResult struct {
	Success      bool   `json:"success"`
	Translation  string `json:"translation"`
	Message      string `json:"message"`
	From         string `json:"from"`
	To           string `json:"to"`
	OriginalText string `json:"original_text"`
}

// 支持的语言映射
var languageMap = map[string]string{
	"en": TRANSLATE_LANGUAGE_ENGLISH,
	"zh": TRANSLATE_LANGUAGE_CHINESE,
	"ja": TRANSLATE_LANGUAGE_JAPANESE,
	"ko": TRANSLATE_LANGUAGE_KOREAN,
	"fr": TRANSLATE_LANGUAGE_FRENCH,
	"de": TRANSLATE_LANGUAGE_GERMAN,
	"es": TRANSLATE_LANGUAGE_SPANISH,
	"it": TRANSLATE_LANGUAGE_ITALIAN,
	"ru": TRANSLATE_LANGUAGE_RUSSIAN,
	"ar": TRANSLATE_LANGUAGE_ARABIC,
}

// 生成请求ID
func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// 创建HTTP客户端并设置签名
func createSignedRequest(appID, appKey string, formData map[string]string) (*http.Request, error) {
	// 准备签名参数
	params := make(map[string]string)
	for k, v := range formData {
		params[k] = v
	}

	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	params["appId"] = appID
	params["timestamp"] = timestamp

	// 生成签名
	signature := generateSignature(params, appKey)
	params["signature"] = signature

	// 创建表单数据
	values := url.Values{}
	for k, v := range params {
		values.Set(k, v)
	}

	// 创建请求
	req, err := http.NewRequest("POST", "https://api-ai.vivo.com.cn/translation/query/self", strings.NewReader(values.Encode()))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

// 生成签名
func generateSignature(params map[string]string, appKey string) string {
	// 按键排序
	keys := make([]string, 0, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 构建签名字符串
	var signStr strings.Builder
	for _, k := range keys {
		if signStr.Len() > 0 {
			signStr.WriteString("&")
		}
		signStr.WriteString(k)
		signStr.WriteString("=")
		signStr.WriteString(params[k])
	}
	signStr.WriteString("&key=")
	signStr.WriteString(appKey)

	// MD5加密
	hash := md5.Sum([]byte(signStr.String()))
	return hex.EncodeToString(hash[:])
}

// 翻译处理器
func TranslationHandler(c *gin.Context) {
	var req TranslationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, TranslationResult{
			Success: false,
			Message: "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查输入文本是否为空
	if strings.TrimSpace(req.Text) == "" {
		c.JSON(http.StatusBadRequest, TranslationResult{
			Success: false,
			Message: "翻译文本不能为空",
		})
		return
	}

	// 映射语言代码
	fromLang, ok := languageMap[req.From]
	if !ok {
		fromLang = req.From // 如果映射不存在，直接使用原值
	}

	toLang, ok := languageMap[req.To]
	if !ok {
		toLang = req.To // 如果映射不存在，直接使用原值
	}

	// 如果源语言和目标语言相同，直接返回原文
	if fromLang == toLang {
		c.JSON(http.StatusOK, TranslationResult{
			Success:      true,
			Translation:  req.Text,
			From:         req.From,
			To:           req.To,
			OriginalText: req.Text,
			Message:      "源语言和目标语言相同，直接返回原文",
		})
		return
	}

	// 这里需要从配置中获取vivo的AppID和AppKey
	appID := "your_vivo_app_id"   // 需要替换为实际的AppID
	appKey := "your_vivo_app_key" // 需要替换为实际的AppKey

	// 如果没有配置vivo的凭据，使用模拟翻译
	if appID == "your_vivo_app_id" || appKey == "your_vivo_app_key" {
		logrus.Warn("Vivo API credentials not configured, using mock translation")
		mockTranslation := getMockTranslation(req.Text, req.From, req.To)
		c.JSON(http.StatusOK, TranslationResult{
			Success:      true,
			Translation:  mockTranslation,
			From:         req.From,
			To:           req.To,
			OriginalText: req.Text,
			Message:      "使用模拟翻译服务",
		})
		return
	}

	// 准备请求数据
	formData := map[string]string{
		"from":      fromLang,
		"to":        toLang,
		"text":      req.Text,
		"app":       "auralab",
		"requestId": generateRequestID(),
	}

	// 创建带签名的请求
	httpReq, err := createSignedRequest(appID, appKey, formData)
	if err != nil {
		logrus.Error("创建HTTP请求失败:", err)
		c.JSON(http.StatusInternalServerError, TranslationResult{
			Success: false,
			Message: "创建请求失败",
		})
		return
	}

	// 发送请求
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		logrus.Error("发送翻译请求失败:", err)
		c.JSON(http.StatusInternalServerError, TranslationResult{
			Success: false,
			Message: "翻译服务请求失败",
		})
		return
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logrus.Error("读取响应失败:", err)
		c.JSON(http.StatusInternalServerError, TranslationResult{
			Success: false,
			Message: "读取响应失败",
		})
		return
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		logrus.Error("HTTP请求失败, 状态码:", resp.StatusCode, "响应:", string(body))
		c.JSON(http.StatusInternalServerError, TranslationResult{
			Success: false,
			Message: "翻译服务返回错误",
		})
		return
	}

	// 解析响应
	var translationResp TranslationResponse
	if err := json.Unmarshal(body, &translationResp); err != nil {
		logrus.Error("解析响应失败:", err)
		c.JSON(http.StatusInternalServerError, TranslationResult{
			Success: false,
			Message: "解析响应失败",
		})
		return
	}

	// 检查业务状态码
	if translationResp.Code != 0 {
		logrus.Error("翻译失败:", translationResp.Msg)
		c.JSON(http.StatusOK, TranslationResult{
			Success: false,
			Message: translationResp.Msg,
		})
		return
	}

	// 返回翻译结果
	c.JSON(http.StatusOK, TranslationResult{
		Success:      true,
		Translation:  translationResp.Data.Translation,
		From:         req.From,
		To:           req.To,
		OriginalText: req.Text,
		Message:      "翻译成功",
	})
}

// 模拟翻译函数（当没有配置vivo API时使用）
func getMockTranslation(text, from, to string) string {
	// 这里可以接入其他翻译服务或返回模拟数据
	mockTranslations := map[string]map[string]string{
		"hello": {
			"zh": "你好",
			"ja": "こんにちは",
			"ko": "안녕하세요",
			"fr": "bonjour",
			"de": "hallo",
		},
		"world": {
			"zh": "世界",
			"ja": "世界",
			"ko": "세계",
			"fr": "monde",
			"de": "welt",
		},
		"你好": {
			"en": "hello",
			"ja": "こんにちは",
			"ko": "안녕하세요",
		},
	}

	lowerText := strings.ToLower(text)
	if translations, ok := mockTranslations[lowerText]; ok {
		if translation, ok := translations[to]; ok {
			return translation
		}
	}

	return fmt.Sprintf("[模拟翻译] %s -> %s: %s", from, to, text)
}

// 获取支持的语言列表
func GetSupportedLanguagesHandler(c *gin.Context) {
	languages := map[string]string{
		"auto": "自动检测",
		"zh":   "中文",
		"en":   "英语",
		"ja":   "日语",
		"ko":   "韩语",
		"fr":   "法语",
		"de":   "德语",
		"es":   "西班牙语",
		"it":   "意大利语",
		"ru":   "俄语",
		"ar":   "阿拉伯语",
	}

	c.JSON(http.StatusOK, gin.H{
		"success":   true,
		"languages": languages,
	})
}
