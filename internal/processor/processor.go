package processor

import (
	"fmt"

	"github.com/EstebanForge/md-over-here/internal/cache"
	"github.com/EstebanForge/md-over-here/internal/converter"
	"github.com/EstebanForge/md-over-here/internal/extractor"
	"github.com/EstebanForge/md-over-here/internal/fetcher"
)

// Result represents the result of processing a URL
type Result struct {
	URL      string
	Markdown string
	Metadata extractor.Metadata
	Error    error
	Cached   bool // Whether result came from cache
}

// Options configures processor behavior
type Options struct {
	NoCache bool // Disable cache entirely
	Verbose bool // Verbose logging
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
		result.Error = fmt.Errorf("fetching URL: %w", err)
		return result
	}

	// Extract main content and metadata
	extractResult, err := p.extractor.Extract(fetchResult.HTML, fetchResult.FinalURL)
	if err != nil {
		result.Error = fmt.Errorf("extracting content: %w", err)
		return result
	}

	// Convert to markdown
	markdown, err := p.converter.Convert(extractResult.HTML, extractResult.Metadata, fetchResult.FinalURL)
	if err != nil {
		result.Error = fmt.Errorf("converting to markdown: %w", err)
		return result
	}

	result.Markdown = markdown
	result.Metadata = extractResult.Metadata

	// Cache the result unless caching is disabled
	if !opts.NoCache && p.cache != nil {
		// Don't fail the request if caching fails - we silently continue
		_ = p.cache.Set(url, markdown, extractResult.Metadata)
	}

	return result
}
