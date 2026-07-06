// 领域模型：租户、API Key、用量记录及 GORM 映射。
package model

import "time"

const (
	TenantStatusInactive int8 = 0 // 禁用：其下 Key 代理请求返回 403
	TenantStatusActive   int8 = 1 // 正常
)

// Tenant 多租户实体，一个租户可拥有多个 API Key。
type Tenant struct {
	ID        uint64    `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:128;uniqueIndex;not null" json:"name"`
	Status    int8      `gorm:"not null;default:1" json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Tenant) TableName() string {
	return "tenants"
}
