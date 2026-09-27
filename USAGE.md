# goJsonDiff Usage Guide

This guide provides comprehensive command-line usage examples, flag configurations, CI/CD integration patterns, and troubleshooting steps for `gojsondiff`.

---

## Command Syntax

```bash
gojsondiff [options] <fileA.json> <fileB.json>
```

### Arguments
- `<fileA.json>`: Path to the original or baseline JSON document. Pass `-` to read from standard input.
- `<fileB.json>`: Path to the modified or incoming JSON document. Pass `-` to read from standard input.

*(Note: Reading both documents from standard input simultaneously is unsupported.)*

---

## Options & Flags

| Flag | Type | Default | Description |
| :--- | :--- | :--- | :--- |
| `-type` | `bool` | `false` | Display data type mutations (e.g. `~ type: number → string`). |
| `-inline` | `bool` | `false` | Render output as an ASCII hierarchical tree view with change indicators. |
| `-color` | `string` | `auto` | Terminal coloring: `auto` (detects TTY), `always`, or `never`. |
| `-format` | `string` | `text` | Diff output formatting: `text` (human-readable) or `json`. |
| `-summary` | `bool` | `true` | Show or omit aggregate change statistics in text output. |
| `-v`, `-version` | `bool` | `false` | Display version information and exit. |
| `-h`, `-help` | `bool` | `false` | Display command help and exit. |

---

## Exit Codes

`gojsondiff` returns semantic exit codes suitable for automated scripting and CI/CD pipelines:

- **`0`**: The two JSON documents are structurally and data identical.
- **`1`**: Differences were detected between the two documents.
- **`2`**: An operational error occurred (e.g. invalid arguments, malformed JSON, or missing files).

---

## Practical Examples

### 1. Default Structural Diff
```bash
gojsondiff original.json modified.json
```

**Output:**
```
JSON DIFF
  age
    - 40
    + 41
  city
    - "Delhi"
  country
    + "India"

Summary: 3 total changes (1 added, 1 removed, 1 modified, 0 type changed)
```

---

### 2. Nested Objects & Array Diff
For nested objects and array elements, `gojsondiff` pinpoints exact paths:

```bash
gojsondiff old_users.json new_users.json
```

**Output:**
```
JSON DIFF
  users[1].name
    - "B"
    + "Bob"
  users[2]
    + {
      "id": 3,
      "name": "C"
    }

Summary: 2 total changes (1 added, 0 removed, 1 modified, 0 type changed)
```

---

### 3. Type Mutation Tracking (`-type`)
Track when types change between documents (especially useful for API schema changes):

```bash
gojsondiff -type old_api.json new_api.json
```

**Output:**
```
JSON DIFF
  age
    ~ type: number → string
    - 40
    + "40"
  enabled
    ~ type: boolean → null
    - true
    + null
```

---

### 4. Hierarchical Tree / Inline View (`-inline`)
Visualize changes across large, deeply nested JSON documents in an ASCII tree:

```bash
gojsondiff -inline old_profile.json new_profile.json
```

**Output:**
```
JSON DIFF
user
├── name
│   └── = "Atur"
├── age
│   └── ~ 40 → 41
├── address
│   ├── city
│   │   └── ~ "Delhi" → "Noida"
│   └── zip
│       └── - "110001"
└── email
    └── + "atur@example.com"
```
Where:
- `+` Field added
- `-` Field removed
- `~` Value or type changed
- `=` Unchanged sibling in modified parent

---

### 5. Reading from Stdin (Pipes & Subshells)

Comparing a local file against an API response:
```bash
curl -s https://api.service.internal/config | gojsondiff expected_config.json -
```

Comparing two API endpoints using process substitution (Bash/Zsh):
```bash
gojsondiff <(curl -s https://api.prod/v1/user/1) <(curl -s https://api.staging/v1/user/1)
```

---

### 6. Machine-Readable JSON Output

Emit full change metadata in JSON format for parsing with `jq` or feeding into automated downstream tools:

```bash
gojsondiff -format json old.json new.json
```

---

### 7. CI/CD Integration

Fail a GitHub Actions or GitLab pipeline if an API response deviates from its schema/fixture:

```bash
#!/usr/bin/env bash
set -e

gojsondiff fixtures/payload.json actual_response.json
status=$?

if [ $status -eq 0 ]; then
  echo "Payload verified successfully."
elif [ $status -eq 1 ]; then
  echo "Payload difference detected!"
  exit 1
else
  echo "Error running diff."
  exit 2
fi
```

---

## Sample Diff Files (GOJSONDIFF-6)

Pre-generated reference diff files demonstrating each output mode are located in [`examples/`](file:///home/mint/golang_toolshed/goJsonDiff/examples/):

- [`examples/sample_structural_diff.txt`](file:///home/mint/golang_toolshed/goJsonDiff/examples/sample_structural_diff.txt): Default structural format with `JSON DIFF` header, path indicators, and summary.
- [`examples/sample_type_diff.txt`](file:///home/mint/golang_toolshed/goJsonDiff/examples/sample_type_diff.txt): Type mutation format using `-type`.
- [`examples/sample_inline_tree_diff.txt`](file:///home/mint/golang_toolshed/goJsonDiff/examples/sample_inline_tree_diff.txt): Hierarchical tree representation using `-inline`.
- [`examples/sample_diff.diff`](file:///home/mint/golang_toolshed/goJsonDiff/examples/sample_diff.diff): Diff output formatted with `.diff` extension for diff viewers.
