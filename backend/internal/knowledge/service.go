package knowledge

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/tester-platform/backend/internal/ollama"
	"golang.org/x/text/encoding/simplifiedchinese"
)

// Service 知识库服务
type Service struct {
	DB         *sql.DB
	Ollama     *ollama.Client
	EmbedModel string
	ChatModel  string
	StoreDir   string
	RAG        *RAGClient // LlamaIndex Python RAG 服务
}

// NewService 创建知识库服务
func NewService(db *sql.DB, client *ollama.Client, embedModel, chatModel, storeDir string) *Service {
	return &Service{
		DB:         db,
		Ollama:     client,
		EmbedModel: embedModel,
		ChatModel:  chatModel,
		StoreDir:   storeDir,
	}
}

// SetRAGClient 设置 LlamaIndex RAG 客户端
func (s *Service) SetRAGClient(client *RAGClient) {
	s.RAG = client
}

// Document 文档记录
type Document struct {
	ID          int64     `json:"id"`
	Title       string    `json:"title"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	FileName    string    `json:"file_name"`
	FileType    string    `json:"file_type"`
	FileSize    int64     `json:"file_size"`
	Status      string    `json:"status"`
	ChunkCount  int       `json:"chunk_count"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	StoragePath string    `json:"storage_path,omitempty"`
}

// FAQ 问题库条目
type FAQ struct {
	ID        int64     `json:"id"`
	Question  string    `json:"question"`
	Answer    string    `json:"answer"`
	Category  string    `json:"category"`
	Tags      []string  `json:"tags"`
	CreatedBy string    `json:"created_by"`
	CreatedAt time.Time `json:"created_at"`
}

// SearchResult 检索结果
type SearchResult struct {
	DocumentID   int64   `json:"document_id"`
	DocumentName string  `json:"document_name"`
	Content      string  `json:"content"`
	Category     string  `json:"category"`
	Score        float64 `json:"score"`
	FAQID        int64   `json:"faq_id,omitempty"`
	Question     string  `json:"question,omitempty"`
	Answer       string  `json:"answer,omitempty"`
	Source       string  `json:"source"` // document / faq
}

