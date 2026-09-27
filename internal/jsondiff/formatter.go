package jsondiff

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

const (
	ansiReset   = "\033[0m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiMagenta = "\033[35m"
	ansiCyan    = "\033[36m"
	ansiDim     = "\033[2m"
	ansiBold    = "\033[1m"
)

// FormatOptions controls output formatting options.
type FormatOptions struct {
	Colorize    bool
	ShowSummary bool
	ShowTypes   bool
	Inline      bool
}

// Format renders differences according to FormatOptions.
func Format(result *DiffResult, opts FormatOptions) string {
	if result.Equal {
		msg := "No differences found. JSON documents are identical in structure and data."
		if opts.Colorize {
			return ansiGreen + msg + ansiReset + "\n"
		}
		return msg + "\n"
	}

	if opts.Inline {
		return FormatInline(result, opts)
	}

	return FormatDefault(result, opts)
}

// FormatDefault renders structural diff conforming to GOJSONDIFF-3 specification:
// JSON DIFF
//
//	age
//	  - 40
//	  + 41
//	city
//	  - "Delhi"
//	country
//	  + "India"
func FormatDefault(result *DiffResult, opts FormatOptions) string {
	if result.Equal {
		return "No differences found.\n"
	}

	var sb strings.Builder
	sb.WriteString("JSON DIFF\n")

	for _, d := range result.Differences {
		sb.WriteString(formatDiffEntry(d, opts))
	}

	if opts.ShowSummary {
		sb.WriteString("\n")
		sb.WriteString(formatSummary(result.Summary, opts.Colorize))
		sb.WriteString("\n")
	}

	return sb.String()
}

func formatDiffEntry(d Difference, opts FormatOptions) string {
	var sb strings.Builder
	// Path header indented by 2 spaces
	pathStr := fmt.Sprintf("  %s\n", d.Path)
	if opts.Colorize {
		sb.WriteString(ansiBold + pathStr + ansiReset)
	} else {
		sb.WriteString(pathStr)
	}

	// Show type diff if -type flag is passed or if type changed
	if opts.ShowTypes && d.OldType != "" && d.NewType != "" && d.OldType != d.NewType {
		typeStr := fmt.Sprintf("    ~ type: %s → %s\n", d.OldType, d.NewType)
		if opts.Colorize {
			sb.WriteString(ansiMagenta + typeStr + ansiReset)
		} else {
			sb.WriteString(typeStr)
		}
	} else if opts.ShowTypes && (d.Kind == KindModified || d.Kind == KindAdded || d.Kind == KindRemoved) {
		tName := d.OldType
		if tName == "" {
			tName = d.NewType
		}
		typeStr := fmt.Sprintf("    ~ type: %s\n", tName)
		if opts.Colorize {
			sb.WriteString(ansiDim + typeStr + ansiReset)
		} else {
			sb.WriteString(typeStr)
		}
	} else if d.Kind == KindTypeChanged {
		typeStr := fmt.Sprintf("    ~ type: %s → %s\n", d.OldType, d.NewType)
		if opts.Colorize {
			sb.WriteString(ansiMagenta + typeStr + ansiReset)
		} else {
			sb.WriteString(typeStr)
		}
	}

	switch d.Kind {
	case KindAdded:
		sb.WriteString(formatValueLines("+", d.NewValue, opts.Colorize, ansiGreen))

	case KindRemoved:
		sb.WriteString(formatValueLines("-", d.OldValue, opts.Colorize, ansiRed))

	case KindModified, KindTypeChanged:
		sb.WriteString(formatValueLines("-", d.OldValue, opts.Colorize, ansiRed))
		sb.WriteString(formatValueLines("+", d.NewValue, opts.Colorize, ansiGreen))
	}

	return sb.String()
}

func formatValueLines(prefix string, val any, colorize bool, colorCode string) string {
	formatted := formatPrettyValue(val)
	lines := strings.Split(formatted, "\n")

	var sb strings.Builder
	for i, line := range lines {
		var out string
		if i == 0 {
			out = fmt.Sprintf("    %s %s\n", prefix, line)
		} else {
			out = fmt.Sprintf("    %s\n", line)
		}

		if colorize {
			sb.WriteString(colorCode + out + ansiReset)
		} else {
			sb.WriteString(out)
		}
	}
	return sb.String()
}

func formatPrettyValue(v any) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case map[string]any, []any:
		b, err := json.MarshalIndent(val, "", "  ")
		if err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

func formatSummary(s DiffSummary, colorize bool) string {
	text := fmt.Sprintf("Summary: %d total changes (%d added, %d removed, %d modified, %d type changed)",
		s.TotalChanges, s.Added, s.Removed, s.Modified, s.TypeChanged)
	if colorize {
		return ansiBold + text + ansiReset
	}
	return text
}

// FormatJSON renders differences as formatted JSON.
func FormatJSON(result *DiffResult, pretty bool) (string, error) {
	var bytes []byte
	var err error
	if pretty {
		bytes, err = json.MarshalIndent(result, "", "  ")
	} else {
		bytes, err = json.Marshal(result)
	}
	if err != nil {
		return "", fmt.Errorf("failed to format JSON: %w", err)
	}
	return string(bytes), nil
}

// --- Inline / Tree Format Implementation ---

type treeNode struct {
	name     string
	kind     DiffKind
	oldVal   any
	newVal   any
	oldType  string
	newType  string
	children []*treeNode
}

// FormatInline renders an ASCII tree of changes:
// JSON DIFF
// user
// ├── name
// │   └── = "Atur"
// ├── age
// │   └── ~ 40 → 41
// ├── address
// │   ├── city
// │   │   └── ~ "Delhi" → "Noida"
// │   └── zip
// │       └── - "110001"
// └── email
//
//	└── + "atur@example.com"
func FormatInline(result *DiffResult, opts FormatOptions) string {
	if result.Equal {
		return "No differences found.\n"
	}

	rootNodes := buildDiffTree(result.Original, result.Modified)

	var sb strings.Builder
	sb.WriteString("JSON DIFF\n")
	for i, node := range rootNodes {
		isLast := (i == len(rootNodes)-1)
		renderTreeNode(&sb, node, "", isLast, true, opts)
	}

	return sb.String()
}

func buildDiffTree(a, b any) []*treeNode {
	mapA, okA := a.(map[string]any)
	mapB, okB := b.(map[string]any)

	if okA && okB {
		allKeys := make(map[string]struct{})
		for k := range mapA {
			allKeys[k] = struct{}{}
		}
		for k := range mapB {
			allKeys[k] = struct{}{}
		}

		sortedKeys := make([]string, 0, len(allKeys))
		for k := range allKeys {
			sortedKeys = append(sortedKeys, k)
		}
		sort.Strings(sortedKeys)

		var nodes []*treeNode
		for _, k := range sortedKeys {
			vA, hasA := mapA[k]
			vB, hasB := mapB[k]

			if hasA && !hasB {
				nodes = append(nodes, &treeNode{
					name:   k,
					kind:   KindRemoved,
					oldVal: vA,
				})
			} else if !hasA && hasB {
				nodes = append(nodes, &treeNode{
					name:   k,
					kind:   KindAdded,
					newVal: vB,
				})
			} else {
				// Present in both
				childMapA, isChildMapA := vA.(map[string]any)
				childMapB, isChildMapB := vB.(map[string]any)
				if isChildMapA && isChildMapB {
					childNodes := buildDiffTree(childMapA, childMapB)
					nodes = append(nodes, &treeNode{
						name:     k,
						children: childNodes,
					})
				} else if !scalarEqual(vA, vB) {
					tA := TypeName(vA)
					tB := TypeName(vB)
					kKind := KindModified
					if tA != tB {
						kKind = KindTypeChanged
					}
					nodes = append(nodes, &treeNode{
						name:    k,
						kind:    kKind,
						oldVal:  vA,
						newVal:  vB,
						oldType: tA,
						newType: tB,
					})
				} else {
					// Unchanged leaf
					nodes = append(nodes, &treeNode{
						name:   k,
						kind:   KindUnchanged,
						oldVal: vA,
					})
				}
			}
		}
		return nodes
	}

	// Fallback for non-object root
	return []*treeNode{
		{
			name:   "$",
			kind:   KindModified,
			oldVal: a,
			newVal: b,
		},
	}
}

func renderTreeNode(sb *strings.Builder, node *treeNode, prefix string, isLast bool, isRoot bool, opts FormatOptions) {
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	if isRoot {
		// Print root node header (e.g. user)
		if len(node.children) > 0 {
			sb.WriteString(node.name + "\n")
			for i, child := range node.children {
				renderTreeNode(sb, child, "", i == len(node.children)-1, false, opts)
			}
			return
		}
	}

	// Non-root intermediate or leaf
	sb.WriteString(prefix + connector + node.name + "\n")

	if len(node.children) > 0 {
		for i, child := range node.children {
			renderTreeNode(sb, child, childPrefix, i == len(node.children)-1, false, opts)
		}
	} else {
		// Leaf value change
		leafConnector := childPrefix + "└── "
		var symbol, desc, color string
		switch node.kind {
		case KindAdded:
			symbol = "+"
			color = ansiGreen
			desc = fmt.Sprintf("+ %s", formatSingleLineValue(node.newVal))
		case KindRemoved:
			symbol = "-"
			color = ansiRed
			desc = fmt.Sprintf("- %s", formatSingleLineValue(node.oldVal))
		case KindModified, KindTypeChanged:
			symbol = "~"
			color = ansiYellow
			if opts.ShowTypes && node.oldType != node.newType && node.oldType != "" {
				desc = fmt.Sprintf("~ type: %s → %s (%s → %s)", node.oldType, node.newType, formatSingleLineValue(node.oldVal), formatSingleLineValue(node.newVal))
			} else {
				desc = fmt.Sprintf("~ %s → %s", formatSingleLineValue(node.oldVal), formatSingleLineValue(node.newVal))
			}
		case KindUnchanged:
			symbol = "="
			color = ansiDim
			desc = fmt.Sprintf("= %s", formatSingleLineValue(node.oldVal))
		}
		_ = symbol

		if opts.Colorize && color != "" {
			sb.WriteString(leafConnector + color + desc + ansiReset + "\n")
		} else {
			sb.WriteString(leafConnector + desc + "\n")
		}
	}
}

func formatSingleLineValue(v any) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	default:
		b, err := json.Marshal(val)
		if err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", val)
	}
}
