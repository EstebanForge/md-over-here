package toon

import (
	"bytes"
	"io"
)

// Encoder writes TOON format output to an io.Writer
type Encoder struct {
	w      io.Writer
	indent string
}

// NewEncoder creates a new TOON encoder
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{
		w:      w,
		indent: "  ", // 2 spaces per TOON spec
	}
}

// SetIndent sets the indentation string (default: 2 spaces)
func (e *Encoder) SetIndent(indent string) {
	e.indent = indent
}

// Encode writes the TOON encoding of v to the stream
func (e *Encoder) Encode(v interface{}) error {
	// Implementation will use reflection and type assertion
	return encodeValue(e.w, v, 0, e.indent)
}

// Marshal returns the TOON encoding of v
func Marshal(v interface{}) ([]byte, error) {
	var b builder
	if err := encodeValue(&b, v, 0, "  "); err != nil {
		return nil, err
	}
	// TOON spec: no trailing newline at end of document
	data := b.Bytes()
	data = bytes.TrimRight(data, "\n")
	return data, nil
}

// MarshalIndent returns the TOON encoding of v with indentation
func MarshalIndent(v interface{}, prefix string) ([]byte, error) {
	var b builder
	if err := encodeValue(&b, v, 0, "  "); err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// Unmarshal parses the TOON-encoded data and stores the result
func Unmarshal(data []byte, v interface{}) error {
	// TODO: Implement decoder
	return nil
}

// builder is a simple bytes buffer builder for TOON encoding
type builder struct {
	buf []byte
}

func (b *builder) Write(p []byte) (n int, err error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *builder) WriteString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *builder) WriteByte(c byte) error {
	b.buf = append(b.buf, c)
	return nil
}

func (b *builder) WriteRune(r rune) {
	b.buf = append(b.buf, string(r)...)
}

func (b *builder) Bytes() []byte {
	return b.buf
}

// encodeValue recursively encodes a value to TOON format
func encodeValue(w io.Writer, v interface{}, depth int, indent string) error {
	// Handle different types
	switch val := v.(type) {
	case map[string]interface{}:
		return encodeObject(w, val, depth, indent)
	case []interface{}:
		return encodeArray(w, val, depth, indent)
	case string:
		return encodeString(w, val)
	case int, int8, int16, int32, int64:
		return encodeNumber(w, val)
	case uint, uint8, uint16, uint32, uint64:
		return encodeNumber(w, val)
	case float32, float64:
		return encodeNumber(w, val)
	case bool:
		return encodeBool(w, val)
	case map[string]string:
		// Convert map[string]string to map[string]interface{} for encoding
		obj := make(map[string]interface{}, len(val))
		for k, v := range val {
			obj[k] = v
		}
		return encodeObject(w, obj, depth, indent)
	case nil:
		return encodeNull(w)
	default:
		// For unknown types, treat as string
		return encodeString(w, toString(v))
	}
}
