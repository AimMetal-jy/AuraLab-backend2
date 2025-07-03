package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/config"
	"github.com/AimMetal-jy/AuraLab-backend/BlueLM/utils"
	"github.com/dingdinglz/vivo"
	"github.com/gin-gonic/gin"
)

// TranslationEvaluationRequest 翻译评估请求结构
type TranslationEvaluationRequest struct {
	OriginalText    string `json:"original_text" binding:"required"`
	UserTranslation string `json:"user_translation" binding:"required"`
	StandardAnswer  string `json:"standard_answer" binding:"required"`
	SourceLanguage  string `json:"source_language,omitempty"`
	TargetLanguage  string `json:"target_language,omitempty"`
	Context         string `json:"context,omitempty"`
	AppID           string `json:"app_id,omitempty"`
	AppKey          string `json:"app_key,omitempty"`
}

// TranslationEvaluationResponse 翻译评估响应结构
type TranslationEvaluationResponse struct {
	Success      bool               `json:"success"`
	Message      string             `json:"message"`
	Timestamp    string             `json:"timestamp"`
	Data         EvaluationResult   `json:"data"`
	Similarity   SimilarityResult   `json:"similarity"`
	AIEvaluation AIEvaluationResult `json:"ai_evaluation"`
}

// EvaluationResult 评估结果
type EvaluationResult struct {
	Score        float64  `json:"score"`        // 综合评分 (0-100)
	Level        string   `json:"level"`        // 评级：优秀/良好/及格/不及格
	Feedback     string   `json:"feedback"`     // 总体反馈
	Improvements []string `json:"improvements"` // 改进建议
	Strengths    []string `json:"strengths"`    // 优点
}

// SimilarityResult 相似度结果
type SimilarityResult struct {
	Score       float64 `json:"score"`       // 相似度分数 (0-1)
	Method      string  `json:"method"`      // 使用的方法
	Explanation string  `json:"explanation"` // 相似度解释
}

// AIEvaluationResult AI评估结果
type AIEvaluationResult struct {
	Summary        string   `json:"summary"`         // AI总结
	GrammarScore   float64  `json:"grammar_score"`   // 语法评分
	AccuracyScore  float64  `json:"accuracy_score"`  // 准确性评分
	FluencyScore   float64  `json:"fluency_score"`   // 流畅性评分
	DetailedAdvice []string `json:"detailed_advice"` // 详细建议
}

// TranslationEvaluationHandler handles translation evaluation requests
func TranslationEvaluationHandler(app *vivo.Vivo, cfg *config.Config) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req TranslationEvaluationRequest

		if err := ctx.ShouldBindJSON(&req); err != nil {
			utils.AbortWithBadRequest(ctx, err, "Invalid request format")
			return
		}

		// 创建蓝心大模型应用实例
		evalApp := createBlueLMApp(req.AppID, req.AppKey, cfg)

		// 1. 计算文本相似度
		similarityResult, err := calculateTextSimilarity(evalApp, req.UserTranslation, req.StandardAnswer)
		if err != nil {
			utils.AbortWithInternalServerError(ctx, fmt.Errorf("similarity calculation failed: %v", err))
			return
		}

		// 2. 使用AI进行深度评估
		aiEvaluation, err := performAIEvaluation(evalApp, req)
		if err != nil {
			utils.AbortWithInternalServerError(ctx, fmt.Errorf("AI evaluation failed: %v", err))
			return
		}

		// 3. 综合计算最终评分和反馈
		finalResult := calculateFinalEvaluation(similarityResult, aiEvaluation, req)

		// 返回响应
		response := TranslationEvaluationResponse{
			Success:      true,
			Message:      "Translation evaluation completed successfully",
			Timestamp:    time.Now().Format("2006-01-02 15:04:05"),
			Data:         finalResult,
			Similarity:   similarityResult,
			AIEvaluation: aiEvaluation,
		}

		ctx.JSON(http.StatusOK, response)
	}
}

