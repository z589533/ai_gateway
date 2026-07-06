// 统一 HTTP 响应封装：管理面 Envelope 与 OpenAI 兼容 error 格式。
package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Envelope 管理面 API 统一响应结构 { code, message, data }。
type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// OpenAIErrorBody OpenAI 兼容错误响应 { error: { message, type, param, code } }。
type OpenAIErrorBody struct {
	Error OpenAIError `json:"error"`
}

// OpenAIError 对齐 OpenAI 错误体：message/type/param/code。
// param 默认 null，表示错误不归属某个具体请求参数（如鉴权、限流、上游错误）。
type OpenAIError struct {
	Message string      `json:"message"`
	Type    string      `json:"type"`
	Param   interface{} `json:"param"`
	Code    string      `json:"code"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Message: "ok", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Code: 0, Message: "ok", Data: data})
}

// Error 管理面错误响应，HTTP 状态码与 body.code 一致。
func Error(c *gin.Context, status int, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	c.JSON(status, Envelope{Code: status, Message: message})
}

// OpenAIErrorJSON 数据面错误响应，供 /v1/* 代理接口使用。
// type 按 OpenAI 风格随 HTTP 状态码分类：429→rate_limit_error，5xx→server_error，其余 4xx→invalid_request_error。
func OpenAIErrorJSON(c *gin.Context, status int, code string, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	c.JSON(status, OpenAIErrorBody{
		Error: OpenAIError{
			Message: message,
			Type:    OpenAIErrorType(status),
			Param:   nil,
			Code:    code,
		},
	})
}

// OpenAIErrorType 按 HTTP 状态码映射 OpenAI 风格的 error.type。
func OpenAIErrorType(status int) string {
	switch {
	case status == http.StatusTooManyRequests:
		return "rate_limit_error"
	case status >= 500:
		return "server_error"
	default:
		return "invalid_request_error"
	}
}
