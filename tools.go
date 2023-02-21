//go:build tools
// +build tools

package main

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint" // Go linter.
	_ "golang.org/x/tools/cmd/goimports"                    // Go import formatter.
)
