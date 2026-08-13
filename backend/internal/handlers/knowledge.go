package handlers

import (
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/knowledge"
	"github.com/tester-platform/backend/pkg/response"
)

// KnowledgeHandler 知识库处理器
type KnowledgeHandler struct {
	Service *knowledge.Service
}

// NewKnowledgeHandler 创建知识库处理器
func NewKnowledgeHandler(svc *knowledge.Service) *KnowledgeHandler {
	return &KnowledgeHandler{Service: svc}
}

// Stats 知识库统计 + 模型状态
func (h *KnowledgeHandler) Stats(c *gin.Context) {
	stats, err := h.Service.Stats(c.Request.Context())
	if err != nil {
		response.ServerError(c, "获取统计失败: "+err.Error())
		return
	}
	response.Success(c, stats)
}

// Upload 上传文档并入库
func (h *KnowledgeHandler) Upload(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的文件")
		return
	}
	defer file.Close()

	// 限制文件大小 50MB
	if header.Size > 50*1024*1024 {
		response.BadRequest(c, "文件过大，最大支持 50MB")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		response.ServerError(c, "读取文件失败: "+err.Error())
		return
	}

	title := c.PostForm("title")
	category := c.PostForm("category")
	description := c.PostForm("description")

	if title == "" {
		title = header.Filename
	}
	if category == "" {
		category = "other"
	}

	// 解析文档内容
	content, err := knowledge.ParseDocument(header.Filename, data)
	if err != nil {
		response.BadRequest(c, "文档解析失败: "+err.Error())
		return
	}

	// 保存原始文件
	storagePath := h.Service.GetStoragePath(header.Filename)
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		response.ServerError(c, "创建存储目录失败: "+err.Error())
		return
	}
	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		response.ServerError(c, "保存文件失败: "+err.Error())
		return
	}

	// 入库（向量化）
	userID, _ := c.Get("username")
	doc := &knowledge.Document{
		Title:       title,
		Category:    category,
		Description: description,
		FileName:    header.Filename,
		FileType:    filepath.Ext(header.Filename),
		FileSize:    header.Size,
		CreatedBy:   userID.(string),
	}

	if err := h.Service.SaveDocument(c.Request.Context(), doc, content, storagePath); err != nil {
		_ = os.Remove(storagePath) // 清理失败文件
		response.ServerError(c, "文档入库失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "文档上传并入库成功", doc)
}

// List 文档列表
func (h *KnowledgeHandler) List(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	category := c.Query("category")

	docs, total, err := h.Service.ListDocuments(c.Request.Context(), category, current, size)
	if err != nil {
		response.ServerError(c, "查询文档失败: "+err.Error())
		return
	}
	response.Page(c, docs, total, current, size)
}

// Get 文档详情
func (h *KnowledgeHandler) Get(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	doc, err := h.Service.GetDocument(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "文档不存在")
		return
	}
	response.Success(c, doc)
}

// Delete 删除文档
func (h *KnowledgeHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.Service.DeleteDocument(c.Request.Context(), id); err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// Search 语义搜索
func (h *KnowledgeHandler) Search(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		query = c.PostForm("q")
	}
	if query == "" {
		response.BadRequest(c, "请输入搜索内容")
		return
	}

	topK, _ := strconv.Atoi(c.DefaultQuery("top_k", "8"))
	minScore, _ := strconv.ParseFloat(c.DefaultQuery("min_score", "0.3"), 64)

	results, err := h.Service.Search(c.Request.Context(), query, topK, minScore)
	if err != nil {
		response.ServerError(c, "搜索失败: "+err.Error())
		return
	}
	response.Success(c, gin.H{
		"query":   query,
		"results": results,
		"count":   len(results),
	})
}

// CreateFAQ 新增问题库条目
func (h *KnowledgeHandler) CreateFAQ(c *gin.Context) {
	var req struct {
		Question string   `json:"question" binding:"required"`
		Answer   string   `json:"answer" binding:"required"`
		Category string   `json:"category"`
		Tags     []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请填写问题与答案")
		return
	}
	if req.Category == "" {
		req.Category = "other"
	}

	userID, _ := c.Get("username")
	faq := &knowledge.FAQ{
		Question:  req.Question,
		Answer:    req.Answer,
		Category:  req.Category,
		Tags:      req.Tags,
		CreatedBy: userID.(string),
	}

	if err := h.Service.SaveFAQ(c.Request.Context(), faq); err != nil {
		response.ServerError(c, "FAQ 入库失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "问题已入库", faq)
}

// ListFAQs 问题库列表
func (h *KnowledgeHandler) ListFAQs(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
	category := c.Query("category")

	faqs, total, err := h.Service.ListFAQs(c.Request.Context(), category, current, size)
	if err != nil {
		response.ServerError(c, "查询问题库失败: "+err.Error())
		return
	}
	response.Page(c, faqs, total, current, size)
}

// DeleteFAQ 删除问题库条目
func (h *KnowledgeHandler) DeleteFAQ(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if err := h.Service.DeleteFAQ(c.Request.Context(), id); err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
