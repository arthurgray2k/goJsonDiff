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

### 1. Basic Comparison of Local Files
```bash
gojsondiff config.dev.json config.prod.json
```

**Output:**
```diff
- $.debug: true
~ $.host: "localhost" -> "prod.internal"
* $.port: 3000 (number) -> "3000" (string)
+ $.ssl: true

Summary: 4 total changes (1 added, 1 removed, 1 modified, 1 type changed)
```

---

### 2. Reading from Stdin (Pipes & Subshells)

Comparing a local file against an API response:
```bash
curl -s https://api.service.internal/config | gojsondiff expected_config.json -
```

Comparing two API endpoints using process substitution (Bash/Zsh):
```bash
gojsondiff <(curl -s https://api.prod/v1/user/1) <(curl -s https://api.staging/v1/user/1)
```

---

### 3. Machine-Readable JSON Output

Emit full change metadata in JSON format for parsing with `jq` or feeding into automated downstream tools:

```bash
gojsondiff -format json old.json new.json
```

**Output:**
```json
{
  "equal": false,
  "differences": [
    {
      "path": "$.rate_limit",
      "kind": "modified",
      "old_value": 1000,
      "new_value": 5000,
      "old_type": "number",
      "new_type": "number"
    }
  ],
  "summary": {
    "total_changes": 1,
    "added": 0,
    "removed": 0,
    "modified": 1,
    "type_changed": 0
  }
}
```

---

### 4. CI/CD Integration

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

### 5. Color Highlighting Control

Force colorization when piping to a pager (like `less -R`):
```bash
gojsondiff -color always docA.json docB.json | less -R
```

Disable colorization explicitly for log files:
```bash
gojsondiff -color never docA.json docB.json > diff.log
```
