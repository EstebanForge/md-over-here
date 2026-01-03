package fetcher

import (
	"fmt"
	"io"
	"net/http"
	"time"
)

// Fetcher defines the interface for fetching content from URLs
// This abstraction allows for multiple implementations (HTTP, Chrome/headless browser)
type Fetcher interface {
	Fetch(url string) (*FetchResult, error)
}

// FetchResult contains the result of a successful fetch operation
type FetchResult struct {
	HTML     string      // Raw HTML content
	FinalURL string      // Final URL after following redirects
	Headers  http.Header // Response headers
}

// HTTPFetcher implements Fetcher using standard HTTP client
type HTTPFetcher struct {
	client    *http.Client
	timeout   time.Duration
	userAgent string
}

// NewHTTPFetcher creates a new HTTP-based fetcher
func NewHTTPFetcher(timeout time.Duration, userAgent string) *HTTPFetcher {
	if timeout == 0 {
		timeout = 30 * time.Second
	}
	if userAgent == "" {
		userAgent = "md-over-here/1.0 (+https://github.com/EstebanForge/md-over-here)"
	}

	return &HTTPFetcher{
		client: &http.Client{
			Timeout: timeout,
			// Use default redirect policy (follows up to 10 redirects)
			CheckRedirect: nil,
		},
		timeout:   timeout,
		userAgent: userAgent,
	}
}

// Fetch retrieves content from the given URL
func (f *HTTPFetcher) Fetch(url string) (*FetchResult, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", f.userAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	return &FetchResult{
		HTML:     string(body),
		FinalURL: resp.Request.URL.String(), // URL after redirects
		Headers:  resp.Header,
	}, nil
}
