// 用量记录 GORM 仓储：写入与条件查询 + SQL 聚合。
package repository

import (
	"context"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	"gorm.io/gorm"
)

// UsageQuery 用量查询过滤条件。
type UsageQuery struct {
	TenantID uint64
	APIKeyID uint64
	From     *time.Time
	To       *time.Time
	Page     int
	PageSize int
}

type UsageRepository struct {
	db *gorm.DB
}

func NewUsageRepository(db *gorm.DB) *UsageRepository {
	return &UsageRepository{db: db}
}

func (r *UsageRepository) Create(ctx context.Context, usage *model.UsageRecord) error {
	return r.db.WithContext(ctx).Create(usage).Error
}

// Query 返回分页明细、总数与 summary（token 求和 + success/error 计数）。
func (r *UsageRepository) Query(ctx context.Context, q UsageQuery) ([]model.UsageRecord, int64, model.UsageSummary, error) {
	query := r.db.WithContext(ctx).Model(&model.UsageRecord{})
	query = applyUsageQuery(query, q)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, model.UsageSummary{}, err
	}

	var summary model.UsageSummary
	var rows []struct {
		PromptTokens     int64
		CompletionTokens int64
		TotalTokens      int64
		SuccessCount     int64
		ErrorCount       int64
	}
	if err := query.Select(`
		COALESCE(SUM(prompt_tokens), 0) AS prompt_tokens,
		COALESCE(SUM(completion_tokens), 0) AS completion_tokens,
		COALESCE(SUM(total_tokens), 0) AS total_tokens,
		COALESCE(SUM(CASE WHEN status = 'success' THEN 1 ELSE 0 END), 0) AS success_count,
		COALESCE(SUM(CASE WHEN status = 'error' THEN 1 ELSE 0 END), 0) AS error_count`).Scan(&rows).Error; err != nil {
		return nil, 0, model.UsageSummary{}, err
	}
	if len(rows) > 0 {
		summary = model.UsageSummary(rows[0])
	}

	var items []model.UsageRecord
	err := applyUsageQuery(r.db.WithContext(ctx).Model(&model.UsageRecord{}), q).
		Order("requested_at desc, id desc").
		Limit(q.PageSize).
		Offset((q.Page - 1) * q.PageSize).
		Find(&items).Error
	return items, total, summary, err
}

func applyUsageQuery(db *gorm.DB, q UsageQuery) *gorm.DB {
	if q.TenantID > 0 {
		db = db.Where("tenant_id = ?", q.TenantID)
	}
	if q.APIKeyID > 0 {
		db = db.Where("api_key_id = ?", q.APIKeyID)
	}
	if q.From != nil {
		db = db.Where("requested_at >= ?", *q.From)
	}
	if q.To != nil {
		db = db.Where("requested_at <= ?", *q.To)
	}
	return db
}
