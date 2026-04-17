package main

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/EstebanForge/md-over-here/internal/aggregator"
	"github.com/EstebanForge/md-over-here/internal/cache"
	"github.com/EstebanForge/md-over-here/internal/extractor"
	"github.com/EstebanForge/md-over-here/internal/fetcher"
	"github.com/EstebanForge/md-over-here/internal/hooks"
	"github.com/EstebanForge/md-over-here/internal/processor"
	"github.com/EstebanForge/md-over-here/internal/toon"
	"github.com/spf13/cobra"
)

var (
	outputPath     string
	noCache        bool
	cacheDir       string
	verbose        bool
	timeout        time.Duration
	userAgent      string
	outputFormat   string
	fields         string
	truncate       int
	full           bool
	showAggregates bool
	noHelp         bool
)

var rootCmd = &cobra.Command{
	Use:   "md-over-here <url> [url...]",
	Short: "Fetch URLs and convert to clean markdown for LLM consumption",
	Long: `md-over-here fetches web pages and converts them to clean markdown,
optimized for feeding to Large Language Models. It extracts the main
content while removing navigation, ads, and other clutter.`,
	Args: cobra.ArbitraryArgs,
	RunE: run,
}

func init() {
	rootCmd.Flags().StringVarP(&outputPath, "save", "s", "", "Save to file (combines multiple URLs)")
	rootCmd.Flags().BoolVar(&noCache, "no-cache", false, "Disable caching for this request")
	rootCmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Custom cache directory (default: ~/.config/md-over-here/cache)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Show metadata and cache status")
	rootCmd.Flags().DurationVar(&timeout, "timeout", 30*time.Second, "HTTP timeout")
	rootCmd.Flags().StringVar(&userAgent, "user-agent", "", "Custom User-Agent header")
	rootCmd.Flags().StringVar(&outputFormat, "format", "toon", "Output format: toon, markdown, json (default 'toon')")
	rootCmd.Flags().StringVar(&fields, "fields", "", "Comma-separated fields to output (for toon format)")
	rootCmd.Flags().IntVar(&truncate, "truncate", 0, "Maximum content length in bytes (0 = no limit)")
	rootCmd.Flags().BoolVar(&full, "full", false, "Bypass truncation and show full content")
	rootCmd.Flags().BoolVar(&showAggregates, "aggregates", false, "Show aggregate statistics for batch operations")
	rootCmd.Flags().BoolVar(&noHelp, "no-help", false, "Suppress help suggestions")

	// Hook subcommands
	hookCmd := &cobra.Command{
		Use:   "hook [command]",
		Short: "Manage shell hooks for ambient cache status",
	}

	initCmd := &cobra.Command{
		Use:   "init",
		Short: "Install shell hooks",
		RunE:  runHookInit,
	}

	statusCmd := &cobra.Command{
		Use:   "status",
		Short: "Check hook installation status",
		RunE:  runHookStatus,
	}

	uninstallCmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Remove shell hooks",
		RunE:  runHookUninstall,
	}

	hookCmd.AddCommand(initCmd, statusCmd, uninstallCmd)
	rootCmd.AddCommand(hookCmd)

	// Cache command already exists in cache.go
}

// generateFilename creates a safe filename from URL and metadata
func generateFilename(rawURL string, metadata extractor.Metadata) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return "article.md"
	}

	domain := strings.ReplaceAll(parsed.Host, ".", "-")

	title := metadata.Title
	if title == "" {
		title = "article"
	}

	title = strings.ToLower(title)
	reg := regexp.MustCompile(`[^a-z0-9]+`)
	title = reg.ReplaceAllString(title, "-")
	title = strings.Trim(title, "-")

	if len(title) > 100 {
		title = title[:100]
	}

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

	// If no args, show dashboard
	if len(urls) == 0 {
		return runDashboard()
	}

	// Validate URLs
	for _, u := range urls {
		if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
			return fmt.Errorf("invalid URL (must start with http:// or https://): %s", u)
		}
	}

	// Parse fields flag
	var fieldList []string
	if fields != "" {
		fieldList = strings.Split(fields, ",")
		for i := range fieldList {
			fieldList[i] = strings.TrimSpace(fieldList[i])
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
			c = nil
		}
	}

	// Setup fetcher
	f := fetcher.NewHTTPFetcher(timeout, userAgent)

	// Setup processor
	p := processor.NewProcessor(f, c)

	// Process options
	opts := processor.Options{
		NoCache:  noCache,
		Verbose:  verbose,
		Truncate: truncate,
		Full:     full,
	}

	// Process all URLs
	results := make([]processor.Result, len(urls))
	for i, u := range urls {
		if verbose {
			fmt.Fprintf(os.Stderr, "Processing: %s\n", u)
		}
		results[i] = p.Process(u, opts)
	}

	// Compute aggregates if requested
	var agg aggregator.AggregateStats
	if showAggregates {
		agg = aggregator.ComputeAggregates(results)
	}

	// Output based on format
	switch outputFormat {
	case "toon":
		return outputTOON(results, agg, fieldList)
	case "json":
		return outputJSON(results, agg, fieldList)
	case "markdown":
		return outputMarkdown(results, fieldList)
	default:
		return fmt.Errorf("unsupported format: %s (valid: toon, json, markdown)", outputFormat)
	}
}

