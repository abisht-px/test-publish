# Test Documentation Tool

A simple tool to generate test documentation from comments in test files. It produces a JSON data.

## Publish Test Documents

1. The tool prints the output if publish flag is marked as false.
2. It can also publish the test doc to testrail when publish flag is marked as 'true'.
3. User and Api key secrets can be retrieved from AWS secrets manager object 'testrail-api-secret'.
4. Current section id is 9074 where all the test cases are uploaded.

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
    "title": "",
    "description": "",
    "steps": "",
    "expected": ""
  }
]
```

## Usage

```shell
make doc publish=false
make doc publish=true testrailusername='' testrailapikey=''

```
