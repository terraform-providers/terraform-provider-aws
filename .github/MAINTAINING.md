# Maintaining the Terraform AWS Provider

<!-- TOC depthFrom:2 -->

- [Pull Requests](#pull-requests)
    - [Pull Request Review Process](#pull-request-review-process)
        - [Dependency Updates](#dependency-updates)
            - [AWS Go SDK Updates](#aws-go-sdk-updates)
            - [golangci-lint Updates](#golangci-lint-updates)
            - [Terraform Plugin SDK Updates](#terraform-plugin-sdk-updates)
            - [tfproviderdocs Updates](#tfproviderdocs-updates)
            - [tfproviderlint Updates](#tfproviderlint-updates)
            - [yaml.v2 Updates](#yaml-v2-updates)
    - [Pull Request Merge Process](#pull-request-merge-process)
    - [Pull Request Types to CHANGELOG](#pull-request-types-to-changelog)
- [Release Process](#release-process)

<!-- /TOC -->

## Pull Requests

### Pull Request Review Process

Notes for each type of pull request are (or will be) available in subsections below.

- If you plan to be responsible for the pull request through the merge/closure process, assign it to yourself
- Add `bug`, `enhancement`, `new-data-source`, `new-resource`, or `technical-debt` labels to match expectations from change
- Perform a quick scan of open issues and ensure they are referenced in the pull request description (e.g. `Closes #1234`, `Relates #5678`). Edit the description yourself and mention this to the author:

```markdown
This pull request appears to be related to/solve #1234, so I have edited the pull request description to denote the issue reference.
```

- Review the contents of the pull request and ensure the change follows the relevant section of the [Contributing Guide](https://github.com/terraform-providers/terraform-provider-aws/blob/master/.github/CONTRIBUTING.md#checklists-for-contribution)
- If the change is not acceptable, leave a long form comment about the reasoning and close the pull request
- If the change is acceptable with modifications, leave a pull request review marked using the `Request Changes` option (for maintainer pull requests with minor modification requests, giving feedback with the `Approve` option is recommended so they do not need to wait for another round of review)
- If the author is unresponsive for changes (by default we give two weeks), determine importance and level of effort to finish the pull request yourself including their commits or close the pull request
- Run relevant acceptance testing ([locally](https://github.com/terraform-providers/terraform-provider-aws/blob/master/.github/CONTRIBUTING.md#running-an-acceptance-test) or in TeamCity) against AWS Commercial and AWS GovCloud (US) to ensure no new failures are being introduced
- Approve the pull request with a comment outlining what steps you took that ensure the change is acceptable, e.g. acceptance testing output

``````markdown
Looks good, thanks @username! :rocket:

Output from acceptance testing in AWS Commercial:

```
--- PASS: TestAcc...
--- PASS: TestAcc...
```

Output from acceptance testing in AWS GovCloud (US):

```
--- PASS: TestAcc...
--- PASS: TestAcc...
```
``````

#### Dependency Updates

##### AWS Go SDK Updates

Almost exclusively, `github.com/aws/aws-sdk-go` updates are additive in nature. It is generally safe to only scan through them before approving and merging. If you have any concerns about any of the service client updates such as suspicious code removals in the update, or deprecations introduced, run the acceptance testing for potentially affected resources before merging.

Authentication changes:

Occassionally, there will be changes listed in the authentication pieces of the AWS Go SDK codebase, e.g. changes to `aws/session`. The AWS Go SDK `CHANGELOG` should include a relevant description of these changes under a heading such as `SDK Enhancements` or `SDK Bug Fixes`. If they seem worthy of a callout in the Terraform AWS Provider `CHANGELOG`, then upon merging we should include a similar message prefixed with the `provider` subsystem, e.g. `* provider: ...`.

Additionally, if a `CHANGELOG` addition seemed appropriate, this dependency and version should also be updated in the Terraform S3 Backend, which currently lives in Terraform Core. An example of this can be found with https://github.com/terraform-providers/terraform-provider-aws/pull/9305 and https://github.com/hashicorp/terraform/pull/22055.

CloudFront changes:

CloudFront service client updates have previously caused an issue when a new field introduced in the SDK was not included with Terraform and caused all requests to error (https://github.com/terraform-providers/terraform-provider-aws/issues/4091). As a precaution, if you see CloudFront updates, run all the CloudFront resource acceptance testing before merging (`TestAccAWSCloudFront`).

New Regions:

These are added to the AWS Go SDK `aws/endpoints/defaults.go` file and generally noted in the AWS Go SDK `CHANGELOG` as `aws/endpoints: Updated Regions`. Since April 2019, new regions added to AWS now require being explicitly enabled before they can be used. Examples of this can be found when `me-south-1` was announced:

- [Terraform AWS Provider issue](https://github.com/terraform-providers/terraform-provider-aws/issues/9545)
- [Terraform AWS Provider AWS Go SDK update pull request](https://github.com/terraform-providers/terraform-provider-aws/pull/9538)
- [Terraform AWS Provider data source update pull request](https://github.com/terraform-providers/terraform-provider-aws/pull/9547)
- [Terraform S3 Backend issue](https://github.com/hashicorp/terraform/issues/22254)
- [Terraform S3 Backend pull request](https://github.com/hashicorp/terraform/pull/22253)

Typically our process for new regions is as follows:

- Create new (if not existing) Terraform AWS Provider issue: Support Automatic Region Validation for `XX-XXXXX-#` (Location)
- Create new (if not existing) Terraform S3 Backend issue: backend/s3: Support Automatic Region Validation for `XX-XXXXX-#` (Location)
- [Enable the new region in an AWS testing account](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable) and verify AWS Go SDK update works with the new region with `export AWS_DEFAULT_REGION=XX-XXXXX-#` with the new region and run the `TestAccDataSourceAwsRegion_` acceptance testing or by building the provider and testing a configuration like the following:

```hcl
provider "aws" {
  region = "me-south-1"
}

data "aws_region" "current" {}

output "region" {
  value = data.aws_region.current.name
}
```

- Merge AWS Go SDK update in Terraform AWS Provider and close issue with the following information:

``````markdown
Support for automatic validation of this new region has been merged and will release with version <x.y.z> of the Terraform AWS Provider, later this week.

---

Please note that this new region requires [a manual process to enable](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). Once enabled in the console, it takes a few minutes for everything to work properly.

If the region is not enabled properly, or the enablement process is still in progress, you can receive errors like these:

```console
$ terraform apply

Error: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid.
    status code: 403, request id: 142f947b-b2c3-11e9-9959-c11ab17bcc63

  on main.tf line 1, in provider "aws":
   1: provider "aws" {
```

---

To use this new region before support has been added to Terraform AWS Provider version in use, you must disable the provider's automatic region validation via:

```hcl
provider "aws" {
  # ... potentially other configuration ...

  region                 = "me-south-1"
  skip_region_validation = true
}
```
``````

- Update the Terraform AWS Provider `CHANGELOG` with the following:

```markdown
NOTES:

* provider: Region validation now automatically supports the new `XX-XXXXX-#` (Location) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform AWS Provider will return errors during credential validation (e.g. `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`) or AWS operations will throw their own errors (e.g. `data.aws_availability_zones.current: Error fetching Availability Zones: AuthFailure: AWS was not able to validate the provided access credentials`). [GH-####]

ENHANCEMENTS:

* provider: Support automatic region validation for `XX-XXXXX-#` [GH-####]
```

- Follow the [Contributing Guide](https://github.com/terraform-providers/terraform-provider-aws/blob/master/.github/CONTRIBUTING.md#new-region) to submit updates for various data sources to support the new region
- Submit the dependency update to the Terraform S3 Backend by running the following:

```shell
go get github.com/aws/aws-sdk-go@v#.#.#
go mod tidy
go mod vendor
```

- Create a S3 Bucket in the new region and verify AWS Go SDK update works with new region by building the Terraform S3 Backend and testing a configuration like the following:

```hcl
terraform {
  backend "s3" {
    bucket = "XXX"
    key    = "test"
    region = "me-south-1"
  }
}

output "test" {
  value = timestamp()
}
```

- After approval, merge AWS Go SDK update in Terraform S3 Backend and close issue with the following information:

``````markdown
Support for automatic validation of this new region has been merged and will release with the next version of the Terraform.

This was verified on a build of Terraform with the update:

```hcl
terraform {
  backend "s3" {
    bucket = "XXX"
    key    = "test"
    region = "me-south-1"
  }
}

output "test" {
  value = timestamp()
}
```

Outputs:

```console
$ terraform init
...
Terraform has been successfully initialized!
```

---

Please note that this new region requires [a manual process to enable](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). Once enabled in the console, it takes a few minutes for everything to work properly.

If the region is not enabled properly, or the enablement process is still in progress, you can receive errors like these:

```console
$ terraform init

Initializing the backend...

Error: error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid.
```

---

To use this new region before this update is released, you must disable the Terraform S3 Backend's automatic region validation via:

```hcl
terraform {
  # ... potentially other configuration ...

  backend "s3" {
    # ... other configuration ...

    region                 = "me-south-1"
    skip_region_validation = true
  }
}
```
``````

- Update the Terraform S3 Backend `CHANGELOG` with the following:

```markdown
NOTES:

* backend/s3: Region validation now automatically supports the new `XX-XXXXX-#` (Location) region. For AWS operations to work in the new region, the region must be explicitly enabled as outlined in the [AWS Documentation](https://docs.aws.amazon.com/general/latest/gr/rande-manage.html#rande-manage-enable). When the region is not enabled, the Terraform S3 Backend will return errors during credential validation (e.g. `error validating provider credentials: error calling sts:GetCallerIdentity: InvalidClientTokenId: The security token included in the request is invalid`). [GH-####]

ENHANCEMENTS:

* backend/s3: Support automatic region validation for `XX-XXXXX-#` [GH-####]
```

##### golangci-lint Updates

Merge if CI passes.

##### Terraform Plugin SDK Updates

Except for trivial changes, run the full acceptance testing suite against the pull request and verify there are no new or unexpected failures.

##### tfproviderdocs Updates

Merge if CI passes.

##### tfproviderlint Updates

Merge if CI passes.

##### yaml.v2 Updates

Run the acceptance testing pattern, `TestAccAWSCloudFormationStack(_dataSource)?_yaml`, and merge if passing.

### Pull Request Merge Process

- Add this pull request to the upcoming release milestone
- Add any linked issues that will be closed by the pull request to the same upcoming release milestone
- Merge the pull request
- Delete the branch (if the branch is on this repository)
- Determine if the pull request should have a CHANGELOG entry by reviewing the [Pull Request Types to CHANGELOG section](#pull-request-types-to-changelog). If so, update the repository `CHANGELOG.md` by directly committing to the `master` branch (e.g. editing the file in the GitHub web interface). See also the [Extending Terraform documentation](https://www.terraform.io/docs/extend/best-practices/versioning.html) for more information about the expected CHANGELOG format.
- Leave a comment on any issues closed by the pull request noting that it has been merged and when to expect the release containing it, e.g.

```markdown
The fix for this has been merged and will release with version X.Y.Z of the Terraform AWS Provider, expected in the XXX timeframe.
```

### Pull Request Types to CHANGELOG

The CHANGELOG is intended to show operator-impacting changes to the codebase for a particular version. If every change or commit to the code resulted in an entry, the CHANGELOG would become less useful for operators. The lists below are general guidelines on when a decision needs to be made to decide whether a change should have an entry.

Changes that should have a CHANGELOG entry:

- New Resources and Data Sources
- New full-length documentation guides (e.g. EKS Getting Started Guide, IAM Policy Documents with Terraform)
- Resource and provider bug fixes
- Resource and provider enhancements
- Deprecations
- Removals

Changes that may have a CHANGELOG entry:

- Dependency updates: If the update contains relevant bug fixes or enhancements that affect operators, those should be called out.

Changes that should _not_ have a CHANGELOG entry:

- Resource and provider documentation updates
- Testing updates

## Release Process

- Create a milestone for the next release after this release (generally, the next milestone will be a minor version increase unless previously decided for a major or patch version)
- Check the existing release milestone for open items and either work through them or move them to the next milestone
- Run the HashiCorp (non-OSS) TeamCity release job with the `DEPLOYMENT_TARGET_VERSION` matching the expected release milestone and `DEPLOYMENT_NEXT_VERSION` matching the next release milestone
- Wait for the TeamCity release job and CircleCI website deployment jobs to complete either by watching the build logs or Slack notifications
- Close the release milestone
- Create a new GitHub release with the release title exactly matching the tag and milestone (e.g. `v2.22.0`) and copy the notes from the CHANGELOG to the release notes. This will trigger [HashiBot](https://github.com/apps/hashibot) release comments.
