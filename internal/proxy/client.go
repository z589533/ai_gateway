package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

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
