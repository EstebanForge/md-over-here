package fetcher

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewHTTPFetcher(t *testing.T) {
	fetcher := NewHTTPFetcher(30*time.Second, "test-agent")

	if fetcher.userAgent != "test-agent" {
		t.Errorf("Expected user agent 'test-agent', got %q", fetcher.userAgent)
	}

	if fetcher.client == nil {
		t.Error("HTTP client is nil")
	}
}

func TestNewHTTPFetcherDefaults(t *testing.T) {
	fetcher := NewHTTPFetcher(0, "")

	if !strings.Contains(fetcher.userAgent, "md-over-here") {
		t.Errorf("Expected default user agent to contain 'md-over-here', got %q", fetcher.userAgent)
	}
}

func TestHTTPFetcherSuccess(t *testing.T) {
	// Create test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html><body>Test content</body></html>"))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(5*time.Second, "test-agent")
	result, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if result == nil {
		t.Fatal("Result is nil")
	}

	if !strings.Contains(result.HTML, "Test content") {
		t.Error("HTML does not contain expected content")
	}

	if result.FinalURL != server.URL {
		t.Errorf("Expected final URL %q, got %q", server.URL, result.FinalURL)
	}

	if result.Headers == nil {
		t.Error("Headers is nil")
	}
}

func TestHTTPFetcherUserAgent(t *testing.T) {
	receivedUA := ""

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	customUA := "custom-user-agent/1.0"
	fetcher := NewHTTPFetcher(5*time.Second, customUA)
	_, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	if receivedUA != customUA {
		t.Errorf("Expected User-Agent %q, server received %q", customUA, receivedUA)
	}
}

func TestHTTPFetcherHTTPError(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
	}{
		{"404 Not Found", http.StatusNotFound},
		{"500 Internal Server Error", http.StatusInternalServerError},
		{"403 Forbidden", http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			fetcher := NewHTTPFetcher(5*time.Second, "test")
			_, err := fetcher.Fetch(server.URL)

			if err == nil {
				t.Errorf("Expected error for status %d, got nil", tt.statusCode)
			}

			if !strings.Contains(err.Error(), "HTTP") {
				t.Errorf("Error should mention HTTP status, got: %v", err)
			}
		})
	}
}

func TestHTTPFetcherRedirect(t *testing.T) {
	// Server that redirects
	redirectServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("Final destination"))
	}))
	defer redirectServer.Close()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, redirectServer.URL, http.StatusFound)
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(5*time.Second, "test")
	result, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Should follow redirect and get final content
	if !strings.Contains(result.HTML, "Final destination") {
		t.Error("Did not receive content from redirected URL")
	}

	// Final URL should be the redirected URL
	if result.FinalURL != redirectServer.URL {
		t.Errorf("Expected final URL %q, got %q", redirectServer.URL, result.FinalURL)
	}
}

func TestHTTPFetcherInvalidURL(t *testing.T) {
	fetcher := NewHTTPFetcher(5*time.Second, "test")
	_, err := fetcher.Fetch("not-a-valid-url")

	if err == nil {
		t.Error("Expected error for invalid URL, got nil")
	}
}

func TestHTTPFetcherTimeout(t *testing.T) {
	// Server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	// Very short timeout
	fetcher := NewHTTPFetcher(100*time.Millisecond, "test")
	_, err := fetcher.Fetch(server.URL)

	if err == nil {
		t.Error("Expected timeout error, got nil")
	}
}

func TestHTTPFetcherHeaders(t *testing.T) {
	receivedHeaders := http.Header{}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(5*time.Second, "test-agent")
	_, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify expected headers were sent
	if receivedHeaders.Get("User-Agent") == "" {
		t.Error("User-Agent header not sent")
	}

	if receivedHeaders.Get("Accept") == "" {
		t.Error("Accept header not sent")
	}

	if receivedHeaders.Get("Accept-Language") == "" {
		t.Error("Accept-Language header not sent")
	}
}

func TestHTTPFetcherResponseHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("X-Custom-Header", "test-value")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	}))
	defer server.Close()

	fetcher := NewHTTPFetcher(5*time.Second, "test")
	result, err := fetcher.Fetch(server.URL)

	if err != nil {
		t.Fatalf("Fetch failed: %v", err)
	}

	// Verify response headers are captured
	if result.Headers.Get("Content-Type") == "" {
		t.Error("Content-Type header not captured")
	}

	if result.Headers.Get("X-Custom-Header") != "test-value" {
		t.Error("Custom header not captured")
	}
}

func TestFetcherInterface(t *testing.T) {
	// Verify HTTPFetcher implements Fetcher interface
	var _ Fetcher = (*HTTPFetcher)(nil)
}