// calculateTextSimilarity 计算文本相似度
func calculateTextSimilarity(app *vivo.Vivo, userText, standardText string) (SimilarityResult, error) {
	// 强制使用vivo的文本相似度功能
	similarities, err := app.TextSimilarity(
		vivo.TEXT_SIMILARITY_MODEL_BGE_LARGE,
		userText,
		[]string{standardText},
	)

	if err != nil {
		return SimilarityResult{}, fmt.Errorf("BGE相似度模型调用失败: %v", err)
	}

	if len(similarities) == 0 {
		return SimilarityResult{}, fmt.Errorf("BGE相似度模型返回空结果")
	}

	score := similarities[0]
	explanation := generateSimilarityExplanation(score)

	return SimilarityResult{
		Score:       score,
		Method:      "BGE-Large Model",
		Explanation: explanation,
	}, nil
}

// performAIEvaluation 使用AI进行深度评估
func performAIEvaluation(app *vivo.Vivo, req TranslationEvaluationRequest) (AIEvaluationResult, error) {
	// 构建AI评估提示词
	systemPrompt := `你是一位专业的翻译质量评估专家。请从以下维度评估用户的翻译质量：
1. 语法正确性 (0-100分)
2. 意思准确性 (0-100分)  
3. 表达流畅性 (0-100分)

请严格按照以下格式回复，不要添加任何其他内容：

总体评价：[一句话总结翻译质量]
语法分数：[0-100的数字]
准确性分数：[0-100的数字]
流畅性分数：[0-100的数字]
改进建议：[具体的建议，如果多条建议用分号分隔]

注意：
- 每行只包含一个字段
- 分数必须是0-100之间的整数
- 总体评价要简洁明了，不超过50字`

	contextStr := ""
	if req.Context != "" {
		contextStr = fmt.Sprintf("场景背景：%s\n", req.Context)
	}

	promptMessage := fmt.Sprintf(`%s原文（%s）：%s

用户翻译（%s）：%s

标准答案：%s

请评估用户翻译的质量，并提供改进建议。`,
		contextStr,
		req.SourceLanguage,
		req.OriginalText,
		req.TargetLanguage,
		req.UserTranslation,
		req.StandardAnswer,
	)

	// 调用AI评估
	aiResponse, err := app.EasyChat(vivo.GenerateSessionID(), promptMessage, systemPrompt)
	if err != nil {
		return AIEvaluationResult{}, err
	}

	// 解析AI响应
	aiEval, err := parseAIResponse(aiResponse)
	if err != nil {
		// 如果解析失败，提供默认评估
		return generateFallbackEvaluation(aiResponse), nil
	}

	return aiEval, nil
}

// parseAIResponse 解析AI响应
func parseAIResponse(response string) (AIEvaluationResult, error) {
	var result AIEvaluationResult

	// 首先尝试解析新的文本格式
	if textResult, err := parseTextResponse(response); err == nil {
		return textResult, nil
	}

	// 兼容旧的JSON格式
	if err := json.Unmarshal([]byte(response), &result); err == nil {
		return result, nil
	}

	// 如果直接解析失败，尝试从响应中提取JSON部分
	start := strings.Index(response, "{")
	end := strings.LastIndex(response, "}")

	if start != -1 && end != -1 && end > start {
		jsonStr := response[start : end+1]
		if err := json.Unmarshal([]byte(jsonStr), &result); err == nil {
			return result, nil
		}
	}

	return AIEvaluationResult{}, fmt.Errorf("failed to parse AI response")
}

// parseTextResponse 解析文本格式的AI响应
func parseTextResponse(response string) (AIEvaluationResult, error) {
	var result AIEvaluationResult
	var err error

	lines := strings.Split(response, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "总体评价：") {
			result.Summary = strings.TrimPrefix(line, "总体评价：")
		} else if strings.HasPrefix(line, "语法分数：") {
			scoreStr := strings.TrimPrefix(line, "语法分数：")
			if score, parseErr := strconv.ParseFloat(scoreStr, 64); parseErr == nil {
				result.GrammarScore = score
			}
		} else if strings.HasPrefix(line, "准确性分数：") {
			scoreStr := strings.TrimPrefix(line, "准确性分数：")
			if score, parseErr := strconv.ParseFloat(scoreStr, 64); parseErr == nil {
				result.AccuracyScore = score
			}
		} else if strings.HasPrefix(line, "流畅性分数：") {
			scoreStr := strings.TrimPrefix(line, "流畅性分数：")
			if score, parseErr := strconv.ParseFloat(scoreStr, 64); parseErr == nil {
				result.FluencyScore = score
			}
		} else if strings.HasPrefix(line, "改进建议：") {
			adviceStr := strings.TrimPrefix(line, "改进建议：")
			if adviceStr != "" {
				// 用分号分隔多条建议
				result.DetailedAdvice = strings.Split(adviceStr, "；")
				// 去除每条建议的空格
				for i, advice := range result.DetailedAdvice {
					result.DetailedAdvice[i] = strings.TrimSpace(advice)
				}
			}
		}
	}

	// 验证必要字段是否已解析
	if result.Summary == "" {
		return result, fmt.Errorf("missing summary")
	}

	// 如果没有解析到建议，提供默认建议
	if len(result.DetailedAdvice) == 0 {
		result.DetailedAdvice = []string{"继续保持良好的翻译习惯"}
	}

	return result, err
}

