package prompts

import (
	"errors"
	"fmt"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"golang.org/x/exp/maps"
)

// ErrInvalidTemplateFormat is the error when the template format is invalid and
// not supported.
var ErrInvalidTemplateFormat = errors.New("invalid template format")

// TemplateFormat is the format of the template.
type TemplateFormat string

const (
	// TemplateFormatGoTemplate is the format for go-template.
	TemplateFormatGoTemplate TemplateFormat = "go-template"
)

// interpolator is the function that interpolates the given template with the given values.
type interpolator func(template string, values map[string]any) (string, error)

// defaultFormatterMapping is the default mapping of TemplateFormat to interpolator.
var defaultformatterMapping = map[TemplateFormat]interpolator{ //nolint:gochecknoglobals
	TemplateFormatGoTemplate: interpolateGoTemplate,
}

// interpolateGoTemplate interpolates the given template with the given values by using
// text/template.
func interpolateGoTemplate(tmpl string, values map[string]any) (string, error) {
	parsedTmpl, err := template.New("template").
		Option("missingkey=error").
		Funcs(sprig.FuncMap()).
		Parse(tmpl)
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
	return formatter(tmpl, values)
}
