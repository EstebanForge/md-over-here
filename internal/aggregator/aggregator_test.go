package aggregator

import (
	"testing"

	"github.com/EstebanForge/md-over-here/internal/errors"
	"github.com/EstebanForge/md-over-here/internal/extractor"
	"github.com/EstebanForge/md-over-here/internal/processor"
)

func TestComputeAggregates(t *testing.T) {
	tests := []struct {
		name          string
		results       []Result
		wantTotal     int
		wantSuccess   int
		wantFailed    int
		wantCacheHits int
		wantTruncated int
	}{
		{
			name: "all successful",
			results: []Result{
				{
					Markdown:   "Content 1",
					FullLength: 100,
					Cached:     false,
					Truncated:  false,
				},
				{
					Markdown:   "Content 2",
					FullLength: 200,
					Cached:     true,
					Truncated:  true,
				},
			},
			wantTotal:     2,
			wantSuccess:   2,
			wantFailed:    0,
			wantCacheHits: 1,
			wantTruncated: 1,
		},
		{
			name: "mixed success and failure",
			results: []Result{
				{
					Markdown:   "Content 1",
					FullLength: 100,
					Cached:     false,
				},
				{
					Error: errors.NewNetworkError("https://example.com", nil),
				},
				{
					Markdown:   "Content 2",
					FullLength: 200,
					Cached:     true,
				},
			},
			wantTotal:     3,
			wantSuccess:   2,
			wantFailed:    1,
			wantCacheHits: 1,
		},
		{
			name: "all failed",
			results: []Result{
				{
					Error: errors.NewNetworkError("https://example.com", nil),
				},
				{
					Error: errors.NewExtractionError("https://example.org", nil),
				},
			},
			wantTotal:   2,
			wantSuccess: 0,
			wantFailed:  2,
		},
		{
			name:        "empty results",
			results:     []Result{},
			wantTotal:   0,
			wantSuccess: 0,
			wantFailed:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Convert to processor.Results
			procResults := make([]processor.Result, len(tt.results))
			for i, r := range tt.results {
				procResults[i] = processor.Result{
					Markdown:   r.Markdown,
					Metadata:   extractor.Metadata{},
					Error:      r.Error,
					Cached:     r.Cached,
					Truncated:  r.Truncated,
					FullLength: r.FullLength,
				}
			}

			stats := ComputeAggregates(procResults)

			if stats.Total != tt.wantTotal {
				t.Errorf("Total = %d, want %d", stats.Total, tt.wantTotal)
			}
			if stats.Success != tt.wantSuccess {
				t.Errorf("Success = %d, want %d", stats.Success, tt.wantSuccess)
			}
			if stats.Failed != tt.wantFailed {
				t.Errorf("Failed = %d, want %d", stats.Failed, tt.wantFailed)
			}
			if stats.CacheHits != tt.wantCacheHits {
				t.Errorf("CacheHits = %d, want %d", stats.CacheHits, tt.wantCacheHits)
			}
			if stats.TruncatedCount != tt.wantTruncated {
				t.Errorf("TruncatedCount = %d, want %d", stats.TruncatedCount, tt.wantTruncated)
			}
		})
	}
}

func TestAggregateTotals(t *testing.T) {
	results := []processor.Result{
		{
			Markdown:   "Content 1",
			FullLength: 1000,
			Cached:     false,
		},
		{
			Markdown:   "Content 2",
			FullLength: 2000,
			Cached:     true,
		},
		{
			Markdown:   "Content 3",
			FullLength: 1500,
			Cached:     false,
		},
	}

	stats := ComputeAggregates(results)

	// Check totals
	if stats.TotalLength != 4500 {
		t.Errorf("TotalLength = %d, want 4500", stats.TotalLength)
	}

	// Check average (4500 / 3 = 1500)
	if stats.AverageLength != 1500 {
		t.Errorf("AverageLength = %d, want 1500", stats.AverageLength)
	}
}

func TestAggregateToTOON(t *testing.T) {
	stats := AggregateStats{
		Total:          3,
		Success:        2,
		Failed:         1,
		CacheHits:      1,
		TotalLength:    4500,
		AverageLength:  1500,
		TruncatedCount: 1,
	}

	output, err := stats.ToTOON()
	if err != nil {
		t.Fatalf("ToTOON error: %v", err)
	}

	// Check that output contains expected fields
	expectedFields := []string{
		"total: 3",
		"success: 2",
		"failed: 1",
		"cache_hits: 1",
		"total_length: 4500",
		"average_length: 1500",
		"truncated_count: 1",
	}

	for _, field := range expectedFields {
		if !containsString(output, field) {
			t.Errorf("TOON output missing field: %s\nGot: %s", field, output)
		}
	}
}

// Helper types for testing
type Result struct {
	Markdown   string
	FullLength int
	Cached     bool
	Truncated  bool
	Error      *errors.StructuredError
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
