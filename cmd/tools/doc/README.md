# Test Documentation Tool

A simple tool to generate test documentation from comments in test files. It produces a JSON data and prints it when publish flag is marked as 'false'. 

## Publish Test Documents

It can also publish the test doc to testrail when publish flag is marked as 'true'. User and Api key secrets can be retrieved from AWS secrets manager object 'testrail-api-secret'.

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