func outputTOON(results []processor.Result, agg aggregator.AggregateStats, fieldList []string) error {
	var output strings.Builder

	// Output aggregates first if present
	if showAggregates {
		toonStr, err := agg.ToTOON()
		if err != nil {
			return fmt.Errorf("marshaling aggregates: %w", err)
		}
		output.WriteString(toonStr)
		output.WriteString("\n")
	}

	// Default fields if none specified
	if len(fieldList) == 0 {
		fieldList = []string{"url", "title", "length", "cached"}
	}

	// Output each result
	for _, result := range results {
		if result.Error != nil {
			errMap := map[string]interface{}{
				"url":   result.URL,
				"error": result.Error.Message,
				"code":  string(result.Error.Code),
			}
			data, err := toon.Marshal(errMap)
			if err != nil {
				return fmt.Errorf("marshaling error: %w", err)
			}
			output.Write(data)
			output.WriteString("\n")
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", result.URL, result.Error)
			continue
		}

		if verbose {
			status := "fetched"
			if result.Cached {
				status = "cached"
			}
			fmt.Fprintf(os.Stderr, "✓ %s (%s)\n", result.URL, status)
		}

		toonStr, err := result.ToTOON(fieldList)
		if err != nil {
			return fmt.Errorf("converting result to TOON: %w", err)
		}
		output.WriteString(toonStr)
		output.WriteString("\n")
	}

	// Write output
	if outputPath != "" {
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
		fmt.Print(output.String())
	}

	// Show contextual help if not suppressed
	if !noHelp {
		printContextualHelp(fieldList, len(results) > 1)
	}

	return nil
}

func outputJSON(results []processor.Result, agg aggregator.AggregateStats, fieldList []string) error {
	return fmt.Errorf("JSON format not yet implemented")
}

func outputMarkdown(results []processor.Result, fieldList []string) error {
	var output strings.Builder

	for _, result := range results {
		if result.Error != nil {
			fmt.Fprintf(os.Stderr, "Error processing %s: %v\n", result.URL, result.Error)
			continue
		}

		if verbose {
			status := "fetched"
			if result.Cached {
				status = "cached"
			}
			fmt.Fprintf(os.Stderr, "✓ %s (%s)\n", result.URL, status)
		}

		if output.Len() > 0 {
			output.WriteString("\n---\n## Next Article\n---\n\n")
		}

		output.WriteString(result.Markdown)
	}

	if outputPath != "" {
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
		fmt.Print(output.String())
	}

	return nil
}

func runDashboard() error {
	dir, err := getCacheDir()
	if err != nil {
		return fmt.Errorf("getting cache directory: %w", err)
	}

	// Get cache stats
	var cacheCount int
	var totalSize int64

	if entries, err := os.ReadDir(dir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				cacheCount++
				filePath := filepath.Join(dir, entry.Name())
				if info, err := os.Stat(filePath); err == nil {
					totalSize += info.Size()
				}
			}
		}
	}

	binPath, err := os.Executable()
	if err != nil {
		binPath = "unknown"
	}

	dashboard := map[string]interface{}{
		"version":     "1.0.0",
		"cache_count": cacheCount,
		"cache_size":  formatBytes(totalSize),
		"location":    dir,
		"binary_path": binPath,
	}

	data, err := toon.Marshal(dashboard)
	if err != nil {
		return fmt.Errorf("marshaling dashboard: %w", err)
	}

	fmt.Println(string(data))
	fmt.Println("\nCommon commands:")
	fmt.Println("  md-over-here https://example.com     # Fetch a URL")
	fmt.Println("  md-over-here --fields url,title      # Select specific fields")
	fmt.Println("  md-over-here --truncate 5000 <url>   # Limit content length")
	fmt.Println("  md-over-here cache stats             # View cache statistics")
	fmt.Println("  md-over-here hook init               # Install shell hooks")

	return nil
}

