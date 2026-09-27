package jsondiff

import (
	"strings"
	"testing"
)

func TestFormatDefault_Equal(t *testing.T) {
	result := &DiffResult{Equal: true}
	text := Format(result, FormatOptions{Colorize: false, ShowSummary: true})

	if !strings.Contains(text, "No differences found") {
		t.Errorf("expected 'No differences found', got: %s", text)
	}
}

func TestFormatDefault_Structure(t *testing.T) {
	result := &DiffResult{
		Equal: false,
		Differences: []Difference{
			{Path: "age", Kind: KindModified, OldValue: 40, NewValue: 41, OldType: "number", NewType: "number"},
			{Path: "city", Kind: KindRemoved, OldValue: "Delhi", OldType: "string"},
			{Path: "country", Kind: KindAdded, NewValue: "India", NewType: "string"},
		},
		Summary: DiffSummary{
			TotalChanges: 3,
			Added:        1,
			Removed:      1,
			Modified:     1,
		},
	}

	text := Format(result, FormatOptions{Colorize: false, ShowSummary: true})

	if !strings.Contains(text, "JSON DIFF") {
		t.Errorf("missing 'JSON DIFF' header: %s", text)
	}
	if !strings.Contains(text, "  age\n    - 40\n    + 41") {
		t.Errorf("missing age diff: %s", text)
	}
	if !strings.Contains(text, "  city\n    - \"Delhi\"") {
		t.Errorf("missing city diff: %s", text)
	}
	if !strings.Contains(text, "  country\n    + \"India\"") {
		t.Errorf("missing country diff: %s", text)
	}
}

func TestFormatDefault_WithType(t *testing.T) {
	result := &DiffResult{
		Equal: false,
		Differences: []Difference{
			{Path: "age", Kind: KindTypeChanged, OldValue: 40, NewValue: "40", OldType: "number", NewType: "string"},
		},
		Summary: DiffSummary{
			TotalChanges: 1,
			TypeChanged:  1,
		},
	}

	text := Format(result, FormatOptions{Colorize: false, ShowSummary: false, ShowTypes: true})

	if !strings.Contains(text, "~ type: number → string") {
		t.Errorf("missing type diff line: %s", text)
	}
	if !strings.Contains(text, "    - 40\n    + \"40\"") {
		t.Errorf("missing values diff: %s", text)
	}
}

func TestFormatInline_Tree(t *testing.T) {
	docA := map[string]any{
		"user": map[string]any{
			"name": "Atur",
			"age":  40,
			"address": map[string]any{
				"city": "Delhi",
				"zip":  "110001",
			},
		},
	}
	docB := map[string]any{
		"user": map[string]any{
			"name": "Atur",
			"age":  41,
			"address": map[string]any{
				"city": "Noida",
			},
			"email": "atur@example.com",
		},
	}

	res := Compare(docA, docB)
	tree := FormatInline(res, FormatOptions{Colorize: false})

	if !strings.Contains(tree, "JSON DIFF") {
		t.Errorf("missing JSON DIFF in tree: %s", tree)
	}
	if !strings.Contains(tree, "user") {
		t.Errorf("missing user in tree: %s", tree)
	}
	if !strings.Contains(tree, `= "Atur"`) {
		t.Errorf("missing = \"Atur\" in tree: %s", tree)
	}
	if !strings.Contains(tree, `~ 40 → 41`) {
		t.Errorf("missing ~ 40 -> 41 in tree: %s", tree)
	}
	if !strings.Contains(tree, `- "110001"`) {
		t.Errorf("missing - \"110001\" in tree: %s", tree)
	}
	if !strings.Contains(tree, `+ "atur@example.com"`) {
		t.Errorf("missing + \"atur@example.com\" in tree: %s", tree)
	}
}

func TestFormatJSON(t *testing.T) {
	result := &DiffResult{
		Equal: false,
		Differences: []Difference{
			{Path: "score", Kind: KindModified, OldValue: 10, NewValue: 20},
		},
		Summary: DiffSummary{TotalChanges: 1, Modified: 1},
	}

	jsonStr, err := FormatJSON(result, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(jsonStr, `"path": "score"`) {
		t.Errorf("expected json to contain path score, got: %s", jsonStr)
	}
}
