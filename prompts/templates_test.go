package prompts

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"
)

//nolint:funlen // TestInterpolateGoTemplate requires comprehensive coverage
func TestInterpolateGoTemplate(t *testing.T) {
	t.Parallel()

	type tests struct {
		name           string
		template       string
		templateValues map[string]any
		expected       string
		errValue       string
	}

	testCases := []tests{
		{
			name:     "Single",
			template: "Hello {{ .key }}",
			templateValues: map[string]any{
				"key": "world",
			},
			expected: "Hello world",
		},
		{
			name:     "Multiple",
			template: "Hello {{ .key1 }} and {{ .key2 }}",
			templateValues: map[string]any{
				"key1": "world",
				"key2": "universe",
			},
			expected: "Hello world and universe",
		},
		{
			name:     "Nested",
			template: "Hello {{ .key1.key2 }}",
			templateValues: map[string]any{
				"key1": map[string]any{
					"key2": "world",
				},
			},
			expected: "Hello world",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Run("go/template", func(t *testing.T) {
				t.Parallel()

				actual, err := interpolateGoTemplate(tc.template, tc.templateValues)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if actual != tc.expected {
					t.Errorf("expected %q, got %q", tc.expected, actual)
				}
			})
			t.Run("jinja2", func(t *testing.T) {
				t.Parallel()

				actual, err := interpolateJinja2(strings.ReplaceAll(tc.template, "{{ .", "{{ "), tc.templateValues)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if actual != tc.expected {
					t.Errorf("expected %q, got %q", tc.expected, actual)
				}
			})
		})
	}

	errTestCases := []tests{
		{
			name:     "ParseErrored",
			template: "Hello {{{ .key1 }}",
			expected: "",
			errValue: "template parse failure: template: template:1: unexpected \"{\" in command",
		},
		{
			name:     "ExecuteErrored",
			template: "Hello {{ .key1 .key2 }}",
			expected: "",
			errValue: "template execution failure: template: template:1:9: executing \"template\" at <.key1>: key1 is not a method but has arguments",
		},
	}

	for _, tc := range errTestCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, err := interpolateGoTemplate(tc.template, map[string]any{})
			if err == nil {
				t.Errorf("expected error, got nil")
			} else if err.Error() != tc.errValue {
				t.Errorf("expected error %q, got %q", tc.errValue, err.Error())
			}
		})
	}
}

func TestCheckValidTemplate(t *testing.T) {
	t.Parallel()

	t.Run("NoTemplateAvailable", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {test}", "unknown", []string{"test"})
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !errors.Is(err, ErrInvalidTemplateFormat) {
			t.Errorf("expected ErrInvalidTemplateFormat, got %v", err)
		} else if err.Error() != "invalid template format, got: unknown, should be one of [f-string go-template jinja2]" {
			t.Errorf("expected specific error message, got %q", err.Error())
		}
	})

	t.Run("TemplateErrored", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {{{ test }}", TemplateFormatGoTemplate, []string{"test"})
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if err.Error() != "template parse failure: template: template:1: unexpected \"{\" in command" {
			t.Errorf("expected specific error message, got %q", err.Error())
		}
	})

	t.Run("TemplateValid", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {{ .test }}", TemplateFormatGoTemplate, []string{"test"})
		if err != nil {
			t.Errorf("unexpected error: %v", err)
		}
	})
}

