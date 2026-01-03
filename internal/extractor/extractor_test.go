package extractor

import (
	"strings"
	"testing"
)

func TestNewExtractor(t *testing.T) {
	ext := NewExtractor()
	if ext == nil {
		t.Fatal("NewExtractor returned nil")
	}
}

func TestExtractBasicArticle(t *testing.T) {
	ext := NewExtractor()

	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Test Article</title>
		</head>
		<body>
			<article>
				<h1>Test Article Title</h1>
				<p>This is the main content of the article.</p>
				<p>It has multiple paragraphs.</p>
			</article>
		</body>
		</html>
	`

	sourceURL := "https://example.com/article"

	result, err := ext.Extract(html, sourceURL)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if result == nil {
		t.Fatal("Extract returned nil result")
	}

	// Check that HTML content was extracted
	if result.HTML == "" {
		t.Error("Extracted HTML is empty")
	}

	// Check metadata
	if result.Metadata.Title == "" {
		t.Error("Title was not extracted")
	}
}

func TestExtractWithMetadata(t *testing.T) {
	ext := NewExtractor()

	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Article with Metadata</title>
			<meta name="author" content="John Doe">
			<meta name="description" content="This is a test description">
		</head>
		<body>
			<article>
				<h1>Article Title</h1>
				<p class="byline">By Jane Smith</p>
				<p>Article content here.</p>
			</article>
		</body>
		</html>
	`

	sourceURL := "https://example.com/article"

	result, err := ext.Extract(html, sourceURL)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should extract some metadata (exact behavior depends on readability algorithm)
	if result.Metadata.Title == "" {
		t.Error("Expected title to be extracted")
	}
}

func TestExtractInvalidHTML(t *testing.T) {
	ext := NewExtractor()

	html := `<p>Not a valid HTML document</p>`
	sourceURL := "https://example.com/invalid"

	// Should not fail, but fall back to returning the HTML as-is
	result, err := ext.Extract(html, sourceURL)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if result == nil {
		t.Fatal("Extract returned nil result")
	}

	// Should have some content even if extraction fails
	if result.HTML == "" {
		t.Error("Expected some HTML content")
	}
}

func TestExtractEmptyHTML(t *testing.T) {
	ext := NewExtractor()

	html := ""
	sourceURL := "https://example.com/empty"

	result, err := ext.Extract(html, sourceURL)

	// Empty HTML should still return a result (fallback behavior)
	// but might fail extraction
	if result == nil && err == nil {
		t.Fatal("Extract returned nil result and nil error")
	}

	// As long as we don't panic, the test passes
	// The extractor handles empty HTML gracefully
}

func TestExtractInvalidURL(t *testing.T) {
	ext := NewExtractor()

	html := `<html><body><p>Content</p></body></html>`
	// Go's url.Parse is very lenient, so most strings parse successfully
	// This tests that extraction still works with minimal URLs
	sourceURL := "not-a-valid-url"

	result, err := ext.Extract(html, sourceURL)

	// url.Parse accepts most strings, so this may not error
	// Just verify we get a result
	if err != nil {
		// If it errors, that's fine too
		return
	}

	if result == nil {
		t.Fatal("Expected result even with simple URL")
	}
}

func TestExtractComplexHTML(t *testing.T) {
	ext := NewExtractor()

	// HTML with navigation, ads, footer (noise that should be filtered)
	html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Complex Page</title>
		</head>
		<body>
			<nav>
				<a href="/">Home</a>
				<a href="/about">About</a>
			</nav>

			<div class="ad">Advertisement here</div>

			<article>
				<h1>Main Article</h1>
				<p>This is the important content.</p>
				<p>More important content here.</p>
			</article>

			<aside>
				<h2>Related Articles</h2>
				<ul><li>Article 1</li></ul>
			</aside>

			<footer>
				<p>Copyright 2025</p>
			</footer>
		</body>
		</html>
	`

	sourceURL := "https://example.com/complex"

	result, err := ext.Extract(html, sourceURL)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	if result == nil {
		t.Fatal("Extract returned nil result")
	}

	// The extracted content should focus on the article
	// Exact behavior depends on readability algorithm
	if !strings.Contains(result.HTML, "important content") {
		t.Error("Main content not found in extracted HTML")
	}
}

func TestExtractResultStructure(t *testing.T) {
	ext := NewExtractor()

	html := `
		<!DOCTYPE html>
		<html>
		<head><title>Test</title></head>
		<body><article><p>Content</p></article></body>
		</html>
	`
	sourceURL := "https://example.com/test"

	result, err := ext.Extract(html, sourceURL)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Verify result structure
	if result.HTML == "" {
		t.Error("HTML field is empty")
	}

	// Metadata should exist (even if fields are empty)
	_ = result.Metadata.Title
	_ = result.Metadata.Author
	_ = result.Metadata.PublishDate
	_ = result.Metadata.Description
}

func TestExtractFallback(t *testing.T) {
	ext := NewExtractor()

	// Minimal HTML that might not extract well
	html := `<div>Just some text</div>`
	sourceURL := "https://example.com/minimal"

	result, err := ext.Extract(html, sourceURL)
	if err != nil {
		t.Fatalf("Extract failed: %v", err)
	}

	// Should still return a result (fallback behavior)
	if result == nil {
		t.Fatal("Expected fallback result, got nil")
	}

	// Should have some HTML content (original or extracted)
	if result.HTML == "" {
		t.Error("Expected HTML content in fallback")
	}
}
