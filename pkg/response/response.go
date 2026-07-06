package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type OpenAIErrorBody struct {
	Error OpenAIError `json:"error"`
}

type OpenAIError struct {
	Message string `json:"message"`
	Type    string `json:"type"`
	Code    string `json:"code"`
}

func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Envelope{Code: 0, Message: "ok", Data: data})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Envelope{Code: 0, Message: "ok", Data: data})
}

func Error(c *gin.Context, status int, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	c.JSON(status, Envelope{Code: status, Message: message})
}

func OpenAIErrorJSON(c *gin.Context, status int, code string, message string) {
	if message == "" {
		message = http.StatusText(status)
	}
	c.JSON(status, OpenAIErrorBody{
		Error: OpenAIError{
			Message: message,
			Type:    "invalid_request_error",
			Code:    code,
		},
	})
}
