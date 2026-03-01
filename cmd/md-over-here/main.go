package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/EstebanForge/md-over-here/internal/cache"
	"github.com/EstebanForge/md-over-here/internal/extractor"
	"github.com/EstebanForge/md-over-here/internal/fetcher"
	"github.com/EstebanForge/md-over-here/internal/processor"
	"github.com/spf13/cobra"
)

var (
	outputPath string
	noCache    bool
	cacheDir   string
	verbose    bool
	timeout    time.Duration
	userAgent  string
)

var rootCmd = &cobra.Command{
	Use:   "md-over-here <url> [url...]",
	Short: "Fetch URLs and convert to clean markdown for LLM consumption",
	Long: `md-over-here fetches web pages and converts them to clean markdown,
optimized for feeding to Large Language Models. It extracts the main
content while removing navigation, ads, and other clutter.`,
	Args: cobra.MinimumNArgs(1),
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputPath, "save", "s", "", "Save to file (combines multiple URLs)")
	rootCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable caching for this request")
	rootCmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Custom cache directory (default: ~/.config/md-over-here/cache)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show metadata and cache status")
	rootCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP timeout")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "", "Custom User-Agent header")
}

// generateFilename creates a safe filename from URL and metadata
// Format: domain-sanitized-title.md
func generateFilename(rawURL string, metadata extractor.Metadata) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "article.md"
	}

	// Extract domain and replace dots with hyphens
	domain := strings.ReplaceAll(parsed.Host, ".", "-")

	// Sanitize title for filename
	title := metadata.Title
	if title == "" {
		title = "article"
	}

	// Remove or replace invalid characters
	// Convert to lowercase
	title = strings.ToLower(title)

	// Replace spaces and special chars with hyphens
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	title = reg.ReplaceAllString(title, "-")

	// Remove leading/trailing hyphens
	title = strings.Trim(title, "-")

	// Limit length to avoid filesystem issues
	if len(title) > 100 {
		title = title[:100]
	}

	// Combine domain and title (only if domain exists)
	var filename string
	if domain != "" && domain != "-" {
		filename = fmt.Sprintf("%s-%s.md", domain, title)
	} else {
		filename = fmt.Sprintf("%s.md", title)
	}

	return filename
}

func run(cmd *cobra.Command, args []string) error {
	urls := args

	// Validate URLs
	for _, url := range urls {
		if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
			return fmt.Errorf("invalid URL (must start with http:// or https://): %s", url)
		}
	}

	// Setup cache
	var c *cache.Cache
	if !noCache {
		var err error
		c, err = cache.NewCache(cacheDir, 24*time.Hour)
		if err != nil {
			if verbose {
				fmt.Fprintf(os.Stderr, "Warning: cache initialization failed: %v\n", err)
			}
			// Continue without cache
			c = nil
		}
	}

	// Setup fetcher
	f := fetcher.NewHTTPFetcher(timeout, userAgent)

	// Setup processor
	p := processor.NewProcessor(f, c)

	// Process options
	opts := processor.Options{
		NoCache: noCache,
		Verbose: verbose,
	}

	// Process all URLs
	results := make([]processor.Result, len(urls))
	for i, url := range urls {
		if verbose {
			fmt.Fprintf(os.Stderr, "Processing: %s\n", url)
		}
		results[i] = p.Process(url, opts)
	}

	// Collect output and errors
	var output strings.Builder
	hasErrors := false

	for i, result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", result.URL, result.Error)
			hasErrors = true
			continue
		}

		// Show cache status in verbose mode
		if verbose {
			status := "fetched"
			if result.Cached {
				status = "cached"
			}
			fmt.Fprintf(os.Stderr, "✓ %s (%s)\n", result.URL, status)
		}

		// Add separator between multiple URLs
		if i > 0 {
			output.WriteString("\n---\n## Next Article\n---\n\n")
		}

		output.WriteString(result.Markdown)
	}

	// Write output
	if outputPath != "" {
		// --save flag: save to file (combines multiple URLs)
		// Create parent directories if they don't exist
		dir := filepath.Dir(outputPath)
		if dir != "." && dir != "" {
			if err := os.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("creating output directory: %w", err)
			}
		}

		if err := os.WriteFile(outputPath, []byte(output.String()), 0644); err != nil {
			return fmt.Errorf("writing output file: %w", err)
		}
		if verbose {
			fmt.Fprintf(os.Stderr, "Saved to: %s\n", outputPath)
		}
	} else {
		// Default: print to stdout (for agent/LLM consumption)
		fmt.Print(output.String())
	}

	if hasErrors {
		os.Exit(1)
	}

	return nil
}

