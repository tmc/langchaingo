package prompts

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	"golang.org/x/exp/maps"
)

// ErrInvalidTemplateFormat is the error when the template format is invalid and
// not supported.
var ErrInvalidTemplateFormat = errors.New("invalid template format")

// TemplateFormat is the format of the template, options: go-template, f-string, jinja2.
type TemplateFormat string

const (
	// TemplateFormatGoTemplate is the format for Go Template
	// TemplateFormatGoTemplate.
	TemplateFormatGoTemplate TemplateFormat = "go-template"
	// TemplateFormatFString is the format for f-string
	// TODO: Add support for f-string.
	TemplateFormatFString TemplateFormat = "f-string"
	// TemplateFormatJinja2 is the format for jinja2
	// TODO: Add support for jinja2.
	TemplateFormatJinja2 TemplateFormat = "jinja2"
)

// Interpolator is the function that interpolates the given template with the given values.
type Interpolator func(string, TemplateFormat, map[string]any) (string, error)

// defaultFormatterMapping is the default mapping of TemplateFormat to Interpolator.
var defaultformatterMapping = map[TemplateFormat]Interpolator{ //nolint:gochecknoglobals
	TemplateFormatGoTemplate: interpolateGoTemplate,
}

// interpolateGoTemplate interpolates the given template with the given values by using
// text/template.
func interpolateGoTemplate(tmpl string, _ TemplateFormat, values map[string]any) (string, error) {
	parsedTmpl, err := template.New("template").Parse(tmpl)
	if err != nil {
		return "", err
	}

	sb := new(strings.Builder)
	err = parsedTmpl.Execute(sb, values)
	if err != nil {
		return "", err
	}

	return sb.String(), nil
}

func newInvalidTemplateError(gotTemplateFormat TemplateFormat) error {
	return fmt.Errorf("%w, got: %s, should be one of %s",
		ErrInvalidTemplateFormat,
		gotTemplateFormat,
		maps.Keys(defaultformatterMapping),
	)
}

// CheckValidTemplate checks if the template is valid through checking whether the given
// TemplateFormat is available and whether the template can be rendered.
func CheckValidTemplate(template string, templateFormat TemplateFormat, inputVariables []string) error {
	_, ok := defaultformatterMapping[templateFormat]
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
func RenderTemplate(tmpl string, tmplFormat TemplateFormat, values map[string]any) (string, error) {
	formatter, ok := defaultformatterMapping[tmplFormat]
	if !ok {
		return "", newInvalidTemplateError(tmplFormat)
	}

	return formatter(tmpl, tmplFormat, values)
}
