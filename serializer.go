package pim

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// JsonSerializer provides advanced JSON serialization capabilities
type JsonSerializer struct {
	options SerializerOptions
}

// SerializerOptions configures the JSON serializer behavior
type SerializerOptions struct {
	// Basic formatting options
	Indent      int  // Number of spaces for indentation
	PrettyPrint bool // Whether to pretty print the JSON
	OmitEmpty   bool // Whether to omit empty fields
	OmitZero    bool // Whether to omit zero values
	UseNumber   bool // Use json.Number for numbers instead of float64

	// Field mapping and transformation
	FieldMappings     map[string]string           // Map struct field names to JSON field names
	FieldTransformers map[string]FieldTransformer // Custom field transformers
	FieldValidators   map[string]FieldValidator   // Custom field validators

	// Advanced options
	IncludeUnexported bool   // Whether to include unexported fields
	TagName           string // JSON tag name (default: "json")
	TimeFormat        string // Time format for time.Time fields
	NullEmptyStrings  bool   // Convert empty strings to null
	EscapeHTML        bool   // Escape HTML characters

	// Output options
	Colored bool // Whether to use colors in output
	Raw     bool // Whether to print raw (unformatted) JSON
}

// FieldTransformer is a function that transforms a field value before serialization
type FieldTransformer func(interface{}) (interface{}, error)

// FieldValidator is a function that validates a field value
type FieldValidator func(interface{}) error

// DefaultSerializerOptions provides sensible defaults
var DefaultSerializerOptions = SerializerOptions{
	Indent:            2,
	PrettyPrint:       true,
	OmitEmpty:         true,
	OmitZero:          false,
	UseNumber:         false,
	FieldMappings:     make(map[string]string),
	FieldTransformers: make(map[string]FieldTransformer),
	FieldValidators:   make(map[string]FieldValidator),
	IncludeUnexported: false,
	TagName:           "json",
	TimeFormat:        time.RFC3339,
	NullEmptyStrings:  false,
	EscapeHTML:        true,
	Colored:           true,
	Raw:               false,
}

// NewJsonSerializer creates a new JSON serializer with the given options
func NewJsonSerializer(opts ...SerializerOptions) *JsonSerializer {
	options := DefaultSerializerOptions
	if len(opts) > 0 {
		options = opts[0]
	}

	// Initialize maps if not provided
	if options.FieldMappings == nil {
		options.FieldMappings = make(map[string]string)
	}
	if options.FieldTransformers == nil {
		options.FieldTransformers = make(map[string]FieldTransformer)
	}
	if options.FieldValidators == nil {
		options.FieldValidators = make(map[string]FieldValidator)
	}

	return &JsonSerializer{options: options}
}

// Marshal serializes a value to JSON with advanced options
func (js *JsonSerializer) Marshal(value interface{}) ([]byte, error) {
	// Apply field transformations and validations
	transformedValue, err := js.transformValue(value)
	if err != nil {
		return nil, fmt.Errorf("transformation error: %w", err)
	}

	// Use standard json.Marshal with custom options
	var jsonData []byte
	var err2 error

	if js.options.Indent > 0 && js.options.PrettyPrint {
		jsonData, err2 = json.MarshalIndent(transformedValue, "", strings.Repeat(" ", js.options.Indent))
	} else {
		jsonData, err2 = json.Marshal(transformedValue)
	}

	if err2 != nil {
		return nil, fmt.Errorf("encoding error: %w", err2)
	}

	return jsonData, nil
}

// MarshalToString serializes a value to a JSON string
func (js *JsonSerializer) MarshalToString(value interface{}) (string, error) {
	data, err := js.Marshal(value)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// Unmarshal deserializes JSON data with advanced options
func (js *JsonSerializer) Unmarshal(data []byte, target interface{}) error {
	decoder := json.NewDecoder(bytes.NewReader(data))

	if js.options.UseNumber {
		decoder.UseNumber()
	}

	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("decoding error: %w", err)
	}

	// Apply reverse transformations if needed
	return js.reverseTransformValue(target)
}