// vectorLiteral 将 []float32 转为 pgvector 字面量
func vectorLiteral(vec []float32) string {
	parts := make([]string, len(vec))
	for i, v := range vec {
		parts[i] = fmt.Sprintf("%g", v)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// SaveDocument 保存文档并生成向量（同步执行）
func (s *Service) SaveDocument(ctx context.Context, doc *Document, content string, filePath string) error {
	// 1. 插入文档记录
	var docID int64
	var createdAt time.Time
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO public.knowledge_documents
		 (title, category, description, file_name, file_type, file_size, storage_path, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'processing', $8)
		 RETURNING id, created_at`,
		doc.Title, doc.Category, doc.Description, doc.FileName, doc.FileType, doc.FileSize, filePath, doc.CreatedBy,
	).Scan(&docID, &createdAt)
	if err != nil {
		return fmt.Errorf("插入文档失败: %w", err)
	}
	doc.CreatedAt = createdAt

	// 2. 调用 LlamaIndex Python 服务：分块 + 向量化 + 入库
	if s.RAG != nil {
		chunkCount, err := s.RAG.IndexDocument(ctx, docID, content)
		if err != nil {
			// RAG 失败时降级到 Go 本地分块 + 向量化，不阻断上传
			fmt.Printf("[WARN] LlamaIndex 索引失败，降级到本地向量化: %v\n", err)
		} else {
			doc.ID = docID
			doc.Status = "ready"
			doc.ChunkCount = chunkCount
			return nil
		}
	}

	// 3. 回退：使用 Go 原有分块 + 向量化
	chunks := ChunkText(content, DefaultChunkOptions())
	if len(chunks) == 0 {
		_, _ = s.DB.ExecContext(ctx,
			`UPDATE public.knowledge_documents SET status='failed', updated_at=NOW() WHERE id=$1`, docID)
		return fmt.Errorf("文档内容为空，无法分块")
	}

	for i, chunk := range chunks {
		vec, err := s.Ollama.Embedding(ctx, s.EmbedModel, chunk)
		if err != nil {
			_, _ = s.DB.ExecContext(ctx,
				`UPDATE public.knowledge_documents SET status='failed', updated_at=NOW() WHERE id=$1`, docID)
			return fmt.Errorf("向量化第 %d 块失败: %w", i+1, err)
		}

		_, err = s.DB.ExecContext(ctx,
			`INSERT INTO public.knowledge_chunks (document_id, chunk_index, content, embedding)
			 VALUES ($1, $2, $3, $4::vector)`,
			docID, i, chunk, vectorLiteral(vec))
		if err != nil {
			_, _ = s.DB.ExecContext(ctx,
				`UPDATE public.knowledge_documents SET status='failed', updated_at=NOW() WHERE id=$1`, docID)
			return fmt.Errorf("保存第 %d 块失败: %w", i+1, err)
		}
	}

	_, err = s.DB.ExecContext(ctx,
		`UPDATE public.knowledge_documents
		 SET status='ready', chunk_count=$2, updated_at=NOW()
		 WHERE id=$1`, docID, len(chunks))
	if err != nil {
		return fmt.Errorf("更新文档状态失败: %w", err)
	}

	doc.ID = docID
	doc.Status = "ready"
	doc.ChunkCount = len(chunks)
	return nil
}

// GetDocument 获取文档
func (s *Service) GetDocument(ctx context.Context, id int64) (*Document, error) {
	var d Document
	var storagePath string
	err := s.DB.QueryRowContext(ctx,
		`SELECT id, title, category, description, file_name, file_type, file_size, status, chunk_count, created_by, created_at, COALESCE(storage_path,'')
		 FROM public.knowledge_documents WHERE id=$1`, id,
	).Scan(&d.ID, &d.Title, &d.Category, &d.Description, &d.FileName, &d.FileType,
		&d.FileSize, &d.Status, &d.ChunkCount, &d.CreatedBy, &d.CreatedAt, &storagePath)
	if err != nil {
		return nil, err
	}
	d.StoragePath = storagePath
	return &d, nil
}

// GetDocumentContent 获取文档的文本内容（用于在线编辑）
func (s *Service) GetDocumentContent(ctx context.Context, id int64) (string, error) {
	var storagePath, fileType string
	err := s.DB.QueryRowContext(ctx,
		`SELECT COALESCE(storage_path,''), file_type FROM public.knowledge_documents WHERE id=$1`, id,
	).Scan(&storagePath, &fileType)
	if err != nil {
		return "", fmt.Errorf("文档不存在: %w", err)
	}
	if storagePath == "" {
		return "", fmt.Errorf("文档文件不存在")
	}

	data, err := os.ReadFile(storagePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}

	// 对 txt/md/json/csv 等文本文件处理编码后返回
	switch strings.ToLower(fileType) {
	case ".txt", ".md", ".markdown", ".log", ".json", ".csv":
		return decodeToUTF8(data), nil
	case ".docx":
		return parseDocx(data)
	case ".xlsx", ".xls":
		return parseXLSX(data)
	case ".pdf":
		return parsePDF(data)
	default:
		return decodeToUTF8(data), nil
	}
}

// decodeToUTF8 检测文件编码并转换为 UTF-8
func decodeToUTF8(data []byte) string {
	// 1. 先去除 UTF-8 BOM
	if len(data) >= 3 && data[0] == 0xEF && data[1] == 0xBB && data[2] == 0xBF {
		data = data[3:]
	}
	// 2. 如果已经是合法 UTF-8，直接返回
	if utf8.Valid(data) {
		return string(data)
	}
	// 3. 尝试 GB18030（兼容 GBK/GB2312）解码为 UTF-8
	decoded, err := simplifiedchinese.GB18030.NewDecoder().Bytes(data)
	if err != nil {
		return string(data)
	}
	return string(decoded)
}

// UpdateDocumentContent 更新文档内容并重新分块向量化
func (s *Service) UpdateDocumentContent(ctx context.Context, id int64, newContent string) error {
	// 1. 获取存储路径和文件类型
	var storagePath, fileType string
	err := s.DB.QueryRowContext(ctx,
		`SELECT COALESCE(storage_path,''), file_type FROM public.knowledge_documents WHERE id=$1`, id,
	).Scan(&storagePath, &fileType)
	if err != nil {
		return fmt.Errorf("文档不存在: %w", err)
	}

	// 2. 写回文件
	if err := os.WriteFile(storagePath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("保存文件失败: %w", err)
	}

	// 3. 优先使用 LlamaIndex 重新索引
	if s.RAG != nil {
		chunkCount, err := s.RAG.IndexDocument(ctx, id, newContent)
		if err != nil {
			// RAG 失败时降级到 Go 本地分块 + 向量化
			fmt.Printf("[WARN] LlamaIndex 重新索引失败，降级到本地向量化: %v\n", err)
		} else {
			_, err = s.DB.ExecContext(ctx,
				`UPDATE public.knowledge_documents SET status='ready', chunk_count=$2, updated_at=NOW() WHERE id=$1`,
				id, chunkCount)
			return err
		}
	}

	// 4. 回退：Go 原有分块 + 向量化
	_, _ = s.DB.ExecContext(ctx, `DELETE FROM public.knowledge_chunks WHERE document_id=$1`, id)

	chunks := ChunkText(newContent, DefaultChunkOptions())
	if len(chunks) == 0 {
		_, _ = s.DB.ExecContext(ctx,
			`UPDATE public.knowledge_documents SET status='failed', chunk_count=0, updated_at=NOW() WHERE id=$1`, id)
		return fmt.Errorf("内容为空，无法分块")
	}

	for i, chunk := range chunks {
		vec, err := s.Ollama.Embedding(ctx, s.EmbedModel, chunk)
		if err != nil {
			_, _ = s.DB.ExecContext(ctx,
				`UPDATE public.knowledge_documents SET status='failed', updated_at=NOW() WHERE id=$1`, id)
			return fmt.Errorf("向量化第 %d 块失败: %w", i+1, err)
		}
		_, err = s.DB.ExecContext(ctx,
			`INSERT INTO public.knowledge_chunks (document_id, chunk_index, content, embedding)
			 VALUES ($1, $2, $3, $4::vector)`,
			id, i, chunk, vectorLiteral(vec))
		if err != nil {
			_, _ = s.DB.ExecContext(ctx,
				`UPDATE public.knowledge_documents SET status='failed', updated_at=NOW() WHERE id=$1`, id)
			return fmt.Errorf("保存第 %d 块失败: %w", i+1, err)
		}
	}

	_, err = s.DB.ExecContext(ctx,
		`UPDATE public.knowledge_documents
		 SET status='ready', chunk_count=$2, updated_at=NOW()
		 WHERE id=$1`, id, len(chunks))
	return err
}

// SaveImage 保存图片文档（不解析文本内容，只存储和记录）
func (s *Service) SaveImage(ctx context.Context, doc *Document, filePath string, description string) error {
	var docID int64
	var createdAt time.Time
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO public.knowledge_documents
		 (title, category, description, file_name, file_type, file_size, storage_path, status, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, 'ready', $8)
		 RETURNING id, created_at`,
		doc.Title, doc.Category, description, doc.FileName, doc.FileType, doc.FileSize, filePath, doc.CreatedBy,
	).Scan(&docID, &createdAt)
	if err != nil {
		return fmt.Errorf("插入图片记录失败: %w", err)
	}
	doc.ID = docID
	doc.CreatedAt = createdAt
	doc.Status = "ready"
	doc.ChunkCount = 0
	return nil
}

// ListDocuments 文档列表
func (s *Service) ListDocuments(ctx context.Context, category string, current, size int) ([]Document, int, error) {
	where := ""
	args := []interface{}{}
	if category != "" && category != "all" {
		where = " WHERE category = $1"
		args = append(args, category)
	}

	var total int
	countQ := `SELECT COUNT(*) FROM public.knowledge_documents` + where
	if err := s.DB.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (current - 1) * size
	query := `SELECT id, title, category, description, file_name, file_type, file_size, status, chunk_count, created_by, created_at
		FROM public.knowledge_documents` + where +
		` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) +
		` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, size, offset)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var docs []Document
	for rows.Next() {
		var d Document
		if err := rows.Scan(&d.ID, &d.Title, &d.Category, &d.Description, &d.FileName, &d.FileType,
			&d.FileSize, &d.Status, &d.ChunkCount, &d.CreatedBy, &d.CreatedAt); err != nil {
			return nil, 0, err
		}
		docs = append(docs, d)
	}
	return docs, total, nil
}

