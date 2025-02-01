package console

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// KeyValueOptions configures the key-value output format
type KeyValueOptions struct {
	Colored     bool   // Whether to use colors in the output
	Indent      int    // Number of spaces for indentation
	Separator   string // Separator between key and value (default: ": ")
	BraceStyle  bool   // Whether to use braces { } around the output
	QuoteKeys   bool   // Whether to quote keys
	QuoteValues bool   // Whether to quote string values
}

// DefaultKeyValueOptions provides default formatting options
var DefaultKeyValueOptions = KeyValueOptions{
	Colored:     true,
	Indent:      2,
	Separator:   ": ",
	BraceStyle:  true,
	QuoteKeys:   false,
	QuoteValues: false,
}

// KeyValue prints a map or struct as formatted key-value pairs
func KeyValue(value interface{}, opts ...KeyValueOptions) error {
	options := DefaultKeyValueOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	output := formatKeyValue(value, options)
	LogWithTimestamp(JsonPrefix, output, InfoLevel)
	return nil
}

// KeyValueInline prints key-value pairs in a single line
func KeyValueInline(value interface{}, opts ...KeyValueOptions) error {
	options := DefaultKeyValueOptions
	if len(opts) > 0 {
		options = opts[0]
	}
	options.Indent = 0 // Force single line

	output := formatKeyValue(value, options)
	// Remove newlines for inline format
	output = strings.ReplaceAll(output, "\n", " ")
	output = strings.ReplaceAll(output, "  ", " ")
	LogWithTimestamp(JsonPrefix, output, InfoLevel)
	return nil
}

// formatKeyValue handles the actual formatting of key-value pairs
func formatKeyValue(value interface{}, opts KeyValueOptions) string {
	var buf bytes.Buffer

	if opts.BraceStyle {
		buf.WriteString("{\n")
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.Map {
		formatMap(&buf, v, opts)
	} else if v.Kind() == reflect.Struct {
		formatStruct(&buf, v, opts)
	}

	if opts.BraceStyle {
		if opts.Indent > 0 {
			buf.WriteString("}")
		} else {
			buf.WriteString(" }")
		}
	}

	return buf.String()
}

// formatMap handles formatting for map types
func formatMap(buf *bytes.Buffer, v reflect.Value, opts KeyValueOptions) {
	indent := strings.Repeat(" ", opts.Indent)

	for _, key := range v.MapKeys() {
		val := v.MapIndex(key)

		// Format key
		if opts.Indent > 0 {
			buf.WriteString(indent)
		}

		keyStr := fmt.Sprint(key.Interface())
		if opts.QuoteKeys {
			keyStr = fmt.Sprintf("%q", keyStr)
		}
		if opts.Colored {
			keyStr = ColorKey + keyStr + ColorReset
		}

		buf.WriteString(keyStr)
		buf.WriteString(opts.Separator)

		// Format value
		valStr := formatValue(val.Interface(), opts)
		buf.WriteString(valStr)

		if opts.Indent > 0 {
			buf.WriteString("\n")
		} else {
			buf.WriteString(" ")
		}
	}
}

// formatStruct handles formatting for struct types
func formatStruct(buf *bytes.Buffer, v reflect.Value, opts KeyValueOptions) {
	t := v.Type()
	indent := strings.Repeat(" ", opts.Indent)

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		// Skip unexported fields
		if !field.IsExported() {
			continue
		}

		// Format key (field name)
		if opts.Indent > 0 {
			buf.WriteString(indent)
		}

		keyStr := field.Name
		if opts.QuoteKeys {
			keyStr = fmt.Sprintf("%q", keyStr)
		}
		if opts.Colored {
			keyStr = ColorKey + keyStr + ColorReset
		}

		buf.WriteString(keyStr)
		buf.WriteString(opts.Separator)

		// Format value
		valStr := formatValue(value.Interface(), opts)
		buf.WriteString(valStr)

		if opts.Indent > 0 {
			buf.WriteString("\n")
		} else {
			buf.WriteString(" ")
		}
	}
}

// formatValue formats a single value with appropriate coloring and quoting
func formatValue(v interface{}, opts KeyValueOptions) string {
	var valStr string

	switch val := v.(type) {
	case string:
		if opts.QuoteValues {
			valStr = fmt.Sprintf("%q", val)
		} else {
			valStr = val
		}
		if opts.Colored {
			valStr = ColorString + valStr + ColorReset
		}
	case json.Number:
		valStr = string(val)
		if opts.Colored {
			valStr = ColorNumber + valStr + ColorReset
		}
	case float64, float32, int, int64, int32, uint, uint64, uint32:
		valStr = fmt.Sprint(val)
		if opts.Colored {
			valStr = ColorNumber + valStr + ColorReset
		}
	case bool:
		valStr = fmt.Sprint(val)
		if opts.Colored {
			valStr = ColorBool + valStr + ColorReset
		}
	case nil:
		valStr = "null"
		if opts.Colored {
			valStr = ColorNull + valStr + ColorReset
		}
	default:
		// For complex types, use standard JSON marshaling
		bytes, err := json.Marshal(val)
		if err != nil {
			valStr = fmt.Sprint(val)
		} else {
			valStr = string(bytes)
		}
	}

	return valStr
}
