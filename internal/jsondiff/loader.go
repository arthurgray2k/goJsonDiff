package jsondiff

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// LoadFromFile reads and decodes JSON from a file path.
// If path is "-", it reads from os.Stdin.
func LoadFromFile(path string) (any, error) {
	if path == "-" {
		return LoadFromReader(os.Stdin)
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %q: %w", path, err)
	}

	return LoadFromBytes(data)
}

// LoadFromBytes decodes JSON from a byte slice into an any value,
// using json.Number to preserve exact number formatting without precision loss.
func LoadFromBytes(data []byte) (any, error) {
	return LoadFromReader(bytes.NewReader(data))
}

// LoadFromReader decodes JSON from an io.Reader into an any value.
func LoadFromReader(r io.Reader) (any, error) {
	decoder := json.NewDecoder(r)
	decoder.UseNumber()

	var val any
	if err := decoder.Decode(&val); err != nil {
		if err == io.EOF {
			return nil, fmt.Errorf("empty JSON input")
		}
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}

	// Ensure there is no trailing content (other than whitespace)
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return nil, fmt.Errorf("unexpected content after JSON document: %w", err)
	}

	return val, nil
}
