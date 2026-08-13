package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client Ollama API 客户端
type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewClient 创建 Ollama 客户端
func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: strings.TrimRight(baseURL, "/"),
		HTTPClient: &http.Client{
			Timeout: 10 * time.Minute,
		},
	}
}

// EmbeddingRequest 嵌入请求
type EmbeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

// EmbeddingResponse 嵌入响应
type EmbeddingResponse struct {
	Embeddings [][]float32 `json:"embeddings"`
}

// Embedding 获取文本嵌入向量
func (c *Client) Embedding(ctx context.Context, model, text string) ([]float32, error) {
	payload, _ := json.Marshal(EmbeddingRequest{
		Model: model,
		Input: text,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/embed", bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama 嵌入请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Ollama 嵌入返回状态 %d: %s", resp.StatusCode, string(body))
	}

	var result EmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析嵌入响应失败: %w", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("Ollama 未返回嵌入向量")
	}
	return result.Embeddings[0], nil
}

// ChatMessage 对话消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest 对话请求
type ChatRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

// ChatResponse 对话响应
type ChatResponse struct {
	Message ChatMessage `json:"message"`
	Done    bool        `json:"done"`
}

// Chat 普通对话（非流式）
func (c *Client) Chat(ctx context.Context, model string, messages []ChatMessage) (string, error) {
	payload, _ := json.Marshal(ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return "", fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama 对话请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("Ollama 对话返回状态 %d: %s", resp.StatusCode, string(body))
	}

	var result ChatResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("解析对话响应失败: %w", err)
	}
	return result.Message.Content, nil
}

// ChatStreamCallback 流式对话回调（每收到一段内容调用一次）
type ChatStreamCallback func(chunk string) error

// ChatStream 流式对话
func (c *Client) ChatStream(ctx context.Context, model string, messages []ChatMessage, cb ChatStreamCallback) error {
	payload, _ := json.Marshal(ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	})

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.BaseURL+"/api/chat", bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("Ollama 对话请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Ollama 对话返回状态 %d: %s", resp.StatusCode, string(body))
	}

	scanner := bufio.NewScanner(resp.Body)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(bytes.TrimSpace(line)) == 0 {
			continue
		}

		var chunk ChatResponse
		if err := json.Unmarshal(line, &chunk); err != nil {
			// 跳过无法解析的行（可能是空行或注释）
			continue
		}

		if chunk.Message.Content != "" {
			if err := cb(chunk.Message.Content); err != nil {
				return err
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取流式响应失败: %w", err)
	}
	return nil
}

// ListModels 列出已安装的模型
type ModelInfo struct {
	Name    string `json:"name"`
	Size    int64  `json:"size"`
	Digest  string `json:"digest"`
}

type listModelsResponse struct {
	Models []ModelInfo `json:"models"`
}

// ListModels 列出已安装的模型
func (c *Client) ListModels(ctx context.Context) ([]ModelInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama 模型列表请求失败: %w", err)
	}
	defer resp.Body.Close()

	var result listModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("解析模型列表失败: %w", err)
	}
	return result.Models, nil
}

// Ping 检测 Ollama 是否可用
func (c *Client) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.BaseURL+"/api/tags", nil)
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Ollama 状态码: %d", resp.StatusCode)
	}
	return nil
}
