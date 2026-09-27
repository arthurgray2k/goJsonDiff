package jsondiff

import (
	"strings"
	"testing"
)

func TestFormatText_Equal(t *testing.T) {
	result := &DiffResult{Equal: true}
	text := FormatText(result, FormatOptions{Colorize: false, ShowSummary: true})

	if !strings.Contains(text, "No differences found") {
		t.Errorf("expected 'No differences found', got: %s", text)
	}
}

func TestFormatText_Differences(t *testing.T) {
	result := &DiffResult{
		Equal: false,
		Differences: []Difference{
			{Path: "$.name", Kind: KindAdded, NewValue: "Bob"},
			{Path: "$.age", Kind: KindRemoved, OldValue: 25},
			{Path: "$.status", Kind: KindModified, OldValue: "pending", NewValue: "active"},
			{Path: "$.count", Kind: KindTypeChanged, OldValue: 1, NewValue: "1", OldType: "number", NewType: "string"},
		},
		Summary: DiffSummary{
			TotalChanges: 4,
			Added:        1,
			Removed:      1,
			Modified:     1,
			TypeChanged:  1,
		},
	}

	text := FormatText(result, FormatOptions{Colorize: false, ShowSummary: true})

	if !strings.Contains(text, "+ $.name: \"Bob\"") {
		t.Errorf("missing added line: %s", text)
	}
	if !strings.Contains(text, "- $.age: 25") {
		t.Errorf("missing removed line: %s", text)
	}
	if !strings.Contains(text, "~ $.status: \"pending\" -> \"active\"") {
		t.Errorf("missing modified line: %s", text)
	}
	if !strings.Contains(text, "* $.count: 1 (number) -> \"1\" (string)") {
		t.Errorf("missing type changed line: %s", text)
	}
	if !strings.Contains(text, "Summary: 4 total changes") {
		t.Errorf("missing summary: %s", text)
	}
}

func TestFormatText_Colorized(t *testing.T) {
	result := &DiffResult{
		Equal: false,
		Differences: []Difference{
			{Path: "$.item", Kind: KindAdded, NewValue: "foo"},
		},
		Summary: DiffSummary{TotalChanges: 1, Added: 1},
	}

	colored := FormatText(result, FormatOptions{Colorize: true, ShowSummary: true})
	if !strings.Contains(colored, "\033[32m") {
		t.Errorf("expected ANSI green color code in output: %s", colored)
	}
}

func TestFormatJSON(t *testing.T) {
	result := &DiffResult{
		Equal: false,
		Differences: []Difference{
			{Path: "$.score", Kind: KindModified, OldValue: 10, NewValue: 20},
		},
		Summary: DiffSummary{TotalChanges: 1, Modified: 1},
	}

	jsonStr, err := FormatJSON(result, true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(jsonStr, `"path": "$.score"`) {
		t.Errorf("expected json to contain path $.score, got: %s", jsonStr)
	}
}
