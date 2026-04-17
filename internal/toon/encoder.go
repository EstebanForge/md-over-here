package toon

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"unicode"
)

// encodeObject encodes a map[string]interface{} to TOON format
func encodeObject(w io.Writer, obj map[string]interface{}, depth int, indent string) error {
	// Encode as key: value pairs
	prefix := strings.Repeat(indent, depth)
	i := 0
	for key, value := range obj {
		// Write key (quoted only if needed)
		var keyStr string
		if needsQuoting(key) {
			keyStr = quoteString(key)
		} else {
			keyStr = key
		}
		if _, err := io.WriteString(w, prefix+keyStr+": "); err != nil {
			return err
		}

		// Encode value
		if err := encodeValue(w, value, 0, indent); err != nil {
			return err
		}

		// Add newline after each entry except the last
		if i < len(obj)-1 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
		i++
	}

	return nil
}

// encodeArray encodes an []interface{} to TOON format
func encodeArray(w io.Writer, arr []interface{}, depth int, indent string) error {
	if len(arr) == 0 {
		if _, err := io.WriteString(w, "0"); err != nil {
			return err
		}
		return nil
	}

	// Check if this is an array of objects (tabular format)
	isObjectArray := true
	for _, item := range arr {
		if _, ok := item.(map[string]interface{}); !ok {
			isObjectArray = false
			break
		}
	}

	if isObjectArray {
		return encodeObjectArray(w, arr, depth, indent)
	}

	// Primitive array or mixed array
	prefix := strings.Repeat(indent, depth)
	if _, err := io.WriteString(w, fmt.Sprintf("%d:\n", len(arr))); err != nil {
		return err
	}

	for i, item := range arr {
		switch v := item.(type) {
		case map[string]interface{}:
			// Nested object in mixed array
			if _, err := io.WriteString(w, prefix+indent+"- key: value\n"); err != nil {
				return err
			}
			if err := encodeObject(w, v, depth+2, indent); err != nil {
				return err
			}
		default:
			// Primitive value
			if _, err := io.WriteString(w, prefix+indent+"- "); err != nil {
				return err
			}
			if err := encodeValue(w, v, 0, indent); err != nil {
				return err
			}
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}

		if i < len(arr)-1 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

// encodeObjectArray encodes an array of objects in tabular format
func encodeObjectArray(w io.Writer, arr []interface{}, depth int, indent string) error {
	if len(arr) == 0 {
		return nil
	}

	// Get field names from first object to preserve order
	firstObj, ok := arr[0].(map[string]interface{})
	if !ok || len(firstObj) == 0 {
		return encodeArray(w, arr, depth, indent)
	}

	// Get field order from first object
	fields := make([]string, 0, len(firstObj))
	for key := range firstObj {
		fields = append(fields, key)
	}

	// Write array header with count and field names
	prefix := strings.Repeat(indent, depth)
	if _, err := io.WriteString(w, prefix); err != nil {
		return err
	}
	if _, err := io.WriteString(w, fmt.Sprintf("results[%d]{", len(arr))); err != nil {
		return err
	}
	for i, field := range fields {
		if i > 0 {
			if _, err := io.WriteString(w, ","); err != nil {
				return err
			}
		}
		if _, err := io.WriteString(w, field); err != nil {
			return err
		}
	}
	if _, err := io.WriteString(w, "}:\n"); err != nil {
		return err
	}

	// Write each row
	dataPrefix := prefix + indent
	for i, item := range arr {
		obj, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		if _, err := io.WriteString(w, dataPrefix); err != nil {
			return err
		}
		for j, field := range fields {
			if j > 0 {
				if _, err := io.WriteString(w, ","); err != nil {
					return err
				}
			}

			value := obj[field]
			if err := encodePrimitive(w, value); err != nil {
				return err
			}
		}
		// Add newline after each row except the last
		if i < len(arr)-1 {
			if _, err := io.WriteString(w, "\n"); err != nil {
				return err
			}
		}
	}

	return nil
}

// encodeString encodes a string value
func encodeString(w io.Writer, s string) error {
	// Check if string needs quoting
	if needsQuoting(s) {
		if _, err := io.WriteString(w, quoteString(s)); err != nil {
			return err
		}
	} else {
		if _, err := io.WriteString(w, s); err != nil {
			return err
		}
	}
	return nil
}

// encodeNumber encodes a numeric value in canonical format
func encodeNumber(w io.Writer, v interface{}) error {
	var s string
	switch val := v.(type) {
	case int:
		s = strconv.FormatInt(int64(val), 10)
	case int8:
		s = strconv.FormatInt(int64(val), 10)
	case int16:
		s = strconv.FormatInt(int64(val), 10)
	case int32:
		s = strconv.FormatInt(int64(val), 10)
	case int64:
		s = strconv.FormatInt(val, 10)
	case uint:
		s = strconv.FormatUint(uint64(val), 10)
	case uint8:
		s = strconv.FormatUint(uint64(val), 10)
	case uint16:
		s = strconv.FormatUint(uint64(val), 10)
	case uint32:
		s = strconv.FormatUint(uint64(val), 10)
	case uint64:
		s = strconv.FormatUint(val, 10)
	case float32:
		s = formatFloat(float64(val))
	case float64:
		s = formatFloat(val)
	default:
		s = toString(v)
	}

	if _, err := io.WriteString(w, s); err != nil {
		return err
	}
	return nil
}

// encodeBool encodes a boolean value
func encodeBool(w io.Writer, b bool) error {
	if b {
		if _, err := io.WriteString(w, "true"); err != nil {
			return err
		}
	} else {
		if _, err := io.WriteString(w, "false"); err != nil {
			return err
		}
	}
	return nil
}

// encodeNull encodes a null value
func encodeNull(w io.Writer) error {
	if _, err := io.WriteString(w, "null"); err != nil {
		return err
	}
	return nil
}

// encodePrimitive encodes a primitive value for array rows
func encodePrimitive(w io.Writer, v interface{}) error {
	switch val := v.(type) {
	case string:
		return encodeString(w, val)
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, float32, float64:
		return encodeNumber(w, val)
	case bool:
		return encodeBool(w, val)
	case nil:
		return encodeNull(w)
	default:
		return encodeString(w, toString(val))
	}
}

// needsQuoting determines if a string needs to be quoted
func needsQuoting(s string) bool {
	if s == "" {
		return true
	}

	// Check for leading/trailing whitespace
	if len(s) > 0 && (unicode.IsSpace(rune(s[0])) || unicode.IsSpace(rune(s[len(s)-1]))) {
		return true
	}

	for _, r := range s {
		// Quote if contains delimiter (comma by default)
		if r == ',' || r == '|' || r == '\t' {
			return true
		}
		// Quote if contains colon
		if r == ':' {
			return true
		}
		// Quote if contains structural characters
		if r == '[' || r == ']' || r == '{' || r == '}' || r == '-' {
			return true
		}
		// Quote if contains special escape sequences
		if r == '\\' || r == '"' || r == '\n' || r == '\r' || r == '\t' {
			return true
		}
	}

	return false
}

// quoteString quotes and escapes a string
func quoteString(s string) string {
	var buf strings.Builder
	buf.WriteByte('"')

	for _, r := range s {
		switch r {
		case '\\':
			buf.WriteString(`\\`)
		case '"':
			buf.WriteString(`\"`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			buf.WriteRune(r)
		}
	}

	buf.WriteByte('"')
	return buf.String()
}

// formatFloat formats a float in canonical form (no exponential, no trailing decimal)
func formatFloat(f float64) string {
	// Handle special cases
	if f != f { // NaN
		return "null"
	}
	if f > 1.7976931348623157e+308 { // Infinity
		return "null"
	}

	// Format without exponential notation
	s := strconv.FormatFloat(f, 'f', -1, 64)

	// Remove trailing decimal point
	s = strings.TrimSuffix(s, ".")

	return s
}

// toString converts any value to string representation
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("%v", v)
}
