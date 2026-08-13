package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tester-platform/backend/internal/models"
	"github.com/tester-platform/backend/pkg/response"
)

// PipelineHandler AI 研发流水线处理器（REQ-2026-001）
type PipelineHandler struct {
	DB *sql.DB
}

func NewPipelineHandler(db *sql.DB) *PipelineHandler {
	return &PipelineHandler{DB: db}
}

// 流水线七个环节（员工 A~G）
var pipelineStageDefs = []string{"A", "B", "C", "D", "E", "F", "G"}

// 需求状态机（key=当前状态，value=允许流转到的状态）
var requirementStatusTransitions = map[string][]string{
	"draft":    {"backlog"},
	"backlog":  {"prd", "rejected"},
	"prd":      {"dev", "rejected"},
	"dev":      {"test", "rejected"},
	"test":     {"verify", "rejected"},
	"verify":   {"review", "rejected"},
	"review":   {"done"},
	"rework":   {"backlog", "prd", "dev", "test", "verify", "review"},
	"rejected": {},
	"done":     {},
}

// 状态 → 各环节状态计划（同步到 ai_pipeline_stages）
func pipelineStagePlan(status string) map[string]string {
	switch status {
	case "draft":
		return map[string]string{"A": "doing", "B": "todo", "C": "todo", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}
	case "backlog":
		return map[string]string{"A": "done", "B": "todo", "C": "todo", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}
	case "prd":
		return map[string]string{"A": "done", "B": "doing", "C": "todo", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}
	case "dev":
		return map[string]string{"A": "done", "B": "done", "C": "doing", "D": "doing", "E": "todo", "F": "todo", "G": "todo"}
	case "test":
		return map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "doing", "G": "todo"}
	case "verify":
		return map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "done", "G": "doing"}
	case "review", "done", "rejected":
		return map[string]string{"A": "done", "B": "done", "C": "done", "D": "done", "E": "na", "F": "done", "G": "done"}
	default:
		return map[string]string{"A": "todo", "B": "todo", "C": "todo", "D": "todo", "E": "todo", "F": "todo", "G": "todo"}
	}
}

func containsString(list []string, v string) bool {
	for _, s := range list {
		if s == v {
			return true
		}
	}
	return false
}

// parseAarrrImpacts 将存储的 JSON 字符串解析为数组
func parseAarrrImpacts(m map[string]interface{}, field string) {
	if s, ok := m[field].(string); ok {
		var arr []string
		if err := json.Unmarshal([]byte(s), &arr); err == nil {
			m[field] = arr
		} else {
			m[field] = []string{}
		}
	}
}

// formatRowTimes 将 time.Time 格式化为可读字符串
func formatRowTimes(m map[string]interface{}, fields ...string) {
	for _, f := range fields {
		if t, ok := m[f].(time.Time); ok {
			m[f] = t.Local().Format("2006-01-02 15:04:05")
		}
	}
}

const pipelineReqSelect = `SELECT id, req_no AS "reqNo", title, description, kano_category AS "kanoCategory",
	aarrr_impacts AS "aarrrImpacts", priority, status, source, created_by AS "createdBy",
	created_at AS "createdAt", updated_at AS "updatedAt"
	FROM public.ai_pipeline_requirements`

// nextReqNo 生成需求编号 REQ-YYYY-NNN
func (h *PipelineHandler) nextReqNo(db *sql.DB) (string, error) {
	prefix := "REQ-" + time.Now().Format("2006") + "-"
	var last string
	err := db.QueryRow(`SELECT req_no FROM public.ai_pipeline_requirements
		WHERE req_no LIKE $1 ORDER BY req_no DESC LIMIT 1`, prefix+"%").Scan(&last)
	if err != nil && err != sql.ErrNoRows {
		return "", err
	}
	seq := 1
	if last != "" {
		if parts := strings.Split(last, "-"); len(parts) > 0 {
			if n, e := strconv.Atoi(parts[len(parts)-1]); e == nil {
				seq = n + 1
			}
		}
	}
	return fmt.Sprintf("%s%03d", prefix, seq), nil
}

