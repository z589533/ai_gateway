package model

import "time"

const (
	APIKeyStatusDisabled int8 = 0 // 禁用，鉴权返回 403 key_disabled
	APIKeyStatusEnabled  int8 = 1 // 正常
	APIKeyStatusDeleted  int8 = 2 // 软删除，鉴权视为 401
)

// APIKey 租户 API Key；库内仅存 hash，明文 secret 不持久化。
type APIKey struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64     `gorm:"index;not null" json:"tenant_id"`
	Tenant    Tenant     `gorm:"foreignKey:TenantID" json:"-"`
	Name      string     `gorm:"size:128;not null" json:"name"`
	KeyPrefix string     `gorm:"size:16;not null" json:"key_prefix"` // 展示用前缀，非 secret
	KeyHash   string     `gorm:"size:64;uniqueIndex;not null" json:"-"`
	Scopes    StringList `gorm:"type:json;not null" json:"scopes"`
	Status    int8       `gorm:"not null;default:1" json:"status"`
	ExpiresAt *time.Time `json:"expires_at"`
	DeletedAt *time.Time `gorm:"index" json:"deleted_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (APIKey) TableName() string {
	return "api_keys"
}
