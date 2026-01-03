package converter

import (
	"strings"
	"testing"

	"github.com/EstebanForge/md-over-here/internal/extractor"
)

func TestNewConverter(t *testing.T) {
	conv := NewConverter()
	if conv == nil {
		t.Fatal("NewConverter returned nil")
	}
	if conv.converter == nil {
		t.Error("Converter's internal converter is nil")
	}
}

func TestConvertBasicHTML(t *testing.T) {
	conv := NewConverter()
	html := "<p>This is a test paragraph.</p>"
	metadata := extractor.Metadata{
		Title: "Test Article",
	}
	sourceURL := "https://example.com/test"

	result, err := conv.Convert(html, metadata, sourceURL)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Check that result contains expected elements
	if !strings.Contains(result, "# Test Article") {
		t.Error("Result missing title heading")
	}

	if !strings.Contains(result, "**URL:** https://example.com/test") {
		t.Error("Result missing URL metadata")
	}

	if !strings.Contains(result, "This is a test paragraph") {
		t.Error("Result missing converted content")
	}

	if !strings.Contains(result, "<!-- Fetched:") {
		t.Error("Result missing fetch timestamp")
	}
}

func TestConvertWithFullMetadata(t *testing.T) {
	conv := NewConverter()
	html := "<p>Content</p>"
	metadata := extractor.Metadata{
		Title:       "Full Metadata Test",
		Author:      "John Doe",
		PublishDate: "2025-01-01",
		Description: "This is a test description",
	}
	sourceURL := "https://example.com/article"

	result, err := conv.Convert(html, metadata, sourceURL)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify all metadata is present
	expectedStrings := []string{
		"# Full Metadata Test",
		"**URL:** https://example.com/article",
		"**Author:** John Doe",
		"**Published:** 2025-01-01",
		"**Description:** This is a test description",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(result, expected) {
			t.Errorf("Result missing expected string: %q", expected)
		}
	}
}

func TestConvertWithPartialMetadata(t *testing.T) {
	conv := NewConverter()
	html := "<p>Content</p>"
	metadata := extractor.Metadata{
		Title: "Partial Metadata",
		// Author, PublishDate, Description are empty
	}
	sourceURL := "https://example.com/article"

	result, err := conv.Convert(html, metadata, sourceURL)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Should have title and URL
	if !strings.Contains(result, "# Partial Metadata") {
		t.Error("Result missing title")
	}

	if !strings.Contains(result, "**URL:**") {
		t.Error("Result missing URL")
	}

	// Should NOT have empty metadata fields
	if strings.Contains(result, "**Author:**") {
		t.Error("Result should not include empty Author field")
	}

	if strings.Contains(result, "**Published:**") {
		t.Error("Result should not include empty PublishDate field")
	}

	if strings.Contains(result, "**Description:**") {
		t.Error("Result should not include empty Description field")
	}
}

func TestConvertHTMLElements(t *testing.T) {
	conv := NewConverter()

	tests := []struct {
		name     string
		html     string
		contains string
	}{
		{
			name:     "heading conversion",
			html:     "<h2>Heading 2</h2>",
			contains: "## Heading 2",
		},
		{
			name:     "bold text",
			html:     "<strong>Bold text</strong>",
			contains: "**Bold text**",
		},
		{
			name:     "italic text",
			html:     "<em>Italic text</em>",
			contains: "Italic text", // Can be *text* or _text_ in markdown
		},
		{
			name:     "link",
			html:     `<a href="https://example.com">Link text</a>`,
			contains: "[Link text](https://example.com)",
		},
		{
			name:     "list",
			html:     "<ul><li>Item 1</li><li>Item 2</li></ul>",
			contains: "- Item 1",
		},
	}

	metadata := extractor.Metadata{Title: "Test"}
	sourceURL := "https://example.com/test"

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := conv.Convert(tt.html, metadata, sourceURL)
			if err != nil {
				t.Fatalf("Convert failed: %v", err)
			}

			if !strings.Contains(result, tt.contains) {
				t.Errorf("Expected result to contain %q, got:\n%s", tt.contains, result)
			}
		})
	}
}

func TestConvertEmptyHTML(t *testing.T) {
	conv := NewConverter()
	metadata := extractor.Metadata{Title: "Empty Content"}
	sourceURL := "https://example.com/empty"

	result, err := conv.Convert("", metadata, sourceURL)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Should still have metadata even with empty content
	if !strings.Contains(result, "# Empty Content") {
		t.Error("Result missing title")
	}

	if !strings.Contains(result, "**URL:**") {
		t.Error("Result missing URL")
	}
}

func TestConvertStructure(t *testing.T) {
	conv := NewConverter()
	html := "<p>Content</p>"
	metadata := extractor.Metadata{Title: "Structure Test"}
	sourceURL := "https://example.com/test"

	result, err := conv.Convert(html, metadata, sourceURL)
	if err != nil {
		t.Fatalf("Convert failed: %v", err)
	}

	// Verify structure: title, metadata, separator, content, footer
	lines := strings.Split(result, "\n")

	// First line should be title
	if !strings.HasPrefix(lines[0], "#") {
		t.Error("First line should be title heading")
	}

	// Should contain separators (---)
	separatorCount := 0
	for _, line := range lines {
		if strings.TrimSpace(line) == "---" {
			separatorCount++
		}
	}

	if separatorCount < 2 {
		t.Errorf("Expected at least 2 separators (---), found %d", separatorCount)
	}

	// Last lines should contain timestamp comment
	lastLines := strings.Join(lines[len(lines)-5:], "\n")
	if !strings.Contains(lastLines, "<!-- Fetched:") {
		t.Error("Missing timestamp comment at the end")
	}
}