// DeleteDocument 删除文档（级联删除分块）
func (s *Service) DeleteDocument(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM public.knowledge_documents WHERE id=$1`, id)
	return err
}

// Search 语义搜索（文档分块 + FAQ）
func (s *Service) Search(ctx context.Context, query string, topK int, minScore float64) ([]SearchResult, error) {
	if topK <= 0 {
		topK = 8
	}
	if minScore <= 0 {
		minScore = 0.3
	}

	// 优先使用 LlamaIndex Python RAG 服务
	if s.RAG != nil {
		pyResults, err := s.RAG.Search(ctx, query, topK, minScore)
		if err != nil {
			// Python 服务失败时回退到 Go 本地检索
			fmt.Printf("[WARN] LlamaIndex 搜索失败，回退本地: %v\n", err)
		} else {
			return pyResults, nil
		}
	}

	// 回退：Go 本地 pgvector 检索
	vec, err := s.Ollama.Embedding(ctx, s.EmbedModel, query)
	if err != nil {
		return nil, fmt.Errorf("查询向量化失败: %w", err)
	}
	vecLit := vectorLiteral(vec)

	results := []SearchResult{}

	// 1. 检索文档分块
	rows, err := s.DB.QueryContext(ctx, `
		SELECT c.document_id, d.title, d.category, c.content, c.embedding <=> $1::vector AS distance
		FROM public.knowledge_chunks c
		JOIN public.knowledge_documents d ON d.id = c.document_id
		WHERE d.status = 'ready'
		ORDER BY c.embedding <=> $1::vector
		LIMIT $2`, vecLit, topK)
	if err != nil {
		return nil, fmt.Errorf("检索文档失败: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r SearchResult
		var distance float64
		if err := rows.Scan(&r.DocumentID, &r.DocumentName, &r.Category, &r.Content, &distance); err != nil {
			return nil, err
		}
		r.Score = 1 - distance
		r.Source = "document"
		if r.Score >= minScore {
			results = append(results, r)
		}
	}

	// 2. 检索 FAQ
	faqRows, err := s.DB.QueryContext(ctx, `
		SELECT id, question, answer, category, embedding <=> $1::vector AS distance
		FROM public.knowledge_faqs
		ORDER BY embedding <=> $1::vector
		LIMIT $2`, vecLit, topK)
	if err == nil {
		defer faqRows.Close()
		for faqRows.Next() {
			var r SearchResult
			var distance float64
			if err := faqRows.Scan(&r.FAQID, &r.Question, &r.Answer, &r.Category, &distance); err != nil {
				continue
			}
			r.Score = 1 - distance
			r.Source = "faq"
			r.Content = r.Question + "\n" + r.Answer
			if r.Score >= minScore {
				results = append(results, r)
			}
		}
	}

	// 按分数降序，截取 topK
	sortResults(results)
	if len(results) > topK {
		results = results[:topK]
	}
	return results, nil
}

// sortResults 按分数降序
func sortResults(results []SearchResult) {
	for i := 1; i < len(results); i++ {
		for j := i; j > 0 && results[j].Score > results[j-1].Score; j-- {
			results[j], results[j-1] = results[j-1], results[j]
		}
	}
}

// SaveFAQ 新增 FAQ
func (s *Service) SaveFAQ(ctx context.Context, faq *FAQ) error {
	tagsJSON, _ := json.Marshal(faq.Tags)

	// 优先使用 LlamaIndex RAG 服务做向量化
	if s.RAG != nil {
		// 先插入获取 ID
		err := s.DB.QueryRowContext(ctx,
			`INSERT INTO public.knowledge_faqs (question, answer, category, tags, embedding, created_by)
			 VALUES ($1, $2, $3, $4, '[0]'::vector, $5) RETURNING id, created_at`,
			faq.Question, faq.Answer, faq.Category, string(tagsJSON), faq.CreatedBy,
		).Scan(&faq.ID, &faq.CreatedAt)
		if err != nil {
			return fmt.Errorf("插入 FAQ 失败: %w", err)
		}
		// 让 Python 服务生成正确的 embedding 并更新
		if err := s.RAG.IndexFAQ(ctx, faq.ID, faq.Question, faq.Answer); err != nil {
			fmt.Printf("[WARN] LlamaIndex FAQ 向量化失败: %v\n", err)
		}
		return nil
	}

	// 回退：Go 本地向量化
	vec, err := s.Ollama.Embedding(ctx, s.EmbedModel, faq.Question+"\n"+faq.Answer)
	if err != nil {
		return fmt.Errorf("FAQ 向量化失败: %w", err)
	}

	err = s.DB.QueryRowContext(ctx,
		`INSERT INTO public.knowledge_faqs (question, answer, category, tags, embedding, created_by)
		 VALUES ($1, $2, $3, $4, $5::vector, $6) RETURNING id, created_at`,
		faq.Question, faq.Answer, faq.Category, string(tagsJSON), vectorLiteral(vec), faq.CreatedBy,
	).Scan(&faq.ID, &faq.CreatedAt)
	return err
}

// ListFAQs FAQ 列表
func (s *Service) ListFAQs(ctx context.Context, category string, current, size int) ([]FAQ, int, error) {
	where := ""
	args := []interface{}{}
	if category != "" && category != "all" {
		where = " WHERE category = $1"
		args = append(args, category)
	}

	var total int
	countQ := `SELECT COUNT(*) FROM public.knowledge_faqs` + where
	if err := s.DB.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (current - 1) * size
	query := `SELECT id, question, answer, category, tags, created_by, created_at
		FROM public.knowledge_faqs` + where +
		` ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", len(args)+1) +
		` OFFSET $` + fmt.Sprintf("%d", len(args)+2)
	args = append(args, size, offset)

	rows, err := s.DB.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var faqs []FAQ
	for rows.Next() {
		var f FAQ
		var tagsJSON string
		if err := rows.Scan(&f.ID, &f.Question, &f.Answer, &f.Category, &tagsJSON, &f.CreatedBy, &f.CreatedAt); err != nil {
			return nil, 0, err
		}
		_ = json.Unmarshal([]byte(tagsJSON), &f.Tags)
		faqs = append(faqs, f)
	}
	return faqs, total, nil
}

