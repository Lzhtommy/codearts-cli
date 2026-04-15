// Package client implements a minimal Huawei Cloud API client with AK/SK
// signing (SDK-HMAC-SHA256).
//
// Reference: https://support.huaweicloud.com/devg-apisign/api-sign-algorithm.html
//
// We don't pull in github.com/huaweicloud/huaweicloud-sdk-go-v3 because the
// all-services module is heavy (~50MB of transitive deps) for a CLI. The
// signing algorithm is stable and well documented — implementing it directly
// keeps the binary small and the integration portable across CodeArts
// endpoints (pipeline, projectman, etc.).
package client

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

// Signer signs HTTP requests using Huawei Cloud's SDK-HMAC-SHA256 scheme.
type Signer struct {
	AK string
	SK string
}

const (
	algorithm       = "SDK-HMAC-SHA256"
	sdkDateHeader   = "X-Sdk-Date"
	sdkDateLayout   = "20060102T150405Z"
	emptyBodySHA256 = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
)

// Sign mutates req by adding Host, X-Sdk-Date and Authorization headers.
// body is the already-read request body (may be nil for GET). If req has a
// body set via req.Body, callers must ensure it is re-readable (e.g. by
// using http.NewRequest with a bytes.Reader) — Sign itself does not read
// req.Body; it uses the body parameter to compute the payload hash.
func (s *Signer) Sign(req *http.Request, body []byte) error {
	if s.AK == "" || s.SK == "" {
		return fmt.Errorf("signer: AK/SK not configured")
	}
	now := time.Now().UTC().Format(sdkDateLayout)
	req.Header.Set(sdkDateHeader, now)
	if req.Header.Get("Host") == "" {
		req.Header.Set("Host", req.URL.Host)
	}
	if req.Header.Get("Content-Type") == "" && len(body) > 0 {
		req.Header.Set("Content-Type", "application/json")
	}

	payloadHash := hashHex(body)
	signedHeaders, canonicalHeaders := canonicalHeaders(req)
	canonicalRequest := strings.Join([]string{
		req.Method,
		canonicalURI(req.URL),
		canonicalQuery(req.URL),
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")

	stringToSign := strings.Join([]string{
		algorithm,
		now,
		hashHex([]byte(canonicalRequest)),
	}, "\n")

	sig := hmacHex(s.SK, stringToSign)
	auth := fmt.Sprintf("%s Access=%s, SignedHeaders=%s, Signature=%s",
		algorithm, s.AK, signedHeaders, sig)
	req.Header.Set("Authorization", auth)
	return nil
}

func hashHex(b []byte) string {
	if len(b) == 0 {
		return emptyBodySHA256
	}
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func hmacHex(key, msg string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}

// canonicalURI returns the URL path, percent-encoded per RFC3986, and
// normalized to always end with a "/" per Huawei Cloud's signing spec.
//
// Reference: https://support.huaweicloud.com/devg-apisign/api-sign-algorithm-002.html
// The spec requires CanonicalURI to terminate with "/" even when the
// request path does not (e.g. `/v5/.../run` becomes `/v5/.../run/` for
// hashing purposes). The on-the-wire request path is NOT modified —
// only the value fed into the canonical request string is.
func canonicalURI(u *url.URL) string {
	p := u.EscapedPath()
	if p == "" {
		return "/"
	}
	if !strings.HasSuffix(p, "/") {
		p += "/"
	}
	return p
}

// canonicalQuery returns the query string with keys sorted and both keys
// and values percent-encoded.
func canonicalQuery(u *url.URL) string {
	if u.RawQuery == "" {
		return ""
	}
	q := u.Query()
	keys := make([]string, 0, len(q))
	for k := range q {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	first := true
	for _, k := range keys {
		vs := q[k]
		sort.Strings(vs)
		for _, v := range vs {
			if !first {
				b.WriteByte('&')
			}
			first = false
			b.WriteString(url.QueryEscape(k))
			b.WriteByte('=')
			b.WriteString(url.QueryEscape(v))
		}
	}
	return b.String()
}

// canonicalHeaders returns (signedHeaders, canonicalHeaderBlock) per the
// Huawei signing spec. All request headers are signed.
func canonicalHeaders(req *http.Request) (string, string) {
	type kv struct{ k, v string }
	headers := make([]kv, 0, len(req.Header)+1)
	// Ensure Host is present even if http.Request tracks it separately.
	if req.Header.Get("Host") == "" && req.URL.Host != "" {
		headers = append(headers, kv{"host", strings.TrimSpace(req.URL.Host)})
	}
	for k, vs := range req.Header {
		lk := strings.ToLower(k)
		// Join multi-value headers with commas (trimmed).
		parts := make([]string, 0, len(vs))
		for _, v := range vs {
			parts = append(parts, strings.TrimSpace(v))
		}
		headers = append(headers, kv{lk, strings.Join(parts, ",")})
	}
	sort.Slice(headers, func(i, j int) bool { return headers[i].k < headers[j].k })

	var block strings.Builder
	signed := make([]string, 0, len(headers))
	for _, h := range headers {
		block.WriteString(h.k)
		block.WriteByte(':')
		block.WriteString(h.v)
		block.WriteByte('\n')
		signed = append(signed, h.k)
	}
	return strings.Join(signed, ";"), block.String()
}

