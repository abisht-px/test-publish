# Test Documentation Tool

A simple tool to generate test documentation from comments in test files. It produces a JSON data

Expected format of comments -

```go
package testdoc

import(
    "testing"
)

// TestName <description>>.
// Steps:
// <test steps>
// Expected:
// <test expectations>
func TestName(t *testing.T) {}
```

Output Format -

```json
[
  {
    "name": "",
    "description": "",
    "steps": "",
    "expected": ""
  }
]
```

## Usage

```shell
make doc
```