// DeleteFAQ 删除 FAQ
func (s *Service) DeleteFAQ(ctx context.Context, id int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM public.knowledge_faqs WHERE id=$1`, id)
	return err
}

// BuildContext 构建 RAG 上下文提示词
func (s *Service) BuildContext(ctx context.Context, query string) (string, []SearchResult, error) {
	results, err := s.Search(ctx, query, 6, 0.3)
	if err != nil {
		return "", nil, err
	}

	var sb strings.Builder
	for i, r := range results {
		sb.WriteString(fmt.Sprintf("[参考资料 %d]\n", i+1))
		if r.Source == "faq" {
			sb.WriteString(fmt.Sprintf("问题: %s\n答案: %s\n", r.Question, r.Answer))
		} else {
			sb.WriteString(fmt.Sprintf("来源文档: %s\n内容: %s\n", r.DocumentName, r.Content))
		}
		sb.WriteString("\n")
	}
	return sb.String(), results, nil
}

// SaveConversation 保存对话
func (s *Service) SaveConversation(ctx context.Context, userID string) (int64, error) {
	var id int64
	err := s.DB.QueryRowContext(ctx,
		`INSERT INTO public.knowledge_conversations (title, user_id) VALUES ('新对话', $1) RETURNING id`,
		userID).Scan(&id)
	return id, err
}

// UpdateConversationTitle 更新对话标题
func (s *Service) UpdateConversationTitle(ctx context.Context, conversationID int64, title string) error {
	_, err := s.DB.ExecContext(ctx,
		`UPDATE public.knowledge_conversations SET title = $1, updated_at = NOW() WHERE id = $2`,
		title, conversationID)
	return err
}

// SaveMessage 保存消息
func (s *Service) SaveMessage(ctx context.Context, conversationID int64, role, content string, sources []SearchResult) error {
	sourcesJSON := "[]"
	if len(sources) > 0 {
		simple := make([]map[string]interface{}, 0, len(sources))
		for _, s := range sources {
			item := map[string]interface{}{
				"source":    s.Source,
				"document":  s.DocumentName,
				"question":  s.Question,
				"score":     s.Score,
				"content":   truncateString(s.Content, 200),
			}
			simple = append(simple, item)
		}
		b, err := json.Marshal(simple)
		if err == nil {
			sourcesJSON = string(b)
		}
	}

	_, err := s.DB.ExecContext(ctx,
		`INSERT INTO public.knowledge_messages (conversation_id, role, content, sources)
		 VALUES ($1, $2, $3, $4)`,
		conversationID, role, content, sourcesJSON)
	return err
}

// GetConversationMessages 获取对话历史
func (s *Service) GetConversationMessages(ctx context.Context, conversationID int64) ([]map[string]interface{}, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT role, content FROM public.knowledge_messages
		 WHERE conversation_id=$1 ORDER BY id ASC`, conversationID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []map[string]interface{}
	for rows.Next() {
		var role, content string
		if err := rows.Scan(&role, &content); err != nil {
			return nil, err
		}
		messages = append(messages, map[string]interface{}{"role": role, "content": content})
	}
	return messages, nil
}

