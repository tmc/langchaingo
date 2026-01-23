package prompts

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
)

// renderGoTemplateWithFS renders a Go template from the filesystem.
func renderGoTemplateWithFS(fsys fs.FS, name string, values map[string]any) (string, error) {
	tmpl, err := template.New(name).
		Option("missingkey=error").
		Funcs(sprig.TxtFuncMap()).
		ParseFS(fsys, name)
	if err != nil {
		// Check if it's a file not found error
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("template file %q not found: %w", name, err)
		}
		return "", fmt.Errorf("failed to parse template %q: %w", name, err)
	}

	sb := new(strings.Builder)
	err = tmpl.Execute(sb, values)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", name, err)
	}

	return sb.String(), nil
}
