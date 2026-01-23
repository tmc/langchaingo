package sanitization

import (
	"html"
	"strings"
	"unicode"
)

// ValidateAndSanitize validates and sanitizes template data internally.
// This is the only function exposed to the templates package.
func ValidateAndSanitize(data map[string]any) (map[string]any, error) {
	sanitized := make(map[string]any)

	for key, value := range data {
		// Validate key names
		if !isValidVariableName(key) {
			return nil, &ValidationError{Key: key, Message: "invalid variable name"}
		}

		// Sanitize value
		sanitized[key] = sanitizeValue(value)
	}

	return sanitized, nil
}

// ValidationError represents a validation failure.
type ValidationError struct {
	Key     string
	Message string
}

func (e *ValidationError) Error() string {
	return "template validation failure: " + e.Message + ": " + e.Key
}

// sanitizeValue recursively sanitizes values.
func sanitizeValue(value any) any {
	switch v := value.(type) {
	case string:
		return sanitizeString(v)
	case []string:
		sanitized := make([]string, len(v))
		for i, s := range v {
			sanitized[i] = sanitizeString(s)
		}
		return sanitized
	case []any:
		sanitized := make([]any, len(v))
		for i, item := range v {
			sanitized[i] = sanitizeValue(item)
		}
		return sanitized
	case map[string]any:
		result, _ := ValidateAndSanitize(v)
		return result
	default:
		// For other types (numbers, bools, etc.), return as-is
		return value
	}
}

// sanitizeString applies string-specific sanitization.
func sanitizeString(s string) string {
	// HTML escape for safety
	return html.EscapeString(s)
}

// isValidVariableName checks if a variable name is safe to use in templates.
func isValidVariableName(name string) bool {
	if name == "" {
		return false
	}

	// Reject names with null bytes
	if strings.Contains(name, "\x00") {
		return false
	}

	// Support dotted notation by validating each part
	parts := strings.Split(name, ".")
	for _, part := range parts {
		if !isValidIdentifier(part) {
			return false
		}
	}

	return true
}

// isValidIdentifier checks if a single identifier is valid.
// We only enforce minimal validation - the template engine handles its own syntax.
func isValidIdentifier(name string) bool {
	if name == "" {
		return false
	}

	// Must start with letter or underscore
	if !unicode.IsLetter(rune(name[0])) && name[0] != '_' {
		return false
	}

	// Rest must be alphanumeric or underscore
	for _, r := range name[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
			return false
		}
	}

	return true
}

