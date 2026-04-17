# TOON Schema Specification for md-over-here

**Purpose:** Official TOON v3.0 specification requirements for implementation
**Based on:** https://toonformat.dev/reference/spec.html
**Version:** TOON v3.0 (2025-11-24)
**Status:** Working Draft

## Overview

TOON (Token-Oriented Object Notation) is a compact, human-readable encoding of JSON data model optimized for LLM prompts. It provides ~40% token savings over JSON while maintaining lossless serialization.

## Official Specification

- **Specification:** https://toonformat.dev/reference/spec.html
- **GitHub:** https://github.com/toon-format/spec
- **Reference Impl:** https://github.com/toon-format/toon (TypeScript)
- **NPM Package:** `@toon-format/toon`
- **Test Suite:** https://github.com/toon-format/spec/tree/main/tests

## Media Type & File Extension

```http
Content-Type: text/toon
```

- **Media Type:** `text/toon` (provisional, pending IANA registration)
- **File Extension:** `.toon`
- **Charset:** UTF-8 only
- **Line Endings:** LF (`\n`) only, never CRLF

## TOON Syntax Structure

### Basic Formats

```toon
# Primitive value
status: success

# Array with header
results[3]{url,title}:
  url1,title1
  url2,title2
  url3,title3

# Object with nested fields
metadata:
  count: 42
  totalSize: 12345
  lastUpdated: 2025-04-16T12:00:00Z

# Mixed structure
results[2]{url,title,length}:
  https://example.com,"Example Title",12345
  https://example.org,"Another Title",67890
aggregate[total,success,failed]:
  2,2,0
```

## Critical Schema Requirements

### 1. Indentation Rules

✅ **MUST:**
- Use consistent spaces (default: 2 spaces)
- Use same indentation depth throughout document
- Indent child elements +2 spaces from parent

❌ **MUST NOT:**
- Use tabs for indentation
- Mix spaces and tabs
- Use inconsistent indentation

```go
// ✅ Correct
results[1]{url}:
  https://example.com

// ❌ Wrong (tabs)
results[1]{url}:
	https://example.com
```

### 2. Line Ending Rules

✅ **MUST:**
- Use LF (`\n`) line endings only
- No trailing newline at end of document

❌ **MUST NOT:**
- Use CRLF (`\r\n`)
- Add trailing newline

```go
// ✅ Correct
results[1]{url}:
  https://example.com
// EOF (no trailing newline)

// ❌ Wrong
results[1]{url}:
  https://example.com\n
// (has trailing newline)
```

### 3. Array Header Format

```toon
key[N<delim?>]{fields}:
```

- `N`: Exact count of array elements (MUST match)
- `delim?`: Optional delimiter (`,` or `|` or `\t`)
- `{fields}`: Optional field names for object arrays

```go
// ✅ Correct
results[3]{url,title}:
  url1,title1
  url2,title2
  url3,title3

// ❌ Wrong (count mismatch)
results[3]{url,title}:
  url1,title1
  url2,title2
// Missing 3rd element!
```

### 4. String Quoting Rules

**Quote strings that contain:**
- Active delimiter (`,` or `|` or `\t`)
- Colon (`:`)
- Structural characters (`[`, `]`, `{`, `}`, `-`)
- Leading/trailing whitespace
- Special escape sequences

```go
// ✅ Correct quoting
title: "Article with: colon"
url: "https://example.com/path?query=value&foo=bar"
tags: tag1,tag2,"tag,with,commas"
description: "Line 1\nLine 2"

// ❌ Wrong (unquoted when should be quoted)
title: Article with: colon
tags: tag,with,commas
```

### 5. Escape Sequences

**Only these 5 escape sequences are valid:**

| Escape | Meaning | ASCII |
|--------|---------|-------|
| `\\` | Backslash | 92 |
| `\"` | Double quote | 34 |
| `\n` | Newline | 10 |
| `\r` | Carriage return | 13 |
| `\t` | Tab | 9 |

❌ **All other escapes are INVALID** (no `\uXXXX`, `\b`, `\f`, etc.)

```go
// ✅ Correct escapes
text: "Line 1\nLine 2"
path: "C:\\Users\\name"
quote: "He said \"hello\""

// ❌ Wrong (invalid escapes)
text: "Line 1\u000ALine 2"  // Use \n instead
path: "C:\Users\name"       // Use \\ instead
```

### 6. Number Formatting

**Canonical number rules:**
- No exponential notation (`1.5e10`)
- No leading zeros (`0123` → `123`)
- No trailing decimal point (`1.` → `1.0`)
- No `+` prefix (`+123` → `123`)
- `-0` → `0`
- `NaN`, `Infinity` → `null`

```go
// ✅ Correct numbers
count: 42
price: 19.99
negative: -123
zero: 0

// ❌ Wrong (non-canonical)
count: 042          // Leading zero
price: 1.99e2       // Exponential
negative: +123      // Plus sign
special: NaN        // Use null instead
```

### 7. Array Types

#### Primitive Arrays (Inline)
```toon
tags[3]:
  tag1,tag2,tab3
```

#### Object Arrays (Tabular)
```toon
results[2]{url,title,length}:
  https://example.com,"Title 1",12345
  https://example.org,"Title 2",67890
```

#### Mixed Arrays (List)
```toon
items[3]:
  - primitive value
  - key: nested object
    field: value
  - another primitive
```

