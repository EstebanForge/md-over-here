package aggregator

import (
	"github.com/EstebanForge/md-over-here/internal/processor"
	"github.com/EstebanForge/md-over-here/internal/toon"
)

// AggregateStats contains computed statistics for batch operations
type AggregateStats struct {
	Total          int // Total URLs processed
	Success        int // Successful fetches
	Failed         int // Failed fetches
	CacheHits      int // Results from cache
	TotalLength    int // Total content length in characters
	AverageLength  int // Average content length
	TruncatedCount int // Number of truncated results
}

// ComputeAggregates calculates statistics from results
func ComputeAggregates(results []processor.Result) AggregateStats {
	stats := AggregateStats{
		Total: len(results),
	}

	var totalLength int
	var successCount int

	for _, result := range results {
		if result.Error != nil {
			stats.Failed++
			continue
		}

		stats.Success++
		successCount++

		if result.Cached {
			stats.CacheHits++
		}

		// Use full length for accurate totals
		if result.FullLength > 0 {
			totalLength += result.FullLength
		} else {
			totalLength += len(result.Markdown)
		}

		if result.Truncated {
			stats.TruncatedCount++
		}
	}

	stats.TotalLength = totalLength

	// Calculate average length from successful results
	if successCount > 0 {
		stats.AverageLength = totalLength / successCount
	}

	return stats
}

// ToTOON converts aggregates to TOON format
func (a *AggregateStats) ToTOON() (string, error) {
	obj := map[string]interface{}{
		"total":           a.Total,
		"success":         a.Success,
		"failed":          a.Failed,
		"cache_hits":      a.CacheHits,
		"total_length":    a.TotalLength,
		"average_length":  a.AverageLength,
		"truncated_count": a.TruncatedCount,
	}

	data, err := toon.Marshal(obj)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
