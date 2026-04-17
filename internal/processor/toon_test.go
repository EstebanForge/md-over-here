package processor

import (
	"testing"

	"github.com/EstebanForge/md-over-here/internal/extractor"
)

func TestResultToTOON(t *testing.T) {
	tests := []struct {
		name   string
		result Result
		fields []string
		// We check that output contains these strings
		contains []string
	}{
		{
			name: "minimal fields",
			result: Result{
				URL:      "https://example.com",
				Markdown: "# Test Content\n\nThis is test content.",
				Metadata: extractor.Metadata{
					Title: "Test Article",
				},
				Cached: false,
			},
			fields: []string{"url", "title", "length", "cached"},
			contains: []string{
				`url: "https://example.com"`,
				"title: Test Article",
				"length:",
				"cached: false",
			},
		},
		{
			name: "all fields",
			result: Result{
				URL:      "https://example.com/article",
				Markdown: "# Article\n\nFull article content here.",
				Metadata: extractor.Metadata{
					Title:       "Article Title",
					Author:      "John Doe",
					PublishDate: "2025-04-16",
					Description: "Test description",
				},
				Cached: true,
			},
			fields: []string{"url", "title", "author", "date", "description", "length", "cached"},
			contains: []string{
				`url: "https://example.com/article"`,
				"title: Article Title",
				"author: John Doe",
				`date: "2025-04-16"`,
				"description: Test description",
				"cached: true",
			},
		},
		{
			name: "with excerpt",
			result: Result{
				URL:      "https://example.com",
				Markdown: "This is a long article content that should be truncated to 200 characters for the excerpt field. " + string(make([]byte, 200)),
				Metadata: extractor.Metadata{
					Title: "Long Article",
				},
				Cached: false,
			},
			fields: []string{"url", "title", "excerpt"},
			contains: []string{
				"url:",
				"title: Long Article",
				"excerpt:",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := tt.result.ToTOON(tt.fields)
			if err != nil {
				t.Fatalf("ToTOON error: %v", err)
			}

			// Check that output contains expected strings
			for _, expected := range tt.contains {
				if !contains(output, expected) {
					t.Errorf("output does not contain %q\nGot: %s", expected, output)
				}
			}

			// Verify no trailing newline
			if len(output) > 0 && output[len(output)-1] == '\n' {
				t.Error("output has trailing newline, should have none")
			}
		})
	}
}

func TestResultExcerpt(t *testing.T) {
	tests := []struct {
		name     string
		markdown string
		expected string
	}{
		{
			name:     "short content",
			markdown: "Short content",
			expected: "Short content",
		},
		{
			name:     "long content",
			markdown: string(make([]byte, 300)),
			expected: string(make([]byte, 200)),
		},
		{
			name:     "exact 200 chars",
			markdown: string(make([]byte, 200)),
			expected: string(make([]byte, 200)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Result{
				Markdown: tt.markdown,
			}
			got := result.excerpt()
			if len(got) != len(tt.expected) {
				t.Errorf("excerpt length = %d, want %d", len(got), len(tt.expected))
			}
		})
	}
}

// Helper function
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
