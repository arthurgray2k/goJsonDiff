package jsondiff

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadFromBytes_Valid(t *testing.T) {
	input := []byte(`{"name": "Alice", "age": 30, "active": true}`)
	val, err := LoadFromBytes(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := val.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", val)
	}

	if m["name"] != "Alice" {
		t.Errorf("expected Alice, got %v", m["name"])
	}
}

func TestLoadFromBytes_Invalid(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"empty", ""},
		{"malformed", `{"name": "Alice"`},
		{"trailing data", `{"name": "Alice"} extra`},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			_, err := LoadFromBytes([]byte(tc.input))
			if err == nil {
				t.Fatalf("expected error for %s, got nil", tc.name)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "test.json")

	content := `{"key": "value", "count": 42}`
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}

	val, err := LoadFromFile(filePath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	m, ok := val.(map[string]any)
	if !ok || m["key"] != "value" {
		t.Errorf("unexpected loaded value: %v", val)
	}

	// Non-existent file
	_, err = LoadFromFile(filepath.Join(tmpDir, "missing.json"))
	if err == nil {
		t.Errorf("expected error for non-existent file, got nil")
	}
}

func TestLoadFromReader(t *testing.T) {
	r := strings.NewReader(`[1, 2, 3]`)
	val, err := LoadFromReader(r)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	arr, ok := val.([]any)
	if !ok || len(arr) != 3 {
		t.Errorf("expected slice of length 3, got %v", val)
	}
}
