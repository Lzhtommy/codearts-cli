package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Lzhtommy/codearts-cli/internal/core"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) (*Client, *httptest.Server) {
	t.Helper()
	srv := httptest.NewServer(handler)
	cfg := &core.Config{
		AK:     "TESTAK",
		SK:     "TESTSK",
		Region: "cn-south-1",
	}
	cli, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}
	return cli, srv
}

func TestDo_Success(t *testing.T) {
	cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	})
	defer srv.Close()

	out := map[string]interface{}{}
	err := cli.Do(context.Background(), "GET", srv.URL, "/test", nil, nil, &out)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if out["status"] != "ok" {
		t.Errorf("Do() response = %v, want status=ok", out)
	}
}

func TestDo_APIError_WithErrorCode(t *testing.T) {
	cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]string{
			"error_code": "PM.02177003",
			"error_msg":  "非目标项目成员",
		})
	})
	defer srv.Close()

	err := cli.Do(context.Background(), "POST", srv.URL, "/test", nil, map[string]string{}, nil)
	if err == nil {
		t.Fatal("Do() should return error on 400")
	}
	apiErr, ok := err.(*APIError)
	if !ok {
		t.Fatalf("error should be *APIError, got %T", err)
	}
	if apiErr.StatusCode != 400 {
		t.Errorf("StatusCode = %d, want 400", apiErr.StatusCode)
	}
	if apiErr.ErrorCode != "PM.02177003" {
		t.Errorf("ErrorCode = %q, want PM.02177003", apiErr.ErrorCode)
	}
}

func TestDo_APIError_401_Hint(t *testing.T) {
	cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		w.Write([]byte(`{"error_code":"APIGW.0301","error_msg":"verify ak sk signature failed"}`))
	})
	defer srv.Close()

	err := cli.Do(context.Background(), "GET", srv.URL, "/test", nil, nil, nil)
	if err == nil {
		t.Fatal("Do() should return error on 401")
	}
	msg := err.Error()
	if !contains(msg, "APIGW.0301") {
		t.Errorf("error should contain error_code, got: %s", msg)
	}
}

func TestDo_APIError_RawBody(t *testing.T) {
	cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		w.Write([]byte("Internal Server Error"))
	})
	defer srv.Close()

	err := cli.Do(context.Background(), "GET", srv.URL, "/test", nil, nil, nil)
	if err == nil {
		t.Fatal("Do() should return error on 500")
	}
	msg := err.Error()
	if !contains(msg, "Internal Server Error") {
		t.Errorf("error should contain raw body, got: %s", msg)
	}
}

func TestDo_EmptyResponse(t *testing.T) {
	cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})
	defer srv.Close()

	out := map[string]interface{}{}
	err := cli.Do(context.Background(), "POST", srv.URL, "/test", nil, map[string]string{}, &out)
	if err != nil {
		t.Fatalf("Do() should not error on empty 200 response, got: %v", err)
	}
	if len(out) != 0 {
		t.Errorf("out should be empty, got %v", out)
	}
}

func TestDo_PostBody(t *testing.T) {
	var receivedContentType string
	var receivedBody []byte
	cli, srv := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		receivedContentType = r.Header.Get("Content-Type")
		buf := make([]byte, 4096)
		n, _ := r.Body.Read(buf)
		receivedBody = buf[:n]
		w.WriteHeader(200)
		w.Write([]byte("{}"))
	})
	defer srv.Close()

	body := map[string]string{"key": "value"}
	err := cli.Do(context.Background(), "POST", srv.URL, "/test", nil, body, nil)
	if err != nil {
		t.Fatalf("Do() error: %v", err)
	}
	if receivedContentType != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", receivedContentType)
	}
	if len(receivedBody) == 0 {
		t.Error("request body should not be empty")
	}
}

func TestNew_MissingCredentials(t *testing.T) {
	_, err := New(&core.Config{Region: "cn-south-1"})
	if err == nil {
		t.Error("New() should fail with empty AK/SK")
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(s) > 0 && containsStr(s, sub))
}

func containsStr(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