// createSymlink creates a symlink at ~/.local/bin/mdoh pointing to the brew installation
func createSymlink() error {
	// Common Homebrew installation paths (in order of preference)
	brewPaths := []string{
		"/opt/homebrew/bin/md-over-here",                                    // macOS ARM
		"/usr/local/bin/md-over-here",                                       // macOS Intel
		"/home/linuxbrew/.linuxbrew/bin/md-over-here",                       // Linux
		"/home/linuxbrew/.linuxbrew/Cellar/md-over-here/*/bin/md-over-here", // Linux alternative
	}

	// Find Homebrew installation
	var brewExecutable string
	for _, path := range brewPaths {
		// Use glob for wildcard paths
		matches, err := filepath.Glob(path)
		if err == nil && len(matches) > 0 {
			// Check if the file exists and is executable
			if info, err := os.Stat(matches[0]); err == nil && !info.IsDir() {
				brewExecutable = matches[0]
				break
			}
		}
		// Also check non-glob paths
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			brewExecutable = path
			break
		}
	}

	// If no Homebrew installation found, use current executable
	executable := brewExecutable
	if executable == "" {
		exec, err := os.Executable()
		if err != nil {
			return fmt.Errorf("getting executable path: %w", err)
		}
		executable = exec
	}

	// Resolve symlink if executable itself is a symlink
	resolvedExecutable, err := filepath.EvalSymlinks(executable)
	if err != nil {
		resolvedExecutable = executable
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("getting home directory: %w", err)
	}

	binDir := filepath.Join(homeDir, ".local", "bin")
	symlinkPath := filepath.Join(binDir, "mdoh")

	// Check if symlink already exists
	if fileInfo, err := os.Lstat(symlinkPath); err == nil {
		// File/symlink exists
		if fileInfo.Mode()&os.ModeSymlink != 0 {
			// It's a symlink, check if it points to our executable
			target, err := os.Readlink(symlinkPath)
			if err == nil {
				// Check if target matches either the resolved or original executable path
				if target == resolvedExecutable || target == executable {
					// Symlink is correct
					return nil
				}
			}
			// Remove stale/broken symlink
			_ = os.Remove(symlinkPath)
		} else {
			// Regular file exists (not a symlink), don't overwrite
			return fmt.Errorf("file exists and is not a symlink: %s", symlinkPath)
		}
	}

	// Create ~/.local/bin directory if it doesn't exist
	if err := os.MkdirAll(binDir, 0755); err != nil {
		return fmt.Errorf("creating bin directory: %w", err)
	}

	// Create symlink pointing to the resolved executable path
	if err := os.Symlink(resolvedExecutable, symlinkPath); err != nil {
		return fmt.Errorf("creating symlink: %w", err)
	}

	fmt.Fprintf(os.Stderr, "Created symlink: %s -> %s\n", symlinkPath, resolvedExecutable)
	fmt.Fprintf(os.Stderr, "You can now use: mdoh <url>\n")

	return nil
}

func main() {
	// Create symlink on first run
	if err := createSymlink(); err != nil {
		// Silently ignore symlink errors - not critical for functionality
		_ = err
	}
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
