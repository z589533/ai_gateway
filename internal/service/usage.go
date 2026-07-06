package service

import (
	"context"
	"time"

	"github.com/z589533/ai_gateway/internal/model"
	"github.com/z589533/ai_gateway/internal/repository"
)

type UsageService struct {
	repo UsageRepo
	now  Clock
}

type UsageList struct {
	Items    []model.UsageRecord `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
	Summary  model.UsageSummary  `json:"summary"`
}

type RecordUsageInput struct {
	TenantID         uint64
	APIKeyID         uint64
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	LatencyMs        int
	Status           string
	RequestedAt      time.Time
}

func NewUsageService(repo UsageRepo) *UsageService {
	return &UsageService{repo: repo, now: realClock}
}

func (s *UsageService) Record(ctx context.Context, input RecordUsageInput) error {
	// 代理请求完成后写入用量记录（成功/失败均记录）
	if input.RequestedAt.IsZero() {
		input.RequestedAt = s.now()
	}
	status := input.Status
	if status == "" {
		status = model.UsageStatusSuccess
	}
	return s.repo.Create(ctx, &model.UsageRecord{
		TenantID:         input.TenantID,
		APIKeyID:         input.APIKeyID,
		Model:            input.Model,
		PromptTokens:     input.PromptTokens,
		CompletionTokens: input.CompletionTokens,
		TotalTokens:      input.TotalTokens,
		LatencyMs:        input.LatencyMs,
		Status:           status,
		RequestedAt:      input.RequestedAt,
	})
}

func (s *UsageService) Query(ctx context.Context, q repository.UsageQuery) (*UsageList, error) {
	q.Page, q.PageSize = normalizePage(q.Page, q.PageSize)
	items, total, summary, err := s.repo.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	return &UsageList{Items: items, Total: total, Page: q.Page, PageSize: q.PageSize, Summary: summary}, nil
}