func TestRenderTemplate(t *testing.T) {
	t.Parallel()

	t.Run("TemplateAvailable", func(t *testing.T) {
		t.Parallel()

		actual, err := RenderTemplate(
			"Hello {{ .key }}",
			TemplateFormatGoTemplate,
			map[string]any{
				"key": "world",
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if actual != "Hello world" {
			t.Errorf("expected %q, got %q", "Hello world", actual)
		}
	})

	t.Run("TemplateNotAvailable", func(t *testing.T) {
		t.Parallel()

		_, err := RenderTemplate(
			"Hello {key}",
			"unknown",
			map[string]any{
				"key": "world",
			},
		)
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !errors.Is(err, ErrInvalidTemplateFormat) {
			t.Errorf("expected ErrInvalidTemplateFormat, got %v", err)
		}
	})
}

//nolint:funlen // TestRenderTemplateFS requires comprehensive filesystem tests
func TestRenderTemplateFS(t *testing.T) {
	t.Parallel()

	t.Run("InvalidTemplateFormat", func(t *testing.T) {
		t.Parallel()

		fsys := &fstest.MapFS{
			"template.txt": &fstest.MapFile{
				Data: []byte("Hello {{ name }}"),
			},
		}

		_, err := RenderTemplateFS(fsys, "template.txt", "unknown", map[string]any{"name": "world"})
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !errors.Is(err, ErrInvalidTemplateFormat) {
			t.Errorf("expected error to be ErrInvalidTemplateFormat, got %v", err)
		}
	})

	t.Run("FileNotFound", func(t *testing.T) {
		t.Parallel()

		fsys := &fstest.MapFS{}

		_, err := RenderTemplateFS(fsys, "nonexistent.txt", TemplateFormatJinja2, map[string]any{})
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "nonexistent.txt") {
			t.Errorf("expected error to contain %q, got %q", "nonexistent.txt", err.Error())
		}
	})

	t.Run("Jinja2WithMapFS", func(t *testing.T) {
		t.Parallel()

		fsys := &fstest.MapFS{
			"main.j2": &fstest.MapFile{
				Data: []byte("Hello {{ name }}! {% include 'greeting.j2' %}"),
			},
			"greeting.j2": &fstest.MapFile{
				Data: []byte("Welcome to {{ company }}."),
			},
		}

		result, err := RenderTemplateFS(fsys, "main.j2", TemplateFormatJinja2, map[string]any{
			"name":    "Alice",
			"company": "Acme Corp",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Hello Alice! Welcome to Acme Corp." {
			t.Errorf("expected %q, got %q", "Hello Alice! Welcome to Acme Corp.", result)
		}
	})

	t.Run("Jinja2WithExtends", func(t *testing.T) {
		t.Parallel()

		fsys := &fstest.MapFS{
			"base.j2": &fstest.MapFile{
				Data: []byte("{% block content %}Default content{% endblock %}"),
			},
			"child.j2": &fstest.MapFile{
				Data: []byte("{% extends 'base.j2' %}{% block content %}Custom content for {{ name }}{% endblock %}"),
			},
		}

		result, err := RenderTemplateFS(fsys, "child.j2", TemplateFormatJinja2, map[string]any{
			"name": "Alice",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Custom content for Alice" {
			t.Errorf("expected %q, got %q", "Custom content for Alice", result)
		}
	})

	t.Run("GoTemplateWithMapFS", func(t *testing.T) {
		t.Parallel()

		fsys := &fstest.MapFS{
			"template.gotmpl": &fstest.MapFile{
				Data: []byte("Hello {{ .name }}! Score: {{ .score }}%"),
			},
		}

		result, err := RenderTemplateFS(fsys, "template.gotmpl", TemplateFormatGoTemplate, map[string]any{
			"name":  "Bob",
			"score": 95,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Hello Bob! Score: 95%" {
			t.Errorf("expected %q, got %q", "Hello Bob! Score: 95%", result)
		}
	})

	t.Run("FStringWithMapFS", func(t *testing.T) {
		t.Parallel()

		fsys := &fstest.MapFS{
			"template.fstring": &fstest.MapFile{
				Data: []byte("Hello {name}! Your score is {score}%."),
			},
		}

		result, err := RenderTemplateFS(fsys, "template.fstring", TemplateFormatFString, map[string]any{
			"name":  "Charlie",
			"score": 88,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Hello Charlie! Your score is 88%." {
			t.Errorf("expected %q, got %q", "Hello Charlie! Your score is 88%.", result)
		}
	})
}

//nolint:funlen // TestMigrationPatterns requires comprehensive migration examples
func TestMigrationPatterns(t *testing.T) {
	t.Parallel()

	t.Run("SafeBasicTemplating", func(t *testing.T) {
		t.Parallel()

		// Simple variable interpolation works the same with both APIs
		template := "Hello {{ name }}! Your score is {{ score }}%."
		data := map[string]any{
			"name":  "Alice",
			"score": 92,
		}

		// OLD API (still works for safe templates)
		result1, err := RenderTemplate(template, TemplateFormatJinja2, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// NEW API with explicit filesystem (works the same for inline templates)
		fsys := &fstest.MapFS{
			"template.j2": &fstest.MapFile{
				Data: []byte(template),
			},
		}
		result2, err := RenderTemplateFS(fsys, "template.j2", TemplateFormatJinja2, data)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Both should produce the same result
		expected := "Hello Alice! Your score is 92%."
		if result1 != expected {
			t.Errorf("result1: expected %q, got %q", expected, result1)
		}
		if result2 != expected {
			t.Errorf("result2: expected %q, got %q", expected, result2)
		}
	})

	t.Run("ComplexTemplateInheritance", func(t *testing.T) {
		t.Parallel()

		// Demonstrate secure template inheritance patterns
		fsys := &fstest.MapFS{
			"layouts/base.j2": &fstest.MapFile{
				Data: []byte(`
<!DOCTYPE html>
<html>
<head>
    <title>{% block title %}Default Title{% endblock %}</title>
</head>
<body>
    {% block content %}{% endblock %}
</body>
</html>`),
			},
			"pages/user_profile.j2": &fstest.MapFile{
				Data: []byte(`
{% extends 'layouts/base.j2' %}
{% block title %}{{ user.name }}'s Profile{% endblock %}
{% block content %}
<h1>Welcome, {{ user.name }}!</h1>
<p>Email: {{ user.email }}</p>
<p>Role: {{ user.role }}</p>
{% endblock %}`),
			},
		}

		result, err := RenderTemplateFS(fsys, "pages/user_profile.j2", TemplateFormatJinja2, map[string]any{
			"user": map[string]any{
				"name":  "Alice Johnson",
				"email": "alice@example.com",
				"role":  "Senior Developer",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		// Verify the template inheritance worked correctly
		if !strings.Contains(result, "<title>Alice Johnson's Profile</title>") {
			t.Errorf("expected title tag in result")
		}
		if !strings.Contains(result, "<h1>Welcome, Alice Johnson!</h1>") {
			t.Errorf("expected welcome header in result")
		}
		if !strings.Contains(result, "Email: alice@example.com") {
			t.Errorf("expected email in result")
		}
		if !strings.Contains(result, "Role: Senior Developer") {
			t.Errorf("expected role in result")
		}
		if !strings.Contains(result, "<!DOCTYPE html>") {
			t.Errorf("expected DOCTYPE in result")
		}
	})

	t.Run("EmbedFSIntegration", func(t *testing.T) {
		t.Parallel()

		// Demonstrate how to migrate to using embed.FS for production use
		// This test shows the pattern without actually embedding files

		// Simulate an embed.FS structure
		fsys := &fstest.MapFS{
			"templates/email/welcome.j2": &fstest.MapFile{
				Data: []byte(`
Subject: Welcome to {{ company }}!

Dear {{ user.name }},

Welcome to {{ company }}! We're excited to have you join our {{ user.department }} team.

Your account details:
- Username: {{ user.username }}
- Email: {{ user.email }}
- Department: {{ user.department }}

Best regards,
The {{ company }} Team`),
			},
		}

		result, err := RenderTemplateFS(fsys, "templates/email/welcome.j2", TemplateFormatJinja2, map[string]any{
			"company": "Acme Corp",
			"user": map[string]any{
				"name":       "Bob Smith",
				"username":   "bsmith",
				"email":      "bob@example.com",
				"department": "Engineering",
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if !strings.Contains(result, "Subject: Welcome to Acme Corp!") {
			t.Errorf("expected 'Subject: Welcome to Acme Corp!' in result")
		}
		if !strings.Contains(result, "Dear Bob Smith,") {
			t.Errorf("expected 'Dear Bob Smith,' in result")
		}
		if !strings.Contains(result, "Username: bsmith") {
			t.Errorf("expected 'Username: bsmith' in result")
		}
		if !strings.Contains(result, "The Acme Corp Team") {
			t.Errorf("expected 'The Acme Corp Team' in result")
		}
	})
}
