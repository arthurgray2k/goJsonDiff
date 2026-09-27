package jsondiff

import (
	"encoding/json"
	"testing"
)

func parse(t *testing.T, s string) any {
	t.Helper()
	v, err := LoadFromBytes([]byte(s))
	if err != nil {
		t.Fatalf("failed to parse JSON: %v", err)
	}
	return v
}

func TestCompare_Identical(t *testing.T) {
	// Same data, different formatting and key order
	jsonA := `{
		"name": "Antigravity",
		"version": 2,
		"tags": ["go", "tooling"],
		"meta": { "active": true, "env": "prod" }
	}`
	jsonB := `{
		"meta": { "env": "prod", "active": true },
		"tags": ["go", "tooling"],
		"version": 2,
		"name": "Antigravity"
	}`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if !res.Equal {
		t.Fatalf("expected documents to be equal, got differences: %+v", res.Differences)
	}
	if res.Summary.TotalChanges != 0 {
		t.Errorf("expected 0 total changes, got %d", res.Summary.TotalChanges)
	}
}

func TestCompare_AddedFields(t *testing.T) {
	jsonA := `{"a": 1}`
	jsonB := `{"a": 1, "b": "new", "c": [1, 2]}`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if res.Equal {
		t.Fatalf("expected documents to differ")
	}
	if res.Summary.Added != 2 {
		t.Errorf("expected 2 added fields, got %d", res.Summary.Added)
	}
}

func TestCompare_RemovedFields(t *testing.T) {
	jsonA := `{"a": 1, "b": "old", "c": true}`
	jsonB := `{"a": 1}`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if res.Equal {
		t.Fatalf("expected documents to differ")
	}
	if res.Summary.Removed != 2 {
		t.Errorf("expected 2 removed fields, got %d", res.Summary.Removed)
	}
}

func TestCompare_ModifiedValues(t *testing.T) {
	jsonA := `{"count": 10, "label": "alpha", "flag": true}`
	jsonB := `{"count": 20, "label": "beta", "flag": false}`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if res.Equal {
		t.Fatalf("expected documents to differ")
	}
	if res.Summary.Modified != 3 {
		t.Errorf("expected 3 modified values, got %d", res.Summary.Modified)
	}
}

func TestCompare_TypeChanged(t *testing.T) {
	jsonA := `{"val1": 123, "val2": "hello", "val3": null, "val4": {"nested": 1}}`
	jsonB := `{"val1": "123", "val2": 456, "val3": "now-string", "val4": [1, 2]}`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if res.Equal {
		t.Fatalf("expected documents to differ")
	}
	if res.Summary.TypeChanged != 4 {
		t.Errorf("expected 4 type changes, got %d", res.Summary.TypeChanged)
	}
}

func TestCompare_Arrays(t *testing.T) {
	jsonA := `{"items": [1, 2, 3, {"name": "a"}]}`
	jsonB := `{"items": [1, 99, 3, {"name": "b"}, 5]}`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if res.Equal {
		t.Fatalf("expected documents to differ")
	}

	// Index 1 modified: 2 -> 99
	// Index 3 nested object modified: name "a" -> "b"
	// Index 4 added: 5
	if res.Summary.Modified != 2 {
		t.Errorf("expected 2 modifications in array, got %d", res.Summary.Modified)
	}
	if res.Summary.Added != 1 {
		t.Errorf("expected 1 addition in array, got %d", res.Summary.Added)
	}
}

func TestCompare_ArrayTruncation(t *testing.T) {
	jsonA := `[1, 2, 3, 4]`
	jsonB := `[1, 2]`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if res.Equal {
		t.Fatalf("expected documents to differ")
	}
	if res.Summary.Removed != 2 {
		t.Errorf("expected 2 removed array elements, got %d", res.Summary.Removed)
	}
}

func TestCompare_SpecialKeyNames(t *testing.T) {
	jsonA := `{"key-with-dashes": 1, "key with spaces": 2}`
	jsonB := `{"key-with-dashes": 2, "key with spaces": 3}`

	res := Compare(parse(t, jsonA), parse(t, jsonB))
	if res.Equal {
		t.Fatalf("expected documents to differ")
	}

	expectedPaths := map[string]bool{
		`$["key-with-dashes"]`: false,
		`$["key with spaces"]`: false,
	}
	for _, d := range res.Differences {
		if _, ok := expectedPaths[d.Path]; ok {
			expectedPaths[d.Path] = true
		}
	}
	for path, found := range expectedPaths {
		if !found {
			t.Errorf("expected path %s not found in differences", path)
		}
	}
}

func TestCompare_Numbers(t *testing.T) {
	// Comparing json.Number representation with int/float
	var a, b any
	_ = json.Unmarshal([]byte(`{"n": 42}`), &a)
	_ = json.Unmarshal([]byte(`{"n": 42.0}`), &b)

	res := Compare(a, b)
	if !res.Equal {
		t.Errorf("expected 42 and 42.0 to be equal, got diff: %+v", res.Differences)
	}
}

func TestTypeName(t *testing.T) {
	tests := []struct {
		val      any
		expected string
	}{
		{nil, "null"},
		{true, "boolean"},
		{123, "number"},
		{float64(45.67), "number"},
		{"text", "string"},
		{[]any{1, 2}, "array"},
		{map[string]any{"k": "v"}, "object"},
	}

	for _, tc := range tests {
		actual := TypeName(tc.val)
		if actual != tc.expected {
			t.Errorf("TypeName(%v) = %q, expected %q", tc.val, actual, tc.expected)
		}
	}
}

func BenchmarkCompare_DeepObject(b *testing.B) {
	jsonA := `{"a": {"b": {"c": {"d": {"e": 100, "f": [1, 2, 3]}}}}}`
	jsonB := `{"a": {"b": {"c": {"d": {"e": 200, "f": [1, 2, 4]}}}}}`

	docA, _ := LoadFromBytes([]byte(jsonA))
	docB, _ := LoadFromBytes([]byte(jsonB))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Compare(docA, docB)
	}
}
