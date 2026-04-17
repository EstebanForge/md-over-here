package errors

import (
	"fmt"
)

// Exit codes
const (
	ExitSuccess         = 0
	ExitGenericError    = 1
	ExitNetworkError    = 2
	ExitExtractionError = 3
	ExitCacheError      = 4
)

// ErrorType represents the category of error
type ErrorType string

const (
	ErrorTypeGeneric    ErrorType = "generic_error"
	ErrorTypeNetwork    ErrorType = "fetch_failed"
	ErrorTypeExtraction ErrorType = "extraction_failed"
	ErrorTypeCache      ErrorType = "cache_error"
)

// StructuredError represents an error in a structured format
type StructuredError struct {
	Code    ErrorType `json:"code" toon:"code"`
	Message string    `json:"message" toon:"message"`
	URL     string    `json:"url,omitempty" toon:"url,omitempty"`
	Context string    `json:"context,omitempty" toon:"context,omitempty"`
}

// Error returns the error message
func (e *StructuredError) Error() string {
	if e.URL != "" {
		return fmt.Sprintf("%s: %s (%s)", e.Code, e.Message, e.URL)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// ExitCode returns the appropriate exit code for this error type
func (e *StructuredError) ExitCode() int {
	switch e.Code {
	case ErrorTypeNetwork:
		return ExitNetworkError
	case ErrorTypeExtraction:
		return ExitExtractionError
	case ErrorTypeCache:
		return ExitCacheError
	default:
		return ExitGenericError
	}
}

// NewNetworkError creates a new network/fetch error
func NewNetworkError(url string, err error) *StructuredError {
	return &StructuredError{
		Code:    ErrorTypeNetwork,
		Message: fmt.Sprintf("HTTP request failed: %v", err),
		URL:     url,
		Context: "",
	}
}

// NewExtractionError creates a new content extraction error
func NewExtractionError(url string, err error) *StructuredError {
	return &StructuredError{
		Code:    ErrorTypeExtraction,
		Message: fmt.Sprintf("Content extraction failed: %v", err),
		URL:     url,
		Context: "",
	}
}

// NewCacheError creates a new cache error
func NewCacheError(operation string, err error) *StructuredError {
	return &StructuredError{
		Code:    ErrorTypeCache,
		Message: fmt.Sprintf("Cache %s failed: %v", operation, err),
		Context: operation,
	}
}

// NewGenericError creates a generic error
func NewGenericError(message string) *StructuredError {
	return &StructuredError{
		Code:    ErrorTypeGeneric,
		Message: message,
	}
}