#### Nested Arrays
```toon
matrix[2][3]:
  1,2,3
  4,5,6
```

### 8. Object Encoding

```toon
# Flat object
metadata:
  count: 42
  total: 12345

# Nested object
result:
  url: https://example.com
  metadata:
    title: "Example"
    date: 2025-04-16
```

**Rules:**
- Key order MUST be preserved
- Keys use same quoting rules as strings
- Colon (`:`) separates key and value
- Indentation shows nesting

## Delimiter Handling

### Active Delimiter

The delimiter specified in array header applies to that scope:

```toon
# Comma is active delimiter
results[2]{url,title}:
  url1,title1
  url2,title2

# Pipe delimiter in different scope
tags[2]|
  tag1|with|commas
  tag2|also|complex
```

### Delimiter Scoping

- Document-level delimiter: First delimiter in document
- Active delimiter: Current array's delimiter
- Quoting overrides delimiter in scope

## Strict Mode Validation

### Required Strict Checks

1. **Array Count Mismatch**
   ```toon
   # ❌ Error: Declared 3, have 2
   results[3]{url}:
     url1
     url2
   ```

2. **Indentation Errors**
   ```toon
   # ❌ Error: Inconsistent indentation
   results[1]{url}:
     https://example.com
      wrong-indent: true
   ```

3. **Invalid Escape Sequences**
   ```toon
   # ❌ Error: \u not allowed
   text: "Unicode\u0041"
   ```

4. **Delimiter Mismatch**
   ```toon
   # ❌ Error: Header has comma, row uses pipe
   results[2]{url,title}:
     url1|title1
     url2,title2
   ```

## Implementation Examples

### Go Struct to TOON

```go
type Result struct {
    URL     string `toon:"url"`
    Title   string `toon:"title"`
    Length  int    `toon:"length"`
    Cached  bool   `toon:"cached"`
}

type Output struct {
    Results []Result `toon:"results"`
    Total   int      `toon:"total"`
}

// Encodes to:
results[2]{url,title,length,cached}:
  https://example.com,"Example Title",12345,true
  https://example.org,"Another Title",67890,false
total:
  2
```

### Error Handling

```go
// ✅ Correct error structure
errors[1]{code,message,url}:
  fetch_failed,"HTTP 404: Not Found",https://invalid.com

// Exit code 2 for network errors
```

## Conformance Checklist

### Encoder Requirements

- [ ] UTF-8 encoding with LF line endings
- [ ] Consistent 2-space indentation
- [ ] Only valid escape sequences
- [ ] Proper string quoting
- [ ] Array counts match actual
- [ ] Preserve key order
- [ ] Canonical number formatting
- [ ] No trailing spaces/newlines

### Decoder Requirements

- [ ] Parse array headers correctly
- [ ] Split using active delimiter
- [ ] Unescape only valid sequences
- [ ] Type unquoted primitives
- [ ] Enforce strict mode rules
- [ ] Preserve array/key order

### Test Suite Validation

Use official test suite: https://github.com/toon-format/spec/tree/main/tests

```bash
# Clone test fixtures
git clone https://github.com/toon-format/spec.git
cd spec/tests

# Run against your implementation
# - tests/fixtures/encode/ - JSON → TOON tests
# - tests/fixtures/decode/ - TOON → JSON tests
# - tests/fixtures/errors/ - Error case tests
```

## Performance Considerations

### Token Efficiency
- ~40% token savings vs JSON
- Minimal whitespace
- No unnecessary quotes
- Compact array notation

### Memory Efficiency
- Streaming-friendly (line-oriented)
- No deep nesting required
- Efficient parsing (single-pass)

## Migration from JSON

### JSON → TOON Examples

```json
{
  "results": [
    {"url": "https://example.com", "title": "Example", "length": 12345}
  ],
  "total": 1
}
```

```toon
results[1]{url,title,length}:
  https://example.com,Example,12345
total:
  1
```

## References

- **Full Spec:** https://toonformat.dev/reference/spec.html
- **GitHub:** https://github.com/toon-format/spec
- **Examples:** https://github.com/toon-format/spec/tree/main/examples
- **Test Suite:** https://github.com/toon-format/spec/tree/main/tests
- **Reference Impl:** https://github.com/johannschopplich/toon

## Implementation Notes for md-over-here

### Recommended Go Package

Create `internal/toon/` package with:

1. **Encoder** (`encoder.go`)
   - `Encode(v interface{}) (string, error)`
   - `EncodeIndent(v interface{}, prefix string) (string, error)`
   - Field tagging support

2. **Decoder** (`decoder.go`)
   - `Decode(s string) (interface{}, error)`
   - Strict mode validation
   - Delimiter handling

3. **Validator** (`validator.go`)
   - `Validate(s string) error`
   - Conformance checking
   - Detailed error reporting

4. **Tests** (`toon_test.go`)
   - Official test suite integration
   - Edge case coverage
   - Performance benchmarks

### Integration Points

```go
// In processor/output.go
func (r *Result) ToTOON() (string, error) {
    return toon.Encode(r)
}

// In main.go flag handling
if formatFlag == "toon" {
    output, err := result.ToTOON()
    // ...
}
```

---

**Last Updated:** 2025-04-16
**Spec Version:** v3.0
**Conformance:** Target full encoder/decoder compliance
