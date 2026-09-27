package jsondiff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
)

// DiffKind represents the classification of a difference between two JSON nodes.
type DiffKind string

const (
	KindAdded       DiffKind = "added"
	KindRemoved     DiffKind = "removed"
	KindModified    DiffKind = "modified"
	KindTypeChanged DiffKind = "type_changed"
	KindUnchanged   DiffKind = "unchanged"
)

// Difference describes a single change detected between two JSON structures.
type Difference struct {
	Path     string   `json:"path"`
	Kind     DiffKind `json:"kind"`
	OldValue any      `json:"old_value,omitempty"`
	NewValue any      `json:"new_value,omitempty"`
	OldType  string   `json:"old_type,omitempty"`
	NewType  string   `json:"new_type,omitempty"`
}

// DiffSummary provides aggregate statistics of detected differences.
type DiffSummary struct {
	TotalChanges int `json:"total_changes"`
	Added        int `json:"added"`
	Removed      int `json:"removed"`
	Modified     int `json:"modified"`
	TypeChanged  int `json:"type_changed"`
}

// DiffResult contains the complete output of comparing two JSON structures.
type DiffResult struct {
	Equal       bool         `json:"equal"`
	Differences []Difference `json:"differences"`
	Summary     DiffSummary  `json:"summary"`
	Original    any          `json:"-"`
	Modified    any          `json:"-"`
}

// Options configures the diff engine behavior.
type Options struct {
	// RootPath overrides the root path if specified. Defaults to "".
	RootPath string
}

var identRegex = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// Compare structurally compares two JSON documents (already parsed into any)
// and returns the differences and a summary.
func Compare(a, b any, opts ...Options) *DiffResult {
	opt := Options{RootPath: ""}
	if len(opts) > 0 {
		opt.RootPath = opts[0].RootPath
	}

	var diffs []Difference
	diffRecursive(opt.RootPath, a, b, &diffs)

	summary := DiffSummary{
		TotalChanges: len(diffs),
	}
	for _, d := range diffs {
		switch d.Kind {
		case KindAdded:
			summary.Added++
		case KindRemoved:
			summary.Removed++
		case KindModified:
			summary.Modified++
		case KindTypeChanged:
			summary.TypeChanged++
		}
	}

	return &DiffResult{
		Equal:       len(diffs) == 0,
		Differences: diffs,
		Summary:     summary,
		Original:    a,
		Modified:    b,
	}
}

// TypeName returns a standardized JSON type name for a given Go value.
func TypeName(v any) string {
	if v == nil {
		return "null"
	}
	switch v.(type) {
	case bool:
		return "boolean"
	case json.Number, float64, float32, int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64:
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return reflect.TypeOf(v).String()
	}
}

func diffRecursive(path string, a, b any, diffs *[]Difference) {
	typeA := TypeName(a)
	typeB := TypeName(b)

	if typeA != typeB {
		*diffs = append(*diffs, Difference{
			Path:     path,
			Kind:     KindTypeChanged,
			OldValue: a,
			NewValue: b,
			OldType:  typeA,
			NewType:  typeB,
		})
		return
	}

	switch valA := a.(type) {
	case map[string]any:
		valB := b.(map[string]any)
		diffObjects(path, valA, valB, diffs)

	case []any:
		valB := b.([]any)
		diffArrays(path, valA, valB, diffs)

	default:
		// Scalar values: number, string, boolean, null
		if !scalarEqual(a, b) {
			*diffs = append(*diffs, Difference{
				Path:     path,
				Kind:     KindModified,
				OldValue: a,
				NewValue: b,
				OldType:  typeA,
				NewType:  typeB,
			})
		}
	}
}

func diffObjects(path string, a, b map[string]any, diffs *[]Difference) {
	allKeys := make(map[string]struct{})
	for k := range a {
		allKeys[k] = struct{}{}
	}
	for k := range b {
		allKeys[k] = struct{}{}
	}

	sortedKeys := make([]string, 0, len(allKeys))
	for k := range allKeys {
		sortedKeys = append(sortedKeys, k)
	}
	sort.Strings(sortedKeys)

	for _, k := range sortedKeys {
		childPath := appendObjectPath(path, k)
		inA, hasA := a[k]
		inB, hasB := b[k]

		if hasA && !hasB {
			*diffs = append(*diffs, Difference{
				Path:     childPath,
				Kind:     KindRemoved,
				OldValue: inA,
				OldType:  TypeName(inA),
			})
		} else if !hasA && hasB {
			*diffs = append(*diffs, Difference{
				Path:     childPath,
				Kind:     KindAdded,
				NewValue: inB,
				NewType:  TypeName(inB),
			})
		} else {
			diffRecursive(childPath, inA, inB, diffs)
		}
	}
}

func diffArrays(path string, a, b []any, diffs *[]Difference) {
	lenA := len(a)
	lenB := len(b)
	minLen := lenA
	if lenB < minLen {
		minLen = lenB
	}

	// Compare overlapping elements
	for i := 0; i < minLen; i++ {
		elemPath := appendArrayPath(path, i)
		diffRecursive(elemPath, a[i], b[i], diffs)
	}

	// Extra elements in B (Added)
	for i := minLen; i < lenB; i++ {
		elemPath := appendArrayPath(path, i)
		*diffs = append(*diffs, Difference{
			Path:     elemPath,
			Kind:     KindAdded,
			NewValue: b[i],
			NewType:  TypeName(b[i]),
		})
	}

	// Missing elements from A (Removed)
	for i := minLen; i < lenA; i++ {
		elemPath := appendArrayPath(path, i)
		*diffs = append(*diffs, Difference{
			Path:     elemPath,
			Kind:     KindRemoved,
			OldValue: a[i],
			OldType:  TypeName(a[i]),
		})
	}
}

func scalarEqual(a, b any) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}

	numA, isNumA := toNumeric(a)
	numB, isNumB := toNumeric(b)
	if isNumA && isNumB {
		return numA == numB
	}

	return reflect.DeepEqual(a, b)
}

func toNumeric(v any) (float64, bool) {
	switch val := v.(type) {
	case json.Number:
		f, err := val.Float64()
		if err == nil {
			return f, true
		}
		return 0, false
	case float64:
		return val, true
	case float32:
		return float64(val), true
	case int:
		return float64(val), true
	case int64:
		return float64(val), true
	default:
		return 0, false
	}
}

func appendObjectPath(base, key string) string {
	var part string
	if identRegex.MatchString(key) {
		part = key
	} else {
		part = fmt.Sprintf("[%q]", key)
	}

	if base == "" {
		return part
	}
	if strings.HasPrefix(part, "[") {
		return base + part
	}
	return base + "." + part
}

func appendArrayPath(base string, idx int) string {
	if base == "" {
		return fmt.Sprintf("[%d]", idx)
	}
	return fmt.Sprintf("%s[%d]", base, idx)
}
