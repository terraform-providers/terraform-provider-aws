module github.com/terraform-providers/terraform-provider-aws/tools

go 1.15

require (
	github.com/bflad/tfproviderdocs v0.8.0
	github.com/client9/misspell v0.3.4
	github.com/golangci/golangci-lint v1.35.2
	github.com/hashicorp/go-changelog v0.0.0-20201005170154-56335215ce3a
	github.com/katbyte/terrafmt v0.2.1-0.20200913185704-5ff4421407b4
	github.com/pavius/impi v0.0.3 // indirect
	github.com/terraform-linters/tflint v0.20.3
)

replace github.com/katbyte/terrafmt => github.com/gdavison/terrafmt v0.2.1-0.20201026181004-a896893cd6af

replace github.com/hashicorp/go-changelog => github.com/breathingdust/go-changelog v0.0.0-20210127001721-f985d5709c15
