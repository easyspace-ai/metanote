package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client OpenAI 兼容 API 客户端
type Client struct {
	baseURL string
	apiKey  string
	model   string
	client  *http.Client
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
	Stream   bool          `json:"stream,omitempty"`
}

type chatChoice struct {
	Message struct {
		Role    string `json:"role"`
		Content string `json:"content"`
	} `json:"message"`
}

type chatCompletionResponse struct {
	Choices []chatChoice `json:"choices"`
}

// NewFromEnv 从环境变量创建客户端
func NewFromEnv() *Client {
	base := os.Getenv("OPENAI_COMPAT_BASE_URL")
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	key := os.Getenv("OPENAI_COMPAT_API_KEY")
	model := os.Getenv("OPENAI_COMPAT_MODEL")
	if model == "" {
		model = "gpt-4.1-mini"
	}
	return &Client{
		baseURL: base,
		apiKey:  key,
		model:   model,
		client:  &http.Client{Timeout: 60 * time.Second},
	}
}

// Chat 发送单轮对话并返回助手回复
func (c *Client) Chat(ctx context.Context, userMessage string, modelOverride *string) (string, error) {
	model := c.model
	if modelOverride != nil && *modelOverride != "" {
		model = *modelOverride
	}
	reqBody := chatCompletionRequest{
		Model: model,
		Messages: []chatMessage{{Role: "user", Content: userMessage}},
	}
	data, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}
	url := c.baseURL + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("build chat request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("call chat api: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", fmt.Errorf("chat api status %d", resp.StatusCode)
	}
	var parsed chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}
	if len(parsed.Choices) == 0 {
		return "", fmt.Errorf("empty choices in chat response")
	}
	return parsed.Choices[0].Message.Content, nil
}
