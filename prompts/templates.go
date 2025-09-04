// Package prompts provides template rendering with support for multiple formats.
//
// Template Formats:
// - Jinja2: Full Jinja2 syntax with controlled filesystem access
// - Go Templates: Standard text/template with sprig functions
// - F-Strings: Python-style string formatting
//
// Basic Usage:
//
//	// Simple template rendering
//	result, err := prompts.RenderTemplate(
//	    "Hello {{ name }}!",
//	    prompts.TemplateFormatJinja2,
//	    map[string]any{"name": "World"},
//	)
//
// Template Files:
//
//	// For templates with includes/inheritance, use RenderTemplateFS
//	//go:embed templates/*
//	var templateFS embed.FS
//	result, err := prompts.RenderTemplateFS(
//	    templateFS,
//	    "welcome.j2",
//	    prompts.TemplateFormatJinja2,
//	    data,
//	)
//
// The RenderTemplate function is designed for inline templates and simple use cases.
// For template files that need to include other templates, use RenderTemplateFS with
// an explicit fs.FS parameter to define the template file boundary.
package prompts

import (
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"slices"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/tmc/langchaingo/prompts/internal/fstring"
	sanitization "github.com/tmc/langchaingo/prompts/internal/sanitization"
)

// ErrInvalidTemplateFormat is the error when the template format is invalid and
// not supported.
var ErrInvalidTemplateFormat = errors.New("invalid template format")

// TemplateFormat is the format of the template.
type TemplateFormat string

const (
	// TemplateFormatGoTemplate uses Go's text/template with sprig functions.
	// This is the recommended format for Go applications and is the default
	// format used by [NewPromptTemplate].
	TemplateFormatGoTemplate TemplateFormat = "go-template"
	// TemplateFormatJinja2 uses Jinja2-style templating with filters and inheritance.
	TemplateFormatJinja2 TemplateFormat = "jinja2"
	// TemplateFormatFString uses Python-style f-string variable substitution.
	TemplateFormatFString TemplateFormat = "f-string"
)

// interpolator is the function that interpolates the given template with the given values.
type interpolator func(template string, values map[string]any) (string, error)

// defaultFormatterMapping is the default mapping of TemplateFormat to interpolator.
var defaultFormatterMapping = map[TemplateFormat]interpolator{ //nolint:gochecknoglobals
	TemplateFormatGoTemplate: interpolateGoTemplate,
	TemplateFormatJinja2:     interpolateJinja2,
	TemplateFormatFString:    fstring.Format,
}

// interpolateGoTemplate interpolates the given template with the given values by using
// text/template.
func interpolateGoTemplate(tmpl string, values map[string]any) (string, error) {
	parsedTmpl, err := template.New("template").
		Option("missingkey=error").
		Funcs(sprig.TxtFuncMap()).
		Parse(tmpl)
	if err != nil {
		return "", fmt.Errorf("template parse failure: %w", err)
	}
	sb := new(strings.Builder)
	err = parsedTmpl.Execute(sb, values)
	if err != nil {
		return "", fmt.Errorf("template execution failure: %w", err)
	}
	return sb.String(), nil
}

func newInvalidTemplateError(gotTemplateFormat TemplateFormat) error {
	formats := slices.AppendSeq(make([]TemplateFormat, 0, len(defaultFormatterMapping)), maps.Keys(defaultFormatterMapping))
	slices.Sort(formats)
	return fmt.Errorf("%w, got: %s, should be one of %s",
		ErrInvalidTemplateFormat,
		gotTemplateFormat,
		formats,
	)
}

// CheckValidTemplate checks if the template is valid through checking whether the given
// TemplateFormat is available and whether the template can be rendered.
//
// Note: This function blocks filesystem access for security. Templates using
// include, extends, import, or from statements will fail. Use RenderTemplateFS
// for controlled filesystem access if needed.
func CheckValidTemplate(template string, templateFormat TemplateFormat, inputVariables []string) error {
	_, ok := defaultFormatterMapping[templateFormat]
	if !ok {
		return newInvalidTemplateError(templateFormat)
	}

	dummyInputs := make(map[string]any, len(inputVariables))
	for _, v := range inputVariables {
		dummyInputs[v] = "foo"
	}

	_, err := RenderTemplate(template, templateFormat, dummyInputs)
	return err
}

