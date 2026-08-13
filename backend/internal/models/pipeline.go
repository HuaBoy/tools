package models

// ===== AI 研发流水线（REQ-2026-001）请求模型 =====

// RequirementCreateRequest 新建需求请求
type RequirementCreateRequest struct {
	Title        string   `json:"title" binding:"required"`
	Description  string   `json:"description"`
	KanoCategory string   `json:"kano_category"` // M/O/A/I/R
	AarrrImpacts []string `json:"aarrr_impacts"` // 获取/激活/留存/收入/传播
	Priority     string   `json:"priority"`      // P0/P1/P2
	Source       string   `json:"source"`
}

// RequirementUpdateRequest 编辑需求请求
type RequirementUpdateRequest struct {
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	KanoCategory string   `json:"kano_category"`
	AarrrImpacts []string `json:"aarrr_impacts"`
	Priority     string   `json:"priority"`
	Source       string   `json:"source"`
}

// StatusChangeRequest 状态流转请求
type StatusChangeRequest struct {
	Status string `json:"status" binding:"required"`
}

// DeliverableCreateRequest 交付物登记请求
type DeliverableCreateRequest struct {
	Stage    string `json:"stage" binding:"required"` // A~G
	Title    string `json:"title" binding:"required"`
	FileType string `json:"file_type"` // markdown/html/image/other
	FilePath string `json:"file_path"`
	URL      string `json:"url"`
}

// ReviewCreateRequest 审核请求
type ReviewCreateRequest struct {
	Action      string `json:"action" binding:"required"` // approve / reject
	TargetStage string `json:"target_stage"`              // 打回时必填：A~G
	Reason      string `json:"reason"`
}