// ListRequirements 需求池分页查询
func (h *PipelineHandler) ListRequirements(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	where := []string{"1=1"}
	args := []interface{}{}
	next := 1
	addCond := func(cond string, val interface{}) {
		where = append(where, fmt.Sprintf(cond, next))
		args = append(args, val)
		next++
	}

	if kw := strings.TrimSpace(c.Query("keyword")); kw != "" {
		addCond(`(title ILIKE $%d OR description ILIKE $%d OR req_no ILIKE $%d)`, "%"+kw+"%")
	}
	if v := c.Query("status"); v != "" {
		addCond(`status = $%d`, v)
	}
	if v := c.Query("kano_category"); v != "" {
		addCond(`kano_category = $%d`, v)
	}
	if v := c.Query("priority"); v != "" {
		addCond(`priority = $%d`, v)
	}
	if v := c.Query("aarrr"); v != "" {
		addCond(`aarrr_impacts ILIKE $%d`, "%"+v+"%")
	}
	if v := c.Query("source"); v != "" {
		addCond(`source = $%d`, v)
	}
	whereSQL := strings.Join(where, " AND ")

	var total int
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM public.ai_pipeline_requirements WHERE `+whereSQL, args...).Scan(&total); err != nil {
		response.ServerError(c, "查询需求失败")
		return
	}

	sql := pipelineReqSelect + ` WHERE ` + whereSQL + fmt.Sprintf(` ORDER BY id DESC LIMIT $%d OFFSET $%d`, next, next+1)
	args = append(args, size, (page-1)*size)

	maps, err := queryRowsToMaps(h.DB, sql, args...)
	if err != nil {
		response.ServerError(c, "查询需求失败")
		return
	}
	for _, m := range maps {
		parseAarrrImpacts(m, "aarrrImpacts")
		formatRowTimes(m, "createdAt", "updatedAt")
	}
	response.Page(c, maps, total, page, size)
}

// GetRequirement 需求详情（含环节、交付物、审核记录）
func (h *PipelineHandler) GetRequirement(c *gin.Context) {
	id := c.Param("id")
	maps, err := queryRowsToMaps(h.DB, pipelineReqSelect+` WHERE id = $1`, id)
	if err != nil {
		response.ServerError(c, "查询需求失败")
		return
	}
	if len(maps) == 0 {
		response.NotFound(c, "需求不存在")
		return
	}
	req := maps[0]
	parseAarrrImpacts(req, "aarrrImpacts")
	formatRowTimes(req, "createdAt", "updatedAt")

	stages, err := queryRowsToMaps(h.DB, `SELECT id, req_id AS "reqId", stage, status,
		to_char(started_at, 'YYYY-MM-DD HH24:MI') AS "startedAt",
		to_char(finished_at, 'YYYY-MM-DD HH24:MI') AS "finishedAt", note
		FROM public.ai_pipeline_stages WHERE req_id = $1 ORDER BY stage`, id)
	if err != nil {
		response.ServerError(c, "查询环节失败")
		return
	}
	deliverables, err := queryRowsToMaps(h.DB, `SELECT id, req_id AS "reqId", stage, title, file_type AS "fileType",
		file_path AS "filePath", url, created_by AS "createdBy", created_at AS "createdAt"
		FROM public.ai_pipeline_deliverables WHERE req_id = $1 ORDER BY id DESC`, id)
	if err != nil {
		response.ServerError(c, "查询交付物失败")
		return
	}
	for _, m := range deliverables {
		formatRowTimes(m, "createdAt")
	}
	reviews, err := queryRowsToMaps(h.DB, `SELECT id, req_id AS "reqId", action, target_stage AS "targetStage",
		reason, reviewed_by AS "reviewedBy", reviewed_name AS "reviewedName", created_at AS "createdAt"
		FROM public.ai_pipeline_reviews WHERE req_id = $1 ORDER BY id DESC`, id)
	if err != nil {
		response.ServerError(c, "查询审核记录失败")
		return
	}
	for _, m := range reviews {
		formatRowTimes(m, "createdAt")
	}

	response.Success(c, gin.H{
		"requirement":  req,
		"stages":       stages,
		"deliverables": deliverables,
		"reviews":      reviews,
	})
}

// CreateRequirement 新建需求（自动编号 + 初始化 A~G 环节）
func (h *PipelineHandler) CreateRequirement(c *gin.Context) {
	var req models.RequirementCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "标题不能为空")
		return
	}
	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		response.BadRequest(c, "标题不能为空")
		return
	}
	if req.KanoCategory == "" {
		req.KanoCategory = "O"
	}
	if req.Priority == "" {
		req.Priority = "P2"
	}
	impacts, _ := json.Marshal(req.AarrrImpacts)
	userID := c.GetString("userID")
	uid, _ := strconv.Atoi(userID)

	reqNo, err := h.nextReqNo(h.DB)
	if err != nil {
		response.ServerError(c, "生成需求编号失败")
		return
	}

	var newID int
	err = h.DB.QueryRow(`INSERT INTO public.ai_pipeline_requirements
		(req_no, title, description, kano_category, aarrr_impacts, priority, status, source, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,'draft',$7,$8) RETURNING id`,
		reqNo, req.Title, req.Description, req.KanoCategory, string(impacts), req.Priority, req.Source, uid).Scan(&newID)
	if err != nil {
		response.ServerError(c, "创建需求失败")
		return
	}

	// 初始化 7 个环节
	for _, stage := range pipelineStageDefs {
		st := "todo"
		if stage == "A" {
			st = "doing"
		}
		h.DB.Exec(`INSERT INTO public.ai_pipeline_stages (req_id, stage, status, note) VALUES ($1,$2,$3,'')`, newID, stage, st)
	}

	response.SuccessWithMessage(c, "需求创建成功", gin.H{"id": newID, "reqNo": reqNo})
}

// UpdateRequirement 编辑需求
func (h *PipelineHandler) UpdateRequirement(c *gin.Context) {
	id := c.Param("id")
	var req models.RequirementUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "请求参数有误")
		return
	}
	var exists int
	if err := h.DB.QueryRow(`SELECT COUNT(*) FROM public.ai_pipeline_requirements WHERE id = $1`, id).Scan(&exists); err != nil || exists == 0 {
		response.NotFound(c, "需求不存在")
		return
	}
	impacts, _ := json.Marshal(req.AarrrImpacts)
	_, err := h.DB.Exec(`UPDATE public.ai_pipeline_requirements SET
		title = COALESCE(NULLIF($1,''), title),
		description = COALESCE(NULLIF($2,''), description),
		kano_category = COALESCE(NULLIF($3,''), kano_category),
		aarrr_impacts = CASE WHEN $4::text = '[]' THEN aarrr_impacts ELSE $4 END,
		priority = COALESCE(NULLIF($5,''), priority),
		source = COALESCE(NULLIF($6,''), source),
		updated_at = NOW() WHERE id = $7`,
		req.Title, req.Description, req.KanoCategory, string(impacts), req.Priority, req.Source, id)
	if err != nil {
		response.ServerError(c, "更新需求失败")
		return
	}
	response.SuccessWithMessage(c, "更新成功", nil)
}

// DeleteRequirement 删除需求（仅 draft/backlog 可删）
func (h *PipelineHandler) DeleteRequirement(c *gin.Context) {
	id := c.Param("id")
	var status string
	err := h.DB.QueryRow(`SELECT status FROM public.ai_pipeline_requirements WHERE id = $1`, id).Scan(&status)
	if err == sql.ErrNoRows {
		response.NotFound(c, "需求不存在")
		return
	}
	if err != nil {
		response.ServerError(c, "查询需求失败")
		return
	}
	if status != "draft" && status != "backlog" {
		response.BadRequest(c, "仅「草稿/已评估」状态的需求可以删除")
		return
	}
	if _, err := h.DB.Exec(`DELETE FROM public.ai_pipeline_requirements WHERE id = $1`, id); err != nil {
		response.ServerError(c, "删除需求失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// ChangeStatus 状态流转（按状态机校验 + 同步各环节状态）
func (h *PipelineHandler) ChangeStatus(c *gin.Context) {
	id := c.Param("id")
	var req models.StatusChangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "目标状态不能为空")
		return
	}
	var curStatus string
	err := h.DB.QueryRow(`SELECT status FROM public.ai_pipeline_requirements WHERE id = $1`, id).Scan(&curStatus)
	if err == sql.ErrNoRows {
		response.NotFound(c, "需求不存在")
		return
	}
	if err != nil {
		response.ServerError(c, "查询需求失败")
		return
	}
	allowed, ok := requirementStatusTransitions[curStatus]
	if !ok || !containsString(allowed, req.Status) {
		response.BadRequest(c, "非法状态流转："+curStatus+" → "+req.Status)
		return
	}
	if _, err := h.DB.Exec(`UPDATE public.ai_pipeline_requirements SET status = $1, updated_at = NOW() WHERE id = $2`, req.Status, id); err != nil {
		response.ServerError(c, "更新状态失败")
		return
	}
	// 同步各环节状态
	plan := pipelineStagePlan(req.Status)
	for stage, st := range plan {
		h.DB.Exec(`UPDATE public.ai_pipeline_stages SET status = $1 WHERE req_id = $2 AND stage = $3`, st, id, stage)
	}
	response.SuccessWithMessage(c, "状态已更新", nil)
}

// GetBoard 流水线看板数据（全部需求 + 环节状态 + 统计）
func (h *PipelineHandler) GetBoard(c *gin.Context) {
	reqs, err := queryRowsToMaps(h.DB, pipelineReqSelect+` ORDER BY id DESC LIMIT 500`)
	if err != nil {
		response.ServerError(c, "查询需求失败")
		return
	}
	for _, m := range reqs {
		parseAarrrImpacts(m, "aarrrImpacts")
		formatRowTimes(m, "createdAt", "updatedAt")
	}
	stages, err := queryRowsToMaps(h.DB, `SELECT req_id AS "reqId", stage, status FROM public.ai_pipeline_stages ORDER BY req_id, stage`)
	if err != nil {
		response.ServerError(c, "查询环节失败")
		return
	}

	stats := map[string]int{}
	for _, m := range reqs {
		if s, ok := m["status"].(string); ok {
			stats[s]++
		}
	}
	kanoStats := map[string]int{}
	for _, m := range reqs {
		if k, ok := m["kanoCategory"].(string); ok && k != "" {
			kanoStats[k]++
		}
	}

	response.Success(c, gin.H{
		"requirements": reqs,
		"stages":       stages,
		"stats":        stats,
		"kanoStats":    kanoStats,
	})
}

// GetStages 某需求的环节状态
func (h *PipelineHandler) GetStages(c *gin.Context) {
	id := c.Param("id")
	maps, err := queryRowsToMaps(h.DB, `SELECT id, req_id AS "reqId", stage, status,
		to_char(started_at, 'YYYY-MM-DD HH24:MI') AS "startedAt",
		to_char(finished_at, 'YYYY-MM-DD HH24:MI') AS "finishedAt", note
		FROM public.ai_pipeline_stages WHERE req_id = $1 ORDER BY stage`, id)
	if err != nil {
		response.ServerError(c, "查询环节失败")
		return
	}
	response.Success(c, maps)
}

// GetDeliverables 某需求的交付物列表
func (h *PipelineHandler) GetDeliverables(c *gin.Context) {
	id := c.Param("id")
	maps, err := queryRowsToMaps(h.DB, `SELECT id, req_id AS "reqId", stage, title, file_type AS "fileType",
		file_path AS "filePath", url, created_by AS "createdBy", created_at AS "createdAt"
		FROM public.ai_pipeline_deliverables WHERE req_id = $1 ORDER BY id DESC`, id)
	if err != nil {
		response.ServerError(c, "查询交付物失败")
		return
	}
	for _, m := range maps {
		formatRowTimes(m, "createdAt")
	}
	response.Success(c, maps)
}

// CreateDeliverable 登记交付物
func (h *PipelineHandler) CreateDeliverable(c *gin.Context) {
	id := c.Param("id")
	var req models.DeliverableCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "环节与标题不能为空")
		return
	}
	if !containsString(pipelineStageDefs, req.Stage) {
		response.BadRequest(c, "非法环节："+req.Stage)
		return
	}
	userName := c.GetString("username")
	_, err := h.DB.Exec(`INSERT INTO public.ai_pipeline_deliverables
		(req_id, stage, title, file_type, file_path, url, created_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		id, req.Stage, req.Title, req.FileType, req.FilePath, req.URL, userName)
	if err != nil {
		response.ServerError(c, "登记交付物失败")
		return
	}
	response.SuccessWithMessage(c, "交付物已登记", nil)
}

