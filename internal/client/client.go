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

// Upstream hostnames the CLI signs requests against. These are stable
// identities the gateway matches on (Host header) to route to the correct
// upstream. Requests physically go to cfg.Gateway, but the Host header —
// and therefore the AK/SK signature — must reference these names so Huawei
// Cloud's IAM validator accepts the request after the gateway forwards.
const (
	hostPipeline   = "cloudpipeline-ext.cn-south-1.myhuaweicloud.com"
	hostProjectMan = "projectman-ext.cn-south-1.myhuaweicloud.com"
	hostRepo       = "codehub-ext.cn-south-1.myhuaweicloud.com"
)

// Client is a thin HTTP wrapper for Huawei Cloud CodeArts APIs with AK/SK
// request signing. All services (pipeline, projectman, repo) are reachable
// through a single gateway configured in ~/.codearts-cli/config.json.
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
	hint := ""
	if e.StatusCode == 401 {
		hint = "\nhint: check AK/SK with `codearts-cli config show`, or re-run `codearts-cli config init`"
	} else if e.StatusCode == 403 {
		hint = "\nhint: check IAM permissions for this AK/SK"
	}
	if e.ErrorCode != "" || e.ErrorMsg != "" {
		return fmt.Sprintf("codearts api error [%d]: %s — %s%s",
			e.StatusCode, e.ErrorCode, e.ErrorMsg, hint)
	}
	body := string(e.Body)
	if len(body) > 500 {
		body = body[:500] + "... (truncated)"
	}
	return fmt.Sprintf("codearts api error [%d]: %s%s", e.StatusCode, body, hint)
}

// PipelineEndpoint returns the signing-time base URL for CodeArts Pipeline.
// The scheme+host are fixed because the signature pins them; the actual TCP
// connection goes to cfg.Gateway — see Do for the dial rewrite.
func (c *Client) PipelineEndpoint() string {
	return "https://" + hostPipeline
}

// ProjectManEndpoint returns the signing-time base URL for CodeArts ProjectMan.
func (c *Client) ProjectManEndpoint() string {
	return "https://" + hostProjectMan
}

// RepoEndpoint returns the signing-time base URL for CodeArts Repo.
func (c *Client) RepoEndpoint() string {
	return "https://" + hostRepo
}

// Do builds, signs, and sends a request to the given endpoint+path and
// decodes the JSON response into out. bodyJSON may be nil.
//
// endpoint is the *signing* URL (a Huawei Cloud hostname, e.g. cloudpipeline
// -ext.cn-south-1.myhuaweicloud.com). After signing, the request is redirected
// to cfg.Gateway at the TCP layer while keeping the Host header pointing at
// the Huawei hostname. The gateway then routes by Host to the correct upstream.
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

	// Redirect TCP to the gateway while preserving the Host header (already
	// baked into the signature). Without this the request would try to resolve
	// the Huawei hostname directly, bypassing the gateway.
	gwURL, err := url.Parse(c.cfg.Gateway)
	if err != nil || gwURL.Host == "" {
		return fmt.Errorf("invalid gateway %q in config: want http(s)://host[:port]", c.cfg.Gateway)
	}
	req.Host = req.URL.Host
	req.URL.Scheme = gwURL.Scheme
	req.URL.Host = gwURL.Host

	resp, err := c.http.Do(req)
	if err != nil {
		if os.IsTimeout(err) {
			return fmt.Errorf("request timed out (30s) via gateway %s: check gateway reachability", c.cfg.Gateway)
		}
		return fmt.Errorf("send request via gateway %s: %w", c.cfg.Gateway, err)
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
			preview := string(respBody)
			if len(preview) > 200 {
				preview = preview[:200] + "..."
			}
			return fmt.Errorf("decode response (status %d): %w\nraw: %s", resp.StatusCode, err, preview)
		}
	}
	return nil
}
