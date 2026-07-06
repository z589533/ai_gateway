package proxy

import (
	"context"
	"testing"
	"time"
)

func TestMockProxyModels(t *testing.T) {
	p := NewMockProxy(0, false)
	models := p.Models()
	if len(models.Data) != 1 || models.Data[0].ID != DefaultModel {
		t.Fatalf("models = %+v, want %s", models.Data, DefaultModel)
	}
}

func TestMockProxyChatSuccess(t *testing.T) {
	p := NewMockProxy(0, false)
	p.Now = func() time.Time { return time.Unix(100, 0) }
	p.NewID = func() string { return "chatcmpl-mock-test" }

	resp, err := p.Chat(context.Background(), ChatCompletionRequest{
		Model:    "gpt-test",
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.ID != "chatcmpl-mock-test" || resp.Model != "gpt-test" {
		t.Fatalf("unexpected response: %+v", resp)
	}
	if resp.Usage.TotalTokens == 0 {
		t.Fatal("expected token usage")
	}
}

func TestMockProxyRejectsStream(t *testing.T) {
	p := NewMockProxy(0, false)
	_, err := p.Chat(context.Background(), ChatCompletionRequest{
		Model:    "gpt-test",
		Messages: []Message{{Role: "user", Content: "hello"}},
		Stream:   true,
	})
	proxyErr, ok := err.(*Error)
	if !ok || proxyErr.Code != "stream_not_supported" {
		t.Fatalf("err = %#v", err)
	}
}

func TestMockProxyTimeout(t *testing.T) {
	p := NewMockProxy(50*time.Millisecond, false)
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	_, err := p.Chat(ctx, ChatCompletionRequest{
		Model:    "gpt-test",
		Messages: []Message{{Role: "user", Content: "hello"}},
	})
	proxyErr, ok := err.(*Error)
	if !ok || proxyErr.Code != "gateway_timeout" {
		t.Fatalf("err = %#v", err)
	}
}
