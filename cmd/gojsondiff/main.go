package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/arthurgray2k/goJsonDiff/internal/jsondiff"
)

var Version = "1.0.0"

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("gojsondiff", flag.ContinueOnError)
	fs.SetOutput(stderr)

	var (
		colorFlag   = fs.String("color", "auto", "Enable colorized output: auto, always, never")
		formatFlag  = fs.String("format", "text", "Output format: text, json")
		summaryFlag = fs.Bool("summary", true, "Include aggregate changes summary in text output")
		typeFlag    = fs.Bool("type", false, "Show type mutation details in diff output")
		inlineFlag  = fs.Bool("inline", false, "Render diff in hierarchical tree/inline format")
		versionFlag = fs.Bool("version", false, "Print version information")
		vFlag       = fs.Bool("v", false, "Print version information (shorthand)")
	)

	fs.Usage = func() {
		fmt.Fprintf(stderr, "Usage: gojsondiff [options] <fileA.json> <fileB.json>\n\n")
		fmt.Fprintf(stderr, "A high-performance structural JSON diff tool that compares JSON documents based on data,\n")
		fmt.Fprintf(stderr, "ignoring irrelevant formatting, key order, and indentation.\n\n")
		fmt.Fprintf(stderr, "Arguments:\n")
		fmt.Fprintf(stderr, "  <fileA.json>   Path to original JSON document (or '-' for stdin)\n")
		fmt.Fprintf(stderr, "  <fileB.json>   Path to modified JSON document (or '-' for stdin)\n\n")
		fmt.Fprintf(stderr, "Options:\n")
		fs.PrintDefaults()
		fmt.Fprintf(stderr, "\nExit Codes:\n")
		fmt.Fprintf(stderr, "  0  No differences found\n")
		fmt.Fprintf(stderr, "  1  Differences found\n")
		fmt.Fprintf(stderr, "  2  Error reading files or invalid arguments\n")
	}

	if err := fs.Parse(args); err != nil {
		return 2
	}

	if *versionFlag || *vFlag {
		fmt.Fprintf(stdout, "gojsondiff version %s\n", Version)
		return 0
	}

	positional := fs.Args()
	if len(positional) != 2 {
		fmt.Fprintf(stderr, "Error: exactly two file arguments are required, received %d\n\n", len(positional))
		fs.Usage()
		return 2
	}

	pathA := positional[0]
	pathB := positional[1]

	if pathA == "-" && pathB == "-" {
		fmt.Fprintf(stderr, "Error: cannot read both documents from stdin ('-') simultaneously\n")
		return 2
	}

	docA, err := jsondiff.LoadFromFile(pathA)
	if err != nil {
		fmt.Fprintf(stderr, "Error loading first JSON document (%s): %v\n", pathA, err)
		return 2
	}

	docB, err := jsondiff.LoadFromFile(pathB)
	if err != nil {
		fmt.Fprintf(stderr, "Error loading second JSON document (%s): %v\n", pathB, err)
		return 2
	}

	result := jsondiff.Compare(docA, docB)

	colorize := false
	switch *colorFlag {
	case "always", "true", "1":
		colorize = true
	case "never", "false", "0":
		colorize = false
	case "auto":
		if f, ok := stdout.(*os.File); ok {
			fileInfo, err := f.Stat()
			if err == nil && (fileInfo.Mode()&os.ModeCharDevice) != 0 {
				colorize = true
			}
		}
	default:
		fmt.Fprintf(stderr, "Error: invalid value for -color: %q (expected auto, always, or never)\n", *colorFlag)
		return 2
	}

	switch *formatFlag {
	case "text":
		out := jsondiff.Format(result, jsondiff.FormatOptions{
			Colorize:    colorize,
			ShowSummary: *summaryFlag,
			ShowTypes:   *typeFlag,
			Inline:      *inlineFlag,
		})
		fmt.Fprint(stdout, out)

	case "json":
		out, err := jsondiff.FormatJSON(result, true)
		if err != nil {
			fmt.Fprintf(stderr, "Error formatting JSON output: %v\n", err)
			return 2
		}
		fmt.Fprintln(stdout, out)

	default:
		fmt.Fprintf(stderr, "Error: unsupported output format %q (expected text or json)\n", *formatFlag)
		return 2
	}

	if !result.Equal {
		return 1
	}

	return 0
}
