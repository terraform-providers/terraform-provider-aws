module github.com/terraform-providers/terraform-provider-aws/tools

go 1.15

require (
	github.com/bflad/tfproviderdocs v0.9.1
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.39.0
	github.com/hashicorp/go-changelog v0.0.0-20201005170154-56335215ce3a
	github.com/hashicorp/go-getter v1.5.2 // indirect
	github.com/katbyte/terrafmt v0.3.0
	github.com/pavius/impi v0.0.3
	github.com/terraform-linters/tflint v0.28.1
)

replace github.com/katbyte/terrafmt => github.com/gdavison/terrafmt v0.3.1-0.20210204054728-84242796be99

replace github.com/hashicorp/go-changelog => github.com/breathingdust/go-changelog v0.0.0-20210127001721-f985d5709c15

// v1.5.1 was incorrectly built
exclude github.com/hashicorp/go-getter v1.5.1