func runHookInit(cmd *cobra.Command, args []string) error {
	shellType := hooks.DetectShell()

	dir, err := getCacheDir()
	if err != nil {
		return fmt.Errorf("getting cache directory: %w", err)
	}

	binPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("getting binary path: %w", err)
	}

	config := hooks.HookConfig{
		ShellType: shellType,
		CacheDir:  dir,
		BinPath:   binPath,
	}

	script, err := hooks.GenerateHookScript(config)
	if err != nil {
		return fmt.Errorf("generating hook script: %w", err)
	}

	scriptPath, err := hooks.GetShellScriptPath(shellType)
	if err != nil {
		return fmt.Errorf("getting shell script path: %w", err)
	}

	existing, err := os.ReadFile(scriptPath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("reading shell config: %w", err)
	}

	if strings.Contains(string(existing), "# md-over-here hook integration") {
		fmt.Printf("Hooks already installed in %s\n", scriptPath)
		return nil
	}

	f, err := os.OpenFile(scriptPath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("opening shell config: %w", err)
	}
	defer func() { _ = f.Close() }()

	if _, err := f.WriteString("\n" + script); err != nil {
		return fmt.Errorf("writing hook script: %w", err)
	}

	fmt.Printf("Hooks installed to %s\n", scriptPath)
	fmt.Printf("Run 'source %s' or restart your shell to activate\n", scriptPath)

	return nil
}

func runHookStatus(cmd *cobra.Command, args []string) error {
	shellType := hooks.DetectShell()
	scriptPath, err := hooks.GetShellScriptPath(shellType)
	if err != nil {
		return fmt.Errorf("getting shell script path: %w", err)
	}

	existing, err := os.ReadFile(scriptPath)
	if os.IsNotExist(err) {
		fmt.Printf("Hooks not installed (file not found: %s)\n", scriptPath)
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading shell config: %w", err)
	}

	if strings.Contains(string(existing), "# md-over-here hook integration") {
		fmt.Printf("Hooks installed in %s\n", scriptPath)
		return nil
	}

	fmt.Printf("Hooks not installed in %s\n", scriptPath)
	return nil
}

func runHookUninstall(cmd *cobra.Command, args []string) error {
	shellType := hooks.DetectShell()
	scriptPath, err := hooks.GetShellScriptPath(shellType)
	if err != nil {
		return fmt.Errorf("getting shell script path: %w", err)
	}

	existing, err := os.ReadFile(scriptPath)
	if os.IsNotExist(err) {
		fmt.Printf("Hooks not installed (file not found: %s)\n", scriptPath)
		return nil
	}
	if err != nil {
		return fmt.Errorf("reading shell config: %w", err)
	}

	content := string(existing)
	if !strings.Contains(content, "# md-over-here hook integration") {
		fmt.Printf("Hooks not installed in %s\n", scriptPath)
		return nil
	}

	startIdx := strings.Index(content, "# md-over-here hook integration")
	if startIdx == -1 {
		fmt.Printf("Hooks not installed in %s\n", scriptPath)
		return nil
	}

	endIdx := strings.Index(content[startIdx:], "\n\n")
	if endIdx == -1 {
		endIdx = len(content)
	} else {
		endIdx += startIdx + 2
	}

	newContent := content[:startIdx] + content[endIdx:]

	if err := os.WriteFile(scriptPath, []byte(newContent), 0644); err != nil {
		return fmt.Errorf("writing shell config: %w", err)
	}

	fmt.Printf("Hooks uninstalled from %s\n", scriptPath)
	fmt.Printf("Run 'source %s' or restart your shell to apply changes\n", scriptPath)

	return nil
}

func printContextualHelp(fieldList []string, isBatch bool) {
	fmt.Fprintf(os.Stderr, "\n💡 Tip: ")
	if isBatch {
		fmt.Fprintf(os.Stderr, "Use --aggregates to see batch statistics\n")
	} else if len(fieldList) > 0 && len(fieldList) <= 4 {
		fmt.Fprintf(os.Stderr, "Add --fields url,title,author for more metadata\n")
	} else {
		fmt.Fprintf(os.Stderr, "Use --truncate 5000 to limit content length\n")
	}
	fmt.Fprintf(os.Stderr, "   Run --no-help to suppress these messages\n")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
