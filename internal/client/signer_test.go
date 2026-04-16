package client

import (
	"net/http"
	"net/url"
	"testing"
)

func TestHashHex_Empty(t *testing.T) {
	got := hashHex(nil)
	if got != emptyBodySHA256 {
		t.Errorf("hashHex(nil) = %q, want %q", got, emptyBodySHA256)
	}
	got2 := hashHex([]byte{})
	if got2 != emptyBodySHA256 {
		t.Errorf("hashHex([]) = %q, want %q", got2, emptyBodySHA256)
	}
}

func TestHashHex_NonEmpty(t *testing.T) {
	got := hashHex([]byte("{}"))
	want := "44136fa355b3678a1146ad16f7e8649e94fb4fc21fe77e8310c060f61caaff8a"
	if got != want {
		t.Errorf("hashHex({}) = %q, want %q", got, want)
	}
}

func TestHmacHex(t *testing.T) {
	// HMAC-SHA256("key", "message") — verified against external tool.
	got := hmacHex("key", "message")
	want := "6e9ef29b75fffc5b7abae527d58fdadb2fe42e7219011976917343065f58ed4a"
	if got != want {
		t.Errorf("hmacHex = %q, want %q", got, want)
	}
}

func TestCanonicalURI(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty path", "", "/"},
		{"root", "/", "/"},
		{"trailing slash", "/v5/abc/", "/v5/abc/"},
		{"no trailing slash appends", "/v5/abc/run", "/v5/abc/run/"},
		{"encoded chars preserved", "/v5/abc%20def/run", "/v5/abc%20def/run/"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("https://example.com" + tt.raw)
			got := canonicalURI(u)
			if got != tt.want {
				t.Errorf("canonicalURI(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestCanonicalQuery(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{"empty", "", ""},
		{"single key", "a=1", "a=1"},
		{"sorted keys", "b=2&a=1", "a=1&b=2"},
		{"multi value sorted", "a=2&a=1", "a=1&a=2"},
		{"encoded special chars", "key=hello world", "key=hello+world"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			u, _ := url.Parse("https://example.com/path?" + tt.raw)
			got := canonicalQuery(u)
			if got != tt.want {
				t.Errorf("canonicalQuery(%q) = %q, want %q", tt.raw, got, tt.want)
			}
		})
	}
}

func TestCanonicalHeaders(t *testing.T) {
	req, _ := http.NewRequest("GET", "https://example.com/path", nil)
	req.Header.Set("X-Sdk-Date", "20260415T000000Z")
	req.Header.Set("Host", "example.com")
	req.Header.Set("Content-Type", "application/json")

	signed, block := canonicalHeaders(req)

	// Signed headers must be sorted, semicolon-separated, lowercase.
	wantSigned := "content-type;host;x-sdk-date"
	if signed != wantSigned {
		t.Errorf("signedHeaders = %q, want %q", signed, wantSigned)
	}
	// Block must end with newline per header.
	if block == "" {
		t.Error("canonicalHeaders block is empty")
	}
	// Each line must be "lowercase-key:value\n".
	if block[len(block)-1] != '\n' {
		t.Error("canonicalHeaders block must end with newline")
	}
}

func TestSign_SetsRequiredHeaders(t *testing.T) {
	s := &Signer{AK: "TESTAKID1234567890", SK: "TestSecretKey123456"}
	req, _ := http.NewRequest("POST", "https://example.com/v5/proj/api/pipelines/pid/run", nil)

	if err := s.Sign(req, nil); err != nil {
		t.Fatalf("Sign() error: %v", err)
	}
	if req.Header.Get("X-Sdk-Date") == "" {
		t.Error("Sign() did not set X-Sdk-Date header")
	}
	if req.Header.Get("Authorization") == "" {
		t.Error("Sign() did not set Authorization header")
	}
	auth := req.Header.Get("Authorization")
	if len(auth) < len(algorithm) {
		t.Fatalf("Authorization header too short: %q", auth)
	}
	if auth[:len(algorithm)] != algorithm {
		t.Errorf("Authorization header should start with %q, got %q", algorithm, auth[:len(algorithm)])
	}
}

func TestSign_EmptyCredentials(t *testing.T) {
	s := &Signer{AK: "", SK: ""}
	req, _ := http.NewRequest("GET", "https://example.com/", nil)
	err := s.Sign(req, nil)
	if err == nil {
		t.Error("Sign() with empty AK/SK should return error")
	}
}

func TestSign_BodyAffectsSignature(t *testing.T) {
	s := &Signer{AK: "AK", SK: "SK"}

	req1, _ := http.NewRequest("POST", "https://example.com/path", nil)
	_ = s.Sign(req1, nil)
	sig1 := req1.Header.Get("Authorization")

	req2, _ := http.NewRequest("POST", "https://example.com/path", nil)
	_ = s.Sign(req2, []byte(`{"key":"value"}`))
	sig2 := req2.Header.Get("Authorization")

	if sig1 == sig2 {
		t.Error("Different bodies should produce different signatures")
	}
}