// ListConversations 获取会话列表
func (s *Service) ListConversations(ctx context.Context, userID string) ([]map[string]interface{}, error) {
	rows, err := s.DB.QueryContext(ctx,
		`SELECT c.id, c.title, c.created_at,
		        (SELECT COUNT(*) FROM public.knowledge_messages m WHERE m.conversation_id = c.id) AS msg_count
		 FROM public.knowledge_conversations c
		 WHERE c.user_id = $1
		 ORDER BY c.updated_at DESC
		 LIMIT 50`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var convs []map[string]interface{}
	for rows.Next() {
		var id int64
		var title string
		var createdAt time.Time
		var msgCount int
		if err := rows.Scan(&id, &title, &createdAt, &msgCount); err != nil {
			return nil, err
		}
		convs = append(convs, map[string]interface{}{
			"id":        id,
			"title":     title,
			"created_at": createdAt,
			"msg_count": msgCount,
		})
	}
	return convs, nil
}

// DeleteConversation 删除会话及其消息
func (s *Service) DeleteConversation(ctx context.Context, conversationID int64) error {
	_, err := s.DB.ExecContext(ctx, `DELETE FROM public.knowledge_messages WHERE conversation_id=$1`, conversationID)
	if err != nil {
		return err
	}
	_, err = s.DB.ExecContext(ctx, `DELETE FROM public.knowledge_conversations WHERE id=$1`, conversationID)
	return err
}

// Stats 知识库统计
func (s *Service) Stats(ctx context.Context) (map[string]interface{}, error) {
	result := map[string]interface{}{}

	var docCount, chunkCount, faqCount int
	_ = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM public.knowledge_documents`).Scan(&docCount)
	_ = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM public.knowledge_chunks`).Scan(&chunkCount)
	_ = s.DB.QueryRowContext(ctx, `SELECT COUNT(*) FROM public.knowledge_faqs`).Scan(&faqCount)

	result["documents"] = docCount
	result["chunks"] = chunkCount
	result["faqs"] = faqCount

	// 模型状态
	var ollamaOK bool
	ollamaStatus := "unknown"
	if err := s.Ollama.Ping(ctx); err == nil {
		ollamaOK = true
		ollamaStatus = "online"
	} else {
		ollamaStatus = "offline"
	}
	result["ollama_online"] = ollamaOK
	result["ollama_status"] = ollamaStatus
	result["embed_model"] = s.EmbedModel
	result["chat_model"] = s.ChatModel

	// 已安装模型
	if ollamaOK {
		models, err := s.Ollama.ListModels(ctx)
		if err == nil {
			names := make([]string, 0, len(models))
			for _, m := range models {
				names = append(names, m.Name)
			}
			result["models"] = names
		}
	}
	return result, nil
}

// GetStoragePath 生成文档存储路径
func (s *Service) GetStoragePath(fileName string) string {
	ext := filepath.Ext(fileName)
	base := strings.TrimSuffix(fileName, ext)
	ts := time.Now().Format("20060102150405")
	return filepath.Join(s.StoreDir, fmt.Sprintf("%s_%s%s", base, ts, ext))
}

func truncateString(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
