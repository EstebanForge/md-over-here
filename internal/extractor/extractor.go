package extractor

import (
	"bytes"
	"fmt"
	"net/url"
	"strings"

	readability "codeberg.org/readeck/go-readability/v2"
)

// Metadata contains extracted metadata about the content
type Metadata struct {
	Title       string
	Author      string
	PublishDate string
	Description string
}

// ExtractResult contains the extracted content and metadata
type ExtractResult struct {
	HTML     string   // Cleaned HTML content (main article)
	Metadata Metadata // Extracted metadata
}

// Extractor extracts main content and metadata from HTML
type Extractor struct{}

// NewExtractor creates a new extractor
func NewExtractor() *Extractor {
	return &Extractor{}
}

// Extract processes HTML and returns cleaned content with metadata
func (e *Extractor) Extract(html string, sourceURL string) (*ExtractResult, error) {
	parsedURL, err := url.Parse(sourceURL)
	if err != nil {
		return nil, fmt.Errorf("parsing URL: %w", err)
	}

	// Parse the HTML and extract main content
	article, err := readability.FromReader(strings.NewReader(html), parsedURL)
	if err != nil {
		// If extraction fails, fall back to returning full HTML
		// This ensures we always return something useful
		return &ExtractResult{
			HTML: html,
			Metadata: Metadata{
				Title: "Untitled",
			},
		}, nil
	}

	// Extract published time
	publishDate := ""
	if pubTime, err := article.PublishedTime(); err == nil && !pubTime.IsZero() {
		publishDate = pubTime.Format("2006-01-02")
	}

	// Render article content to HTML
	var htmlBuf bytes.Buffer
	if err := article.RenderHTML(&htmlBuf); err != nil {
		return nil, fmt.Errorf("rendering article HTML: %w", err)
	}

	return &ExtractResult{
		HTML: htmlBuf.String(),
		Metadata: Metadata{
			Title:       article.Title(),
			Author:      article.Byline(),
			PublishDate: publishDate,
			Description: article.Excerpt(),
		},
	}, nil
}
