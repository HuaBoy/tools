package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// RAGClient 调用 Python LlamaIndex RAG 服务的 HTTP 客户端
type RAGClient struct {
	BaseURL    string
	HTTPClient *http.Client
}

// NewRAGClient 创建 RAG 客户端
func NewRAGClient(baseURL string) *RAGClient {
	return &RAGClient{
		BaseURL: baseURL,
		HTTPClient: &http.Client{
			Timeout: 120 * time.Second, // 分块向量化可能较慢
		},
	}
}

// IndexRequest 索引请求
type IndexRequest struct {
	DocumentID int64  `json:"document_id"`
	Content    string `json:"content"`
}

// IndexResponse 索引响应
type IndexResponse struct {
	OK         bool `json:"ok"`
	ChunkCount int  `json:"chunk_count"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query    string  `json:"query"`
	TopK     int     `json:"top_k"`
	MinScore float64 `json:"min_score"`
}

// SearchResponse 搜索响应
type SearchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
	Count   int            `json:"count"`
}

// ContextResponse 上下文响应
type ContextResponse struct {
	Context string         `json:"context"`
	Results []SearchResult `json:"results"`
}

// FAQIndexRequest FAQ 索引请求
type FAQIndexRequest struct {
	ID       int64  `json:"id"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// HealthCheck 健康检查
func (c *RAGClient) HealthCheck(ctx context.Context) error {
	req, _ := http.NewRequestWithContext(ctx, "GET", c.BaseURL+"/health", nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return fmt.Errorf("RAG 服务不可达: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("RAG 服务异常: status %d", resp.StatusCode)
	}
	return nil
}

// IndexDocument 调用 Python 服务分块 + 向量化 + 入库
func (c *RAGClient) IndexDocument(ctx context.Context, docID int64, content string) (int, error) {
	body, _ := json.Marshal(IndexRequest{DocumentID: docID, Content: content})
	resp, err := c.doPost(ctx, "/api/rag/index", body)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return 0, fmt.Errorf("RAG index 失败: %s", string(b))
	}

	var result IndexResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}
	return result.ChunkCount, nil
}

// Search 调用 Python 服务语义搜索
func (c *RAGClient) Search(ctx context.Context, query string, topK int, minScore float64) ([]SearchResult, error) {
	body, _ := json.Marshal(SearchRequest{Query: query, TopK: topK, MinScore: minScore})
	resp, err := c.doPost(ctx, "/api/rag/search", body)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("RAG search 失败: %s", string(b))
	}

	var result SearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}
	return result.Results, nil
}

// BuildContext 调用 Python 服务搜索 + 构建上下文
func (c *RAGClient) BuildContext(ctx context.Context, query string, topK int, minScore float64) (string, []SearchResult, error) {
	body, _ := json.Marshal(SearchRequest{Query: query, TopK: topK, MinScore: minScore})
	resp, err := c.doPost(ctx, "/api/rag/context", body)
	if err != nil {
		return "", nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return "", nil, fmt.Errorf("RAG context 失败: %s", string(b))
	}

	var result ContextResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", nil, err
	}
	return result.Context, result.Results, nil
}

// IndexFAQ 调用 Python 服务向量化 FAQ
func (c *RAGClient) IndexFAQ(ctx context.Context, faqID int64, question, answer string) error {
	body, _ := json.Marshal(FAQIndexRequest{ID: faqID, Question: question, Answer: answer})
	resp, err := c.doPost(ctx, "/api/rag/faq", body)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("RAG FAQ 失败: %s", string(b))
	}
	return nil
}

// DeleteDocument 调用 Python 服务删除文档分块
func (c *RAGClient) DeleteDocument(ctx context.Context, docID int64) error {
	req, _ := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/api/rag/document/%d", c.BaseURL, docID), nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// DeleteFAQ 调用 Python 服务删除 FAQ
func (c *RAGClient) DeleteFAQ(ctx context.Context, faqID int64) error {
	req, _ := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/api/rag/faq/%d", c.BaseURL, faqID), nil)
	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (c *RAGClient) doPost(ctx context.Context, path string, body []byte) (*http.Response, error) {
	req, _ := http.NewRequestWithContext(ctx, "POST", c.BaseURL+path, bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	return c.HTTPClient.Do(req)
}
