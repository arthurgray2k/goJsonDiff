package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRun_IdenticalFiles(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.json")
	f2 := filepath.Join(tmpDir, "b.json")

	_ = os.WriteFile(f1, []byte(`{"a": 1, "b": 2}`), 0644)
	_ = os.WriteFile(f2, []byte(`{"b": 2, "a": 1}`), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{f1, f2}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0, got %d. Stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), "No differences found") {
		t.Errorf("expected 'No differences found', got: %s", stdout.String())
	}
}

func TestRun_DifferentFiles(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.json")
	f2 := filepath.Join(tmpDir, "b.json")

	_ = os.WriteFile(f1, []byte(`{"a": 1}`), 0644)
	_ = os.WriteFile(f2, []byte(`{"a": 2}`), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-format", "json", f1, f2}, &stdout, &stderr)

	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d. Stderr: %s", exitCode, stderr.String())
	}
	if !strings.Contains(stdout.String(), `"equal": false`) {
		t.Errorf("expected json output with 'equal: false', got: %s", stdout.String())
	}
}

func TestRun_ColorFlags(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.json")
	f2 := filepath.Join(tmpDir, "b.json")

	_ = os.WriteFile(f1, []byte(`{"a": 1}`), 0644)
	_ = os.WriteFile(f2, []byte(`{"a": 2}`), 0644)

	// Color always
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-color", "always", f1, f2}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "\033[") {
		t.Errorf("expected ANSI color codes with -color always")
	}

	// Color never
	stdout.Reset()
	stderr.Reset()
	exitCode = run([]string{"-color", "never", "-summary=false", f1, f2}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}
	if strings.Contains(stdout.String(), "\033[") {
		t.Errorf("expected no ANSI color codes with -color never")
	}
	if strings.Contains(stdout.String(), "Summary:") {
		t.Errorf("expected summary omitted with -summary=false")
	}

	// Color auto
	stdout.Reset()
	stderr.Reset()
	exitCode = run([]string{"-color", "auto", f1, f2}, &stdout, &stderr)
	if exitCode != 1 {
		t.Fatalf("expected exit code 1, got %d", exitCode)
	}

	// Invalid color
	stdout.Reset()
	stderr.Reset()
	exitCode = run([]string{"-color", "invalid_choice", f1, f2}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 for invalid color, got %d", exitCode)
	}
}

func TestRun_InvalidFormat(t *testing.T) {
	tmpDir := t.TempDir()
	f1 := filepath.Join(tmpDir, "a.json")
	f2 := filepath.Join(tmpDir, "b.json")

	_ = os.WriteFile(f1, []byte(`{}`), 0644)
	_ = os.WriteFile(f2, []byte(`{}`), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-format", "xml", f1, f2}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 for unsupported format, got %d", exitCode)
	}
}

func TestRun_BothStdinError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-", "-"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 when both are stdin, got %d", exitCode)
	}
}

func TestRun_InvalidArguments(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"only-one-arg.json"}, &stdout, &stderr)

	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
	if !strings.Contains(stderr.String(), "exactly two file arguments are required") {
		t.Errorf("expected error message in stderr, got: %s", stderr.String())
	}
}

func TestRun_FlagParseError(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-unknown-flag-xyz"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2 for unknown flag, got %d", exitCode)
	}
}

func TestRun_Version(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"-version"}, &stdout, &stderr)

	if exitCode != 0 {
		t.Fatalf("expected exit code 0 for -version, got %d", exitCode)
	}
	if !strings.Contains(stdout.String(), "gojsondiff version") {
		t.Errorf("expected version output, got: %s", stdout.String())
	}

	stdout.Reset()
	exitCode = run([]string{"-v"}, &stdout, &stderr)
	if exitCode != 0 || !strings.Contains(stdout.String(), "gojsondiff version") {
		t.Errorf("expected shorthand -v to work")
	}
}

func TestRun_FileError(t *testing.T) {
	tmpDir := t.TempDir()
	validFile := filepath.Join(tmpDir, "valid.json")
	_ = os.WriteFile(validFile, []byte(`{}`), 0644)

	var stdout, stderr bytes.Buffer
	exitCode := run([]string{"/nonexistent/a.json", validFile}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}

	stdout.Reset()
	stderr.Reset()
	exitCode = run([]string{validFile, "/nonexistent/b.json"}, &stdout, &stderr)
	if exitCode != 2 {
		t.Fatalf("expected exit code 2, got %d", exitCode)
	}
}
