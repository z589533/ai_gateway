package proxy

import (
	"context"
	"strings"
	"time"

	"github.com/google/uuid"
)

// MockProxy 本地 mock 上游，不发起真实外网请求，用于 MVP 演示与自测。
type MockProxy struct {
	Latency time.Duration
	Fail    bool
	Now     func() time.Time
	NewID   func() string
}

func NewMockProxy(latency time.Duration, fail bool) *MockProxy {
	return &MockProxy{
		Latency: latency,
		Fail:    fail,
		Now:     func() time.Time { return time.Now().UTC() },
		NewID:   func() string { return "chatcmpl-mock-" + uuid.NewString() },
	}
}

// Chat 模拟 OpenAI chat/completions：校验参数、可选延迟/失败，返回固定内容与估算 token。
func (p *MockProxy) Chat(ctx context.Context, req ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if strings.TrimSpace(req.Model) == "" {
		return nil, InvalidRequest("invalid_request", "model is required")
	}
	if len(req.Messages) == 0 {
		return nil, InvalidRequest("invalid_request", "messages must not be empty")
	}
	if req.Stream {
		return nil, InvalidRequest("stream_not_supported", "stream=true is not supported by this MVP")
	}

	// 模拟上游延迟，支持通过配置测试 504
	if p.Latency > 0 {
		timer := time.NewTimer(p.Latency)
		defer timer.Stop()
		select {
		case <-ctx.Done():
			return nil, GatewayTimeout("upstream request timed out")
		case <-timer.C:
		}
	}
	if p.Fail {
		return nil, BadGateway("mock upstream failed")
	}

	promptTokens := estimatePromptTokens(req.Messages)
	completionTokens := 12
	return &ChatCompletionResponse{
		ID:      p.NewID(),
		Object:  "chat.completion",
		Created: p.Now().Unix(),
		Model:   req.Model,
		Choices: []Choice{{
			Index:        0,
			Message:      Message{Role: "assistant", Content: "This is a mock response from AI Gateway."},
			FinishReason: "stop",
		}},
		Usage: Usage{
			PromptTokens:     promptTokens,
			CompletionTokens: completionTokens,
			TotalTokens:      promptTokens + completionTokens,
		},
	}, nil
}

// Models 返回 mock 支持的模型列表，MVP 统一暴露 DefaultModel（gpt5.5）。
func (p *MockProxy) Models() ModelListResponse {
	return ModelListResponse{
		Object: "list",
		Data: []Model{{
			ID:      DefaultModel,
			Object:  "model",
			Created: 1710000000,
			OwnedBy: "ai-gateway-mock",
		}},
	}
}

// estimatePromptTokens 按字符数/4 粗算 prompt token，非 tiktoken 精确值。
func estimatePromptTokens(messages []Message) int {
	chars := 0
	for _, msg := range messages {
		chars += len([]rune(msg.Role)) + len([]rune(msg.Content))
	}
	if chars == 0 {
		return 0
	}
	tokens := chars / 4
	if chars%4 != 0 {
		tokens++
	}
	if tokens == 0 {
		return 1
	}
	return tokens
}