// DeleteDeliverable 删除交付物
func (h *PipelineHandler) DeleteDeliverable(c *gin.Context) {
	id := c.Param("id")
	if _, err := h.DB.Exec(`DELETE FROM public.ai_pipeline_deliverables WHERE id = $1`, id); err != nil {
		response.ServerError(c, "删除交付物失败")
		return
	}
	response.SuccessWithMessage(c, "删除成功", nil)
}

// GetReviews 审核记录
func (h *PipelineHandler) GetReviews(c *gin.Context) {
	id := c.Param("id")
	maps, err := queryRowsToMaps(h.DB, `SELECT id, req_id AS "reqId", action, target_stage AS "targetStage",
		reason, reviewed_by AS "reviewedBy", reviewed_name AS "reviewedName", created_at AS "createdAt"
		FROM public.ai_pipeline_reviews WHERE req_id = $1 ORDER BY id DESC`, id)
	if err != nil {
		response.ServerError(c, "查询审核记录失败")
		return
	}
	for _, m := range maps {
		formatRowTimes(m, "createdAt")
	}
	response.Success(c, maps)
}

// CreateReview 审核（通过 approve / 打回 reject）
func (h *PipelineHandler) CreateReview(c *gin.Context) {
	id := c.Param("id")
	var req models.ReviewCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "action 必须为 approve 或 reject")
		return
	}
	if req.Action != "approve" && req.Action != "reject" {
		response.BadRequest(c, "action 必须为 approve 或 reject")
		return
	}
	var curStatus string
	err := h.DB.QueryRow(`SELECT status FROM public.ai_pipeline_requirements WHERE id = $1`, id).Scan(&curStatus)
	if err == sql.ErrNoRows {
		response.NotFound(c, "需求不存在")
		return
	}
	if err != nil {
		response.ServerError(c, "查询需求失败")
		return
	}
	if curStatus == "done" || curStatus == "rejected" {
		response.BadRequest(c, "需求已终结，不能再审核")
		return
	}
	if req.Action == "approve" && curStatus != "review" {
		response.BadRequest(c, "仅「待审核」状态的需求可以审核通过")
		return
	}
	if req.Action == "reject" && !containsString(pipelineStageDefs, req.TargetStage) {
		response.BadRequest(c, "打回必须指定目标环节（A~G）")
		return
	}

	userID := c.GetString("userID")
	uid, _ := strconv.Atoi(userID)
	userName := c.GetString("username")

	// 写入审核记录
	if _, err := h.DB.Exec(`INSERT INTO public.ai_pipeline_reviews
		(req_id, action, target_stage, reason, reviewed_by, reviewed_name)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		id, req.Action, req.TargetStage, req.Reason, uid, userName); err != nil {
		response.ServerError(c, "提交审核失败")
		return
	}

	if req.Action == "approve" {
		if _, err := h.DB.Exec(`UPDATE public.ai_pipeline_requirements SET status = 'done', updated_at = NOW() WHERE id = $1`, id); err != nil {
			response.ServerError(c, "更新需求失败")
			return
		}
		h.DB.Exec(`UPDATE public.ai_pipeline_stages SET status = 'done' WHERE req_id = $1`, id)
		response.SuccessWithMessage(c, "审核通过，需求已完结", nil)
		return
	}

	// 打回：状态置 rework，目标环节 doing，其后 todo
	if _, err := h.DB.Exec(`UPDATE public.ai_pipeline_requirements SET status = 'rework', updated_at = NOW() WHERE id = $1`, id); err != nil {
		response.ServerError(c, "更新需求失败")
		return
	}
	idx := 0
	for i, s := range pipelineStageDefs {
		if s == req.TargetStage {
			idx = i
			break
		}
	}
	for i, stage := range pipelineStageDefs {
		st := "done"
		if i == idx {
			st = "doing"
		} else if i > idx {
			st = "todo"
		}
		h.DB.Exec(`UPDATE public.ai_pipeline_stages SET status = $1 WHERE req_id = $2 AND stage = $3`, st, id, stage)
	}
	response.SuccessWithMessage(c, "已打回至环节 "+req.TargetStage, nil)
}
