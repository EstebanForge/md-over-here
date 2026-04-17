package toon

import (
	"bytes"
	"fmt"
	"strings"
	"testing"
)

func TestEncodeString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "simple string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "string with comma",
			input:    "hello,world",
			expected: `"hello,world"`,
		},
		{
			name:     "string with colon",
			input:    "key:value",
			expected: `"key:value"`,
		},
		{
			name:     "string with bracket",
			input:    "test[1]",
			expected: `"test[1]"`,
		},
		{
			name:     "string with newline",
			input:    "line1\nline2",
			expected: `"line1\nline2"`,
		},
		{
			name:     "string with backslash",
			input:    "C:\\Users\\name",
			expected: `"C:\\Users\\name"`,
		},
		{
			name:     "string with quote",
			input:    `he said "hello"`,
			expected: `"he said \"hello\""`,
		},
		{
			name:     "string with leading space",
			input:    "  leading",
			expected: `"  leading"`,
		},
		{
			name:     "empty string",
			input:    "",
			expected: `""`,
		},
		{
			name:     "string with tab",
			input:    "col1\tcol2",
			expected: `"col1\tcol2"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)

			err := encoder.Encode(tt.input)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			got := buf.String()
			// Remove trailing newline for comparison
			got = trimRightNewline(got)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEncodeNumber(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "integer",
			input:    42,
			expected: "42",
		},
		{
			name:     "negative integer",
			input:    -123,
			expected: "-123",
		},
		{
			name:     "zero",
			input:    0,
			expected: "0",
		},
		{
			name:     "float",
			input:    19.99,
			expected: "19.99",
		},
		{
			name:     "float with trailing zero",
			input:    1.0,
			expected: "1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)

			err := encoder.Encode(tt.input)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			got := buf.String()
			got = trimRightNewline(got)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEncodeBool(t *testing.T) {
	tests := []struct {
		name     string
		input    bool
		expected string
	}{
		{
			name:     "true",
			input:    true,
			expected: "true",
		},
		{
			name:     "false",
			input:    false,
			expected: "false",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)

			err := encoder.Encode(tt.input)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			got := buf.String()
			got = trimRightNewline(got)
			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestEncodeObject(t *testing.T) {
	tests := []struct {
		name     string
		input    map[string]interface{}
		contains []string // check that output contains these strings
	}{
		{
			name: "simple object",
			input: map[string]interface{}{
				"url":    "https://example.com",
				"title":  "Article Title",
				"length": 12345,
			},
			contains: []string{
				`url: "https://example.com"`,
				"title: Article Title",
				"length: 12345",
			},
		},
		{
			name: "object with quoted keys",
			input: map[string]interface{}{
				"url:with:colon": "value",
				"normal_key":     "value2",
			},
			contains: []string{
				`"url:with:colon": value`,
				"normal_key: value2",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			encoder := NewEncoder(&buf)

			err := encoder.Encode(tt.input)
			if err != nil {
				t.Fatalf("encode error: %v", err)
			}

			got := buf.String()
			got = trimRightNewline(got)

			// Check that all required strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(got, expected) {
					t.Errorf("output does not contain %q\nGot: %q", expected, got)
				}
			}
		})
	}
}

func TestEncodeObjectArray(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		contains []string // check that output contains these strings
	}{
		{
			name: "object array",
			input: []interface{}{
				map[string]interface{}{
					"url":    "https://example.com",
					"title":  "Article Title",
					"length": 12345,
				},
				map[string]interface{}{
					"url":    "https://example.org",
					"title":  "Another Title",
					"length": 67890,
				},
			},
			contains: []string{
				"results[2]{",
				`"https://example.com"`,
				"Article Title",
				"12345",
				`"https://example.org"`,
				"Another Title",
				"67890",
			},
		},
		{
			name: "object array with special characters",
			input: []interface{}{
				map[string]interface{}{
					"url":    "https://example.com",
					"title":  "Title, with, commas",
					"length": 12345,
				},
			},
			contains: []string{
				"results[1]{",
				`"https://example.com"`,
				"\"Title, with, commas\"",
				"12345",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			got := string(data)
			got = trimRightNewline(got)

			// Check that all required strings are present
			for _, expected := range tt.contains {
				if !strings.Contains(got, expected) {
					t.Errorf("output does not contain %q\nGot: %q", expected, got)
				}
			}

			// Verify array count matches
			if !strings.Contains(got, fmt.Sprintf("results[%d]", len(tt.input))) {
				t.Errorf("array count mismatch, expected results[%d]", len(tt.input))
			}
		})
	}
}

func TestMarshal(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "string",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "number",
			input:    42,
			expected: "42",
		},
		{
			name:     "bool",
			input:    true,
			expected: "true",
		},
		{
			name:     "nil",
			input:    nil,
			expected: "null",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := Marshal(tt.input)
			if err != nil {
				t.Fatalf("marshal error: %v", err)
			}

			got := string(data)
			got = trimRightNewline(got)

			if got != tt.expected {
				t.Errorf("got %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestLineEndings(t *testing.T) {
	// TOON spec requires LF (\n) only, never CRLF (\r\n)
	input := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
	}

	data, err := Marshal(input)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Check for CRLF
	if bytes.Contains(data, []byte("\r\n")) {
		t.Error("output contains CRLF line endings, should be LF only")
	}

	// Verify line endings are LF
	if !bytes.Contains(data, []byte("\n")) {
		t.Error("output should contain LF line endings")
	}
}

func TestNoTrailingNewline(t *testing.T) {
	input := map[string]interface{}{
		"key": "value",
	}

	data, err := Marshal(input)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	// Trim trailing newlines and check if anything remains
	trimmed := bytes.TrimRight(data, "\n")
	if len(trimmed) < len(data) {
		t.Error("output has trailing newline, should have none")
	}
}

// Helper functions

func trimRightNewline(s string) string {
	return strings.TrimRight(s, "\n")
}
