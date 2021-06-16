// +build tools

package main

import (
	_ "github.com/bflad/tfproviderdocs"
	_ "github.com/client9/misspell/cmd/misspell"
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "github.com/hashicorp/go-changelog/cmd/changelog-build"
	_ "github.com/katbyte/terrafmt"
	_ "github.com/pavius/impi/cmd/impi"
	_ "github.com/terraform-linters/tflint"
)
