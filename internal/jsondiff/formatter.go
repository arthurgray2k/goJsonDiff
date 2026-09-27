package jsondiff

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	ansiReset   = "\033[0m"
	ansiRed     = "\033[31m"
	ansiGreen   = "\033[32m"
	ansiYellow  = "\033[33m"
	ansiMagenta = "\033[35m"
	ansiBold    = "\033[1m"
)

// FormatOptions controls output formatting options.
type FormatOptions struct {
	Colorize    bool
	ShowSummary bool
}

// FormatText renders differences in human-readable diff format.
func FormatText(result *DiffResult, opts FormatOptions) string {
	if result.Equal {
		msg := "No differences found. JSON documents are identical in structure and data."
		if opts.Colorize {
			return ansiGreen + msg + ansiReset + "\n"
		}
		return msg + "\n"
	}

	var sb strings.Builder

	for _, d := range result.Differences {
		line := formatDiffLine(d, opts.Colorize)
		sb.WriteString(line)
		sb.WriteString("\n")
	}

	if opts.ShowSummary {
		sb.WriteString("\n")
		sb.WriteString(formatSummary(result.Summary, opts.Colorize))
		sb.WriteString("\n")
	}

	return sb.String()
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

func formatDiffLine(d Difference, colorize bool) string {
	var symbol string
	var color string
	var desc string

	switch d.Kind {
	case KindAdded:
		symbol = "+"
		color = ansiGreen
		desc = fmt.Sprintf("+ %s: %s", d.Path, formatValue(d.NewValue))

	case KindRemoved:
		symbol = "-"
		color = ansiRed
		desc = fmt.Sprintf("- %s: %s", d.Path, formatValue(d.OldValue))

	case KindModified:
		symbol = "~"
		color = ansiYellow
		desc = fmt.Sprintf("~ %s: %s -> %s", d.Path, formatValue(d.OldValue), formatValue(d.NewValue))

	case KindTypeChanged:
		symbol = "*"
		color = ansiMagenta
		desc = fmt.Sprintf("* %s: %s (%s) -> %s (%s)", d.Path, formatValue(d.OldValue), d.OldType, formatValue(d.NewValue), d.NewType)
	}

	_ = symbol
	if colorize && color != "" {
		return color + desc + ansiReset
	}
	return desc
}

func formatSummary(s DiffSummary, colorize bool) string {
	text := fmt.Sprintf("Summary: %d total changes (%d added, %d removed, %d modified, %d type changed)",
		s.TotalChanges, s.Added, s.Removed, s.Modified, s.TypeChanged)
	if colorize {
		return ansiBold + text + ansiReset
	}
	return text
}

func formatValue(v any) string {
	if v == nil {
		return "null"
	}
	switch val := v.(type) {
	case string:
		return fmt.Sprintf("%q", val)
	case map[string]any, []any:
		b, err := json.Marshal(val)
		if err == nil {
			return string(b)
		}
		return fmt.Sprintf("%v", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}
