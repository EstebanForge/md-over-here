package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/EstebanForge/md-over-here/internal/extractor"
)

func TestNewCache(t *testing.T) {
	// Create temp directory for testing
	tempDir := t.TempDir()
	cacheDir := filepath.Join(tempDir, "test-cache")

	cache, err := NewCache(cacheDir, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	if cache.dir != cacheDir {
		t.Errorf("Expected cache dir %s, got %s", cacheDir, cache.dir)
	}

	if cache.ttl != 24*time.Hour {
		t.Errorf("Expected TTL 24h, got %v", cache.ttl)
	}

	// Verify directory was created
	if _, err := os.Stat(cacheDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewCache(tempDir, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	url := "https://example.com/article"
	markdown := "# Test Article\n\nContent here."
	metadata := extractor.Metadata{
		Title:       "Test Article",
		Author:      "John Doe",
		PublishDate: "2025-01-01",
		Description: "Test description",
	}

	// Set cache entry
	err = cache.Set(url, markdown, metadata)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get cache entry
	entry, err := cache.Get(url)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if entry == nil {
		t.Fatal("Expected cache entry, got nil")
	}

	if entry.Markdown != markdown {
		t.Errorf("Expected markdown %q, got %q", markdown, entry.Markdown)
	}

	if entry.Metadata.Title != metadata.Title {
		t.Errorf("Expected title %q, got %q", metadata.Title, entry.Metadata.Title)
	}
}

func TestCacheMiss(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewCache(tempDir, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	// Try to get non-existent entry
	entry, err := cache.Get("https://example.com/nonexistent")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if entry != nil {
		t.Error("Expected nil entry for cache miss, got entry")
	}
}

func TestCacheExpiration(t *testing.T) {
	tempDir := t.TempDir()
	// Set very short TTL for testing
	cache, err := NewCache(tempDir, 1*time.Millisecond)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	url := "https://example.com/article"
	markdown := "# Test"
	metadata := extractor.Metadata{Title: "Test"}

	// Set cache entry
	err = cache.Set(url, markdown, metadata)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Wait for expiration
	time.Sleep(5 * time.Millisecond)

	// Try to get expired entry
	entry, err := cache.Get(url)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if entry != nil {
		t.Error("Expected nil for expired cache entry, got entry")
	}
}

func TestNormalizeURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "lowercase scheme and host",
			input:    "HTTPS://EXAMPLE.COM/path",
			expected: "https://example.com/path",
		},
		{
			name:     "remove fragment",
			input:    "https://example.com/page#section",
			expected: "https://example.com/page",
		},
		{
			name:     "sort query params",
			input:    "https://example.com/page?z=1&a=2&m=3",
			expected: "https://example.com/page?a=2&m=3&z=1",
		},
		{
			name:     "combined normalization",
			input:    "HTTPS://EXAMPLE.COM/Page?z=1&a=2#fragment",
			expected: "https://example.com/Page?a=2&z=1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeURL(tt.input)
			if err != nil {
				t.Fatalf("normalizeURL failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestHashURL(t *testing.T) {
	url := "https://example.com/article"
	hash1 := hashURL(url)
	hash2 := hashURL(url)

	// Same URL should produce same hash
	if hash1 != hash2 {
		t.Error("Same URL produced different hashes")
	}

	// Hash should be hex string of expected length (SHA256 = 64 hex chars)
	if len(hash1) != 64 {
		t.Errorf("Expected hash length 64, got %d", len(hash1))
	}

	// Different URL should produce different hash
	hash3 := hashURL("https://example.com/different")
	if hash1 == hash3 {
		t.Error("Different URLs produced same hash")
	}
}

func TestCacheURLNormalization(t *testing.T) {
	tempDir := t.TempDir()
	cache, err := NewCache(tempDir, 24*time.Hour)
	if err != nil {
		t.Fatalf("NewCache failed: %v", err)
	}

	markdown := "# Test"
	metadata := extractor.Metadata{Title: "Test"}

	// Set with one URL variant
	err = cache.Set("HTTPS://EXAMPLE.COM/page?z=1&a=2#frag", markdown, metadata)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	// Get with normalized equivalent URL
	entry, err := cache.Get("https://example.com/page?a=2&z=1")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if entry == nil {
		t.Fatal("Expected cache hit for normalized URL, got miss")
	}

	if entry.Markdown != markdown {
		t.Error("Retrieved wrong cache entry")
	}
}
