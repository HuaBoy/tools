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

// OverseasShippingHandler 海外发货管理处理器
type OverseasShippingHandler struct {
	DB *sql.DB
}

// NewOverseasShippingHandler 创建处理器
func NewOverseasShippingHandler(db *sql.DB) *OverseasShippingHandler {
	return &OverseasShippingHandler{DB: db}
}

const overseasSelect = "id,product_type,sales_engineer,country,customer_name,controller_sn,handheld_sn,controller_hw_version,controller_upgrade_history,handheld_upgrade_history,last_upgrade_date,assistant,remark,created_at,updated_at"

// overseasShippingColumns OverseaShipping struct 的非自增列（INSERT / UPDATE 用）
var overseasShippingColumns = []string{
	"product_type", "sales_engineer", "country", "customer_name",
	"controller_sn", "handheld_sn", "controller_hw_version",
	"controller_upgrade_history", "handheld_upgrade_history",
	"last_upgrade_date", "assistant", "remark",
}

// List 列表（分页 + 关键字搜索 + 产品类型/国家筛选）
func (h *OverseasShippingHandler) List(c *gin.Context) {
	current, _ := strconv.Atoi(c.DefaultQuery("current", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	keyword := c.Query("keyword")
	productType := c.Query("product_type")
	country := c.Query("country")

	if current < 1 {
		current = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	var conditions []string
	var args []interface{}
	argIdx := 1

	if keyword != "" {
		like := "%" + keyword + "%"
		cols := []string{"product_type", "sales_engineer", "country", "customer_name",
			"controller_sn", "handheld_sn", "controller_upgrade_history",
			"handheld_upgrade_history", "assistant", "remark"}
		var orClauses []string
		for _, col := range cols {
			orClauses = append(orClauses, fmt.Sprintf("%s ILIKE $%d", col, argIdx))
		}
		args = append(args, like)
		conditions = append(conditions, "("+strings.Join(orClauses, " OR ")+")")
		argIdx++
	}
	if productType != "" {
		conditions = append(conditions, fmt.Sprintf("product_type = $%d", argIdx))
		args = append(args, productType)
		argIdx++
	}
	if country != "" {
		conditions = append(conditions, fmt.Sprintf("country = $%d", argIdx))
		args = append(args, country)
		argIdx++
	}

	where := ""
	if len(conditions) > 0 {
		where = " WHERE " + strings.Join(conditions, " AND ")
	}

	// Count
	var total int
	countQuery := "SELECT COUNT(*) FROM public.overseas_shipping" + where
	if err := h.DB.QueryRow(countQuery, args...).Scan(&total); err != nil {
		response.ServerError(c, "查询失败: "+err.Error())
		return
	}

	// Data
	offset := (current - 1) * size
	dataQuery := fmt.Sprintf("SELECT %s FROM public.overseas_shipping%s ORDER BY created_at DESC LIMIT %d OFFSET %d",
		overseasSelect, where, size, offset)

	rows, qerr := queryRowsToMaps(h.DB, dataQuery, args...)
	if qerr != nil {
		response.ServerError(c, "查询失败: "+qerr.Error())
		return
	}

	response.Page(c, rows, total, current, size)
}

// Get 获取单条
func (h *OverseasShippingHandler) Get(c *gin.Context) {
	id := c.Param("id")
	rows, err := queryRowsToMaps(h.DB, fmt.Sprintf("SELECT %s FROM public.overseas_shipping WHERE id = $1", overseasSelect), id)
	if err != nil || len(rows) == 0 {
		response.NotFound(c, "记录不存在")
		return
	}
	response.Success(c, rows[0])
}

// Create 新增
func (h *OverseasShippingHandler) Create(c *gin.Context) {
	var req models.OverseasShippingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	cols := overseasShippingColumns
	placeholders := make([]string, len(cols))
	vals := make([]interface{}, len(cols))
	for i, col := range cols {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		vals[i] = getOverseasValue(col, &req)
	}

	query := fmt.Sprintf("INSERT INTO public.overseas_shipping (%s) VALUES (%s) RETURNING %s",
		strings.Join(cols, ","), strings.Join(placeholders, ","), overseasSelect)

	rows, err := queryRowsToMaps(h.DB, query, vals...)
	if err != nil {
		response.ServerError(c, "创建失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "新增成功", rows[0])
}

// Update 更新
func (h *OverseasShippingHandler) Update(c *gin.Context) {
	id := c.Param("id")
	var req models.OverseasShippingCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数错误: "+err.Error())
		return
	}

	setClauses := make([]string, len(overseasShippingColumns))
	vals := make([]interface{}, len(overseasShippingColumns)+1)
	for i, col := range overseasShippingColumns {
		setClauses[i] = fmt.Sprintf("%s = $%d", col, i+1)
		vals[i] = getOverseasValue(col, &req)
	}
	vals[len(overseasShippingColumns)] = id

	query := fmt.Sprintf("UPDATE public.overseas_shipping SET %s WHERE id = $%d RETURNING %s",
		strings.Join(setClauses, ","), len(overseasShippingColumns)+1, overseasSelect)

	rows, err := queryRowsToMaps(h.DB, query, vals...)
	if err != nil || len(rows) == 0 {
		response.ServerError(c, "更新失败或记录不存在", )
		return
	}
	response.SuccessWithMessage(c, "更新成功", rows[0])
}

// Delete 删除
func (h *OverseasShippingHandler) Delete(c *gin.Context) {
	id := c.Param("id")
	_, err := h.DB.Exec("DELETE FROM public.overseas_shipping WHERE id = $1", id)
	if err != nil {
		response.ServerError(c, "删除失败: "+err.Error())
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// overseasErrMsg （保留旧签名，新版不再使用）
func overseasErrMsg(prefix string, err error, _ ...interface{}) string {
	if err != nil {
		return prefix + ": " + err.Error()
	}
	return prefix
}

// getOverseasValue 从请求结构中取对应列的值
func getOverseasValue(col string, req *models.OverseasShippingCreateRequest) interface{} {
	switch col {
	case "product_type":
		return req.ProductType
	case "sales_engineer":
		return req.SalesEngineer
	case "country":
		return req.Country
	case "customer_name":
		return req.CustomerName
	case "controller_sn":
		return req.ControllerSn
	case "handheld_sn":
		return req.HandheldSn
	case "controller_hw_version":
		return req.ControllerHwVersion
	case "controller_upgrade_history":
		return req.ControllerUpgradeHistory
	case "handheld_upgrade_history":
		return req.HandheldUpgradeHistory
	case "last_upgrade_date":
		if req.LastUpgradeDate == "" {
			return nil
		}
		return req.LastUpgradeDate
	case "assistant":
		return req.Assistant
	case "remark":
		return req.Remark
	}
	return ""
}
