// +build tools

package main

import (
	_ "github.com/bflad/tfproviderdocs"
	_ "github.com/bflad/tfproviderlint/cmd/tfproviderlint"
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/katbyte/terrafmt"
)
