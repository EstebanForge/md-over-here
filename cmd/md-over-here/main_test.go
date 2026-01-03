package main

import (
	"testing"

	"github.com/EstebanForge/md-over-here/internal/extractor"
)

func TestGenerateFilename(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		metadata extractor.Metadata
		expected string
	}{
		{
			name: "simple domain and title",
			url:  "https://example.com/article",
			metadata: extractor.Metadata{
				Title: "Test Article",
			},
			expected: "example-com-test-article.md",
		},
		{
			name: "domain with subdomain",
			url:  "https://blog.example.com/post",
			metadata: extractor.Metadata{
				Title: "My Post",
			},
			expected: "blog-example-com-my-post.md",
		},
		{
			name: "title with special characters",
			url:  "https://example.com/article",
			metadata: extractor.Metadata{
				Title: "Hello, World! How are you?",
			},
			expected: "example-com-hello-world-how-are-you.md",
		},
		{
			name: "title with multiple spaces",
			url:  "https://example.com/article",
			metadata: extractor.Metadata{
				Title: "The   Quick   Brown   Fox",
			},
			expected: "example-com-the-quick-brown-fox.md",
		},
		{
			name: "title with numbers",
			url:  "https://example.com/article",
			metadata: extractor.Metadata{
				Title: "Top 10 Ways to Do Something",
			},
			expected: "example-com-top-10-ways-to-do-something.md",
		},
		{
			name: "empty title",
			url:  "https://example.com/article",
			metadata: extractor.Metadata{
				Title: "",
			},
			expected: "example-com-article.md",
		},
		{
			name:     "invalid URL",
			url:      "not-a-url",
			metadata: extractor.Metadata{},
			expected: "article.md",
		},
		{
			name: "actitud xyz blog",
			url:  "https://actitud.xyz/blog/2025/in-memoriam-os/",
			metadata: extractor.Metadata{
				Title: "In memoriam: Os",
			},
			expected: "actitud-xyz-in-memoriam-os.md",
		},
		{
			name: "youtube revanced",
			url:  "https://actitud.xyz/blog/2022/youtube-revanced-2022/",
			metadata: extractor.Metadata{
				Title: "YouTube ReVanced is in da house!",
			},
			expected: "actitud-xyz-youtube-revanced-is-in-da-house.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := generateFilename(tt.url, tt.metadata)
			if result != tt.expected {
				t.Errorf("Expected filename %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestGenerateFilenameLengthLimit(t *testing.T) {
	// Create a very long title
	longTitle := ""
	for i := 0; i < 200; i++ {
		longTitle += "word"
		if i < 199 {
			longTitle += "-"
		}
	}

	metadata := extractor.Metadata{
		Title: longTitle,
	}

	result := generateFilename("https://example.com/article", metadata)

	// Check that filename isn't too long (should be truncated)
	if len(result) > 150 {
		t.Errorf("Filename too long: %d characters", len(result))
	}

	// Should end with .md
	if len(result) < 3 || result[len(result)-3:] != ".md" {
		t.Errorf("Filename should end with .md, got %q", result)
	}
}
