package handlers

import (
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"

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

// GetContent 获取文档文本内容（用于在线编辑）
func (h *KnowledgeHandler) GetContent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	content, err := h.Service.GetDocumentContent(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, err.Error())
		return
	}
	response.Success(c, gin.H{
		"id":      id,
		"content": content,
	})
}

// UpdateContent 更新文档文本内容（保存后重新分块向量化）
func (h *KnowledgeHandler) UpdateContent(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req struct {
		Content string `json:"content" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请提供文档内容")
		return
	}
	if err := h.Service.UpdateDocumentContent(c.Request.Context(), id, req.Content); err != nil {
		response.ServerError(c, "更新失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "文档已更新并重新向量化", nil)
}

// UploadImage 上传图片（存储文件并记录，不进行向量化）
func (h *KnowledgeHandler) UploadImage(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		response.BadRequest(c, "请选择要上传的图片")
		return
	}
	defer file.Close()

	// 限制文件大小 10MB
	if header.Size > 10*1024*1024 {
		response.BadRequest(c, "图片过大，最大支持 10MB")
		return
	}

	data, err := io.ReadAll(file)
	if err != nil {
		response.ServerError(c, "读取图片失败: "+err.Error())
		return
	}

	// 校验图片类型
	ext := strings.ToLower(filepath.Ext(header.Filename))
	imageTypes := map[string]bool{".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".webp": true, ".bmp": true, ".svg": true}
	if !imageTypes[ext] {
		response.BadRequest(c, "不支持的图片格式，支持: jpg/png/gif/webp/bmp/svg")
		return
	}

	title := c.PostForm("title")
	description := c.PostForm("description")
	category := c.PostForm("category")
	if title == "" {
		title = header.Filename
	}
	if category == "" {
		category = "image"
	}

	// 保存文件
	storagePath := h.Service.GetStoragePath(header.Filename)
	if err := os.MkdirAll(filepath.Dir(storagePath), 0755); err != nil {
		response.ServerError(c, "创建存储目录失败: "+err.Error())
		return
	}
	if err := os.WriteFile(storagePath, data, 0644); err != nil {
		response.ServerError(c, "保存图片失败: "+err.Error())
		return
	}

	userID, _ := c.Get("username")
	doc := &knowledge.Document{
		Title:    title,
		Category: category,
		FileName: header.Filename,
		FileType: ext,
		FileSize: header.Size,
		CreatedBy: userID.(string),
	}

	if err := h.Service.SaveImage(c.Request.Context(), doc, storagePath, description); err != nil {
		_ = os.Remove(storagePath)
		response.ServerError(c, "保存图片记录失败: "+err.Error())
		return
	}

	response.SuccessWithMessage(c, "图片上传成功", doc)
}

// ServeFile 提供文档/图片文件访问
func (h *KnowledgeHandler) ServeFile(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	doc, err := h.Service.GetDocument(c.Request.Context(), id)
	if err != nil {
		response.NotFound(c, "文档不存在")
		return
	}
	if doc.StoragePath == "" {
		response.NotFound(c, "文件不存在")
		return
	}
	// 根据文件扩展名设置 Content-Type
	ext := strings.ToLower(doc.FileType)
	mimeTypes := map[string]string{
		".jpg": "image/jpeg", ".jpeg": "image/jpeg", ".png": "image/png",
		".gif": "image/gif", ".webp": "image/webp", ".bmp": "image/bmp",
		".svg": "image/svg+xml",
		".txt": "text/plain", ".md": "text/markdown", ".markdown": "text/markdown",
		".json": "application/json", ".csv": "text/csv", ".log": "text/plain",
		".pdf": "application/pdf",
		".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		".xls": "application/vnd.ms-excel",
	}
	if mime, ok := mimeTypes[ext]; ok {
		c.Header("Content-Type", mime)
	}
	c.File(doc.StoragePath)
}
