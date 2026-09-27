# goJsonDiff

A high-performance structural JSON diff utility and Go package built with zero third-party dependencies (pure Go standard library, targeting Go 1.26).

Unlike conventional textual or line-by-line diff utilities (such as `diff` or `git diff`), `goJsonDiff` compares JSON documents based on their underlying **data and structure**:
- Ignores key order in objects (`{"a": 1, "b": 2}` equals `{"b": 2, "a": 1}`).
- Ignores whitespace, indentation, and formatting discrepancies.
- Accurately tracks deep hierarchical changes using standard JSONPath notation (`$.user.addresses[0].city`).
- Accurately detects:
  - **Added fields** (`+`)
  - **Removed fields** (`-`)
  - **Modified values** (`~`)
  - **Type mutations** (`*`) (e.g. number changed to string, or object changed to array).
- Provides human-readable output, ANSI color terminal highlighting, and machine-readable structured JSON diff output.

---

## Installation

### Prerequisites
- Go 1.26 or later installed.

### Build from Source
```bash
git clone git@github-golang:arthurgray2k/goJsonDiff.git
cd goJsonDiff
make build
```

Or install directly using the Go toolchain:
```bash
go install ./cmd/gojsondiff
```

---

## Quick Start

### Basic Comparison
```bash
gojsondiff original.json modified.json
```

### Stdin Streaming
Pipe JSON directly via standard input using `-`:
```bash
curl -s https://api.example.com/v1/config | gojsondiff expected.json -
```

### Colorized Terminal Diff
```bash
gojsondiff -color always original.json modified.json
```

### Structured JSON Output
```bash
gojsondiff -format json original.json modified.json
```

---

## Project Structure

```text
goJsonDiff/
├── cmd/
│   └── gojsondiff/
│       ├── main.go              # CLI entry point, argument parsing, exit codes
│       └── main_test.go         # CLI integration and unit tests
├── internal/
│   └── jsondiff/
│       ├── diff.go              # Core recursive comparison engine
│       ├── diff_test.go         # Unit tests and benchmarks (>89% coverage)
│       ├── formatter.go         # Human-readable, color, and JSON formatters
│       ├── formatter_test.go    # Formatters unit tests
│       ├── loader.go            # Stream and file JSON decoding with json.Number
│       └── loader_test.go       # Loader unit tests
├── examples/
│   ├── original.json            # Sample baseline document
│   └── modified.json            # Sample updated document
├── go.mod                       # Go 1.26 module definition
├── Makefile                     # Build and testing targets
├── README.md                    # Project documentation
├── USAGE.md                     # Comprehensive CLI usage guide
└── LICENSE                      # Mozilla Public License 2.0
```

---

## Programmatic Library Usage

`goJsonDiff` can be embedded into Go applications:

```go
package main

import (
	"fmt"
	"log"

	"github.com/arthurgray2k/goJsonDiff/internal/jsondiff"
)

func main() {
	docA, err := jsondiff.LoadFromBytes([]byte(`{"service": "auth", "port": 8080}`))
	if err != nil {
		log.Fatal(err)
	}

	docB, err := jsondiff.LoadFromBytes([]byte(`{"service": "auth", "port": "8080", "ssl": true}`))
	if err != nil {
		log.Fatal(err)
	}

	result := jsondiff.Compare(docA, docB)

	if result.Equal {
		fmt.Println("Documents are identical!")
		return
	}

	// Print human-readable diff
	fmt.Print(jsondiff.FormatText(result, jsondiff.FormatOptions{
		Colorize:    true,
		ShowSummary: true,
	}))
}
```

---

## Testing & Verification

Run tests and check test coverage:
```bash
make test-coverage
```

Current test suite achievements:
- `cmd/gojsondiff`: >90% statement coverage
- `internal/jsondiff`: >89% statement coverage
- Clean `go vet` and `go fmt` compliance.

---

## License

This project is licensed under the Mozilla Public License Version 2.0. See [LICENSE](LICENSE) for details.
