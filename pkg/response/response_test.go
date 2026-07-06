package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOKEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	OK(c, gin.H{"id": 1})

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d", w.Code)
	}
	var body Envelope
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Code != 0 || body.Message != "ok" {
		t.Fatalf("unexpected envelope: %+v", body)
	}
}

func TestOpenAIErrorJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	OpenAIErrorJSON(c, http.StatusUnauthorized, "invalid_api_key", "bad key")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d", w.Code)
	}
	var body OpenAIErrorBody
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "invalid_api_key" {
		t.Fatalf("code = %q", body.Error.Code)
	}
	if body.Error.Type != "invalid_request_error" {
		t.Fatalf("type = %q, want invalid_request_error", body.Error.Type)
	}
	// OpenAI 风格始终包含 param 字段（无归属参数时为 null），校验它确实出现在 JSON 中。
	var raw map[string]map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatal(err)
	}
	param, present := raw["error"]["param"]
	if !present {
		t.Fatalf("param field missing in error response")
	}
	if string(param) != "null" {
		t.Fatalf("param = %s, want null", string(param))
	}
}

func TestOpenAIErrorType(t *testing.T) {
	cases := []struct {
		status int
		want   string
	}{
		{http.StatusBadRequest, "invalid_request_error"},
		{http.StatusUnauthorized, "invalid_request_error"},
		{http.StatusForbidden, "invalid_request_error"},
		{http.StatusNotFound, "invalid_request_error"},
		{http.StatusTooManyRequests, "rate_limit_error"},
		{http.StatusInternalServerError, "server_error"},
		{http.StatusBadGateway, "server_error"},
		{http.StatusGatewayTimeout, "server_error"},
	}
	for _, tc := range cases {
		if got := OpenAIErrorType(tc.status); got != tc.want {
			t.Fatalf("OpenAIErrorType(%d) = %q, want %q", tc.status, got, tc.want)
		}
	}
}

func TestOpenAIErrorJSONTypeByStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	checks := []struct {
		status int
		want   string
	}{
		{http.StatusTooManyRequests, "rate_limit_error"},
		{http.StatusBadGateway, "server_error"},
		{http.StatusGatewayTimeout, "server_error"},
	}
	for _, tc := range checks {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		OpenAIErrorJSON(c, tc.status, "code", "msg")
		var body OpenAIErrorBody
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatal(err)
		}
		if body.Error.Type != tc.want {
			t.Fatalf("status %d type = %q, want %q", tc.status, body.Error.Type, tc.want)
		}
	}
}