// UnmarshalFromString deserializes a JSON string
func (js *JsonSerializer) UnmarshalFromString(jsonStr string, target interface{}) error {
	return js.Unmarshal([]byte(jsonStr), target)
}

// transformValue applies field transformations and validations
func (js *JsonSerializer) transformValue(value interface{}) (interface{}, error) {
	v := reflect.ValueOf(value)

	// Handle nil values
	if !v.IsValid() || v.IsNil() {
		return value, nil
	}

	// Handle different types
	switch v.Kind() {
	case reflect.Struct:
		return js.transformStruct(v)
	case reflect.Map:
		return js.transformMap(v)
	case reflect.Slice, reflect.Array:
		return js.transformSlice(v)
	case reflect.Ptr:
		if v.Elem().IsValid() {
			return js.transformValue(v.Elem().Interface())
		}
		return value, nil
	default:
		return value, nil
	}
}

// transformStruct handles struct transformation
func (js *JsonSerializer) transformStruct(v reflect.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// Skip unexported fields unless explicitly requested
		if !field.IsExported() && !js.options.IncludeUnexported {
			continue
		}

		// Get field name from JSON tag or use field name
		fieldName := js.getFieldName(field)

		// Check if field should be omitted
		if js.shouldOmitField(fieldValue, field) {
			continue
		}

		// Apply field transformation if exists
		transformedValue := fieldValue.Interface()
		if transformer, exists := js.options.FieldTransformers[fieldName]; exists {
			if transformed, err := transformer(transformedValue); err != nil {
				return nil, fmt.Errorf("field %s transformation error: %w", fieldName, err)
			} else {
				transformedValue = transformed
			}
		}

		// Apply field validation if exists
		if validator, exists := js.options.FieldValidators[fieldName]; exists {
			if err := validator(transformedValue); err != nil {
				return nil, fmt.Errorf("field %s validation error: %w", fieldName, err)
			}
		}

		// Handle special types
		transformedValue = js.handleSpecialTypes(transformedValue)

		result[fieldName] = transformedValue
	}

	return result, nil
}

// transformMap handles map transformation
func (js *JsonSerializer) transformMap(v reflect.Value) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for _, key := range v.MapKeys() {
		keyStr := fmt.Sprint(key.Interface())
		value := v.MapIndex(key).Interface()

		// Apply transformations recursively
		if transformed, err := js.transformValue(value); err != nil {
			return nil, fmt.Errorf("map value transformation error: %w", err)
		} else {
			result[keyStr] = transformed
		}
	}

	return result, nil
}

// transformSlice handles slice/array transformation
func (js *JsonSerializer) transformSlice(v reflect.Value) ([]interface{}, error) {
	result := make([]interface{}, v.Len())

	for i := 0; i < v.Len(); i++ {
		value := v.Index(i).Interface()

		// Apply transformations recursively
		if transformed, err := js.transformValue(value); err != nil {
			return nil, fmt.Errorf("slice element transformation error: %w", err)
		} else {
			result[i] = transformed
		}
	}

	return result, nil
}

// getFieldName extracts the JSON field name from struct tags
func (js *JsonSerializer) getFieldName(field reflect.StructField) string {
	// Check for custom field mapping first
	if mappedName, exists := js.options.FieldMappings[field.Name]; exists {
		return mappedName
	}

	// Check for JSON tag
	if tag := field.Tag.Get(js.options.TagName); tag != "" {
		parts := strings.Split(tag, ",")
		if parts[0] != "" && parts[0] != "-" {
			return parts[0]
		}
	}

	return field.Name
}

// shouldOmitField determines if a field should be omitted
func (js *JsonSerializer) shouldOmitField(fieldValue reflect.Value, field reflect.StructField) bool {
	// Check for omitempty tag
	if tag := field.Tag.Get(js.options.TagName); strings.Contains(tag, "omitempty") {
		return js.isEmptyValue(fieldValue)
	}

	// Check global omit options
	if js.options.OmitEmpty && js.isEmptyValue(fieldValue) {
		return true
	}

	if js.options.OmitZero && js.isZeroValue(fieldValue) {
		return true
	}

	return false
}

