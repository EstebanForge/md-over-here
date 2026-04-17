package processor

import (
	"github.com/EstebanForge/md-over-here/internal/cache"
	"github.com/EstebanForge/md-over-here/internal/converter"
	"github.com/EstebanForge/md-over-here/internal/errors"
	"github.com/EstebanForge/md-over-here/internal/extractor"
	"github.com/EstebanForge/md-over-here/internal/fetcher"
	"github.com/EstebanForge/md-over-here/internal/toon"
)

// Result represents the result of processing a URL
type Result struct {
	URL        string
	Markdown   string
	Metadata   extractor.Metadata
	Error      *errors.StructuredError
	Cached     bool // Whether result came from cache
	Truncated  bool // Whether content was truncated
	FullLength int  // Original content length before truncation
}

// ToTOON converts the result to TOON format
func (r *Result) ToTOON(fields []string) (string, error) {
	// Build object with selected fields
	obj := make(map[string]interface{})

	for _, field := range fields {
		switch field {
		case "url":
			obj["url"] = r.URL
		case "title":
			obj["title"] = r.Metadata.Title
		case "author":
			obj["author"] = r.Metadata.Author
		case "date":
			obj["date"] = r.Metadata.PublishDate
		case "description":
			obj["description"] = r.Metadata.Description
		case "length":
			obj["length"] = len(r.Markdown)
		case "cached":
			obj["cached"] = r.Cached
		case "truncated":
			obj["truncated"] = r.Truncated
		case "excerpt":
			obj["excerpt"] = r.excerpt()
		case "content":
			obj["content"] = r.Markdown
		}
	}

	// Marshal to TOON
	data, err := toon.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// excerpt returns the first 200 characters of content
func (r *Result) excerpt() string {
	if len(r.Markdown) > 200 {
		return r.Markdown[:200]
	}
	return r.Markdown
}

// Options configures processor behavior
type Options struct {
	NoCache  bool // Disable cache entirely
	Verbose  bool // Verbose logging
	Truncate int  // Maximum content length (0 = no limit)
	Full     bool // Bypass truncation and show full content
}

// Processor orchestrates fetching, extracting, and converting URLs
type Processor struct {
	fetcher   fetcher.Fetcher
	extractor *extractor.Extractor
	converter *converter.Converter
	cache     *cache.Cache
}

// NewProcessor creates a new processor
func NewProcessor(f fetcher.Fetcher, c *cache.Cache) *Processor {
	return &Processor{
		fetcher:   f,
		extractor: extractor.NewExtractor(),
		converter: converter.NewConverter(),
		cache:     c,
	}
}

// Process fetches, extracts, and converts a URL to markdown
func (p *Processor) Process(url string, opts Options) Result {
	result := Result{
		URL: url,
	}

	// Check cache unless disabled
	if !opts.NoCache && p.cache != nil {
		if cached, err := p.cache.Get(url); err == nil && cached != nil {
			result.Markdown = cached.Markdown
			result.Metadata = cached.Metadata
			result.Cached = true
			return result
		}
		// Cache miss or error - continue with fetch
	}

	// Fetch HTML
	fetchResult, err := p.fetcher.Fetch(url)
	if err != nil {
		result.Error = errors.NewNetworkError(url, err)
		return result
	}

	// Extract main content and metadata
	extractResult, err := p.extractor.Extract(fetchResult.HTML, fetchResult.FinalURL)
	if err != nil {
		result.Error = errors.NewExtractionError(url, err)
		return result
	}

	// Convert to markdown
	markdown, err := p.converter.Convert(extractResult.HTML, extractResult.Metadata, fetchResult.FinalURL)
	if err != nil {
		result.Error = errors.NewExtractionError(url, err)
		return result
	}

	result.Markdown = markdown
	result.Metadata = extractResult.Metadata
	result.FullLength = len(markdown)

	// Apply truncation if requested and not bypassed
	if !opts.Full && opts.Truncate > 0 && len(markdown) > opts.Truncate {
		result.Markdown = truncateMarkdown(markdown, opts.Truncate)
		result.Truncated = true
	}

	// Cache the result unless caching is disabled
	if !opts.NoCache && p.cache != nil {
		// Don't fail the request if caching fails - we silently continue
		_ = p.cache.Set(url, markdown, extractResult.Metadata)
	}

	return result
}

// truncateMarkdown truncates markdown to maxLen with smart boundary detection
func truncateMarkdown(markdown string, maxLen int) string {
	if len(markdown) <= maxLen {
		return markdown
	}

	// Try to find a good break point (sentence or paragraph)
	cutoff := maxLen
	// Look for sentence endings (., !, ?) followed by space or newline
	for i := maxLen - 1; i >= maxLen-100 && i >= 0; i-- {
		if i < len(markdown)-1 && (markdown[i] == '.' || markdown[i] == '!' || markdown[i] == '?') {
			next := i + 1
			if next < len(markdown) && (markdown[next] == ' ' || markdown[next] == '\n') {
				cutoff = next + 1
				break
			}
		}
	}

	// If no sentence boundary found, try paragraph break (double newline)
	if cutoff == maxLen {
		for i := maxLen - 1; i >= maxLen-50 && i >= 1; i-- {
			if markdown[i] == '\n' && markdown[i-1] == '\n' {
				cutoff = i + 1
				break
			}
		}
	}

	return markdown[:cutoff]
}
