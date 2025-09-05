package ai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
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

type Candidate struct {
	Content      Content `json:"content"`
	FinishReason string  `json:"finishReason,omitempty"`
}

// 新增：处理实际API响应的结构
type GeminiAPIResponse struct {
	Candidates []struct {
		Content struct {
			Role  string `json:"role"`
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
		FinishReason string `json:"finishReason"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount int `json:"promptTokenCount"`
		TotalTokenCount  int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
}

// NewGeminiClient 创建新的Gemini客户端
func NewGeminiClient(apiKey, baseURL string) *GeminiClient {
	return &GeminiClient{
		APIKey:  apiKey,
		BaseURL: baseURL,
		client: &http.Client{
			Timeout: 120 * time.Second, // 增加超时时间到120秒
		},
	}
}

// SummarizeActivities 总结用户活动
func (g *GeminiClient) SummarizeActivities(activities []*models.Activity) (string, error) {
	// 构建活动数据的文本描述
	activityText := g.buildActivityText(activities)

	// 构建提示词 - 简化版本
	prompt := fmt.Sprintf(`分析用户活动数据：

%s

请简要总结：
1. 主要应用
2. 使用模式
3. 效率建议

用中文回复，保持简洁。`, activityText)

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
			MaxOutputTokens: 6717, // 增加输出token限制
			TopP:            0.8,
			TopK:            40,
		},
	}

	// 序列化请求
	jsonData, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// 构建API URL - AiHubMix需要在URL中包含key参数
	apiURL := fmt.Sprintf("%s/v1beta/models/gemini-2.5-flash:generateContent?key=%s", g.BaseURL, g.APIKey)

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
	var geminiResp GeminiAPIResponse
	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	// 提取生成的文本
	if len(geminiResp.Candidates) == 0 {
		return "", fmt.Errorf("no candidates in response")
	}

	candidate := geminiResp.Candidates[0]
	if len(candidate.Content.Parts) == 0 {
		// 如果没有parts，可能是因为达到了最大token限制或其他原因
		if candidate.FinishReason == "MAX_TOKENS" {
			return "", fmt.Errorf("response truncated due to max tokens limit")
		}
		return "", fmt.Errorf("no parts in candidate content, finish reason: %s", candidate.FinishReason)
	}

	text := candidate.Content.Parts[0].Text
	if text == "" {
		return "", fmt.Errorf("empty text in response")
	}

	return text, nil
}

// buildActivityText 构建活动数据的文本描述
func (g *GeminiClient) buildActivityText(activities []*models.Activity) string {
	var text string
	for i, activity := range activities {
		if i >= 10 { // 减少到最多分析10条记录
			break
		}
		// 简化输出格式，减少token使用
		text += fmt.Sprintf("%s: %s在%s (持续%d秒)\n",
			activity.Timestamp.Format("15:04"),
			activity.Type,
			activity.AppName,
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

// generateContentStream 流式生成内容
func (g *GeminiClient) generateContentStream(prompt string) (<-chan string, <-chan error) {
	resultChan := make(chan string, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer close(resultChan)
		defer close(errorChan)

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
				MaxOutputTokens: 6717,
				TopP:            0.8,
				TopK:            40,
			},
		}

		// 序列化请求
		jsonData, err := json.Marshal(request)
		if err != nil {
			errorChan <- fmt.Errorf("failed to marshal request: %w", err)
			return
		}

		// 构建API URL
		apiURL := fmt.Sprintf("%s/v1beta/models/gemini-2.5-flash:streamGenerateContent?key=%s", g.BaseURL, g.APIKey)

		// 创建HTTP请求
		req, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
		if err != nil {
			errorChan <- fmt.Errorf("failed to create request: %w", err)
			return
		}

		// 设置请求头
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "text/event-stream")

		// 发送请求
		resp, err := g.client.Do(req)
		if err != nil {
			errorChan <- fmt.Errorf("failed to send request: %w", err)
			return
		}
		defer resp.Body.Close()

		// 检查HTTP状态码
		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			errorChan <- fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
			return
		}

		// 如果不支持流式，回退到普通模式
		contentType := resp.Header.Get("Content-Type")
		if !strings.Contains(contentType, "text/event-stream") {
			// 读取完整响应
			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errorChan <- fmt.Errorf("failed to read response: %w", err)
				return
			}

			// 解析响应
			var geminiResp GeminiResponse
			if err := json.Unmarshal(body, &geminiResp); err != nil {
				errorChan <- fmt.Errorf("failed to unmarshal response: %w", err)
				return
			}

			// 提取生成的文本
			if len(geminiResp.Candidates) > 0 && len(geminiResp.Candidates[0].Content.Parts) > 0 {
				resultChan <- geminiResp.Candidates[0].Content.Parts[0].Text
			}
			return
		}

		// 处理流式响应
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					break
				}

				var chunk GeminiResponse
				if err := json.Unmarshal([]byte(data), &chunk); err == nil {
					if len(chunk.Candidates) > 0 && len(chunk.Candidates[0].Content.Parts) > 0 {
						resultChan <- chunk.Candidates[0].Content.Parts[0].Text
					}
				}
			}
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("error reading stream: %w", err)
		}
	}()

	return resultChan, errorChan
}
