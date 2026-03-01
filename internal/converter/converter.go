package converter

import (
	"fmt"
	"strings"
	"time"

	"github.com/EstebanForge/md-over-here/internal/extractor"
	md "github.com/JohannesKaufmann/html-to-markdown"
)

// Converter converts HTML to markdown
type Converter struct {
	converter *md.Converter
}

// NewConverter creates a new markdown converter
func NewConverter() *Converter {
	converter := md.NewConverter("", true, nil)

	// Configure for clean output
	converter.Keep("mark") // Keep <mark> elements

	return &Converter{
		converter: converter,
	}
}

// Convert transforms HTML to markdown with metadata header
func (c *Converter) Convert(html string, metadata extractor.Metadata, sourceURL string) (string, error) {
	// Convert HTML to markdown
	markdown, err := c.converter.ConvertString(html)
	if err != nil {
		return "", fmt.Errorf("converting HTML to markdown: %w", err)
	}

	// Build the complete output with metadata header
	var output strings.Builder

	// Title as H1
	if metadata.Title != "" {
		fmt.Fprintf(&output, "# %s\n\n", metadata.Title)
	}

	// Metadata section
	fmt.Fprintf(&output, "**URL:** %s\n", sourceURL)

	if metadata.Author != "" {
		fmt.Fprintf(&output, "**Author:** %s\n", metadata.Author)
	}

	if metadata.PublishDate != "" {
		fmt.Fprintf(&output, "**Published:** %s\n", metadata.PublishDate)
	}

	if metadata.Description != "" {
		fmt.Fprintf(&output, "**Description:** %s\n", metadata.Description)
	}

	output.WriteString("\n---\n\n")

	// Main content
	output.WriteString(markdown)

	// Footer with fetch timestamp
	fmt.Fprintf(&output, "\n\n---\n<!-- Fetched: %s -->\n",
		time.Now().UTC().Format(time.RFC3339))

	return output.String(), nil
}
