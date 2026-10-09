package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/knowledge"
	"github.com/tester-platform/backend/internal/ollama"
	"github.com/tester-platform/backend/pkg/response"
)

// ChatHandler AI 对话处理器
type ChatHandler struct {
	Service *knowledge.Service
}

// NewChatHandler 创建对话处理器
func NewChatHandler(svc *knowledge.Service) *ChatHandler {
	return &ChatHandler{Service: svc}
}

// Send 发送对话（流式 SSE）
// POST /api/v1/knowledge/chat
// { question: "...", conversation_id: 0, stream: true }
func (h *ChatHandler) Send(c *gin.Context) {
	var req struct {
		Question       string `json:"question" binding:"required"`
		ConversationID int64  `json:"conversation_id"`
		Stream         bool   `json:"stream"`
		Mode           string `json:"mode"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请输入问题")
		return
	}

	ctx := c.Request.Context()
	userID, _ := c.Get("username")
	username, _ := userID.(string)

	// 1. 构建 RAG 上下文
	contextText, sources, err := h.Service.BuildContext(ctx, req.Question)
	if err != nil {
		response.ServerError(c, "检索知识库失败: "+err.Error())
		return
	}

	// 2. 获取对话历史
	var history []map[string]interface{}
	if req.ConversationID > 0 {
		history, _ = h.Service.GetConversationMessages(ctx, req.ConversationID)
	}

	// 3. 构建消息
	messages := []ollama.ChatMessage{}
	systemPrompt := buildSystemPrompt(req.Mode, contextText)

	messages = append(messages, ollama.ChatMessage{Role: "system", Content: systemPrompt})

	// 追加历史（最多保留最近 6 条）
	start := 0
	if len(history) > 6 {
		start = len(history) - 6
	}
	for _, m := range history[start:] {
		role, _ := m["role"].(string)
		content, _ := m["content"].(string)
		messages = append(messages, ollama.ChatMessage{Role: role, Content: content})
	}
	messages = append(messages, ollama.ChatMessage{Role: "user", Content: req.Question})

	// 4. 保存用户消息 + 创建会话
	var convID int64 = req.ConversationID
	isNew := convID == 0
	if isNew {
		convID, _ = h.Service.SaveConversation(ctx, username)
	}
	_ = h.Service.SaveMessage(ctx, convID, "user", req.Question, nil)

	// 首次消息时，用用户问题前16字更新对话标题
	if isNew {
		title := req.Question
		if len([]rune(title)) > 16 {
			title = string([]rune(title)[:16]) + "..."
		}
		_ = h.Service.UpdateConversationTitle(ctx, convID, title)
	}

	// 5. 流式输出
	if req.Stream {
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("X-Accel-Buffering", "no")

		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Flush()

		// 发送会话信息
		meta, _ := json.Marshal(map[string]interface{}{
			"type":            "meta",
			"conversation_id": convID,
			"source_count":    len(sources),
			"is_new":          isNew,
		})
		writeSSE(c, meta)

		var answer strings.Builder
		streamErr := h.Service.Ollama.ChatStream(ctx, h.Service.ChatModel, messages, func(chunk string) error {
			answer.WriteString(chunk)
			data, err := json.Marshal(map[string]interface{}{
				"type":   "chunk",
				"content": chunk,
			})
			if err != nil {
				return err
			}
			return writeSSE(c, data)
		})

		if streamErr != nil {
			errData, _ := json.Marshal(map[string]interface{}{
				"type":  "error",
				"error": streamErr.Error(),
			})
			writeSSE(c, errData)
			_ = h.Service.SaveMessage(ctx, convID, "assistant", "[生成失败] "+streamErr.Error(), nil)
			return
		}

		// 保存 AI 回答
		_ = h.Service.SaveMessage(ctx, convID, "assistant", answer.String(), sources)

		done, _ := json.Marshal(map[string]interface{}{
			"type":    "done",
			"answer":  answer.String(),
			"sources": sources,
		})
		writeSSE(c, done)
		return
	}

	// 6. 非流式
	answer, err := h.Service.Ollama.Chat(ctx, h.Service.ChatModel, messages)
	if err != nil {
		response.ServerError(c, "AI 调用失败: "+err.Error())
		return
	}
	_ = h.Service.SaveMessage(ctx, convID, "assistant", answer, sources)

	response.Success(c, gin.H{
		"conversation_id": convID,
		"answer":          answer,
		"sources":         sources,
		"is_new":          isNew,
	})
}

// Conversations 会话列表
func (h *ChatHandler) Conversations(c *gin.Context) {
	userID, _ := c.Get("username")
	convs, err := h.Service.ListConversations(c.Request.Context(), userID.(string))
	if err != nil {
		response.ServerError(c, "获取会话失败: "+err.Error())
		return
	}
	response.Success(c, convs)
}

// Delete 删除会话
// DELETE /api/v1/knowledge/conversations/:id
func (h *ChatHandler) Delete(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	if id == 0 {
		response.BadRequest(c, "无效的会话 ID")
		return
	}
	if err := h.Service.DeleteConversation(c.Request.Context(), id); err != nil {
		response.ServerError(c, "删除会话失败: "+err.Error())
		return
	}
	response.Success(c, nil)
}

// History 获取会话消息
func (h *ChatHandler) History(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	messages, err := h.Service.GetConversationMessages(c.Request.Context(), id)
	if err != nil {
		response.ServerError(c, "获取历史失败: "+err.Error())
		return
	}
	response.Success(c, messages)
}

// writeSSE 写入 SSE 数据帧
func writeSSE(c *gin.Context, data []byte) error {
	if _, err := c.Writer.WriteString("data: " + string(data) + "\n\n"); err != nil {
		return err
	}
	c.Writer.Flush()
	return nil
}

// buildSystemPrompt 根据模式构建系统提示词
// mode 为空或 "knowledge" → 企业内部知识助手；"customer_service" → 面向终端用户的智能客服
func buildSystemPrompt(mode, contextText string) string {
	if mode == "customer_service" {
		if contextText == "" {
			return "你是我们产品的智能客服助手，面向终端用户提供支持。请用友好、耐心、专业的中文回答用户问题；" +
				"若参考资料不足，请如实说明并建议联系人工客服。"
		}
		return `你是我们产品的智能客服助手，面向终端用户提供支持。请基于提供的参考资料，用友好、耐心、专业的中文回答客户问题。

回答要求：
1. 优先使用参考资料中的内容回答，不要编造
2. 如资料不足以回答，如实说明并建议联系人工客服
3. 使用清晰、易懂的语言，必要时给出操作步骤
4. 态度亲切，体现服务意识

参考资料：
` + contextText
	}

	// 默认：企业内部知识助手
	if contextText == "" {
		return "你是一个企业内部智能知识助手，请用简洁中文回答用户问题。"
	}
	return `你是一个企业内部的智能知识助手，请基于提供的参考资料回答用户问题。

回答要求：
1. 优先使用参考资料中的内容回答，不要编造
2. 如果参考资料不足以回答，如实说明
3. 使用简洁、清晰的中文回答
4. 可以列出操作步骤

参考资料：
` + contextText
}
