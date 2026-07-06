package model

import "time"

const (
	UsageStatusSuccess = "success"
	UsageStatusError   = "error"
)

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

type UsageSummary struct {
	PromptTokens     int64 `json:"prompt_tokens"`
	CompletionTokens int64 `json:"completion_tokens"`
	TotalTokens      int64 `json:"total_tokens"`
	SuccessCount     int64 `json:"success_count"`
	ErrorCount       int64 `json:"error_count"`
}
