package prompts

import (
	"errors"
	"fmt"
	"io/fs"
	"sync"

	"github.com/nikolalohinski/gonja"
	"github.com/nikolalohinski/gonja/config"
	"github.com/tmc/langchaingo/prompts/internal/loader"
)

var (
	secureGonjaEnv     *gonja.Environment
	secureGonjaEnvOnce sync.Once
)

// getSecureGonjaEnv returns a gonja environment that disables filesystem access.
// This is a singleton that gets initialized once with secure defaults.
func getSecureGonjaEnv() *gonja.Environment {
	secureGonjaEnvOnce.Do(func() {
		cfg := config.NewConfig()
		nilLoader := &loader.NilFSLoader{}
		secureGonjaEnv = gonja.NewEnvironment(cfg, nilLoader)
	})
	return secureGonjaEnv
}

// interpolateJinja2 interpolates the given template with the given values by using
// jinja2(impl by https://github.com/NikolaLohinski/gonja).
//
// Security: This function uses a secure gonja environment that disables filesystem
// access by default to prevent template injection attacks such as:
// - {% include "/etc/passwd" %}
// - {% extends "/sensitive/file" %}
// - {% import "/system/module" %}
// - {% from "/etc/shadow" import passwords %}
//
// The secure loader blocks all filesystem operations, ensuring templates can only
// perform safe variable interpolation, conditionals, loops, and built-in functions.
// For controlled filesystem access, use RenderTemplateFS with an explicit fs.FS.
func interpolateJinja2(tmpl string, values map[string]any) (string, error) {
	env := getSecureGonjaEnv()
	tpl, err := env.FromString(tmpl)
	if err != nil {
		return "", fmt.Errorf("template parse failure: %w", err)
	}
	result, err := tpl.Execute(values)
	if err != nil {
		return "", fmt.Errorf("template execution failure: %w", err)
	}
	return result, nil
}

// renderJinja2WithFS renders a Jinja2 template from the filesystem with controlled access.
func renderJinja2WithFS(fsys fs.FS, name string, values map[string]any) (string, error) {
	cfg := config.NewConfig()
	fsLoader := loader.NewFSLoader(fsys)
	env := gonja.NewEnvironment(cfg, fsLoader)

	tpl, err := env.GetTemplate(name)
	if err != nil {
		// Check if it's a file not found error
		if errors.Is(err, fs.ErrNotExist) {
			return "", fmt.Errorf("template file %q not found: %w", name, err)
		}
		return "", fmt.Errorf("failed to load template %q: %w", name, err)
	}
	result, err := tpl.Execute(values)
	if err != nil {
		return "", fmt.Errorf("failed to execute template %q: %w", name, err)
	}
	return result, nil
}
