// HTTP 上游代理客户端：解码真实上游响应（MVP 主要使用 MockProxy）。
package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

// HTTPClientProxy 通过 HTTP 调用外部 LLM 上游并解码 JSON 响应。
type HTTPClientProxy struct {
	Client  *http.Client
	BaseURL string
}

func NewHTTPClientProxy(baseURL string, timeout time.Duration) *HTTPClientProxy {
	return &HTTPClientProxy{
		Client:  &http.Client{Timeout: timeout},
		BaseURL: baseURL,
	}
}

// DecodeResponse 解析上游 HTTP 响应：5xx → 502，超时/解码失败 → 502/504。
func (p *HTTPClientProxy) DecodeResponse(ctx context.Context, resp *http.Response) (*ChatCompletionResponse, error) {
	if resp.StatusCode >= http.StatusInternalServerError {
		return nil, BadGateway("upstream returned 5xx")
	}
	defer resp.Body.Close()
	var decoded ChatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		if ctx.Err() != nil {
			return nil, GatewayTimeout("upstream request timed out")
		}
		return nil, BadGateway("failed to decode upstream response")
	}
	return &decoded, nil
}
