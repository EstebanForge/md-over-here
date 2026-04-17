package processor

import (
	"testing"

	"github.com/EstebanForge/md-over-here/internal/extractor"
)

func TestTruncation(t *testing.T) {
	tests := []struct {
		name       string
		markdown   string
		truncateTo int
		full       bool
		wantTrunc  bool
		minLength  int
		maxLength  int
	}{
		{
			name:       "short content - no truncation",
			markdown:   "Short content",
			truncateTo: 100,
			full:       false,
			wantTrunc:  false,
			minLength:  13,
			maxLength:  13,
		},
		{
			name:       "long content - truncated",
			markdown:   string(make([]byte, 10000)),
			truncateTo: 5000,
			full:       false,
			wantTrunc:  true,
			minLength:  4900,
			maxLength:  5000,
		},
		{
			name:       "truncation bypassed with full flag",
			markdown:   string(make([]byte, 10000)),
			truncateTo: 5000,
			full:       true,
			wantTrunc:  false,
			minLength:  10000,
			maxLength:  10000,
		},
		{
			name:       "zero limit - no truncation",
			markdown:   string(make([]byte, 1000)),
			truncateTo: 0,
			full:       false,
			wantTrunc:  false,
			minLength:  1000,
			maxLength:  1000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := Options{
				Truncate: tt.truncateTo,
				Full:     tt.full,
			}

			// Create a mock result
			result := Result{
				URL:      "https://example.com",
				Markdown: tt.markdown,
				Metadata: extractor.Metadata{},
			}

			// Apply truncation
			if !opts.Full && opts.Truncate > 0 && len(result.Markdown) > opts.Truncate {
				result.Markdown = truncateMarkdown(result.Markdown, opts.Truncate)
				result.Truncated = true
			}

			// Check truncation flag
			if result.Truncated != tt.wantTrunc {
				t.Errorf("Truncated = %v, want %v", result.Truncated, tt.wantTrunc)
			}

			// Check length
			if len(result.Markdown) < tt.minLength || len(result.Markdown) > tt.maxLength {
				t.Errorf("Length = %d, want between %d and %d", len(result.Markdown), tt.minLength, tt.maxLength)
			}
		})
	}
}

func TestTruncateMarkdown(t *testing.T) {
	tests := []struct {
		name      string
		markdown  string
		maxLen    int
		minLength int
		maxLength int
	}{
		{
			name:      "finds sentence boundary",
			markdown:  "This is sentence one. This is sentence two. This is sentence three.",
			maxLen:    30,
			minLength: 20,
			maxLength: 30,
		},
		{
			name:      "finds paragraph break",
			markdown:  "Paragraph one\n\nParagraph two\n\nParagraph three",
			maxLen:    20,
			minLength: 15,
			maxLength: 20,
		},
		{
			name:      "no boundary found",
			markdown:  "onewordanotherwordyetmorewords",
			maxLen:    15,
			minLength: 15,
			maxLength: 15,
		},
		{
			name:      "shorter than limit",
			markdown:  "Short text",
			maxLen:    100,
			minLength: 10,
			maxLength: 10,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := truncateMarkdown(tt.markdown, tt.maxLen)
			if len(result) < tt.minLength || len(result) > tt.maxLength {
				t.Errorf("Length = %d, want between %d and %d", len(result), tt.minLength, tt.maxLength)
			}
		})
	}
}
