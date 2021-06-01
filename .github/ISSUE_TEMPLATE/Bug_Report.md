---
name: 🐛 Bug Report
about: If something isn't working as expected 🤔.

---

<!---
Please note the following potential times when an issue might be in Terraform core:

* [Configuration Language](https://www.terraform.io/docs/configuration/index.html) or resource ordering issues
* [State](https://www.terraform.io/docs/state/index.html) and [State Backend](https://www.terraform.io/docs/backends/index.html) issues
* [Provisioner](https://www.terraform.io/docs/provisioners/index.html) issues
* [Registry](https://registry.terraform.io/) issues
* Spans resources across multiple providers

If you are running into one of these scenarios, we recommend opening an issue in the [Terraform core repository](https://github.com/hashicorp/terraform/) instead.
--->

<!--- Please keep this note for the community --->

### Community Note

* Please vote on this issue by adding a 👍 [reaction](https://blog.github.com/2016-03-10-add-reactions-to-pull-requests-issues-and-comments/) to the original issue to help the community and maintainers prioritize this request
* Please do not leave "+1" or other comments that do not add relevant new information or questions, they generate extra noise for issue followers and do not help prioritize the request
* If you are interested in working on this issue or have submitted a pull request, please leave a comment

<!--- Thank you for keeping this note for the community --->

### Terraform CLI and Terraform AWS Provider Version

<!--- Please run `terraform -v` to show the Terraform core version and provider version(s). If you are not running the latest version of Terraform or the provider, please upgrade because your issue may have already been fixed. [Terraform documentation on provider versioning](https://www.terraform.io/docs/configuration/providers.html#provider-versions). --->

### Affected Resource(s)

<!--- Please list the affected resources and data sources. --->

* aws_XXXXX

### Terraform Configuration Files

<!--- Information about code formatting: https://help.github.com/articles/basic-writing-and-formatting-syntax/#quoting-code --->

Please include all Terraform configurations required to reproduce the bug. Bug reports without a functional reproduction may be closed without investigation.

```hcl
# Copy-paste your Terraform configurations here - for large Terraform configs,
# please use a service like Dropbox and share a link to the ZIP file. For
# security, you can also encrypt the files using our GPG public key: https://keybase.io/hashicorp
```

### Debug Output

<!---
Please provide a link to a GitHub Gist containing the complete debug output. Please do NOT paste the debug output in the issue; just paste a link to the Gist.

To obtain the debug output, see the [Terraform documentation on debugging](https://www.terraform.io/docs/internals/debugging.html).
--->

### Panic Output

<!--- If Terraform produced a panic, please provide a link to a GitHub Gist containing the output of the `crash.log`. --->

### Expected Behavior

<!--- What should have happened? --->

### Actual Behavior

<!--- What actually happened? --->

### Steps to Reproduce

<!--- Please list the steps required to reproduce the issue. --->

1. `terraform apply`

### Important Factoids

<!--- Are there anything atypical about your accounts that we should know? For example: Running in EC2 Classic? --->

### References

<!---
Information about referencing Github Issues: https://help.github.com/articles/basic-writing-and-formatting-syntax/#referencing-issues-and-pull-requests

Are there any other GitHub issues (open or closed) or pull requests that should be linked here? Vendor documentation? For example:
--->

* #0000
