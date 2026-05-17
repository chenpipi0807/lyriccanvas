// ============================================================
// LyricCanvas — DeepSeek API 客户端
// ============================================================
package services

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type DeepSeekClient struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

func NewDeepSeekClient(apiKey, baseURL string) *DeepSeekClient {
	return &DeepSeekClient{
		apiKey:  apiKey,
		baseURL: strings.TrimRight(baseURL, "/"),
		client:  &http.Client{},
	}
}

// ---------- 请求/响应结构体 ----------

type ChatCompletionRequest struct {
	Model       string          `json:"model"`
	Messages    []ChatMessageDS `json:"messages"`
	Temperature float64         `json:"temperature,omitempty"`
	MaxTokens   int             `json:"max_tokens,omitempty"`
	Stream      bool            `json:"stream"`
}

type ChatMessageDS struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

type ChatStreamChunk struct {
	Choices []struct {
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		FinishReason *string `json:"finish_reason"`
	} `json:"choices"`
}

// Chat 非流式对话
func (c *DeepSeekClient) Chat(model string, messages []ChatMessageDS, temperature float64, maxTokens int) (string, error) {
	if temperature == 0 {
		temperature = 1.0
	}
	if maxTokens == 0 {
		maxTokens = 4096
	}

	req := ChatCompletionRequest{
		Model:       model,
		Messages:    messages,
		Temperature: temperature,
		MaxTokens:   maxTokens,
		Stream:      false,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("序列化请求失败: %w", err)
	}

	httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("请求 DeepSeek API 失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("DeepSeek API 返回错误 (%d): %s", resp.StatusCode, string(respBody))
	}

	var result ChatCompletionResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("解析响应失败: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("DeepSeek 返回空响应")
	}

	return result.Choices[0].Message.Content, nil
}

// ChatStream 流式对话，通过 channel 返回每个 chunk
func (c *DeepSeekClient) ChatStream(model string, messages []ChatMessageDS, temperature float64, maxTokens int) (<-chan string, <-chan error) {
	contentCh := make(chan string, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(contentCh)
		defer close(errCh)

		if temperature == 0 {
			temperature = 1.0
		}
		if maxTokens == 0 {
			maxTokens = 4096
		}

		req := ChatCompletionRequest{
			Model:       model,
			Messages:    messages,
			Temperature: temperature,
			MaxTokens:   maxTokens,
			Stream:      true,
		}

		body, err := json.Marshal(req)
		if err != nil {
			errCh <- fmt.Errorf("序列化请求失败: %w", err)
			return
		}

		httpReq, err := http.NewRequest("POST", c.baseURL+"/chat/completions", bytes.NewReader(body))
		if err != nil {
			errCh <- fmt.Errorf("创建请求失败: %w", err)
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
		httpReq.Header.Set("Accept", "text/event-stream")

		resp, err := c.client.Do(httpReq)
		if err != nil {
			errCh <- fmt.Errorf("请求 DeepSeek API 失败: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			respBody, _ := io.ReadAll(resp.Body)
			errCh <- fmt.Errorf("DeepSeek API 返回错误 (%d): %s", resp.StatusCode, string(respBody))
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" || line == "data: [DONE]" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}

			jsonStr := strings.TrimPrefix(line, "data: ")
			var chunk ChatStreamChunk
			if err := json.Unmarshal([]byte(jsonStr), &chunk); err != nil {
				continue
			}

			if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
				contentCh <- chunk.Choices[0].Delta.Content
			}
		}

		if err := scanner.Err(); err != nil {
			errCh <- fmt.Errorf("读取流式响应失败: %w", err)
		}
	}()

	return contentCh, errCh
}

// ValidateKey 验证 API Key 是否有效
func (c *DeepSeekClient) ValidateKey() bool {
	if c.apiKey == "" {
		return false
	}
	_, err := c.Chat("deepseek-v4-flash", []ChatMessageDS{
		{Role: "user", Content: "Hi"},
	}, 0, 1)
	return err == nil
}

// SetAPIKey 运行时更新 API Key
func (c *DeepSeekClient) SetAPIKey(apiKey string) {
	c.apiKey = apiKey
}
