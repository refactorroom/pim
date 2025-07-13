package pim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// JsonOptions configures the JSON output format
type JsonOptions struct {
	Colored bool // Whether to use colors in the output
	Indent  int  // Number of spaces for indentation
	Raw     bool // Whether to print raw (unformatted) JSON
}

// DefaultJsonOptions provides default formatting options
var DefaultJsonOptions = JsonOptions{
	Colored: true,
	Indent:  2,
	Raw:     false,
}

// Json prints any value as beautifully formatted JSON
func Json(value interface{}, opts ...JsonOptions) error {
	options := DefaultJsonOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// Marshal the value to JSON
	var jsonData []byte
	var err error

	if options.Raw {
		jsonData, err = json.Marshal(value)
	} else if options.Indent > 0 {
		jsonData, err = json.MarshalIndent(value, "", strings.Repeat(" ", options.Indent))
	} else {
		jsonData, err = json.Marshal(value)
	}

	if err != nil {
		Error("Failed to marshal JSON", err)
		return err
	}

	// Format the JSON with colors if enabled
	var output string
	if options.Colored && !options.Raw {
		output = colorizeJson(jsonData)
	} else {
		output = string(jsonData)
	}

	// Log using the existing system
	LogWithTimestamp(JsonPrefix, output, InfoLevel)
	return nil
}

// colorizeJson adds ANSI color codes to format the JSON string
func colorizeJson(data []byte) string {
	var buf bytes.Buffer
	var inKey bool

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()

	for {
		token, err := dec.Token()
		if err != nil {
			break
		}

		switch v := token.(type) {
		case json.Delim:
			switch v {
			case '{', '[':
				buf.WriteString(string(v) + "\n")
			case '}', ']':
				buf.WriteString("\n" + string(v))
			case ':':
				buf.WriteString(string(v) + " ")
				inKey = false
			}
		case string:
			if inKey {
				buf.WriteString(ColorKey + fmt.Sprintf("%q", v) + ColorReset)
			} else {
				buf.WriteString(ColorString + fmt.Sprintf("%q", v) + ColorReset)
			}
			if dec.More() {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		case json.Number:
			buf.WriteString(ColorNumber + v.String() + ColorReset)
			if dec.More() {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		case bool:
			buf.WriteString(ColorBool + fmt.Sprintf("%v", v) + ColorReset)
			if dec.More() {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		case nil:
			buf.WriteString(ColorNull + "null" + ColorReset)
			if dec.More() {
				buf.WriteString(",")
			}
			buf.WriteString("\n")
		}

		// Track if next string will be a key
		if v, ok := token.(json.Delim); ok && v == '{' {
			inKey = true
		}
	}

	return buf.String()
}
