package middleware

import (
	"fmt"
	"net/http"

	sentinel "github.com/alibaba/sentinel-golang/api"
	"github.com/alibaba/sentinel-golang/core/base"
	"github.com/alibaba/sentinel-golang/core/flow"
	"github.com/gin-gonic/gin"
	"github.com/z589533/ai_gateway/pkg/response"
)

// RateLimitConfig 定义全站 / 租户 / 单 Key 三级 QPS 阈值。
type RateLimitConfig struct {
	GlobalQPS float64
	KeyQPS    float64
	TenantQPS float64
}

// InitSentinel 启动时加载全局限流规则。
func InitSentinel(cfg RateLimitConfig) error {
	if err := sentinel.InitDefault(); err != nil {
		return err
	}
	_, err := flow.LoadRules([]*flow.Rule{{
		Resource:               "proxy:global",
		TokenCalculateStrategy: flow.Direct,
		ControlBehavior:        flow.Reject,
		Threshold:              cfg.GlobalQPS,
		StatIntervalInMs:       1000,
	}})
	return err
}

// SentinelRateLimit 代理请求限流：依次检查 global → tenant → key，超限返回 429。
func SentinelRateLimit(cfg RateLimitConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth, ok := AuthResultFromContext(c)
		if !ok {
			c.Next()
			return
		}
		LoadDynamicRateLimitRules(cfg, auth.TenantID, auth.APIKeyID)
		resources := []string{
			"proxy:global",
			fmt.Sprintf("proxy:tenant:%d", auth.TenantID),
			fmt.Sprintf("proxy:key:%d", auth.APIKeyID),
		}
		entries := make([]*base.SentinelEntry, 0, len(resources))
		for _, resource := range resources {
			entry, blockErr := sentinel.Entry(resource, sentinel.WithTrafficType(base.Inbound))
			if blockErr != nil {
				for _, e := range entries {
					e.Exit()
				}
				response.OpenAIErrorJSON(c, http.StatusTooManyRequests, "rate_limit_exceeded", "rate limit exceeded")
				c.Abort()
				return
			}
			entries = append(entries, entry)
		}
		defer func() {
			for _, e := range entries {
				e.Exit()
			}
		}()
		c.Next()
	}
}

// LoadDynamicRateLimitRules 按当前租户与 Key 动态加载限流规则。
func LoadDynamicRateLimitRules(cfg RateLimitConfig, tenantID, keyID uint64) {
	_, _ = flow.LoadRules([]*flow.Rule{
		{
			Resource:               "proxy:global",
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              cfg.GlobalQPS,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               fmt.Sprintf("proxy:tenant:%d", tenantID),
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              cfg.TenantQPS,
			StatIntervalInMs:       1000,
		},
		{
			Resource:               fmt.Sprintf("proxy:key:%d", keyID),
			TokenCalculateStrategy: flow.Direct,
			ControlBehavior:        flow.Reject,
			Threshold:              cfg.KeyQPS,
			StatIntervalInMs:       1000,
		},
	})
}