// isEmptyValue checks if a value is considered empty
func (js *JsonSerializer) isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.String:
		return v.Len() == 0
	case reflect.Array, reflect.Slice, reflect.Map:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

// isZeroValue checks if a value is the zero value for its type
func (js *JsonSerializer) isZeroValue(v reflect.Value) bool {
	return reflect.DeepEqual(v.Interface(), reflect.Zero(v.Type()).Interface())
}

// handleSpecialTypes handles special type conversions
func (js *JsonSerializer) handleSpecialTypes(value interface{}) interface{} {
	switch v := value.(type) {
	case time.Time:
		if js.options.TimeFormat != "" {
			return v.Format(js.options.TimeFormat)
		}
		return v
	case string:
		if js.options.NullEmptyStrings && v == "" {
			return nil
		}
		return v
	}
	return value
}

// reverseTransformValue applies reverse transformations after unmarshaling
func (js *JsonSerializer) reverseTransformValue(target interface{}) error {
	// This is a placeholder for reverse transformations
	// In a real implementation, you might want to apply reverse field mappings
	// or other post-processing
	return nil
}

// AddFieldMapping adds a custom field mapping
func (js *JsonSerializer) AddFieldMapping(structField, jsonField string) {
	js.options.FieldMappings[structField] = jsonField
}

// AddFieldTransformer adds a custom field transformer
func (js *JsonSerializer) AddFieldTransformer(fieldName string, transformer FieldTransformer) {
	js.options.FieldTransformers[fieldName] = transformer
}

// AddFieldValidator adds a custom field validator
func (js *JsonSerializer) AddFieldValidator(fieldName string, validator FieldValidator) {
	js.options.FieldValidators[fieldName] = validator
}

// SetTimeFormat sets the time format for time.Time fields
func (js *JsonSerializer) SetTimeFormat(format string) {
	js.options.TimeFormat = format
}

// SetIndent sets the indentation level
func (js *JsonSerializer) SetIndent(indent int) {
	js.options.Indent = indent
}

// SetPrettyPrint enables or disables pretty printing
func (js *JsonSerializer) SetPrettyPrint(pretty bool) {
	js.options.PrettyPrint = pretty
}

// SetOmitEmpty enables or disables omitting empty fields
func (js *JsonSerializer) SetOmitEmpty(omit bool) {
	js.options.OmitEmpty = omit
}

// SetOmitZero enables or disables omitting zero values
func (js *JsonSerializer) SetOmitZero(omit bool) {
	js.options.OmitZero = omit
}

// SetUseNumber enables or disables using json.Number
func (js *JsonSerializer) SetUseNumber(use bool) {
	js.options.UseNumber = use
}

// SetNullEmptyStrings enables or disables converting empty strings to null
func (js *JsonSerializer) SetNullEmptyStrings(nullify bool) {
	js.options.NullEmptyStrings = nullify
}

// SetEscapeHTML enables or disables HTML escaping
func (js *JsonSerializer) SetEscapeHTML(escape bool) {
	js.options.EscapeHTML = escape
}

// SetIncludeUnexported enables or disables including unexported fields
func (js *JsonSerializer) SetIncludeUnexported(include bool) {
	js.options.IncludeUnexported = include
}

// Convenience functions for common transformations

// StringToUpper transforms a string field to uppercase
func StringToUpper(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return strings.ToUpper(str), nil
	}
	return value, nil
}

// StringToLower transforms a string field to lowercase
func StringToLower(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return strings.ToLower(str), nil
	}
	return value, nil
}

// StringTrim transforms a string field by trimming whitespace
func StringTrim(value interface{}) (interface{}, error) {
	if str, ok := value.(string); ok {
		return strings.TrimSpace(str), nil
	}
	return value, nil
}

// StringReplace transforms a string field by replacing substrings
func StringReplace(old, new string) FieldTransformer {
	return func(value interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			return strings.ReplaceAll(str, old, new), nil
		}
		return value, nil
	}
}