// RenderTemplate renders the template with the given values.
//
// This function is designed for inline templates and simple use cases.
// It supports variable interpolation, conditionals, loops, and filters.
// For templates that need to include other template files, use RenderTemplateFS
// with an explicit fs.FS parameter to specify the template file source.
//
// Supported features:
// - Variable interpolation: {{ variable }}
// - Conditional logic: {% if condition %}...{% endif %}
// - Loops: {% for item in list %}...{% endfor %}
// - Text filters: {{ text | filter }}
//
// Security: By default, this function renders templates without sanitization.
// To enable HTML escaping for untrusted data, use the WithSanitization() option.
// This function always blocks filesystem access for security - templates using
// include, extends, or import statements will fail. Use RenderTemplateFS for
// controlled filesystem access.
//
// Example:
//
//	result, err := RenderTemplate(
//	    "Hello {{ name }}! Score: {{ score }}%",
//	    TemplateFormatJinja2,
//	    map[string]any{"name": "Alice", "score": 95},
//	)
func RenderTemplate(tmpl string, tmplFormat TemplateFormat, values map[string]any, opts ...RenderOption) (string, error) {
	formatter, ok := defaultFormatterMapping[tmplFormat]
	if !ok {
		return "", newInvalidTemplateError(tmplFormat)
	}

	// Apply options
	cfg := applyOptions(opts)

	// Only sanitize if explicitly requested
	valuesToUse := values
	if cfg.enableSanitization {
		// Validate and sanitize input data when requested
		safeValues, err := sanitization.ValidateAndSanitize(values)
		if err != nil {
			return "", fmt.Errorf("template execution failure: %w", err)
		}
		valuesToUse = safeValues
	}

	return formatter(tmpl, valuesToUse)
}

// RenderTemplateFS renders a template loaded from the provided filesystem.
// This enables templates to use include, extends, and import statements
// to compose templates from multiple files.
//
// The fs.FS parameter defines which files the template can access,
// providing a clean boundary for template file organization.
//
// Supported fs.FS implementations:
// - embed.FS: Embed templates at compile time (recommended for production)
// - os.DirFS: Access a specific directory tree
// - testing/fstest.MapFS: In-memory filesystem for testing
// - Custom fs.FS implementations for specialized needs
//
// Security: Like RenderTemplate, this function optionally validates and sanitizes
// input data to prevent template injection attacks. The fsys parameter acts as a
// security boundary, limiting templates to only access files within that filesystem.
//
// Examples:
//
//	// Production deployment with embedded templates:
//	//go:embed templates/*
//	var templateFS embed.FS
//	result, err := RenderTemplateFS(templateFS, "email/welcome.j2", TemplateFormatJinja2, data)
//
//	// Development with directory access:
//	fsys := os.DirFS("./templates")
//	result, err := RenderTemplateFS(fsys, "report.j2", TemplateFormatJinja2, data)
//
// Template composition features:
// - Jinja2: include, extends, import, from statements
// - Go templates: ParseFS functionality for template inheritance
// - F-strings: File reading from the specified filesystem
func RenderTemplateFS(fsys fs.FS, name string, tmplFormat TemplateFormat, values map[string]any, opts ...RenderOption) (string, error) {
	// Apply options
	cfg := applyOptions(opts)

	// Only sanitize if explicitly requested
	valuesToUse := values
	if cfg.enableSanitization {
		// Validate and sanitize input data when requested
		safeValues, err := sanitization.ValidateAndSanitize(values)
		if err != nil {
			return "", fmt.Errorf("template execution failure: %w", err)
		}
		valuesToUse = safeValues
	}
	switch tmplFormat {
	case TemplateFormatJinja2:
		return renderJinja2WithFS(fsys, name, valuesToUse)
	case TemplateFormatGoTemplate:
		return renderGoTemplateWithFS(fsys, name, valuesToUse)
	case TemplateFormatFString:
		// F-String templates don't support filesystem operations
		content, err := fs.ReadFile(fsys, name)
		if err != nil {
			return "", fmt.Errorf("failed to read template file %q: %w", name, err)
		}
		return fstring.Format(string(content), valuesToUse)
	default:
		return "", newInvalidTemplateError(tmplFormat)
	}
}
