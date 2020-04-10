# Contributing to Terraform - AWS Provider

**First:** if you're unsure or afraid of _anything_, ask for help! You can
submit a work in progress (WIP) pull request, or file an issue with the parts
you know. We'll do our best to guide you in the right direction, and let you
know if there are guidelines we will need to follow. We want people to be able
to participate without fear of doing the wrong thing.

Below are our expectations for contributors. Following these guidelines gives us
the best opportunity to work with you, by making sure we have the things we need
in order to make it happen. Doing your best to follow it will speed up our
ability to merge PRs and respond to issues.

<!-- TOC depthFrom:2 -->

- [Issues](#issues)
    - [Issue Reporting Checklists](#issue-reporting-checklists)
        - [Bug Reports](#bug-reports)
        - [Feature Requests](#feature-requests)
        - [Questions](#questions)
    - [Issue Lifecycle](#issue-lifecycle)
- [Pull Requests](#pull-requests)
    - [Pull Request Lifecycle](#pull-request-lifecycle)
    - [Checklists for Contribution](#checklists-for-contribution)
        - [Documentation Update](#documentation-update)
        - [Enhancement/Bugfix to a Resource](#enhancementbugfix-to-a-resource)
        - [Adding Resource Import Support](#adding-resource-import-support)
        - [Adding Resource Name Generation Support](#adding-resource-name-generation-support)
        - [Adding Resource Tagging Support](#adding-resource-tagging-support)
        - [New Resource](#new-resource)
        - [New Service](#new-service)
        - [New Region](#new-region)
    - [Common Review Items](#common-review-items)
        - [Go Coding Style](#go-coding-style)
        - [Resource Contribution Guidelines](#resource-contribution-guidelines)
        - [Acceptance Testing Guidelines](#acceptance-testing-guidelines)
    - [Writing Acceptance Tests](#writing-acceptance-tests)
        - [Acceptance Tests Often Cost Money to Run](#acceptance-tests-often-cost-money-to-run)
        - [Running an Acceptance Test](#running-an-acceptance-test)
        - [Writing an Acceptance Test](#writing-an-acceptance-test)
        - [Writing and running Cross-Account Acceptance Tests](#writing-and-running-cross-account-acceptance-tests)
        - [Writing and running Cross-Region Acceptance Tests](#writing-and-running-cross-region-acceptance-tests)

<!-- /TOC -->

## Issues

### Issue Reporting Checklists

We welcome issues of all kinds including feature requests, bug reports, and
general questions. Below you'll find checklists with guidelines for well-formed
issues of each type.

#### [Bug Reports](https://github.com/terraform-providers/terraform-provider-aws/issues/new?template=Bug_Report.md)

 - [ ] __Test against latest release__: Make sure you test against the latest
   released version. It is possible we already fixed the bug you're experiencing.

 - [ ] __Search for possible duplicate reports__: It's helpful to keep bug
   reports consolidated to one thread, so do a quick search on existing bug
   reports to check if anybody else has reported the same thing. You can [scope
      searches by the label "bug"](https://github.com/terraform-providers/terraform-provider-aws/issues?q=is%3Aopen+is%3Aissue+label%3Abug) to help narrow things down.

 - [ ] __Include steps to reproduce__: Provide steps to reproduce the issue,
   along with your `.tf` files, with secrets removed, so we can try to
   reproduce it. Without this, it makes it much harder to fix the issue.

 - [ ] __For panics, include `crash.log`__: If you experienced a panic, please
   create a [gist](https://gist.github.com) of the *entire* generated crash log
   for us to look at. Double check no sensitive items were in the log.

#### [Feature Requests](https://github.com/terraform-providers/terraform-provider-aws/issues/new?labels=enhancement&template=Feature_Request.md)

 - [ ] __Search for possible duplicate requests__: It's helpful to keep requests
   consolidated to one thread, so do a quick search on existing requests to
   check if anybody else has reported the same thing. You can [scope searches by
      the label "enhancement"](https://github.com/terraform-providers/terraform-provider-aws/issues?q=is%3Aopen+is%3Aissue+label%3Aenhancement) to help narrow things down.

 - [ ] __Include a use case description__: In addition to describing the
   behavior of the feature you'd like to see added, it's helpful to also lay
   out the reason why the feature would be important and how it would benefit
   Terraform users.

#### [Questions](https://github.com/terraform-providers/terraform-provider-aws/issues/new?labels=question&template=Question.md)

 - [ ] __Search for answers in Terraform documentation__: We're happy to answer
   questions in GitHub Issues, but it helps reduce issue churn and maintainer
   workload if you work to [find answers to common questions in the
   documentation](https://www.terraform.io/docs/providers/aws/index.html). Oftentimes Question issues result in documentation updates
   to help future users, so if you don't find an answer, you can give us
   pointers for where you'd expect to see it in the docs.

### Issue Lifecycle

1. The issue is reported.

2. The issue is verified and categorized by a Terraform collaborator.
   Categorization is done via GitHub labels. We generally use a two-label
   system of (1) issue/PR type, and (2) section of the codebase. Type is
   one of "bug", "enhancement", "documentation", or "question", and section
   is usually the AWS service name.

3. An initial triage process determines whether the issue is critical and must
    be addressed immediately, or can be left open for community discussion.

4. The issue is addressed in a pull request or commit. The issue number will be
   referenced in the commit message so that the code that fixes it is clearly
   linked.

5. The issue is closed. Sometimes, valid issues will be closed because they are
   tracked elsewhere or non-actionable. The issue is still indexed and
   available for future viewers, or can be re-opened if necessary.

## Pull Requests

We appreciate direct contributions to the provider codebase. Here's what to
expect:

 * For pull requests that follow the guidelines, we will proceed to reviewing
  and merging, following the provider team's review schedule. There may be some
  internal or community discussion needed before we can complete this.
 * Pull requests that don't follow the guidelines will be commented with what
  they're missing. The person who submits the pull request or another community
  member will need to address those requests before they move forward.

### Pull Request Lifecycle

1. [Fork the GitHub repository](https://help.github.com/en/articles/fork-a-repo),
   modify the code, and [create a pull request](https://help.github.com/en/articles/creating-a-pull-request-from-a-fork).
   You are welcome to submit your pull request for commentary or review before
   it is fully completed by creating a [draft pull request](https://help.github.com/en/articles/about-pull-requests#draft-pull-requests)
   or adding `[WIP]` to the beginning of the pull request title.
   Please include specific questions or items you'd like feedback on.

1. Once you believe your pull request is ready to be reviewed, ensure the
   pull request is not a draft pull request by [marking it ready for review](https://help.github.com/en/articles/changing-the-stage-of-a-pull-request)
   or removing `[WIP]` from the pull request title if necessary, and a
   maintainer will review it. Follow [the checklists below](#checklists-for-contribution)
   to help ensure that your contribution can be easily reviewed and potentially
   merged.

1. One of Terraform's provider team members will look over your contribution and
   either approve it or provide comments letting you know if there is anything
   left to do. We do our best to keep up with the volume of PRs waiting for
   review, but it may take some time depending on the complexity of the work.

1. Once all outstanding comments and checklist items have been addressed, your
   contribution will be merged! Merged PRs will be included in the next
   Terraform release. The provider team takes care of updating the CHANGELOG as
   they merge.

1. In some cases, we might decide that a PR should be closed without merging.
   We'll make sure to provide clear reasoning when this happens.

### Checklists for Contribution

There are several different kinds of contribution, each of which has its own
standards for a speedy review. The following sections describe guidelines for
each type of contribution.

#### Documentation Update

The [Terraform AWS Provider's website source][website] is in this repository
along with the code and tests. Below are some common items that will get
flagged during documentation reviews:

- [ ] __Reasoning for Change__: Documentation updates should include an explanation for why the update is needed.
- [ ] __Prefer AWS Documentation__: Documentation about AWS service features and valid argument values that are likely to update over time should link to AWS service user guides and API references where possible.
- [ ] __Large Example Configurations__: Example Terraform configuration that includes multiple resource definitions should be added to the repository `examples` directory instead of an individual resource documentation page. Each directory under `examples` should be self-contained to call `terraform apply` without special configuration.
- [ ] __Terraform Configuration Language Features__: Individual resource documentation pages and examples should refrain from highlighting particular Terraform configuration language syntax workarounds or features such as `variable`, `local`, `count`, and built-in functions.

#### Enhancement/Bugfix to a Resource

Working on existing resources is a great way to get started as a Terraform
contributor because you can work within existing code and tests to get a feel
for what to do.

In addition to the below checklist, please see the [Common Review
Items](#common-review-items) sections for more specific coding and testing
guidelines.

 - [ ] __Acceptance test coverage of new behavior__: Existing resources each
   have a set of [acceptance tests][acctests] covering their functionality.
   These tests should exercise all the behavior of the resource. Whether you are
   adding something or fixing a bug, the idea is to have an acceptance test that
   fails if your code were to be removed. Sometimes it is sufficient to
   "enhance" an existing test by adding an assertion or tweaking the config
   that is used, but it's often better to add a new test. You can copy/paste an
   existing test and follow the conventions you see there, modifying the test
   to exercise the behavior of your code.
 - [ ] __Documentation updates__: If your code makes any changes that need to
   be documented, you should include those doc updates in the same PR. This
   includes things like new resource attributes or changes in default values.
   The [Terraform website][website] source is in this repo and includes
   instructions for getting a local copy of the site up and running if you'd
   like to preview your changes.
 - [ ] __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`. (The
   Travis CI build will fail if `go fmt` has not been run on incoming code.)
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
 - [ ] __Vendor additions__: Create a separate PR if you are updating the vendor
   folder. This is to avoid conflicts as the vendor versions tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

#### Adding Resource Import Support

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform. This type of enhancement generally requires a small to moderate amount of code changes.

Comprehensive code examples and information about resource import support can be found in the [Extending Terraform documentation](https://www.terraform.io/docs/extend/resources/import.html).

In addition to the below checklist and the items noted in the Extending Terraform documentation, please see the [Common Review Items](#common-review-items) sections for more specific coding and testing guidelines.

- [ ] _Resource Code Implementation_: In the resource code (e.g. `aws/resource_aws_service_thing.go`), implementation of `Importer` `State` function
- [ ] _Resource Acceptance Testing Implementation_: In the resource acceptance testing (e.g. `aws/resource_aws_service_thing_test.go`), implementation of `TestStep`s with `ImportState: true`
- [ ] _Resource Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `Import` documentation section at the bottom of the page

#### Adding Resource Name Generation Support

Terraform AWS Provider resources can use shared logic to support and test name generation, where the operator can choose between an expected naming value, a generated naming value with a prefix, or a fully generated name.

Implementing name generation support for Terraform AWS Provider resources requires the following, each with its own section below:

- [ ] _Resource Name Generation Code Implementation_: In the resource code (e.g. `aws/resource_aws_service_thing.go`), implementation of `name_prefix` attribute, along with handling in `Create` function.
- [ ] _Resource Name Generation Testing Implementation_: In the resource acceptance testing (e.g. `aws/resource_aws_service_thing_test.go`), implementation of new acceptance test functions and configurations to exercise new naming logic.
- [ ] _Resource Name Generation Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `name_prefix` argument and update of `name` argument description.

##### Resource Name Generation Code Implementation

- In the resource Go file (e.g. `aws/resource_aws_service_thing.go`), add the following Go import: `"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"`
- In the resource schema, add the new `name_prefix` attribute and adjust the `name` attribute to be `Optional`, `Computed`, and `ConflictsWith` the `name_prefix` attribute. Ensure to keep any existing schema fields on `name` such as `ValidateFunc`. e.g.

```go
"name": {
  Type:          schema.TypeString,
  Optional:      true,
  Computed:      true,
  ForceNew:      true,
  ConflictsWith: []string{"name_prefix"},
},
"name_prefix": {
  Type:          schema.TypeString,
  Optional:      true,
  ForceNew:      true,
  ConflictsWith: []string{"name"},
},
```

- In the resource `Create` function, switch any calls from `d.Get("name").(string)` to instead use the `naming.Generate()` function, e.g.

```go
name := naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))

// ... in AWS Go SDK Input types, etc. use aws.String(name)
```

##### Resource Name Generation Testing Implementation

- In the resource testing (e.g. `aws/resource_aws_service_thing_test.go`), add the following Go import: `"github.com/terraform-providers/terraform-provider-aws/aws/internal/naming"`
- In the resource testing, implement two new tests named `_Name_Generated` and `_NamePrefix` with associated configurations, that verifies creating the resource without `name` and `name_prefix` arguments (for the former) and with only the `name_prefix` argument (for the latter). e.g.

```go
func TestAccAWSServiceThing_Name_Generated(t *testing.T) {
  var thing service.ServiceThing
  resourceName := "aws_service_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    Providers:    testAccProviders,
    CheckDestroy: testAccCheckAWSServiceThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAWSServiceThingConfigNameGenerated(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAWSServiceThingExists(resourceName, &thing),
          naming.TestCheckResourceAttrNameGenerated(resourceName, "name"),
        ),
      },
      // If the resource supports import:
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func TestAccAWSServiceThing_NamePrefix(t *testing.T) {
  var thing service.ServiceThing
  resourceName := "aws_service_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    Providers:    testAccProviders,
    CheckDestroy: testAccCheckAWSServiceThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAWSServiceThingConfigNamePrefix("tf-acc-test-prefix-"),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAWSServiceThingExists(resourceName, &thing),
          naming.TestCheckResourceAttrNameFromPrefix(resourceName, "name", "tf-acc-test-prefix-"),
        ),
      },
      // If the resource supports import:
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func testAccAWSServiceThingConfigNameGenerated() string {
  return fmt.Sprintf(`
resource "aws_service_thing" "test" {
  # ... other configuration ...
}
`)
}

func testAccAWSServiceThingConfigNamePrefix(namePrefix string) string {
  return fmt.Sprintf(`
resource "aws_service_thing" "test" {
  # ... other configuration ...

  name_prefix = %[1]q
}
`, namePrefix)
}
```

##### Resource Code Generation Documentation Implementation

- In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), add the following to the arguments reference:

```markdown
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
```

- Adjust the existing `name` argument reference to ensure its denoted as `Optional`, includes a mention that it can be generated, and that it conflicts with `name_prefix`:

```markdown
* `name` - (Optional) Name of the thing. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
```

#### Adding Resource Tagging Support

AWS provides key-value metadata across many services and resources, which can be used for a variety of use cases including billing, ownership, and more. See the [AWS Tagging Strategy page](https://aws.amazon.com/answers/account-management/aws-tagging-strategies/) for more information about tagging at a high level.

Implementing tagging support for Terraform AWS Provider resources requires the following, each with its own section below:

- [ ] _Generated Service Tagging Code_: In the internal code generators (e.g. `aws/internal/keyvaluetags`), implementation and customization of how a service handles tagging, which is standardized for the resources.
- [ ] _Resource Tagging Code Implementation_: In the resource code (e.g. `aws/resource_aws_service_thing.go`), implementation of `tags` schema attribute, along with handling in `Create`, `Read`, and `Update` functions.
- [ ] _Resource Tagging Acceptance Testing Implementation_: In the resource acceptance testing (e.g. `aws/resource_aws_service_thing_test.go`), implementation of new acceptance test function and configurations to exercise new tagging logic.
- [ ] _Resource Tagging Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `tags` argument

See also a [full example pull request for implementing EKS tagging](https://github.com/terraform-providers/terraform-provider-aws/pull/10307).

##### Adding Service to Tag Generating Code

This step is only necessary for the first implementation and may have been previously completed. If so, move on to the next section.

More details about this code generation, including fixes for potential error messages in this process, can be found in the [keyvaluetags documentation](../aws/internal/keyvaluetags/README.md).

- Open the AWS Go SDK documentation for the service, e.g. for [`service/eks`](https://docs.aws.amazon.com/sdk-for-go/api/service/eks/). Note: there can be a delay between the AWS announcement and the updated AWS Go SDK documentation.
- Determine the "type" of tagging implementation. Some services will use a simple map style (`map[string]*string` in Go) while others will have a separate structure shape (`[]service.Tag` struct with `Key` and `Value` fields).

  - If the type is a map, add the AWS Go SDK service name (e.g. `eks`) to `mapServiceNames` in `aws/internal/keyvaluetags/generators/servicetags/main.go`
  - Otherwise, if the type is a struct, add the AWS Go SDK service name (e.g. `eks`) to `sliceServiceNames` in `aws/internal/keyvaluetags/generators/servicetags/main.go`. If the struct name is not exactly `Tag`, it can be customized via the `ServiceTagType` function. If the struct key field is not exactly `Key`, it can be customized via the `ServiceTagTypeKeyField` function. If the struct value field is not exactly `Value`, it can be customized via the `ServiceTagTypeValueField` function.

- Determine if the service API includes functionality for listing tags (usually a `ListTags` or `ListTagsForResource` API call) or updating tags (usually `TagResource` and `UntagResource` API calls). If so, add the AWS Go SDK service client information to `ServiceClientType` (along with the new required import) in `aws/internal/keyvaluetags/service_generation_customizations.go`, e.g. for EKS:

  ```go
  case "eks":
    funcType = reflect.TypeOf(eks.New)
  ```

  - If the service API includes functionality for listing tags, add the AWS Go SDK service name (e.g. `eks`) to `serviceNames` in `aws/internal/keyvaluetags/generators/listtags/main.go`.
  - If the service API includes functionality for updating tags, add the AWS Go SDK service name (e.g. `eks`) to `serviceNames` in `aws/internal/keyvaluetags/generators/updatetags/main.go`.

- Run `make gen` (`go generate ./...`) and ensure there are no errors via `make test` (`go test ./...`)

##### Resource Tagging Code Implementation

- In the resource Go file (e.g. `aws/resource_aws_eks_cluster.go`), add the following Go import: `"github.com/terraform-providers/terraform-provider-aws/aws/internal/keyvaluetags"`
- In the resource schema, add `"tags": tagsSchema(),`
- If the API supports tagging on creation (the `Input` struct accepts a `Tags` field), in the resource `Create` function, implement the logic to convert the configuration tags into the service tags, e.g. with EKS Clusters:

  ```go
  input := &eks.CreateClusterInput{
    /* ... other configuration ... */
    Tags: keyvaluetags.New(d.Get("tags").(map[string]interface{})).IgnoreAws().EksTags(),
  }
  ```

  If the service API does not allow passing an empty list, the logic can be adjusted similar to:

  ```go
  input := &eks.CreateClusterInput{
    /* ... other configuration ... */
  }

  if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
    input.Tags = keyvaluetags.New(v).IgnoreAws().EksTags()
  }
  ```

- Otherwise if the API does not support tagging on creation (the `Input` struct does not accept a `Tags` field), in the resource `Create` function, implement the logic to convert the configuration tags into the service API call to tag a resource, e.g. with ElasticSearch Domain:

  ```go
if v := d.Get("tags").(map[string]interface{}); len(v) > 0 {
	if err := keyvaluetags.ElasticsearchserviceUpdateTags(conn, d.Id(), nil, v); err != nil {
		return fmt.Errorf("error adding Elasticsearch Cluster (%s) tags: %s", d.Id(), err)
	}
}
  ```

- Some EC2 resources (for example [`aws_ec2_fleet`](https://www.terraform.io/docs/providers/aws/r/ec2_fleet.html)) have a `TagsSpecification` field in the `InputStruct` instead of a `Tags` field. In these cases the `ec2TagSpecificationsFromMap()` helper function should be used, e.g.:

  ```go
  input := &ec2.CreateFleetInput{
    /* ... other configuration ... */
    TagSpecifications: ec2TagSpecificationsFromMap(d.Get("tags").(map[string]interface{}), ec2.ResourceTypeFleet),
  }
  ```

- In the resource `Read` function, implement the logic to convert the service tags to save them into the Terraform state for drift detection, e.g. with EKS Clusters (which had the tags available in the DescribeCluster API call):

  ```go
  if err := d.Set("tags", keyvaluetags.EksKeyValueTags(cluster.Tags).IgnoreAws().Map()); err != nil {
    return fmt.Errorf("error setting tags: %s", err)
  }
  ```

  If the service API does not return the tags directly from reading the resource and requires a separate API call, its possible to use the `keyvaluetags` functionality like the following, e.g. with Athena Workgroups:

  ```go
  tags, err := keyvaluetags.AthenaListTags(conn, arn.String())

  if err != nil {
    return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
  }

  if err := d.Set("tags", tags.IgnoreAws().Map()); err != nil {
    return fmt.Errorf("error setting tags: %s", err)
  }
  ```

- In the resource `Update` function (this may be the first functionality requiring the creation of the `Update` function), implement the logic to handle tagging updates, e.g. with EKS Clusters:

  ```go
  if d.HasChange("tags") {
    o, n := d.GetChange("tags")
    if err := keyvaluetags.EksUpdateTags(conn, d.Get("arn").(string), o, n); err != nil {
      return fmt.Errorf("error updating tags: %s", err)
    }
  }
  ```

##### Resource Tagging Acceptance Testing Implementation

- In the resource testing (e.g. `aws/resource_aws_eks_cluster_test.go`), verify that existing resources without tagging are unaffected and do not have tags saved into their Terraform state. This should be done in the `_basic` acceptance test by adding a line similar to `resource.TestCheckResourceAttr(resourceName, "tags.%s", "0"),`
- In the resource testing, implement a new test named `_Tags` with associated configurations, that verifies creating the resource with tags and updating tags. e.g. EKS Clusters:

  ```go
  func TestAccAWSEksCluster_Tags(t *testing.T) {
    var cluster1, cluster2, cluster3 eks.Cluster
    rName := acctest.RandomWithPrefix("tf-acc-test")
    resourceName := "aws_eks_cluster.test"

    resource.ParallelTest(t, resource.TestCase{
      PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEks(t) },
      Providers:    testAccProviders,
      CheckDestroy: testAccCheckAWSEksClusterDestroy,
      Steps: []resource.TestStep{
        {
          Config: testAccAWSEksClusterConfigTags1(rName, "key1", "value1"),
          Check: resource.ComposeTestCheckFunc(
            testAccCheckAWSEksClusterExists(resourceName, &cluster1),
            resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
            resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
          ),
        },
        {
          ResourceName:      resourceName,
          ImportState:       true,
          ImportStateVerify: true,
        },
        {
          Config: testAccAWSEksClusterConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
          Check: resource.ComposeTestCheckFunc(
            testAccCheckAWSEksClusterExists(resourceName, &cluster2),
            resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
            resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
            resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
          ),
        },
        {
          Config: testAccAWSEksClusterConfigTags1(rName, "key2", "value2"),
          Check: resource.ComposeTestCheckFunc(
            testAccCheckAWSEksClusterExists(resourceName, &cluster3),
            resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
            resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
          ),
        },
      },
    })
  }

  func testAccAWSEksClusterConfigTags1(rName, tagKey1, tagValue1 string) string {
    return testAccAWSEksClusterConfig_Base(rName) + fmt.Sprintf(`
  resource "aws_eks_cluster" "test" {
    name     = %[1]q
    role_arn = "${aws_iam_role.test.arn}"

    tags = {
      %[2]q = %[3]q
    }

    vpc_config {
      subnet_ids = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
    }

    depends_on = ["aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy", "aws_iam_role_policy_attachment.test-AmazonEKSServicePolicy"]
  }
  `, rName, tagKey1, tagValue1)
  }

  func testAccAWSEksClusterConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
    return testAccAWSEksClusterConfig_Base(rName) + fmt.Sprintf(`
  resource "aws_eks_cluster" "test" {
    name     = %[1]q
    role_arn = "${aws_iam_role.test.arn}"

    tags = {
      %[2]q = %[3]q
      %[4]q = %[5]q
    }

    vpc_config {
      subnet_ids = ["${aws_subnet.test.*.id[0]}", "${aws_subnet.test.*.id[1]}"]
    }

    depends_on = ["aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy", "aws_iam_role_policy_attachment.test-AmazonEKSServicePolicy"]
  }
  `, rName, tagKey1, tagValue1, tagKey2, tagValue2)
  }
  ```

- Verify all acceptance testing passes for the resource (e.g. `make testacc TESTARGS='-run=TestAccAWSEksCluster_'`)

#### Resource Tagging Documentation Implementation

- In the resource documentation (e.g. `website/docs/r/eks_cluster.html.markdown`), add the following to the arguments reference:

  ```markdown
  * `tags` - (Optional) Key-value mapping of resource tags
  ```

#### New Resource

Implementing a new resource is a good way to learn more about how Terraform
interacts with upstream APIs. There are plenty of examples to draw from in the
existing resources, but you still get to implement something completely new.

In addition to the below checklist, please see the [Common Review
Items](#common-review-items) sections for more specific coding and testing
guidelines.

 - [ ] __Minimal LOC__: It's difficult for both the reviewer and author to go
   through long feedback cycles on a big PR with many resources. We ask you to
   only submit **1 resource at a time**.
 - [ ] __Acceptance tests__: New resources should include acceptance tests
   covering their behavior. See [Writing Acceptance
   Tests](#writing-acceptance-tests) below for a detailed guide on how to
   approach these.
 - [ ] __Resource Naming__: Resources should be named `aws_<service>_<name>`,
   using underscores (`_`) as the separator. Resources are namespaced with the
   service name to allow easier searching of related resources, to align
   the resource naming with the service for [Customizing Endpoints](https://www.terraform.io/docs/providers/aws/guides/custom-service-endpoints.html#available-endpoint-customizations),
   and to prevent future conflicts with new AWS services/resources.
   For reference:

   - `service` is the AWS short service name that matches the entry in
     `endpointServiceNames` (created via the [New Service](#new-service)
     section)
   - `name` represents the conceptual infrastructure represented by the
     create, read, update, and delete methods of the service API. It should
     be a singular noun. For example, in an API that has methods such as
     `CreateThing`, `DeleteThing`, `DescribeThing`, and `ModifyThing` the name
     of the resource would end in `_thing`.

 - [ ] __Arguments_and_Attributes__: The HCL for arguments and attributes should
   mimic the types and structs presented by the AWS API. API arguments should be
   converted from `CamelCase` to `camel_case`.
 - [ ] __Documentation__: Each resource gets a page in the Terraform
   documentation. The [Terraform website][website] source is in this
   repo and includes instructions for getting a local copy of the site up and
   running if you'd like to preview your changes. For a resource, you'll want
   to add a new file in the appropriate place and add a link to the sidebar for
   that page.
 - [ ] __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`. (The
   Travis CI build will fail if `go fmt` has not been run on incoming code.)
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
 - [ ] __Vendor updates__: Create a separate PR if you are adding to the vendor
   folder. This is to avoid conflicts as the vendor versions tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

#### New Service

Implementing a new AWS service gives Terraform the ability to manage resources in
a whole new API. It's a larger undertaking, but brings major new functionality
into Terraform.

- [ ] __Service Client__: Before new resources are submitted, we request
  a separate pull request containing just the new AWS Go SDK service client.
  Doing so will pull the AWS Go SDK service code into the project at the
  current version. Since the AWS Go SDK is updated frequently, these pull
  requests can easily have merge conflicts or be out of date. The maintainers
  prioritize reviewing and merging these quickly to prevent those situations.

  To add the AWS Go SDK service client:

  - In `aws/provider.go` Add a new service entry to `endpointServiceNames`.
    This service name should match the AWS Go SDK or AWS CLI service name.
  - In `aws/config.go`: Add a new import for the AWS Go SDK code. e.g.
    `github.com/aws/aws-sdk-go/service/quicksight`
  - In `aws/config.go`: Add a new `{SERVICE}conn` field to the `AWSClient`
    struct for the service client. The service name should match the name
    in `endpointServiceNames`. e.g. `quicksightconn *quicksight.QuickSight`
  - In `aws/config.go`: Create the new service client in the `{SERVICE}conn`
    field in the `AWSClient` instantiation within `Client()`. e.g.
    `quicksightconn: quicksight.New(sess.Copy(&aws.Config{Endpoint: aws.String(c.Endpoints["quicksight"])})),`
  - In `website/allowed-subcategories.txt`: Add a name acceptable for the documentation navigation.
  - In `website/docs/guides/custom-service-endpoints.html.md`: Add the service
    name in the list of customizable endpoints.
  - In `.hashibot.hcl`: Add the new service to automated issue and pull request labeling. e.g. with the `quicksight` service

  ```hcl
  behavior "regexp_issue_labeler_v2" "service_labels" {
    # ... other configuration ...

    label_map = {
      # ... other services ...
      "service/quicksight" = [
        "aws_quicksight_",
      ],
      # ... other services ...
    }
  }

  behavior "pull_request_path_labeler" "service_labels"
    # ... other configuration ...

    label_map = {
      # ... other services ...
      "service/quicksight" = [
        "**/*_quicksight_*",
        "**/quicksight_*",
      ],
      # ... other services ...
    }
  }
  ```

  - Run the following then submit the pull request:

  ```sh
  go test ./aws
  go mod tidy
  go mod vendor
  ```

- [ ] __Initial Resource__: Some services can be big and it can be
  difficult for both reviewer & author to go through long feedback cycles
  on a big PR with many resources. Often feedback items in one resource
  will also need to be applied in other resources. We prefer you to submit
  the necessary minimum in a single PR, ideally **just the first resource**
  of the service.

The initial resource and changes afterwards should follow the other sections
of this guide as appropriate.

#### New Region

While region validation is automatically added with SDK updates, new regions
are generally limited in which services they support. Below are some
manually sourced values from documentation.

 - [ ] Check [Regions and Endpoints ELB regions](https://docs.aws.amazon.com/general/latest/gr/rande.html#elb_region) and add Route53 Hosted Zone ID if available to `aws/data_source_aws_elb_hosted_zone_id.go`
 - [ ] Check [Regions and Endpoints S3 website endpoints](https://docs.aws.amazon.com/general/latest/gr/rande.html#s3_website_region_endpoints) and add Route53 Hosted Zone ID if available to `aws/hosted_zones.go`
 - [ ] Check [CloudTrail Supported Regions docs](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html#cloudtrail-supported-regions) and add AWS Account ID if available to `aws/data_source_aws_cloudtrail_service_account.go`
 - [ ] Check [Elastic Load Balancing Access Logs docs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy) and add Elastic Load Balancing Account ID if available to `aws/data_source_aws_elb_service_account.go`
 - [ ] Check [Redshift Database Audit Logging docs](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) and add AWS Account ID if available to `aws/data_source_aws_redshift_service_account.go`
 - [ ] Check [Regions and Endpoints Elastic Beanstalk](https://docs.aws.amazon.com/general/latest/gr/rande.html#elasticbeanstalk_region) and add Route53 Hosted Zone ID if available to `aws/data_source_aws_elastic_beanstalk_hosted_zone.go`

### Common Review Items

The Terraform AWS Provider follows common practices to ensure consistent and
reliable implementations across all resources in the project. While there may be
older resource and testing code that predates these guidelines, new submissions
are generally expected to adhere to these items to maintain Terraform Provider
quality. For any guidelines listed, contributors are encouraged to ask any
questions and community reviewers are encouraged to provide review suggestions
based on these guidelines to speed up the review and merge process.

#### Go Coding Style

The following Go language resources provide common coding preferences that may be referenced during review, if not automatically handled by the project's linting tools.

- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)

#### Resource Contribution Guidelines

The following resource checks need to be addressed before your contribution can be merged. The exclusion of any applicable check may result in a delayed time to merge.

- [ ] __Passes Testing__: All code and documentation changes must pass unit testing, code linting, and website link testing. Resource code changes must pass all acceptance testing for the resource.
- [ ] __Avoids API Calls Across Account, Region, and Service Boundaries__: Resources should not implement cross-account, cross-region, or cross-service API calls.
- [ ] __Avoids Optional and Required for Non-Configurable Attributes__: Resource schema definitions for read-only attributes should not include `Optional: true` or `Required: true`.
- [ ] __Avoids resource.Retry() without resource.RetryableError()__: Resource logic should only implement [`resource.Retry()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Retry) if there is a retryable condition (e.g. `return resource.RetryableError(err)`).
- [ ] __Avoids Resource Read Function in Data Source Read Function__: Data sources should fully implement their own resource `Read` functionality including duplicating `d.Set()` calls.
- [ ] __Avoids Reading Schema Structure in Resource Code__: The resource `Schema` should not be read in resource `Create`/`Read`/`Update`/`Delete` functions to perform looping or otherwise complex attribute logic. Use [`d.Get()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Get) and [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) directly with individual attributes instead.
- [ ] __Avoids ResourceData.GetOkExists()__: Resource logic should avoid using [`ResourceData.GetOkExists()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.GetOkExists) as its expected functionality is not guaranteed in all scenarios.
- [ ] __Implements Read After Create and Update__: Except where API eventual consistency prohibits immediate reading of resources or updated attributes,  resource `Create` and `Update` functions should return the resource `Read` function.
- [ ] __Implements Immediate Resource ID Set During Create__: Immediately after calling the API creation function, the resource ID should be set with [`d.SetId()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.SetId) before other API operations or returning the `Read` function.
- [ ] __Implements Attribute Refreshes During Read__: All attributes available in the API should have [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) called their values in the Terraform state during the `Read` function.
- [ ] __Implements Error Checks with Non-Primative Attribute Refreshes__: When using [`d.Set()`](https://godoc.org/github.com/hashicorp/terraform/helper/schema#ResourceData.Set) with non-primative types (`schema.TypeList`, `schema.TypeSet`, or `schema.TypeMap`), perform error checking to [prevent issues where the code is not properly able to refresh the Terraform state](https://www.terraform.io/docs/extend/best-practices/detecting-drift.html#error-checking-aggregate-types).
- [ ] __Implements Import Acceptance Testing and Documentation__: Support for resource import (`Importer` in resource schema) must include `ImportState` acceptance testing (see also the [Acceptance Testing Guidelines](#acceptance-testing-guidelines) below) and `## Import` section in resource documentation.
- [ ] __Implements Customizable Timeouts Documentation__: Support for customizable timeouts (`Timeouts` in resource schema) must include `## Timeouts` section in resource documentation.
- [ ] __Implements State Migration When Adding New Virtual Attribute__: For new "virtual" attributes (those only in Terraform and not in the API), the schema should implement [State Migration](https://www.terraform.io/docs/extend/resources.html#state-migrations) to prevent differences for existing configurations that upgrade.
- [ ] __Uses AWS Go SDK Constants__: Many AWS services provide string constants for value enumerations, error codes, and status types. See also the "Constants" sections under each of the service packages in the [AWS Go SDK documentation](https://docs.aws.amazon.com/sdk-for-go/api/).
- [ ] __Uses AWS Go SDK Pointer Conversion Functions__: Many APIs return pointer types and these functions return the zero value for the type if the pointer is `nil`. This prevents potential panics from unchecked `*` pointer dereferences and can eliminate boilerplate `nil` checking in many cases. See also the [`aws` package in the AWS Go SDK documentation](https://docs.aws.amazon.com/sdk-for-go/api/aws/).
- [ ] __Uses AWS Go SDK Types__: Use available SDK structs instead of implementing custom types with indirection.
- [ ] __Uses TypeList and MaxItems: 1__: Configuration block attributes (e.g. `Type: schema.TypeList` or `Type: schema.TypeSet` with `Elem: &schema.Resource{...}`) that can only have one block should use `Type: schema.TypeList` and `MaxItems: 1` in the schema definition.
- [ ] __Uses Existing Validation Functions__: Schema definitions including `ValidateFunc` for attribute validation should use available [Terraform `helper/validation` package](https://godoc.org/github.com/hashicorp/terraform/helper/validation) functions. `All()`/`Any()` can be used for combining multiple validation function behaviors.
- [ ] __Uses isResourceTimeoutError() with resource.Retry()__: Resource logic implementing [`resource.Retry()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Retry) should error check with `isResourceTimeoutError(err error)` and potentially unset the error before returning the error. For example:

  ```go
  var output *kms.CreateKeyOutput
  err := resource.Retry(1*time.Minute, func() *resource.RetryError {
    var err error

    output, err = conn.CreateKey(input)

    /* ... */

    return nil
  })

  if isResourceTimeoutError(err) {
    output, err = conn.CreateKey(input)
  }

  if err != nil {
    return fmt.Errorf("error creating KMS External Key: %s", err)
  }
  ```

- [ ] __Uses resource.NotFoundError__: Custom errors for missing resources should use [`resource.NotFoundError`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#NotFoundError).
- [ ] __Uses resource.UniqueId()__: API fields for concurrency protection such as `CallerReference` and `IdempotencyToken` should use [`resource.UniqueId()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#UniqueId). The implementation includes a monotonic counter which is safer for concurrent operations than solutions such as `time.Now()`.
- [ ] __Skips Exists Function__: Implementing a resource `Exists` function is extraneous as it often duplicates resource `Read` functionality. Ensure `d.SetId("")` is used to appropriately trigger resource recreation in the resource `Read` function.
- [ ] __Skips id Attribute__: The `id` attribute is implicit for all Terraform resources and does not need to be defined in the schema.

The below are style-based items that _may_ be noted during review and are recommended for simplicity, consistency, and quality assurance:

- [ ] __Avoids CustomizeDiff__: Usage of `CustomizeDiff` is generally discouraged.
- [ ] __Implements Error Message Context__: Returning errors from resource `Create`, `Read`, `Update`, and `Delete` functions should include additional messaging about the location or cause of the error for operators and code maintainers by wrapping with [`fmt.Errorf()`](https://godoc.org/golang.org/x/exp/errors/fmt#Errorf).
  - An example `Delete` API error: `return fmt.Errorf("error deleting {SERVICE} {THING} (%s): %s", d.Id(), err)`
  - An example `d.Set()` error: `return fmt.Errorf("error setting {ATTRIBUTE}: %s", err)`
- [ ] __Implements arn Attribute__: APIs that return an Amazon Resource Name (ARN), should implement `arn` as an attribute.
- [ ] __Implements Warning Logging With Resource State Removal__: If a resource is removed outside of Terraform (e.g. via different tool, API, or web UI), `d.SetId("")` and `return nil` can be used in the resource `Read` function to trigger resource recreation. When this occurs, a warning log message should be printed beforehand: `log.Printf("[WARN] {SERVICE} {THING} (%s) not found, removing from state", d.Id())`
- [ ] __Uses isAWSErr() with AWS Go SDK Error Objects__: Use the available `isAWSErr(err error, code string, message string)` helper function instead of the `awserr` package to compare error code and message contents.
- [ ] __Uses %s fmt Verb with AWS Go SDK Objects__: AWS Go SDK objects implement `String()` so using the `%v`, `%#v`, or `%+v` fmt verbs with the object are extraneous or provide unhelpful detail.
- [ ] __Uses Elem with TypeMap__: While provider schema validation does not error when the `Elem` configuration is not present with `Type: schema.TypeMap` attributes, including the explicit `Elem: &schema.Schema{Type: schema.TypeString}` is recommended.
- [ ] __Uses American English for Attribute Naming__: For any ambiguity with attribute naming, prefer American English over British English. e.g. `color` instead of `colour`.
- [ ] __Skips Timestamp Attributes__: Generally, creation and modification dates from the API should be omitted from the schema.
- [ ] __Skips Error() Call with AWS Go SDK Error Objects__: Error objects do not need to have `Error()` called.

#### Acceptance Testing Guidelines

The below are required items that will be noted during submission review and prevent immediate merging:

- [ ] __Implements CheckDestroy__: Resource testing should include a `CheckDestroy` function (typically named `testAccCheckAws{SERVICE}{RESOURCE}Destroy`) that calls the API to verify that the Terraform resource has been deleted or disassociated as appropriate. More information about `CheckDestroy` functions can be found in the [Extending Terraform TestCase documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy).
- [ ] __Implements Exists Check Function__: Resource testing should include a `TestCheckFunc` function (typically named `testAccCheckAws{SERVICE}{RESOURCE}Exists`) that calls the API to verify that the Terraform resource has been created or associated as appropriate. Preferably, this function will also accept a pointer to an API object representing the Terraform resource from the API response that can be set for potential usage in later `TestCheckFunc`. More information about these functions can be found in the [Extending Terraform Custom Check Functions documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/testcase.html#checkdestroy).
- [ ] __Excludes Provider Declarations__: Test configurations should not include `provider "aws" {...}` declarations. If necessary, only the provider declarations in `provider_test.go` should be used for multiple account/region or otherwise specialized testing.
- [ ] __Passes in us-west-2 Region__: Tests default to running in `us-west-2` and at a minimum should pass in that region or include necessary `PreCheck` functions to skip the test when ran outside an expected environment.
- [ ] __Uses resource.ParallelTest__: Tests should utilize [`resource.ParallelTest()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#ParallelTest) instead of [`resource.Test()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#Test) except where serialized testing is absolutely required.
- [ ] __Uses fmt.Sprintf()__: Test configurations preferably should to be separated into their own functions (typically named `testAccAws{SERVICE}{RESOURCE}Config{PURPOSE}`) that call [`fmt.Sprintf()`](https://golang.org/pkg/fmt/#Sprintf) for variable injection or a string `const` for completely static configurations. Test configurations should avoid `var` or other variable injection functionality such as [`text/template`](https://golang.org/pkg/text/template/).
- [ ] __Uses Randomized Infrastructure Naming__: Test configurations that utilize resources where a unique name is required should generate a random name. Typically this is created via `rName := acctest.RandomWithPrefix("tf-acc-test")` in the acceptance test function before generating the configuration.

For resources that support import, the additional item below is required that will be noted during submission review and prevent immediate merging:

- [ ] __Implements ImportState Testing__: Tests should include an additional `TestStep` configuration that verifies resource import via `ImportState: true` and `ImportStateVerify: true`. This `TestStep` should be added to all possible tests for the resource to ensure that all infrastructure configurations are properly imported into Terraform.

The below are style-based items that _may_ be noted during review and are recommended for simplicity, consistency, and quality assurance:

- [ ] __Uses Builtin Check Functions__: Tests should utilize already available check functions, e.g. `resource.TestCheckResourceAttr()`, to verify values in the Terraform state over creating custom `TestCheckFunc`. More information about these functions can be found in the [Extending Terraform Builtin Check Functions documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/teststep.html#builtin-check-functions).
- [ ] __Uses TestCheckResoureAttrPair() for Data Sources__: Tests should utilize [`resource.TestCheckResourceAttrPair()`](https://godoc.org/github.com/hashicorp/terraform/helper/resource#TestCheckResourceAttrPair) to verify values in the Terraform state for data sources attributes to compare them with their expected resource attributes.
- [ ] __Excludes Timeouts Configurations__: Test configurations should not include `timeouts {...}` configuration blocks except for explicit testing of customizable timeouts (typically very short timeouts with `ExpectError`).
- [ ] __Implements Default and Zero Value Validation__: The basic test for a resource (typically named `TestAccAws{SERVICE}{RESOURCE}_basic`) should utilize available check functions, e.g. `resource.TestCheckResourceAttr()`, to verify default and zero values in the Terraform state for all attributes. Empty/missing configuration blocks can be verified with `resource.TestCheckResourceAttr(resourceName, "{ATTRIBUTE}.#", "0")` and empty maps with `resource.TestCheckResourceAttr(resourceName, "{ATTRIBUTE}.%", "0")`

The below are location-based items that _may_ be noted during review and are recommended for consistency with testing flexibility. Resource testing is expected to pass across multiple AWS environments supported by the Terraform AWS Provider (e.g. AWS Standard and AWS GovCloud (US)). Contributors are not expected or required to perform testing outside of AWS Standard, e.g. running only in the `us-west-2` region is perfectly acceptable, however these are provided for reference:

- [ ] __Uses aws_ami Data Source__: Any hardcoded AMI ID configuration, e.g. `ami-12345678`, should be replaced with the [`aws_ami` data source](https://www.terraform.io/docs/providers/aws/d/ami.html) pointing to an Amazon Linux image. A common pattern is a configuration like the below, which will likely be moved into a common configuration function in the future:

  ```hcl
  data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
    most_recent = true
    owners      = ["amazon"]

    filter {
      name = "name"
      values = ["amzn-ami-minimal-hvm-*"]
    }
    filter {
      name = "root-device-type"
      values = ["ebs"]
    }
  }
  ```

- [ ] __Uses aws_availability_zones Data Source__: Any hardcoded AWS Availability Zone configuration, e.g. `us-west-2a`, should be replaced with the [`aws_availability_zones` data source](https://www.terraform.io/docs/providers/aws/d/availability_zones.html). A common pattern is declaring `data "aws_availability_zones" "available" {...}` and referencing it via `data.aws_availability_zones.available.names[0]` or `data.aws_availability_zones.available.names[count.index]` in resources utilizing `count`.

  ```hcl
  data "aws_availability_zones" "available" {
    state = "available"

    filter {
      name   = "opt-in-status"
      values = ["opt-in-not-required"]
    }
  }
  ```

- [ ] __Uses aws_region Data Source__: Any hardcoded AWS Region configuration, e.g. `us-west-2`, should be replaced with the [`aws_region` data source](https://www.terraform.io/docs/providers/aws/d/region.html). A common pattern is declaring `data "aws_region" "current" {}` and referencing it via `data.aws_region.current.name`
- [ ] __Uses aws_partition Data Source__: Any hardcoded AWS Partition configuration, e.g. the `aws` in a `arn:aws:SERVICE:REGION:ACCOUNT:RESOURCE` ARN, should be replaced with the [`aws_partition` data source](https://www.terraform.io/docs/providers/aws/d/partition.html). A common pattern is declaring `data "aws_partition" "current" {}` and referencing it via `data.aws_partition.current.partition`
- [ ] __Uses Builtin ARN Check Functions__: Tests should utilize available ARN check functions, e.g. `testAccMatchResourceAttrRegionalARN()`, to validate ARN attribute values in the Terraform state over `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`
- [ ] __Uses testAccCheckResourceAttrAccountID()__: Tests should utilize the available AWS Account ID check function, `testAccCheckResourceAttrAccountID()` to validate account ID attribute values in the Terraform state over `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`

### Writing Acceptance Tests

Terraform includes an acceptance test harness that does most of the repetitive
work involved in testing a resource. For additional information about testing
Terraform Providers, see the [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/index.html).

#### Acceptance Tests Often Cost Money to Run

Because acceptance tests create real resources, they often cost money to run.
Because the resources only exist for a short period of time, the total amount
of money required is usually a relatively small. Nevertheless, we don't want
financial limitations to be a barrier to contribution, so if you are unable to
pay to run acceptance tests for your contribution, mention this in your
pull request. We will happily accept "best effort" implementations of
acceptance tests and run them for you on our side. This might mean that your PR
takes a bit longer to merge, but it most definitely is not a blocker for
contributions.

#### Running an Acceptance Test

Acceptance tests can be run using the `testacc` target in the Terraform
`Makefile`. The individual tests to run can be controlled using a regular
expression. Prior to running the tests provider configuration details such as
access keys must be made available as environment variables.

For example, to run an acceptance test against the Amazon Web Services
provider, the following environment variables must be set:

```sh
# Using a profile
export AWS_PROFILE=...
# Otherwise
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_DEFAULT_REGION=...
```

Please note that the default region for the testing is `us-west-2` and must be
overriden via the `AWS_DEFAULT_REGION` environment variable, if necessary. This
is especially important for testing AWS GovCloud (US), which requires:

```sh
export AWS_DEFAULT_REGION=us-gov-west-1
```

Tests can then be run by specifying the target provider and a regular
expression defining the tests to run:

```sh
$ make testacc TEST=./aws TESTARGS='-run=TestAccAWSCloudWatchDashboard_update'
==> Checking that code complies with gofmt requirements...
TF_ACC=1 go test ./aws -v -run=TestAccAWSCloudWatchDashboard_update -timeout 120m
=== RUN   TestAccAWSCloudWatchDashboard_update
--- PASS: TestAccAWSCloudWatchDashboard_update (26.56s)
PASS
ok  	github.com/terraform-providers/terraform-provider-aws/aws	26.607s
```

Entire resource test suites can be targeted by using the naming convention to
write the regular expression. For example, to run all tests of the
`aws_cloudwatch_dashboard` resource rather than just the update test, you can start
testing like this:

```sh
$ make testacc TEST=./aws TESTARGS='-run=TestAccAWSCloudWatchDashboard'
==> Checking that code complies with gofmt requirements...
TF_ACC=1 go test ./aws -v -run=TestAccAWSCloudWatchDashboard -timeout 120m
=== RUN   TestAccAWSCloudWatchDashboard_importBasic
--- PASS: TestAccAWSCloudWatchDashboard_importBasic (15.06s)
=== RUN   TestAccAWSCloudWatchDashboard_basic
--- PASS: TestAccAWSCloudWatchDashboard_basic (12.70s)
=== RUN   TestAccAWSCloudWatchDashboard_update
--- PASS: TestAccAWSCloudWatchDashboard_update (27.81s)
PASS
ok  	github.com/terraform-providers/terraform-provider-aws/aws	55.619s
```

#### Writing an Acceptance Test

Terraform has a framework for writing acceptance tests which minimises the
amount of boilerplate code necessary to use common testing patterns. The entry
point to the framework is the `resource.ParallelTest()` function.

Tests are divided into `TestStep`s. Each `TestStep` proceeds by applying some
Terraform configuration using the provider under test, and then verifying that
results are as expected by making assertions using the provider API. It is
common for a single test function to exercise both the creation of and updates
to a single resource. Most tests follow a similar structure.

1. Pre-flight checks are made to ensure that sufficient provider configuration
   is available to be able to proceed - for example in an acceptance test
   targeting AWS, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` must be set prior
   to running acceptance tests. This is common to all tests exercising a single
   provider.

Each `TestStep` is defined in the call to `resource.ParallelTest()`. Most assertion
functions are defined out of band with the tests. This keeps the tests
readable, and allows reuse of assertion functions across different tests of the
same type of resource. The definition of a complete test looks like this:

```go
func TestAccAWSCloudWatchDashboard_basic(t *testing.T) {
	var dashboard cloudwatch.GetDashboardOutput
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchDashboardDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchDashboardConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchDashboardExists("aws_cloudwatch_dashboard.foobar", &dashboard),
					resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foobar", "dashboard_name", testAccAWSCloudWatchDashboardName(rInt)),
				),
			},
		},
	})
}
```

When executing the test, the following steps are taken for each `TestStep`:

1. The Terraform configuration required for the test is applied. This is
   responsible for configuring the resource under test, and any dependencies it
   may have. For example, to test the `aws_cloudwatch_dashboard` resource, a valid configuration with the requisite fields is required. This results in configuration which looks like this:

    ```hcl
    resource "aws_cloudwatch_dashboard" "foobar" {
      dashboard_name = "terraform-test-dashboard-%d"
      dashboard_body = <<EOF
      {
        "widgets": [{
          "type": "text",
          "x": 0,
          "y": 0,
          "width": 6,
          "height": 6,
          "properties": {
            "markdown": "Hi there from Terraform: CloudWatch"
          }
        }]
      }
      EOF
    }
    ```

1. Assertions are run using the provider API. These use the provider API
   directly rather than asserting against the resource state. For example, to
   verify that the `aws_cloudwatch_dashboard` described above was created
   successfully, a test function like this is used:

    ```go
    func testAccCheckCloudWatchDashboardExists(n string, dashboard *cloudwatch.GetDashboardOutput) resource.TestCheckFunc {
      return func(s *terraform.State) error {
        rs, ok := s.RootModule().Resources[n]
        if !ok {
          return fmt.Errorf("Not found: %s", n)
        }

        conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn
        params := cloudwatch.GetDashboardInput{
          DashboardName: aws.String(rs.Primary.ID),
        }

        resp, err := conn.GetDashboard(&params)
        if err != nil {
          return err
        }

        *dashboard = *resp

        return nil
      }
    }
    ```

   Notice that the only information used from the Terraform state is the ID of
   the resource. For computed properties, we instead assert that the value saved in the Terraform state was the
   expected value if possible. The testing framework provides helper functions
   for several common types of check - for example:

    ```go
    resource.TestCheckResourceAttr("aws_cloudwatch_dashboard.foobar", "dashboard_name", testAccAWSCloudWatchDashboardName(rInt)),
    ```

2. The resources created by the test are destroyed. This step happens
   automatically, and is the equivalent of calling `terraform destroy`.

3. Assertions are made against the provider API to verify that the resources
   have indeed been removed. If these checks fail, the test fails and reports
   "dangling resources". The code to ensure that the `aws_cloudwatch_dashboard` shown
   above has been destroyed looks like this:

    ```go
    func testAccCheckAWSCloudWatchDashboardDestroy(s *terraform.State) error {
      conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

      for _, rs := range s.RootModule().Resources {
        if rs.Type != "aws_cloudwatch_dashboard" {
          continue
        }

        params := cloudwatch.GetDashboardInput{
          DashboardName: aws.String(rs.Primary.ID),
        }

        _, err := conn.GetDashboard(&params)
        if err == nil {
          return fmt.Errorf("Dashboard still exists: %s", rs.Primary.ID)
        }
        if !isCloudWatchDashboardNotFoundErr(err) {
          return err
        }
      }

      return nil
    }
    ```

   These functions usually test only for the resource directly under test.

#### Writing and running Cross-Account Acceptance Tests

When testing requires AWS infrastructure in a second AWS account, the below changes to the normal setup will allow the management or reference of resources and data sources across accounts:

- In the `PreCheck` function, include `testAccAlternateAccountPreCheck(t)` to ensure a standardized set of information is required for cross-account testing credentials
- Declare a `providers` variable at the top of the test function: `var providers []*schema.Provider`
- Switch usage of `Providers: testAccProviders` to `ProviderFactories: testAccProviderFactories(&providers)`
- Add `testAccAlternateAccountProviderConfig()` to the test configuration and use `provider = "aws.alternate"` for cross-account resources. The resource that is the focus of the acceptance test should _not_ use the provider alias to simplify the testing setup.
- For any `TestStep` that includes `ImportState: true`, add the `Config` that matches the previous `TestStep` `Config`

An example acceptance test implementation can be seen below:

```go
func TestAccAwsExample_basic(t *testing.T) {
  var providers []*schema.Provider
  resourceName := "aws_example.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck: func() {
      testAccPreCheck(t)
      testAccAlternateAccountPreCheck(t)
    },
    ProviderFactories: testAccProviderFactories(&providers),
    CheckDestroy:      testAccCheckAwsExampleDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleConfig(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleExists(resourceName),
          // ... additional checks ...
        ),
      },
      {
        Config:            testAccAwsExampleConfig(),
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func testAccAwsExampleConfig() string {
  return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
# Cross account resources should be handled by the cross account provider.
# The standardized provider alias is aws.alternate as seen below.
resource "aws_cross_account_example" "test" {
  provider = "aws.alternate"

  # ... configuration ...
}

# The resource that is the focus of the testing should be handled by the default provider,
# which is automatically done by not specifying the provider configuration in the resource.
resource "aws_example" "test" {
  # ... configuration ...
}
`)
}
```

Searching for usage of `testAccAlternateAccountPreCheck` in the codebase will yield real world examples of this setup in action.

Running these acceptance tests is the same as before, except the following additional credential information is required:

```sh
# Using a profile
export AWS_ALTERNATE_PROFILE=...
# Otherwise
export AWS_ALTERNATE_ACCESS_KEY_ID=...
export AWS_ALTERNATE_SECRET_ACCESS_KEY=...
```

#### Writing and running Cross-Region Acceptance Tests

When testing requires AWS infrastructure in a second AWS region, the below changes to the normal setup will allow the management or reference of resources and data sources across regions:

- In the `PreCheck` function, include `testAccMultipleRegionsPreCheck(t)` and `testAccAlternateRegionPreCheck(t)` to ensure a standardized set of information is required for cross-region testing configuration. If the infrastructure in the second AWS region is also in a second AWS account also include `testAccAlternateAccountPreCheck(t)`
- Declare a `providers` variable at the top of the test function: `var providers []*schema.Provider`
- Switch usage of `Providers: testAccProviders` to `ProviderFactories: testAccProviderFactories(&providers)`
- Add `testAccAlternateRegionProviderConfig()` to the test configuration and use `provider = "aws.alternate"` for cross-region resources. The resource that is the focus of the acceptance test should _not_ use the provider alias to simplify the testing setup. If the infrastructure in the second AWS region is also in a second AWS account use `testAccAlternateAccountAlternateRegionProviderConfig()` instead
- For any `TestStep` that includes `ImportState: true`, add the `Config` that matches the previous `TestStep` `Config`

An example acceptance test implementation can be seen below:

```go
func TestAccAwsExample_basic(t *testing.T) {
  var providers []*schema.Provider
  resourceName := "aws_example.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck: func() {
      testAccPreCheck(t)
      testAccMultipleRegionsPreCheck(t)
      testAccAlternateRegionPreCheck(t)
    },
    ProviderFactories: testAccProviderFactories(&providers),
    CheckDestroy:      testAccCheckAwsExampleDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleConfig(),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleExists(resourceName),
          // ... additional checks ...
        ),
      },
      {
        Config:            testAccAwsExampleConfig(),
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

func testAccAwsExampleConfig() string {
  return testAccAlternateRegionProviderConfig() + fmt.Sprintf(`
# Cross region resources should be handled by the cross region provider.
# The standardized provider alias is aws.alternate as seen below.
resource "aws_cross_region_example" "test" {
  provider = "aws.alternate"

  # ... configuration ...
}

# The resource that is the focus of the testing should be handled by the default provider,
# which is automatically done by not specifying the provider configuration in the resource.
resource "aws_example" "test" {
  # ... configuration ...
}
`)
}
```

Searching for usage of `testAccAlternateRegionPreCheck` in the codebase will yield real world examples of this setup in action.

Running these acceptance tests is the same as before, except if an AWS region other than the default alternate region - `us-east-1` - is required,
in which case the following additional configuration information is required:

```sh
export AWS_ALTERNATE_REGION=...
```

[website]: https://github.com/terraform-providers/terraform-provider-aws/tree/master/website
[acctests]: https://github.com/hashicorp/terraform#acceptance-tests
[ml]: https://groups.google.com/group/terraform-tool
