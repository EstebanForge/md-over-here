package processor

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/EstebanForge/md-over-here/internal/cache"
	"github.com/EstebanForge/md-over-here/internal/fetcher"
)

// mockFetcher for testing
type mockFetcher struct {
	html     string
	finalURL string
	err      error
}

func (m *mockFetcher) Fetch(url string) (*fetcher.FetchResult, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &fetcher.FetchResult{
		HTML:     m.html,
		FinalURL: m.finalURL,
		Headers:  http.Header{},
	}, nil
}

func TestNewProcessor(t *testing.T) {
	mockF := &mockFetcher{}
	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)

	proc := NewProcessor(mockF, c)

	if proc == nil {
		t.Fatal("NewProcessor returned nil")
	}

	if proc.fetcher == nil {
		t.Error("Processor fetcher is nil")
	}

	if proc.extractor == nil {
		t.Error("Processor extractor is nil")
	}

	if proc.converter == nil {
		t.Error("Processor converter is nil")
	}

	if proc.cache == nil {
		t.Error("Processor cache is nil")
	}
}

func TestProcessSuccess(t *testing.T) {
	html := `
		<!DOCTYPE html>
		<html>
		<head><title>Test Article</title></head>
		<body><article><h1>Test</h1><p>Content here.</p></article></body>
		</html>
	`

	mockF := &mockFetcher{
		html:     html,
		finalURL: "https://example.com/article",
	}

	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)
	proc := NewProcessor(mockF, c)

	result := proc.Process("https://example.com/article", Options{})

	if result.Error != nil {
		t.Fatalf("Process failed: %v", result.Error)
	}

	if result.Markdown == "" {
		t.Error("Markdown result is empty")
	}

	if result.URL != "https://example.com/article" {
		t.Errorf("Expected URL https://example.com/article, got %s", result.URL)
	}

	if result.Cached {
		t.Error("First fetch should not be cached")
	}

	// Verify markdown contains expected content
	if !strings.Contains(result.Markdown, "Test") {
		t.Error("Markdown missing expected content")
	}
}

func TestProcessWithCache(t *testing.T) {
	html := `<html><body><p>Content</p></body></html>`

	mockF := &mockFetcher{
		html:     html,
		finalURL: "https://example.com/article",
	}

	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)
	proc := NewProcessor(mockF, c)

	url := "https://example.com/article"

	// First fetch
	result1 := proc.Process(url, Options{})
	if result1.Error != nil {
		t.Fatalf("First process failed: %v", result1.Error)
	}

	if result1.Cached {
		t.Error("First fetch should not be cached")
	}

	// Second fetch - should be cached
	result2 := proc.Process(url, Options{})
	if result2.Error != nil {
		t.Fatalf("Second process failed: %v", result2.Error)
	}

	if !result2.Cached {
		t.Error("Second fetch should be cached")
	}

	// Results should be identical
	if result1.Markdown != result2.Markdown {
		t.Error("Cached result differs from original")
	}
}

func TestProcessNoCache(t *testing.T) {
	html := `<html><body><p>Content</p></body></html>`

	mockF := &mockFetcher{
		html:     html,
		finalURL: "https://example.com/article",
	}

	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)
	proc := NewProcessor(mockF, c)

	url := "https://example.com/article"

	// First fetch (will be cached)
	proc.Process(url, Options{})

	// Second fetch with NoCache option
	result := proc.Process(url, Options{NoCache: true})

	if result.Cached {
		t.Error("NoCache option should prevent cache use")
	}
}

func TestProcessFetchError(t *testing.T) {
	mockF := &mockFetcher{
		err: fmt.Errorf("network error"),
	}

	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)
	proc := NewProcessor(mockF, c)

	result := proc.Process("https://example.com/article", Options{})

	if result.Error == nil {
		t.Error("Expected error for fetch failure, got nil")
	}

	if result.Markdown != "" {
		t.Error("Should not have markdown on error")
	}

	if !strings.Contains(result.Error.Error(), "fetching URL") {
		t.Errorf("Error should mention 'fetching URL', got: %v", result.Error)
	}
}

func TestProcessNilCache(t *testing.T) {
	html := `<html><body><p>Content</p></body></html>`

	mockF := &mockFetcher{
		html:     html,
		finalURL: "https://example.com/article",
	}

	// Processor with no cache
	proc := NewProcessor(mockF, nil)

	result := proc.Process("https://example.com/article", Options{})

	if result.Error != nil {
		t.Fatalf("Process failed: %v", result.Error)
	}

	if result.Cached {
		t.Error("Should not be cached when cache is nil")
	}

	if result.Markdown == "" {
		t.Error("Should still produce markdown without cache")
	}
}

func TestProcessRealHTTPFetcher(t *testing.T) {
	// Integration test with real HTTP fetcher and test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		html := `
			<!DOCTYPE html>
			<html>
			<head><title>Integration Test</title></head>
			<body>
				<article>
					<h1>Integration Test Article</h1>
					<p>This is integration test content.</p>
				</article>
			</body>
			</html>
		`
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(html))
	}))
	defer server.Close()

	httpFetcher := fetcher.NewHTTPFetcher(5*time.Second, "test-agent")
	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)
	proc := NewProcessor(httpFetcher, c)

	result := proc.Process(server.URL, Options{})

	if result.Error != nil {
		t.Fatalf("Process failed: %v", result.Error)
	}

	if result.Markdown == "" {
		t.Error("Markdown is empty")
	}

	if !strings.Contains(result.Markdown, "Integration Test") {
		t.Error("Markdown missing expected content")
	}

	// Second fetch should be cached
	result2 := proc.Process(server.URL, Options{})
	if !result2.Cached {
		t.Error("Second fetch should be cached")
	}
}

func TestProcessMetadataExtraction(t *testing.T) {
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Article with Metadata</title>
			<meta name="author" content="Test Author">
			<meta name="description" content="Test description">
		</head>
		<body>
			<article>
				<h1>Main Article</h1>
				<p>Article content here.</p>
			</article>
		</body>
		</html>
	`

	mockF := &mockFetcher{
		html:     html,
		finalURL: "https://example.com/article",
	}

	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)
	proc := NewProcessor(mockF, c)

	result := proc.Process("https://example.com/article", Options{})

	if result.Error != nil {
		t.Fatalf("Process failed: %v", result.Error)
	}

	// Metadata should be populated
	if result.Metadata.Title == "" {
		t.Error("Title metadata not extracted")
	}
}

func TestProcessURLInResult(t *testing.T) {
	mockF := &mockFetcher{
		html:     "<p>Content</p>",
		finalURL: "https://example.com/article",
	}

	tempDir := t.TempDir()
	c, _ := cache.NewCache(tempDir, 24*time.Hour)
	proc := NewProcessor(mockF, c)

	inputURL := "https://example.com/article"
	result := proc.Process(inputURL, Options{})

	if result.URL != inputURL {
		t.Errorf("Expected URL %q in result, got %q", inputURL, result.URL)
	}
}
