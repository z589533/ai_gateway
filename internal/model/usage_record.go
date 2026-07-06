package model

import "time"

const (
	UsageStatusSuccess = "success" // 代理成功
	UsageStatusError   = "error"   // 已识别 Key 但代理失败（502/504 等）
)

// UsageRecord 单次代理调用的用量快照。
type UsageRecord struct {
	ID               uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID         uint64    `gorm:"index;not null" json:"tenant_id"`
	APIKeyID         uint64    `gorm:"index;not null" json:"api_key_id"`
	Model            string    `gorm:"size:64;not null" json:"model"`
	PromptTokens     int       `gorm:"not null;default:0" json:"prompt_tokens"`
	CompletionTokens int       `gorm:"not null;default:0" json:"completion_tokens"`
	TotalTokens      int       `gorm:"not null;default:0" json:"total_tokens"`
	LatencyMs        int       `gorm:"not null;default:0" json:"latency_ms"`
	Status           string    `gorm:"size:16;not null" json:"status"`
	RequestedAt      time.Time `gorm:"index;not null" json:"requested_at"`
	CreatedAt        time.Time `json:"created_at"`
}

func (UsageRecord) TableName() string {
	return "usage_records"
}

// UsageSummary 查询结果的聚合统计。
type UsageSummary struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
	SuccessCount     int64 `json:"success_count"`
	ErrorCount       int64 `json:"error_count"`
}