// StringRegexReplace transforms a string field using regex replacement
func StringRegexReplace(pattern, replacement string) FieldTransformer {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return func(value interface{}) (interface{}, error) {
			return nil, fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	return func(value interface{}) (interface{}, error) {
		if str, ok := value.(string); ok {
			return re.ReplaceAllString(str, replacement), nil
		}
		return value, nil
	}
}

// NumberFormat formats a number with specified precision
func NumberFormat(precision int) FieldTransformer {
	return func(value interface{}) (interface{}, error) {
		switch v := value.(type) {
		case float64:
			return strconv.FormatFloat(v, 'f', precision, 64), nil
		case float32:
			return strconv.FormatFloat(float64(v), 'f', precision, 32), nil
		}
		return value, nil
	}
}

// TimeFormat formats a time.Time with specified format
func TimeFormat(format string) FieldTransformer {
	return func(value interface{}) (interface{}, error) {
		if t, ok := value.(time.Time); ok {
			return t.Format(format), nil
		}
		return value, nil
	}
}

// Convenience functions for common validations

// StringNotEmpty validates that a string is not empty
func StringNotEmpty(value interface{}) error {
	if str, ok := value.(string); ok && str == "" {
		return fmt.Errorf("string cannot be empty")
	}
	return nil
}

// StringMinLength validates minimum string length
func StringMinLength(min int) FieldValidator {
	return func(value interface{}) error {
		if str, ok := value.(string); ok && len(str) < min {
			return fmt.Errorf("string length must be at least %d characters", min)
		}
		return nil
	}
}

// StringMaxLength validates maximum string length
func StringMaxLength(max int) FieldValidator {
	return func(value interface{}) error {
		if str, ok := value.(string); ok && len(str) > max {
			return fmt.Errorf("string length must be at most %d characters", max)
		}
		return nil
	}
}

// StringPattern validates string against regex pattern
func StringPattern(pattern string) FieldValidator {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return func(value interface{}) error {
			return fmt.Errorf("invalid regex pattern: %w", err)
		}
	}

	return func(value interface{}) error {
		if str, ok := value.(string); ok && !re.MatchString(str) {
			return fmt.Errorf("string does not match pattern %s", pattern)
		}
		return nil
	}
}

// NumberMin validates minimum number value
func NumberMin(min float64) FieldValidator {
	return func(value interface{}) error {
		switch v := value.(type) {
		case float64:
			if v < min {
				return fmt.Errorf("number must be at least %f", min)
			}
		case float32:
			if float64(v) < min {
				return fmt.Errorf("number must be at least %f", min)
			}
		case int, int8, int16, int32, int64:
			if float64(reflect.ValueOf(v).Int()) < min {
				return fmt.Errorf("number must be at least %f", min)
			}
		case uint, uint8, uint16, uint32, uint64:
			if float64(reflect.ValueOf(v).Uint()) < min {
				return fmt.Errorf("number must be at least %f", min)
			}
		}
		return nil
	}
}

// NumberMax validates maximum number value
func NumberMax(max float64) FieldValidator {
	return func(value interface{}) error {
		switch v := value.(type) {
		case float64:
			if v > max {
				return fmt.Errorf("number must be at most %f", max)
			}
		case float32:
			if float64(v) > max {
				return fmt.Errorf("number must be at most %f", max)
			}
		case int, int8, int16, int32, int64:
			if float64(reflect.ValueOf(v).Int()) > max {
				return fmt.Errorf("number must be at most %f", max)
			}
		case uint, uint8, uint16, uint32, uint64:
			if float64(reflect.ValueOf(v).Uint()) > max {
				return fmt.Errorf("number must be at most %f", max)
			}
		}
		return nil
	}
}

// Required validates that a value is not nil or empty
func Required(value interface{}) error {
	if value == nil {
		return fmt.Errorf("field is required")
	}

	v := reflect.ValueOf(value)
	if v.Kind() == reflect.String && v.Len() == 0 {
		return fmt.Errorf("field is required")
	}

	return nil
}
