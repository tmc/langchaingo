package pgvector

import (
	"strings"
	"testing"
)

func TestFilterPredicatesKeepCallerTextOutOfTheStatement(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		filter map[string]any
	}{
		{"value closes the literal and appends a tautology", map[string]any{
			"kind": "x' OR '1'='1",
		}},
		{"value ends the statement", map[string]any{
			"kind": "x'; DROP TABLE langchaingo_pg_embedding; --",
		}},
		{"key closes the literal", map[string]any{
			"kind') = 'x' OR '1'='1": "y",
		}},
		{"quote inside an ordinary value", map[string]any{
			"owner": "O'Brien",
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			predicates, args := filterPredicates("data.", tc.filter, 0)
			joined := strings.Join(predicates, " AND ")
			if strings.ContainsAny(joined, "'\";") {
				t.Fatalf("caller text reached the statement: %q", joined)
			}
			for k, v := range tc.filter {
				if !contains(args, k) {
					t.Errorf("key %q must travel as an argument, got args %v", k, args)
				}
				if !contains(args, v) {
					t.Errorf("value %q must travel as an argument, got args %v", v, args)
				}
			}
		})
	}
}

func TestFilterPredicatesNumberFromTheOffset(t *testing.T) {
	t.Parallel()

	predicates, args := filterPredicates("data.", map[string]any{"a": "1", "b": "2"}, 4)
	want := "(data.cmetadata ->> $5) = $6 AND (data.cmetadata ->> $7) = $8"
	if got := strings.Join(predicates, " AND "); got != want {
		t.Errorf("want %q, got %q", want, got)
	}
	if len(args) != 4 {
		t.Errorf("want four arguments beside two pairs, got %v", args)
	}
	if args[0] != "a" || args[2] != "b" {
		t.Errorf("keys must be ordered so the statement is stable, got %v", args)
	}
}

func TestFilterPredicatesOnValuesThatAreNotStrings(t *testing.T) {
	t.Parallel()

	_, args := filterPredicates("t.", map[string]any{"n": 5}, 0)
	if len(args) != 2 || args[1] != "5" {
		t.Errorf("a non-string value must reach the wire as text, got %v", args)
	}
}

func TestFilterPredicatesOnAnEmptyFilter(t *testing.T) {
	t.Parallel()

	predicates, args := filterPredicates("data.", map[string]any{}, 3)
	if len(predicates) != 0 || len(args) != 0 {
		t.Errorf("an empty filter must add nothing, got %v and %v", predicates, args)
	}
}

func contains(args []any, want any) bool {
	for _, a := range args {
		if a == want {
			return true
		}
	}
	return false
}
