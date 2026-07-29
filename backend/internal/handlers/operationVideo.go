package handlers

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/models"
	"github.com/tester-platform/backend/pkg/response"
)

// OperationVideoHandler 操作视频处理器
type OperationVideoHandler struct {
	DB *sql.DB
}

func NewOperationVideoHandler(db *sql.DB) *OperationVideoHandler {
	return &OperationVideoHandler{DB: db}
}

const videoSelect = "id,title,category,description,url,duration,views,date"
const videoOrder = "date DESC"

func (h *OperationVideoHandler) List(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	category := c.Query("category")
	if current < 1 {
		current = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	var conds []string
	var args []interface{}
	idx := 1
	if keyword != "" {
		like := "%" + keyword + "%"
		conds = append(conds, fmt.Sprintf("(title ILIKE $%d OR description ILIKE $%d)", idx, idx))
		args = append(args, like)
		idx++
	}
	if category != "" {
		conds = append(conds, fmt.Sprintf("category = $%d", idx))
		args = append(args, category)
		idx++
	}

	where := ""
	if len(conds) > 0 {
		where = " WHERE " + strings.Join(conds, " AND ")
	}

	var total int
	h.DB.QueryRow("SELECT COUNT(*) FROM public.operation_videos"+where, args...).Scan(&total)

	offset := (current - 1) * size
	query := fmt.Sprintf("SELECT %s FROM public.operation_videos%s ORDER BY %s LIMIT %d OFFSET %d",
		videoSelect, where, videoOrder, size, offset)
	rows, err := queryRowsToMaps(h.DB, query, args...)
	if err != nil {
		response.ServerError(c, "查询失败: "+err.Error())
		return
	}
	response.Page(c, rows, total, current, size)
}

func (h *OperationVideoHandler) Get(c *gin.Context) {
	id := c.Param("id")
	// 访问时增加 views 计数
	h.DB.Exec("UPDATE public.operation_videos SET views = views + 1 WHERE id = $1", id)
	rows, err := queryRowsToMaps(h.DB, fmt.Sprintf("SELECT %s FROM public.operation_videos WHERE id = $1", videoSelect), id)
	if err != nil || len(rows) == 0 {
		response.NotFound(c, "记录不存在")
		return
	}
	response.Success(c, rows[0])
}

func (h *OperationVideoHandler) Create(c *gin.Context) {
	var req models.OperationVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	rows, err := queryRowsToMaps(h.DB,
		`INSERT INTO public.operation_videos (title,category,description,url,duration)
		 VALUES ($1,$2,$3,$4,$5) RETURNING `+videoSelect,
		req.Title, req.Category, req.Description, req.URL, req.Duration)
	if err != nil || len(rows) == 0 {
		response.ServerError(c, "创建失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "创建成功", rows[0])
}

func (h *OperationVideoHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.OperationVideoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}
	rows, err := queryRowsToMaps(h.DB,
		`UPDATE public.operation_videos SET title=$1,category=$2,description=$3,url=$4,duration=$5
		 WHERE id=$6 RETURNING `+videoSelect,
		req.Title, req.Category, req.Description, req.URL, req.Duration, id)
	if err != nil || len(rows) == 0 {
		response.ServerError(c, "更新失败或记录不存在")
		return
	}
	response.SuccessWithMessage(c, "更新成功", rows[0])
}

func (h *OperationVideoHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	_, err := h.DB.Exec("DELETE FROM public.operation_videos WHERE id = $1", id)
	if err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}
