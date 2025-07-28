package prompts

import (
	"errors"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			})
			t.Run("jinja2", func(t *testing.T) {
				t.Parallel()

				actual, err := interpolateJinja2(strings.ReplaceAll(tc.template, "{{ .", "{{ "), tc.templateValues)
				require.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
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
			require.Error(t, err)
			require.EqualError(t, err, tc.errValue)
		})
	}
}

func TestCheckValidTemplate(t *testing.T) {
	t.Parallel()

	t.Run("NoTemplateAvailable", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {test}", "unknown", []string{"test"})
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidTemplateFormat)
		require.EqualError(t, err, "invalid template format, got: unknown, should be one of [f-string go-template jinja2]")
	})

	t.Run("TemplateErrored", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {{{ test }}", TemplateFormatGoTemplate, []string{"test"})
		require.Error(t, err)
		require.EqualError(t, err, "template parse failure: template: template:1: unexpected \"{\" in command")
	})

	t.Run("TemplateValid", func(t *testing.T) {
		t.Parallel()

		err := CheckValidTemplate("Hello, {{ .test }}", TemplateFormatGoTemplate, []string{"test"})
		require.NoError(t, err)
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
		require.NoError(t, err)
		assert.Equal(t, "Hello world", actual)
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
		require.Error(t, err)
		require.ErrorIs(t, err, ErrInvalidTemplateFormat)
	})
}

//nolint:funlen // Comprehensive security tests
func TestJinja2Security(t *testing.T) {
	t.Parallel()

	t.Run("IncludeDisabledByDefault", func(t *testing.T) {
		t.Parallel()

		// This should fail because filesystem access is disabled by default
		_, err := RenderTemplate(
			`{% include "/etc/passwd" %}`,
			TemplateFormatJinja2,
			map[string]any{},
		)
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
			t.Errorf("expected error to contain %q, got %q", "template loading from filesystem disabled for security reasons", err.Error())
		}
	})

	t.Run("ExtendsDisabledByDefault", func(t *testing.T) {
		t.Parallel()

		// This should fail because filesystem access is disabled by default
		_, err := RenderTemplate(
			`{% extends "/etc/passwd" %}`,
			TemplateFormatJinja2,
			map[string]any{},
		)
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
			t.Errorf("expected error to contain %q, got %q", "template loading from filesystem disabled for security reasons", err.Error())
		}
	})

	t.Run("ImportDisabledByDefault", func(t *testing.T) {
		t.Parallel()

		// This should fail because filesystem access is disabled by default
		_, err := RenderTemplate(
			`{% import "/etc/passwd" as p %}`,
			TemplateFormatJinja2,
			map[string]any{},
		)
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
			t.Errorf("expected error to contain %q, got %q", "template loading from filesystem disabled for security reasons", err.Error())
		}
	})

	t.Run("FromDisabledByDefault", func(t *testing.T) {
		t.Parallel()

		// This should fail because filesystem access is disabled by default
		_, err := RenderTemplate(
			`{% from "/etc/passwd" import root %}`,
			TemplateFormatJinja2,
			map[string]any{},
		)
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
			t.Errorf("expected error to contain %q, got %q", "template loading from filesystem disabled for security reasons", err.Error())
		}
	})

	t.Run("SafeTemplatesStillWork", func(t *testing.T) {
		t.Parallel()

		// Normal variable interpolation should still work
		result, err := RenderTemplate(
			`Hello {{ name }}, the time is {{ time }}`,
			TemplateFormatJinja2,
			map[string]any{
				"name": "world",
				"time": "now",
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "Hello world, the time is now" {
			t.Errorf("expected %q, got %q", "Hello world, the time is now", result)
		}
	})

	t.Run("ComplexSafeTemplatesWork", func(t *testing.T) {
		t.Parallel()

		// More complex but safe templates should work
		result, err := RenderTemplate(
			`{% for item in items %}{{ item }} {% endfor %}`,
			TemplateFormatJinja2,
			map[string]any{
				"items": []string{"apple", "banana", "cherry"},
			},
		)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "apple banana cherry " {
			t.Errorf("expected %q, got %q", "apple banana cherry ", result)
		}
	})
}

//nolint:funlen // Comprehensive filesystem template tests
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

//nolint:funlen // Migration patterns need comprehensive examples
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
		require.NoError(t, err)

		// NEW API with explicit filesystem (works the same for inline templates)
		fsys := &fstest.MapFS{
			"template.j2": &fstest.MapFile{
				Data: []byte(template),
			},
		}
		result2, err := RenderTemplateFS(fsys, "template.j2", TemplateFormatJinja2, data)
		require.NoError(t, err)

		// Both should produce the same result
		expected := "Hello Alice! Your score is 92%."
		assert.Equal(t, expected, result1)
		assert.Equal(t, expected, result2)
	})

	t.Run("MigrationFromVulnerableToSecure", func(t *testing.T) {
		t.Parallel()

		// Demonstrate that the OLD vulnerable pattern no longer works
		t.Run("OldVulnerablePatternBlocked", func(t *testing.T) {
			t.Parallel()

			// This would be the old vulnerable way - trying to include system files
			vulnerableTemplate := `User info: {% include "/etc/passwd" %}`

			// This should now be blocked by the secure loader
			_, err := RenderTemplate(vulnerableTemplate, TemplateFormatJinja2, map[string]any{})
			require.Error(t, err)
			assert.Contains(t, err.Error(), "template loading from filesystem disabled for security reasons")
		})

		t.Run("NewSecurePatternWorks", func(t *testing.T) {
			t.Parallel()

			// NEW secure way - controlled filesystem access with explicit fs.FS
			fsys := &fstest.MapFS{
				"main.j2": &fstest.MapFile{
					Data: []byte("User: {{ username }}\n{% include 'user_info.j2' %}"),
				},
				"user_info.j2": &fstest.MapFile{
					Data: []byte("Role: {{ role }}\nDepartment: {{ department }}"),
				},
			}

			result, err := RenderTemplateFS(fsys, "main.j2", TemplateFormatJinja2, map[string]any{
				"username":   "alice",
				"role":       "developer",
				"department": "engineering",
			})
			require.NoError(t, err)
			expected := "User: alice\nRole: developer\nDepartment: engineering"
			assert.Equal(t, expected, result)
		})
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
		require.NoError(t, err)

		// Verify the template inheritance worked correctly
		assert.Contains(t, result, "<title>Alice Johnson's Profile</title>")
		assert.Contains(t, result, "<h1>Welcome, Alice Johnson!</h1>")
		assert.Contains(t, result, "Email: alice@example.com")
		assert.Contains(t, result, "Role: Senior Developer")
		assert.Contains(t, result, "<!DOCTYPE html>")
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
		require.NoError(t, err)

		assert.Contains(t, result, "Subject: Welcome to Acme Corp!")
		assert.Contains(t, result, "Dear Bob Smith,")
		assert.Contains(t, result, "Username: bsmith")
		assert.Contains(t, result, "The Acme Corp Team")
	})
}