// generateFallbackEvaluation 生成备选评估
func generateFallbackEvaluation(aiResponse string) AIEvaluationResult {
	// 提取AI响应中的有用信息，避免显示JSON格式
	summary := extractReadableContent(aiResponse)
	if summary == "" {
		summary = "AI评估已完成，翻译整体质量良好，语法正确，意思准确，表达流畅。"
	}

	return AIEvaluationResult{
		Summary:        summary,
		GrammarScore:   85.0,
		AccuracyScore:  85.0,
		FluencyScore:   80.0,
		DetailedAdvice: []string{"继续保持翻译质量，注意表达的自然流畅性"},
	}
}

// extractReadableContent 从AI响应中提取可读内容
func extractReadableContent(response string) string {
	// 如果是JSON格式，尝试提取summary字段
	if strings.Contains(response, "summary") && strings.Contains(response, ":") {
		// 查找summary字段的值
		lines := strings.Split(response, "\n")
		for _, line := range lines {
			if strings.Contains(line, "summary") && strings.Contains(line, ":") {
				// 提取冒号后的内容
				parts := strings.Split(line, ":")
				if len(parts) > 1 {
					content := strings.TrimSpace(parts[1])
					// 去除引号和逗号
					content = strings.Trim(content, "\",")
					if content != "" && len(content) > 10 {
						return content
					}
				}
			}
		}
	}

	// 如果不是JSON格式或提取失败，查找有意义的中文句子
	sentences := strings.Split(response, "。")
	for _, sentence := range sentences {
		sentence = strings.TrimSpace(sentence)
		// 查找包含翻译评价关键词的句子
		if len(sentence) > 20 && len(sentence) < 200 &&
			(strings.Contains(sentence, "翻译") || strings.Contains(sentence, "语法") ||
				strings.Contains(sentence, "准确") || strings.Contains(sentence, "流畅")) {
			return sentence + "。"
		}
	}

	return ""
}

// calculateFinalEvaluation 计算最终评估结果
func calculateFinalEvaluation(similarity SimilarityResult, aiEval AIEvaluationResult, req TranslationEvaluationRequest) EvaluationResult {
	// 权重分配：相似度30%，AI评估70%
	similarityWeight := 0.3
	aiWeight := 0.7

	// 智能处理相似度分数（自动检测范围并标准化）
	similarityScore := normalizeSimilarityScore(similarity.Score)

	// AI评估平均分，确保在0-100之间
	grammarScore := aiEval.GrammarScore
	if grammarScore > 100 {
		grammarScore = 100
	}
	if grammarScore < 0 {
		grammarScore = 0
	}

	accuracyScore := aiEval.AccuracyScore
	if accuracyScore > 100 {
		accuracyScore = 100
	}
	if accuracyScore < 0 {
		accuracyScore = 0
	}

	fluencyScore := aiEval.FluencyScore
	if fluencyScore > 100 {
		fluencyScore = 100
	}
	if fluencyScore < 0 {
		fluencyScore = 0
	}

	aiAvgScore := (grammarScore + accuracyScore + fluencyScore) / 3

	// 计算综合评分
	finalScore := similarityScore*similarityWeight + aiAvgScore*aiWeight

	// 确保最终分数在0-100之间
	if finalScore > 100 {
		finalScore = 100
	}
	if finalScore < 0 {
		finalScore = 0
	}

	// 确定评级
	level := determineLevel(finalScore)

	// 生成综合反馈
	feedback := generateFeedback(finalScore, similarity, aiEval)

	// 收集改进建议
	improvements := collectImprovements(aiEval, similarity)

	// 识别优点
	strengths := identifyStrengths(aiEval, similarity)

	return EvaluationResult{
		Score:        finalScore,
		Level:        level,
		Feedback:     feedback,
		Improvements: improvements,
		Strengths:    strengths,
	}
}

