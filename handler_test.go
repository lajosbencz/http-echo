package httpecho_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	httpecho "github.com/lajosbencz/http-echo"
	"github.com/rs/zerolog"
)

// mockResponseWriter is a simple implementation of http.ResponseWriter for testing purposes.
type mockResponseWriter struct {
	statusCode int
	header     http.Header
	body       []byte
}

func (m *mockResponseWriter) Header() http.Header {
	if m.header == nil {
		m.header = make(http.Header)
	}
	return m.header
}

func (m *mockResponseWriter) Write(data []byte) (int, error) {
	m.body = append(m.body, data...)
	return len(data), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestHandler(t *testing.T) {
	method := http.MethodPost
	uri := "/test"
	query := "foo=bar"
	req, err := http.NewRequest(method, uri+"?"+query, bytes.NewBufferString("{\"foo\":\"bar\"}"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+jwtToken)

	handler := httpecho.NewHttpEchoHandler(zerolog.Logger{}, "Authorization")

	w := &mockResponseWriter{
		statusCode: 200,
	}
	handler.ServeHTTP(w, req)

	if w.statusCode != 200 {
		t.Error("status code")
	}

	var response httpecho.HttpEchoResponse
	if err := json.Unmarshal(w.body, &response); err != nil {
		t.Error(err)
		t.FailNow()
	}

	if response.Method != method {
		t.Error("method")
	}

	if response.Uri != uri {
		t.Error("uri")
	}

	if response.Query == nil || len(response.Query["foo"]) != 1 || response.Query["foo"][0] != "bar" {
		t.Error("query")
	}

	if response.Json == nil {
		t.Error("json")
		t.FailNow()
	}

	responseJson := *response.Json
	responseMap := responseJson.(map[string]interface{})
	if v, ok := responseMap["foo"].(string); !ok || v != "bar" {
		t.Error("json value")
	}

}
