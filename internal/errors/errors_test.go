package errors

import (
	"errors"
	"testing"
)

func TestStructuredError(t *testing.T) {
	tests := []struct {
		name     string
		err      *StructuredError
		wantCode int
	}{
		{
			name:     "network error",
			err:      NewNetworkError("https://example.com", errors.New("connection refused")),
			wantCode: ExitNetworkError,
		},
		{
			name:     "extraction error",
			err:      NewExtractionError("https://example.com", errors.New("no content found")),
			wantCode: ExitExtractionError,
		},
		{
			name:     "cache error",
			err:      NewCacheError("read", errors.New("file not found")),
			wantCode: ExitCacheError,
		},
		{
			name:     "generic error",
			err:      NewGenericError("something went wrong"),
			wantCode: ExitGenericError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.err.ExitCode(); got != tt.wantCode {
				t.Errorf("ExitCode() = %v, want %v", got, tt.wantCode)
			}

			// Test Error() method returns non-empty string
			if errMsg := tt.err.Error(); errMsg == "" {
				t.Error("Error() returned empty string")
			}

			// Test Code is set
			if tt.err.Code == "" {
				t.Error("Code is empty")
			}

			// Test Message is set
			if tt.err.Message == "" {
				t.Error("Message is empty")
			}
		})
	}
}

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		errType  ErrorType
		expected string
	}{
		{
			name:     "generic error type",
			errType:  ErrorTypeGeneric,
			expected: "generic_error",
		},
		{
			name:     "network error type",
			errType:  ErrorTypeNetwork,
			expected: "fetch_failed",
		},
		{
			name:     "extraction error type",
			errType:  ErrorTypeExtraction,
			expected: "extraction_failed",
		},
		{
			name:     "cache error type",
			errType:  ErrorTypeCache,
			expected: "cache_error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if string(tt.errType) != tt.expected {
				t.Errorf("ErrorType = %v, want %v", tt.errType, tt.expected)
			}
		})
	}
}

func TestExitCodes(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected string
	}{
		{
			name:     "success",
			code:     ExitSuccess,
			expected: "0",
		},
		{
			name:     "generic error",
			code:     ExitGenericError,
			expected: "1",
		},
		{
			name:     "network error",
			code:     ExitNetworkError,
			expected: "2",
		},
		{
			name:     "extraction error",
			code:     ExitExtractionError,
			expected: "3",
		},
		{
			name:     "cache error",
			code:     ExitCacheError,
			expected: "4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.code < 0 || tt.code > 4 {
				t.Errorf("Exit code %d out of range", tt.code)
			}
		})
	}
}
