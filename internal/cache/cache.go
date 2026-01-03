package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/EstebanForge/md-over-here/internal/extractor"
)

// CacheEntry represents a cached result
type CacheEntry struct {
	URL       string             `json:"url"`
	FetchedAt time.Time          `json:"fetchedAt"`
	Markdown  string             `json:"markdown"`
	Metadata  extractor.Metadata `json:"metadata"`
}

// Cache handles caching of fetched and converted content
type Cache struct {
	dir string
	ttl time.Duration
}

// NewCache creates a new cache instance
func NewCache(cacheDir string, ttl time.Duration) (*Cache, error) {
	if cacheDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("getting home directory: %w", err)
		}
		cacheDir = filepath.Join(home, ".config", "md-over-here", "cache")
	}

	if ttl == 0 {
		ttl = 24 * time.Hour
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return nil, fmt.Errorf("creating cache directory: %w", err)
	}

	return &Cache{
		dir: cacheDir,
		ttl: ttl,
	}, nil
}

// Get retrieves a cached entry if it exists and is not expired
func (c *Cache) Get(rawURL string) (*CacheEntry, error) {
	normalized, err := normalizeURL(rawURL)
	if err != nil {
		return nil, err
	}

	hash := hashURL(normalized)
	path := filepath.Join(c.dir, hash+".json")

	// Check if cache file exists
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return nil, nil // Cache miss
	}
	if err != nil {
		return nil, fmt.Errorf("checking cache file: %w", err)
	}

	// Check if cache is expired based on file modification time
	if time.Since(info.ModTime()) > c.ttl {
		return nil, nil // Cache expired
	}

	// Read cache file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading cache file: %w", err)
	}

	var entry CacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		return nil, fmt.Errorf("unmarshaling cache entry: %w", err)
	}

	return &entry, nil
}

// Set stores an entry in the cache
func (c *Cache) Set(rawURL string, markdown string, metadata extractor.Metadata) error {
	normalized, err := normalizeURL(rawURL)
	if err != nil {
		return err
	}

	hash := hashURL(normalized)
	path := filepath.Join(c.dir, hash+".json")

	entry := CacheEntry{
		URL:       normalized,
		FetchedAt: time.Now().UTC(),
		Markdown:  markdown,
		Metadata:  metadata,
	}

	data, err := json.MarshalIndent(entry, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling cache entry: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("writing cache file: %w", err)
	}

	return nil
}

// normalizeURL normalizes a URL for consistent cache keys
// Lowercase scheme and host, sort query parameters, remove fragment
func normalizeURL(rawURL string) (string, error) {
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", fmt.Errorf("parsing URL: %w", err)
	}

	// Lowercase scheme and host
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)

	// Sort query parameters for consistency
	if u.RawQuery != "" {
		query := u.Query()
		// Sort keys
		keys := make([]string, 0, len(query))
		for k := range query {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		// Build sorted query string
		values := url.Values{}
		for _, k := range keys {
			for _, v := range query[k] {
				values.Add(k, v)
			}
		}
		u.RawQuery = values.Encode()
	}

	// Remove fragment
	u.Fragment = ""

	return u.String(), nil
}

// hashURL generates a SHA256 hash of the URL for use as cache key
func hashURL(url string) string {
	hash := sha256.Sum256([]byte(url))
	return hex.EncodeToString(hash[:])
}
