package models

// OperationManual 操作手册
type OperationManual struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	FileURL     string `json:"file_url"`
	FileSize    string `json:"file_size"`
	UpdatedAt   string `json:"updated_at"`
}

// OperationManualRequest 创建/更新请求
type OperationManualRequest struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	FileURL     string `json:"file_url"`
	FileSize    string `json:"file_size"`
}
