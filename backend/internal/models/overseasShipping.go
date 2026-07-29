package models

// OverseasShipping 海外发货记录
type OverseasShipping struct {
	ID                       int64  `json:"id"`
	ProductType              string `json:"product_type"`
	SalesEngineer            string `json:"sales_engineer"`
	Country                  string `json:"country"`
	CustomerName             string `json:"customer_name"`
	ControllerSn             string `json:"controller_sn"`
	HandheldSn               string `json:"handheld_sn"`
	ControllerHwVersion      string `json:"controller_hw_version"`
	ControllerUpgradeHistory string `json:"controller_upgrade_history"`
	HandheldUpgradeHistory   string `json:"handheld_upgrade_history"`
	LastUpgradeDate          string `json:"last_upgrade_date"`
	Assistant                string `json:"assistant"`
	Remark                   string `json:"remark"`
	CreatedAt                string `json:"created_at"`
	UpdatedAt                string `json:"updated_at"`
}

// OverseasShippingCreateRequest 创建/更新请求（前端始终提交完整记录）
type OverseasShippingCreateRequest struct {
	ProductType              string `json:"product_type"`
	SalesEngineer            string `json:"sales_engineer"`
	Country                  string `json:"country"`
	CustomerName             string `json:"customer_name"`
	ControllerSn             string `json:"controller_sn"`
	HandheldSn               string `json:"handheld_sn"`
	ControllerHwVersion      string `json:"controller_hw_version"`
	ControllerUpgradeHistory string `json:"controller_upgrade_history"`
	HandheldUpgradeHistory   string `json:"handheld_upgrade_history"`
	LastUpgradeDate          string `json:"last_upgrade_date"`
	Assistant                string `json:"assistant"`
	Remark                   string `json:"remark"`
}
