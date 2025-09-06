package ai

import (
	"fmt"
	"time"

	"yaml-backend/internal/storage"
	"yaml-backend/pkg/models"
)

// AIService AI服务管理器
type AIService struct {
	storage      *storage.SQLiteStorage
	geminiClient *GeminiClient
}

// NewAIService 创建新的AI服务
func NewAIService(storage *storage.SQLiteStorage, apiKey, baseURL string, timeout time.Duration) *AIService {
	return &AIService{
		storage:      storage,
		geminiClient: NewGeminiClient(apiKey, baseURL, timeout),
	}
}

// GenerateActivitySummary 生成活动总结
func (s *AIService) GenerateActivitySummary(limit int) (*storage.SummaryResult, error) {
	// 获取最近的活动数据
	activities, err := s.storage.GetRecentActivities(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get activities: %w", err)
	}

	if len(activities) == 0 {
		return &storage.SummaryResult{
			Type:      "activity",
			Summary:   "暂无活动数据可供分析",
			DataCount: 0,
			CreatedAt: time.Now(),
		}, nil
	}

	// 调用AI生成总结
	summary, err := s.geminiClient.SummarizeActivities(activities)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// 保存总结结果
	result := &storage.SummaryResult{
		Type:      "activity",
		Summary:   summary,
		DataCount: len(activities),
		CreatedAt: time.Now(),
	}

	if err := s.saveSummary(result); err != nil {
		fmt.Printf("Warning: failed to save summary: %v\n", err)
	}

	return result, nil
}

// GenerateKeyboardSummary 生成键盘输入总结
func (s *AIService) GenerateKeyboardSummary(limit int) (*storage.SummaryResult, error) {
	// 获取最近的键盘输入数据
	inputs, err := s.storage.GetRecentKeyboardInputs(limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get keyboard inputs: %w", err)
	}

	if len(inputs) == 0 {
		return &storage.SummaryResult{
			Type:      "keyboard",
			Summary:   "暂无键盘输入数据可供分析",
			DataCount: 0,
			CreatedAt: time.Now(),
		}, nil
	}

	// 调用AI生成总结
	summary, err := s.geminiClient.SummarizeKeyboardInputs(inputs)
	if err != nil {
		return nil, fmt.Errorf("failed to generate summary: %w", err)
	}

	// 保存总结结果
	result := &storage.SummaryResult{
		Type:      "keyboard",
		Summary:   summary,
		DataCount: len(inputs),
		CreatedAt: time.Now(),
	}

	if err := s.saveSummary(result); err != nil {
		fmt.Printf("Warning: failed to save summary: %v\n", err)
	}

	return result, nil
}

// GetRecentSummaries 获取最近的总结
func (s *AIService) GetRecentSummaries(limit int) ([]*storage.SummaryResult, error) {
	return s.storage.GetRecentSummaries(limit)
}

// saveSummary 保存总结到数据库
func (s *AIService) saveSummary(summary *storage.SummaryResult) error {
	return s.storage.SaveSummary(summary)
}

// StreamActivitySummary 流式生成活动总结
func (s *AIService) StreamActivitySummary(activities []*models.Activity) (<-chan string, <-chan error) {
	// 构建活动数据的文本描述
	activityText := s.buildActivityText(activities)

	// 构建提示词
	prompt := fmt.Sprintf(`请分析以下用户活动数据，并生成一份简洁的总结报告：

%s

请从以下几个方面进行分析：
1. 主要使用的应用程序
2. 活动时间分布
3. 工作效率评估
4. 建议和改进点

请用中文回复，保持简洁明了。`, activityText)

	return s.geminiClient.generateContentStream(prompt)
}

// buildActivityText 构建活动数据的文本描述
func (s *AIService) buildActivityText(activities []*models.Activity) string {
	var text string
	for i, activity := range activities {
		if i >= 20 { // 限制最多分析20条记录
			break
		}
		text += fmt.Sprintf("时间: %s, 类型: %s, 应用: %s, 内容: %s, 持续时间: %d秒\n",
			activity.Timestamp.Format("2006-01-02 15:04:05"),
			activity.Type,
			activity.AppName,
			activity.Content,
			activity.Duration)
	}
	return text
}
