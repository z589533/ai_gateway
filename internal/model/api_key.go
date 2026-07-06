package model

import "time"

const (
	APIKeyStatusDisabled int8 = 0
	APIKeyStatusEnabled  int8 = 1
	APIKeyStatusDeleted  int8 = 2
)

type APIKey struct {
	ID        uint64     `gorm:"primaryKey;autoIncrement" json:"id"`
	TenantID  uint64     `gorm:"index;not null" json:"tenant_id"`
	Tenant    Tenant     `gorm:"foreignKey:TenantID" json:"-"`
	Name      string     `gorm:"size:128;not null" json:"name"`
	KeyPrefix string     `gorm:"size:16;not null" json:"key_prefix"`
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
