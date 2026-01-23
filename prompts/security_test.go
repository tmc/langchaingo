package prompts

import (
	"strings"
	"testing"
	"testing/fstest"
)

// TestJinja2PathTraversalSecurity tests that path traversal attacks are blocked
// by the secure template loader.
//
//nolint:funlen // TestJinja2PathTraversalSecurity requires comprehensive security tests
func TestJinja2PathTraversalSecurity(t *testing.T) {
	t.Parallel()

	t.Run("IncludePathTraversal", func(t *testing.T) {
		t.Parallel()

		// Test various path traversal attempts with include
		testCases := []struct {
			name     string
			template string
		}{
			{"AbsolutePath", `{% include "/etc/passwd" %}`},
			{"RelativePathTraversal", `{% include "../../../etc/passwd" %}`},
			{"ComplexPathTraversal", `{% include "../../../../../../etc/passwd" %}`},
			{"HiddenPathTraversal", `{% include "./../../../etc/passwd" %}`},
			{"WindowsStylePath", `{% include "C:\\Windows\\System32\\config\\SAM" %}`},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := RenderTemplate(
					tc.template,
					TemplateFormatJinja2,
					map[string]any{},
				)
				if err == nil {
					t.Errorf("expected error for template %q, got nil", tc.template)
				} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
					t.Errorf("expected security error, got %q", err.Error())
				}
			})
		}
	})

	t.Run("ExtendsPathTraversal", func(t *testing.T) {
		t.Parallel()

		// Test path traversal with extends
		testCases := []struct {
			name     string
			template string
		}{
			{"AbsolutePath", `{% extends "/etc/passwd" %}`},
			{"RelativePathTraversal", `{% extends "../../../etc/passwd" %}`},
			{"ComplexPathTraversal", `{% extends "../../../../../../etc/passwd" %}`},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := RenderTemplate(
					tc.template,
					TemplateFormatJinja2,
					map[string]any{},
				)
				if err == nil {
					t.Errorf("expected error for template %q, got nil", tc.template)
				} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
					t.Errorf("expected security error, got %q", err.Error())
				}
			})
		}
	})

	t.Run("ImportPathTraversal", func(t *testing.T) {
		t.Parallel()

		// Test path traversal with import
		testCases := []struct {
			name     string
			template string
		}{
			{"AbsolutePath", `{% import "/etc/passwd" as p %}`},
			{"RelativePathTraversal", `{% import "../../../etc/passwd" as p %}`},
			{"ComplexPathTraversal", `{% import "../../../../../../etc/passwd" as p %}`},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := RenderTemplate(
					tc.template,
					TemplateFormatJinja2,
					map[string]any{},
				)
				if err == nil {
					t.Errorf("expected error for template %q, got nil", tc.template)
				} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
					t.Errorf("expected security error, got %q", err.Error())
				}
			})
		}
	})

	t.Run("FromImportPathTraversal", func(t *testing.T) {
		t.Parallel()

		// Test path traversal with from...import
		testCases := []struct {
			name     string
			template string
		}{
			{"AbsolutePath", `{% from "/etc/passwd" import root %}`},
			{"RelativePathTraversal", `{% from "../../../etc/passwd" import root %}`},
			{"ComplexPathTraversal", `{% from "../../../../../../etc/passwd" import root %}`},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				_, err := RenderTemplate(
					tc.template,
					TemplateFormatJinja2,
					map[string]any{},
				)
				if err == nil {
					t.Errorf("expected error for template %q, got nil", tc.template)
				} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
					t.Errorf("expected security error, got %q", err.Error())
				}
			})
		}
	})

	t.Run("VulnerablePatternBlocked", func(t *testing.T) {
		t.Parallel()

		// Test that attempts to access sensitive files are blocked
		vulnerableTemplates := []string{
			`User info: {% include "/etc/passwd" %}`,
			`Config: {% include "/etc/shadow" %}`,
			`Keys: {% include "~/.ssh/id_rsa" %}`,
			`AWS: {% include "~/.aws/credentials" %}`,
		}

		for _, tmpl := range vulnerableTemplates {
			_, err := RenderTemplate(tmpl, TemplateFormatJinja2, map[string]any{})
			if err == nil {
				t.Errorf("expected error for template %q, got nil", tmpl)
				continue
			}
			if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
				t.Errorf("expected security error, got %q", err.Error())
			}
		}
	})
}

// TestSecurityMechanismEffectiveness verifies that our security mechanism
// properly prevents filesystem access while allowing safe operations.
func TestSecurityMechanismEffectiveness(t *testing.T) {
	t.Parallel()

	t.Run("SafeOperationsAllowed", func(t *testing.T) {
		t.Parallel()

		// Ensure normal template operations still work
		safeTemplates := []struct {
			name     string
			template string
			data     map[string]any
			expected string
		}{
			{
				name:     "SimpleVariable",
				template: `Hello {{ name }}!`,
				data:     map[string]any{"name": "World"},
				expected: "Hello World!",
			},
			{
				name:     "ConditionalLogic",
				template: `{% if logged_in %}Welcome back!{% else %}Please login{% endif %}`,
				data:     map[string]any{"logged_in": true},
				expected: "Welcome back!",
			},
			{
				name:     "Loops",
				template: `{% for item in items %}{{ item }} {% endfor %}`,
				data:     map[string]any{"items": []string{"A", "B", "C"}},
				expected: "A B C ",
			},
			{
				name:     "Filters",
				template: `{{ name|upper }}`,
				data:     map[string]any{"name": "test"},
				expected: "TEST",
			},
		}

		for _, tc := range safeTemplates {
			t.Run(tc.name, func(t *testing.T) {
				result, err := RenderTemplate(tc.template, TemplateFormatJinja2, tc.data)
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if result != tc.expected {
					t.Errorf("expected %q, got %q", tc.expected, result)
				}
			})
		}
	})

	t.Run("SecurityBoundary", func(t *testing.T) {
		t.Parallel()

		// Verify that filesystem access is blocked at the boundary
		_, err := RenderTemplate(
			`{% include "local_file.txt" %}`,
			TemplateFormatJinja2,
			map[string]any{},
		)
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
			t.Errorf("expected security error, got %q", err.Error())
		}
	})
}

// TestMigrationFromVulnerableToSecure demonstrates migration from vulnerable patterns
// to secure filesystem-controlled template loading.
func TestMigrationFromVulnerableToSecure(t *testing.T) {
	t.Parallel()

	t.Run("OldVulnerablePatternBlocked", func(t *testing.T) {
		t.Parallel()

		// This would be the old vulnerable way - trying to include system files
		vulnerableTemplate := `User info: {% include "/etc/passwd" %}`

		// This should now be blocked by the secure loader
		_, err := RenderTemplate(vulnerableTemplate, TemplateFormatJinja2, map[string]any{})
		if err == nil {
			t.Errorf("expected error, got nil")
		} else if !strings.Contains(err.Error(), "template loading from filesystem disabled for security reasons") {
			t.Errorf("expected security error, got %q", err.Error())
		}
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
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		expected := "User: alice\nRole: developer\nDepartment: engineering"
		if result != expected {
			t.Errorf("expected %q, got %q", expected, result)
		}
	})
}
