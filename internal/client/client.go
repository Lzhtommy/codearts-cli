package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/Lzhtommy/codearts-cli/internal/core"
)

// Client is a thin HTTP wrapper for Huawei Cloud CodeArts APIs with AK/SK
// request signing and endpoint resolution.
type Client struct {
	cfg    *core.Config
	signer *Signer
	http   *http.Client
}

// New builds a Client from a validated config.
func New(cfg *core.Config) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	return &Client{
		cfg:    cfg,
		signer: &Signer{AK: cfg.AK, SK: cfg.SK},
		http:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

// APIError is returned when the server responds with a non-2xx status.
type APIError struct {
	Status     string
	StatusCode int
	Body       []byte
	// Parsed fields (best-effort; may be empty).
	ErrorCode string `json:"error_code,omitempty"`
	ErrorMsg  string `json:"error_msg,omitempty"`
}

func (e *APIError) Error() string {
	if e.ErrorCode != "" || e.ErrorMsg != "" {
		return fmt.Sprintf("codearts api error [%d %s]: %s %s",
			e.StatusCode, e.Status, e.ErrorCode, e.ErrorMsg)
	}
	return fmt.Sprintf("codearts api error [%d %s]: %s",
		e.StatusCode, e.Status, string(e.Body))
}

// PipelineEndpoint returns the cloudpipeline host for the configured region.
// Precedence: $CODEARTS_PIPELINE_ENDPOINT > regional default.
//
// The regional default follows Huawei Cloud's subdomain convention:
// https://cloudpipeline-ext.<region>.myhuaweicloud.com
func (c *Client) PipelineEndpoint() string {
	if v := os.Getenv("CODEARTS_PIPELINE_ENDPOINT"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fmt.Sprintf("https://cloudpipeline-ext.%s.myhuaweicloud.com", c.cfg.Region)
}

// ProjectManEndpoint returns the host for CodeArts ProjectMan / 工作项管理.
// Override via $CODEARTS_PROJECTMAN_ENDPOINT.
func (c *Client) ProjectManEndpoint() string {
	if v := os.Getenv("CODEARTS_PROJECTMAN_ENDPOINT"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fmt.Sprintf("https://projectman-ext.%s.myhuaweicloud.com", c.cfg.Region)
}

// RepoEndpoint returns the host for CodeArts Repo / 代码托管.
// Override via $CODEARTS_REPO_ENDPOINT.
func (c *Client) RepoEndpoint() string {
	if v := os.Getenv("CODEARTS_REPO_ENDPOINT"); v != "" {
		return strings.TrimRight(v, "/")
	}
	return fmt.Sprintf("https://codehub-ext.%s.myhuaweicloud.com", c.cfg.Region)
}

// Do builds, signs, and sends a request to the given endpoint+path and
// decodes the JSON response into out. bodyJSON may be nil.
func (c *Client) Do(ctx context.Context, method, endpoint, path string, query url.Values, bodyJSON interface{}, out interface{}) error {
	var bodyBytes []byte
	if bodyJSON != nil {
		b, err := json.Marshal(bodyJSON)
		if err != nil {
			return fmt.Errorf("marshal request body: %w", err)
		}
		bodyBytes = b
	}

	full := strings.TrimRight(endpoint, "/") + path
	if len(query) > 0 {
		full += "?" + query.Encode()
	}

	var reqBody io.Reader
	if len(bodyBytes) > 0 {
		reqBody = bytes.NewReader(bodyBytes)
	}
	req, err := http.NewRequestWithContext(ctx, method, full, reqBody)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if len(bodyBytes) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "codearts-cli/0.1")

	if err := c.signer.Sign(req, bodyBytes); err != nil {
		return fmt.Errorf("sign request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &APIError{
			Status:     resp.Status,
			StatusCode: resp.StatusCode,
			Body:       respBody,
		}
		// Best-effort parse of the Huawei error envelope.
		_ = json.Unmarshal(respBody, apiErr)
		return apiErr
	}

	if out != nil && len(respBody) > 0 {
		if err := json.Unmarshal(respBody, out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
	}
	return nil
}
