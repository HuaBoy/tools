package models

// OperationVideo 操作视频
type OperationVideo struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Duration    string `json:"duration"`
	Views       int    `json:"views"`
	Date        string `json:"date"`
}

// OperationVideoRequest 创建/更新请求
type OperationVideoRequest struct {
	Title       string `json:"title"`
	Category    string `json:"category"`
	Description string `json:"description"`
	URL         string `json:"url"`
	Duration    string `json:"duration"`
}
