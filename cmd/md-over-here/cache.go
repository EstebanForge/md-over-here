package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache",
	Short: "Manage cache",
	Long:  `Commands for managing the local cache of fetched content.`,
}

var cacheClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Clear all cached content",
	Long:  `Remove all cached content from the cache directory.`,
	RunE:  runCacheClear,
}

var cacheStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show cache statistics",
	Long:  `Display statistics about the cache (size, number of entries, location).`,
	RunE:  runCacheStats,
}

func init() {
	// Add cache subcommands
	cacheCmd.AddCommand(cacheClearCmd)
	cacheCmd.AddCommand(cacheStatsCmd)

	// Add cache command to root
	rootCmd.AddCommand(cacheCmd)

	// Add cache-dir flag to cache commands
	cacheClearCmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Custom cache directory (default: ~/.config/md-over-here/cache)")
	cacheStatsCmd.Flags().StringVar(&cacheDir, "cache-dir", "", "Custom cache directory (default: ~/.config/md-over-here/cache)")
}

func getCacheDir() (string, error) {
	if cacheDir != "" {
		return cacheDir, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("getting home directory: %w", err)
	}

	return filepath.Join(home, ".config", "md-over-here", "cache"), nil
}

func runCacheClear(cmd *cobra.Command, args []string) error {
	dir, err := getCacheDir()
	if err != nil {
		return err
	}

	// Check if cache directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Cache directory does not exist: %s\n", dir)
		return nil
	}

	// Count files before clearing
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading cache directory: %w", err)
	}

	fileCount := 0
	for _, entry := range entries {
		if !entry.IsDir() {
			fileCount++
		}
	}

	if fileCount == 0 {
		fmt.Printf("Cache is already empty: %s\n", dir)
		return nil
	}

	// Remove all files in cache directory
	for _, entry := range entries {
		if !entry.IsDir() {
			filePath := filepath.Join(dir, entry.Name())
			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("removing cache file %s: %w", entry.Name(), err)
			}
		}
	}

	fmt.Printf("Cleared %d cache entries from: %s\n", fileCount, dir)
	return nil
}

func runCacheStats(cmd *cobra.Command, args []string) error {
	dir, err := getCacheDir()
	if err != nil {
		return err
	}

	// Check if cache directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		fmt.Printf("Cache directory does not exist: %s\n", dir)
		fmt.Println("\nNo cache entries.")
		return nil
	}

	// Read cache directory
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("reading cache directory: %w", err)
	}

	// Calculate stats
	var totalSize int64
	fileCount := 0

	for _, entry := range entries {
		if !entry.IsDir() {
			fileCount++
			filePath := filepath.Join(dir, entry.Name())
			info, err := os.Stat(filePath)
			if err == nil {
				totalSize += info.Size()
			}
		}
	}

	// Display stats
	fmt.Printf("Cache Directory: %s\n", dir)
	fmt.Printf("Entries: %d\n", fileCount)
	fmt.Printf("Total Size: %s\n", formatBytes(totalSize))

	if fileCount > 0 {
		fmt.Printf("Average Entry Size: %s\n", formatBytes(totalSize/int64(fileCount)))
	}

	return nil
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