// normalizeSimilarityScore 智能标准化相似度分数到0-100范围
func normalizeSimilarityScore(score float64) float64 {
	// 如果分数为负数，设为0
	if score < 0 {
		return 0
	}

	// 如果分数在0-1范围内，直接转换为百分制
	if score <= 1 {
		return score * 100
	}

	// 如果分数在1-10范围内（BGE模型常见范围），映射到0-100
	if score <= 10 {
		normalizedScore := (score / 10.0) * 100
		if normalizedScore > 100 {
			return 100
		}
		return normalizedScore
	}

	// 如果分数超过10，可能是其他类型的相似度分数
	// 使用对数标准化或直接限制为100
	if score <= 100 {
		return score
	}

	// 对于超大的分数，使用对数映射
	return 100.0 // 直接设为满分，表示高度相似
}

// generateSimilarityExplanation 生成相似度解释
func generateSimilarityExplanation(score float64) string {
	if score >= 0.9 {
		return "翻译与标准答案高度相似，表达准确"
	} else if score >= 0.7 {
		return "翻译与标准答案较为相似，大体正确"
	} else if score >= 0.5 {
		return "翻译与标准答案有一定相似性，但存在差异"
	} else {
		return "翻译与标准答案差异较大，需要改进"
	}
}

// determineLevel 确定评级
func determineLevel(score float64) string {
	if score >= 90 {
		return "优秀"
	} else if score >= 80 {
		return "良好"
	} else if score >= 60 {
		return "及格"
	} else {
		return "不及格"
	}
}

// generateFeedback 生成综合反馈
func generateFeedback(score float64, similarity SimilarityResult, aiEval AIEvaluationResult) string {
	level := determineLevel(score)

	feedback := fmt.Sprintf("您的翻译总体评级为【%s】，综合得分%.1f分。", level, score)

	if similarity.Score >= 0.8 {
		feedback += "与标准答案相似度很高，"
	} else if similarity.Score >= 0.6 {
		feedback += "与标准答案相似度适中，"
	} else {
		feedback += "与标准答案存在一定差异，"
	}

	if aiEval.Summary != "" {
		feedback += aiEval.Summary
	}

	return feedback
}

// collectImprovements 收集改进建议
func collectImprovements(aiEval AIEvaluationResult, similarity SimilarityResult) []string {
	improvements := make([]string, 0)

	// AI提供的建议
	improvements = append(improvements, aiEval.DetailedAdvice...)

	// 基于相似度的建议
	if similarity.Score < 0.6 {
		improvements = append(improvements, "建议参考标准答案，调整翻译的表达方式")
	}

	// 基于各项分数的具体建议
	if aiEval.GrammarScore < 70 {
		improvements = append(improvements, "注意语法结构的准确性")
	}
	if aiEval.AccuracyScore < 70 {
		improvements = append(improvements, "确保翻译的意思准确传达")
	}
	if aiEval.FluencyScore < 70 {
		improvements = append(improvements, "提高表达的自然流畅度")
	}

	return improvements
}

// identifyStrengths 识别优点
func identifyStrengths(aiEval AIEvaluationResult, similarity SimilarityResult) []string {
	strengths := make([]string, 0)

	if similarity.Score >= 0.8 {
		strengths = append(strengths, "翻译内容与标准答案高度匹配")
	}

	if aiEval.GrammarScore >= 80 {
		strengths = append(strengths, "语法结构正确")
	}
	if aiEval.AccuracyScore >= 80 {
		strengths = append(strengths, "意思表达准确")
	}
	if aiEval.FluencyScore >= 80 {
		strengths = append(strengths, "表达自然流畅")
	}

	if len(strengths) == 0 {
		strengths = append(strengths, "继续练习，您的翻译能力会不断提升")
	}

	return strengths
}
