package ai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"yaml-backend/pkg/models"
)

// GeminiClient Gemini AI客户端
type GeminiClient struct {
	APIKey  string
	BaseURL string
	client  *http.Client
}

// GeminiRequest Gemini API请求结构
type GeminiRequest struct {
	Contents []Content `json:"contents"`
	Config   Config    `json:"generationConfig,omitempty"`
}

// Content 内容结构
type Content struct {
	Role  string `json:"role"`
	Parts []Part `json:"parts"`
}

// Part 内容部分
type Part struct {
	Text string `json:"text"`
}

// Config 生成配置
type Config struct {
	Temperature     float64 `json:"temperature,omitempty"`
	MaxOutputTokens int     `json:"maxOutputTokens,omitempty"`
	TopP            float64 `json:"topP,omitempty"`
	TopK            int     `json:"topK,omitempty"`
}

// GeminiResponse Gemini API响应结构
type GeminiResponse struct {
	Candidates []Candidate `json:"candidates"`
}

// Candidate 候选响应
type Candidate struct {
	Content Content `json:"content"`
}

// NewGeminiClient 创建新的Gemini客户端
func NewGeminiClient(apiKey, baseURL string) *GeminiClient {
	return &GeminiClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// SummarizeActivities 总结用户活动
func (g *GeminiClient) SummarizeActivities(activities []*models.Activity) (string, error) {
	// 构建活动数据的文本描述
	activityText := g.buildActivityText(activities)
	
	// 构建提示词
	prompt := fmt.Sprintf(`请分析以下用户活动数据，并生成一份简洁的总结报告：

%s

请从以下几个方面进行分析：
1. 主要使用的应用程序
2. 活动时间分布
3. 工作效率评估
4. 建议和改进点

请用中文回复，保持简洁明了。`, activityText)

	return g.generateContent(prompt)
}

// SummarizeKeyboardInputs 总结键盘输入
func (g *GeminiClient) SummarizeKeyboardInputs(inputs []*models.KeyboardInput) (string, error) {
	// 构建输入数据的文本描述
	inputText := g.buildInputText(inputs)
	
	// 构建提示词
	prompt := fmt.Sprintf(`请分析以下用户键盘输入数据，并生成一份总结：

%s

请分析：
1. 输入内容的类型和特征
2. 使用频率最高的应用
3. 输入模式和习惯
4. 可能的工作内容推测

请用中文回复，注意保护隐私，不要直接引用具体的输入内容。`, inputText)

	return g.generateContent(prompt)
}

// generateContent 调用Gemini API生成内容
func (g *GeminiClient) generateContent(prompt string) (string, error) {
	// 构建请求
	request := GeminiRequest{
		Contents: []Content{
			{
				Role: "user",
				Parts: []Part{
					{Text: prompt},
				},
			},
		},
		Config: Config{
			Temperature:     0.7,
			MaxOutputTokens: 1000,
			TopP:            0.8,
			TopK:            40,
		},
	}

	// 序列化请求
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建API URL - AiHubMix需要在路径中包含API key
	apiURL := fmt.Sprintf("%s/v1beta/models/gemini-2.0-flash-exp:generateContent?key=%s", g.BaseURL, g.APIKey)

	// 创建HTTP请求
	req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")

	// 发送请求
	resp, err := g.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// 解析响应
	var geminiResp GeminiResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 提取生成的文本
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content generated")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

// buildActivityText 构建活动数据的文本描述
func (g *GeminiClient) buildActivityText(activities []*models.Activity) string {
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

// buildInputText 构建输入数据的文本描述
func (g *GeminiClient) buildInputText(inputs []*models.KeyboardInput) string {
	var text string
	for i, input := range inputs {
		if i >= 15 { // 限制最多分析15条记录
			break
		}
		// 为了隐私保护，只显示输入长度和应用信息
		text += fmt.Sprintf("时间: %s, 应用: %s, 输入长度: %d字符\n",
			input.Timestamp.Format("2006-01-02 15:04:05"),
			input.AppName,
			len(input.Text))
	}
	return text
}