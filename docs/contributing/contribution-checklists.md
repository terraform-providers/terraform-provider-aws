# Contribution Types and Checklists

There are several different kinds of contribution, each of which has its own
standards for a speedy review. The following sections describe guidelines for
each type of contribution.

- [Documentation Update](#documentation-update)
- [Enhancement/Bugfix to a Resource](#enhancementbugfix-to-a-resource)
- [Adding Resource Import Support](#adding-resource-import-support)
- [Adding Resource Name Generation Support](#adding-resource-name-generation-support)
    - [Resource Name Generation Code Implementation](#resource-name-generation-code-implementation)
    - [Resource Name Generation Testing Implementation](#resource-name-generation-testing-implementation)
    - [Resource Code Generation Documentation Implementation](#resource-code-generation-documentation-implementation)
- [Adding Resource Policy Support](#adding-resource-policy-support)
- [Adding Resource Tagging Support](#adding-resource-tagging-support)
    - [Adding Service to Tag Generating Code](#adding-service-to-tag-generating-code)
    - [Resource Tagging Code Implementation](#resource-tagging-code-implementation)
    - [Resource Tagging Acceptance Testing Implementation](#resource-tagging-acceptance-testing-implementation)
    - [Resource Tagging Documentation Implementation](#resource-tagging-documentation-implementation)
- [New Resource](#new-resource)
- [New Service](#new-service)
- [New Region](#new-region)

## Documentation Update

The [Terraform AWS Provider's website source][website] is in this repository
along with the code and tests. Below are some common items that will get
flagged during documentation reviews:

- [ ] __Reasoning for Change__: Documentation updates should include an explanation for why the update is needed.
- [ ] __Prefer AWS Documentation__: Documentation about AWS service features and valid argument values that are likely to update over time should link to AWS service user guides and API references where possible.
- [ ] __Large Example Configurations__: Example Terraform configuration that includes multiple resource definitions should be added to the repository `examples` directory instead of an individual resource documentation page. Each directory under `examples` should be self-contained to call `terraform apply` without special configuration.
- [ ] __Terraform Configuration Language Features__: Individual resource documentation pages and examples should refrain from highlighting particular Terraform configuration language syntax workarounds or features such as `variable`, `local`, `count`, and built-in functions.

## Enhancement/Bugfix to a Resource

Working on existing resources is a great way to get started as a Terraform
contributor because you can work within existing code and tests to get a feel
for what to do.

In addition to the below checklist, please see the [Common Review
Items](pullrequest-submission-and-lifecycle.md#common-review-items) sections for more specific coding and testing
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
   see in the codebase, and ensure your code is formatted with `go fmt`.
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
- [ ] __Dependency updates__: Create a separate PR if you are updating dependencies.
   This is to avoid conflicts as version updates tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

## Adding Resource Import Support

Adding import support for Terraform resources will allow existing infrastructure to be managed within Terraform. This type of enhancement generally requires a small to moderate amount of code changes.

Comprehensive code examples and information about resource import support can be found in the [Extending Terraform documentation](https://www.terraform.io/docs/extend/resources/import.html).

In addition to the below checklist and the items noted in the Extending Terraform documentation, please see the [Common Review Items](pullrequest-submission-and-lifecycle.md#common-review-items) sections for more specific coding and testing guidelines.

- [ ] _Resource Code Implementation_: In the resource code (e.g. `aws/resource_aws_service_thing.go`), implementation of `Importer` `State` function
- [ ] _Resource Acceptance Testing Implementation_: In the resource acceptance testing (e.g. `aws/resource_aws_service_thing_test.go`), implementation of `TestStep`s with `ImportState: true`
- [ ] _Resource Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `Import` documentation section at the bottom of the page

## Adding Resource Name Generation Support

Terraform AWS Provider resources can use shared logic to support and test name generation, where the operator can choose between an expected naming value, a generated naming value with a prefix, or a fully generated name.

Implementing name generation support for Terraform AWS Provider resources requires the following, each with its own section below:

- [ ] _Resource Name Generation Code Implementation_: In the resource code (e.g. `aws/resource_aws_service_thing.go`), implementation of `name_prefix` attribute, along with handling in `Create` function.
- [ ] _Resource Name Generation Testing Implementation_: In the resource acceptance testing (e.g. `aws/resource_aws_service_thing_test.go`), implementation of new acceptance test functions and configurations to exercise new naming logic.
- [ ] _Resource Name Generation Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `name_prefix` argument and update of `name` argument description.

### Resource Name Generation Code Implementation

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
  Computed:      true,
  ForceNew:      true,
  ConflictsWith: []string{"name"},
},
```

- In the resource `Create` function, switch any calls from `d.Get("name").(string)` to instead use the `naming.Generate()` function, e.g.

```go
name := naming.Generate(d.Get("name").(string), d.Get("name_prefix").(string))

// ... in AWS Go SDK Input types, etc. use aws.String(name)
```

- If the resource supports import, in the resource `Read` function add a call to `d.Set("name_prefix", ...)`, e.g.

```go
d.Set("name", resp.Name)
d.Set("name_prefix", naming.NamePrefixFromName(aws.StringValue(resp.Name)))
```

### Resource Name Generation Testing Implementation

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
          resource.TestCheckResourceAttr(resourceName, "name_prefix", "terraform-"),
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
          resource.TestCheckResourceAttr(resourceName, "name_prefix", "tf-acc-test-prefix-"),
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

### Resource Name Generation Documentation Implementation

- In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), add the following to the arguments reference:

```markdown
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
```

- Adjust the existing `name` argument reference to ensure its denoted as `Optional`, includes a mention that it can be generated, and that it conflicts with `name_prefix`:

```markdown
* `name` - (Optional) Name of the thing. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
```

## Adding Resource Policy Support

Some AWS components support [resource-based IAM policies](https://docs.aws.amazon.com/IAM/latest/UserGuide/access_policies_identity-vs-resource.html) to control permissions. When implementing this support in the Terraform AWS Provider, we typically prefer creating a separate resource, `aws_{SERVICE}_{THING}_policy` (e.g. `aws_s3_bucket_policy`). See the [New Resource section](#new-resource) for more information about implementing the separate resource and the [Provider Design page](provider-design.md) for rationale.
## Adding Resource Tagging Support

AWS provides key-value metadata across many services and resources, which can be used for a variety of use cases including billing, ownership, and more. See the [AWS Tagging Strategy page](https://aws.amazon.com/answers/account-management/aws-tagging-strategies/) for more information about tagging at a high level.

Implementing tagging support for Terraform AWS Provider resources requires the following, each with its own section below:

- [ ] _Generated Service Tagging Code_: In the internal code generators (e.g. `aws/internal/keyvaluetags`), implementation and customization of how a service handles tagging, which is standardized for the resources.
- [ ] _Resource Tagging Code Implementation_: In the resource code (e.g. `aws/resource_aws_service_thing.go`), implementation of `tags` schema attribute, along with handling in `Create`, `Read`, and `Update` functions.
- [ ] _Resource Tagging Acceptance Testing Implementation_: In the resource acceptance testing (e.g. `aws/resource_aws_service_thing_test.go`), implementation of new acceptance test function and configurations to exercise new tagging logic.
- [ ] _Resource Tagging Documentation Implementation_: In the resource documentation (e.g. `website/docs/r/service_thing.html.markdown`), addition of `tags` argument

See also a [full example pull request for implementing EKS tagging](https://github.com/hashicorp/terraform-provider-aws/pull/10307).

### Adding Service to Tag Generating Code

This step is only necessary for the first implementation and may have been previously completed. If so, move on to the next section.

More details about this code generation, including fixes for potential error messages in this process, can be found in the [keyvaluetags documentation](../../aws/internal/keyvaluetags/README.md).

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

### Resource Tagging Code Implementation

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
  // Typically declared near conn := /* ... */
  ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

  if err := d.Set("tags", keyvaluetags.EksKeyValueTags(cluster.Tags).IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
    return fmt.Errorf("error setting tags: %w", err)
  }
  ```

  If the service API does not return the tags directly from reading the resource and requires a separate API call, its possible to use the `keyvaluetags` functionality like the following, e.g. with Athena Workgroups:

  ```go
  // Typically declared near conn := /* ... */
  ignoreTagsConfig := meta.(*AWSClient).IgnoreTagsConfig

  tags, err := keyvaluetags.AthenaListTags(conn, arn.String())

  if err != nil {
    return fmt.Errorf("error listing tags for resource (%s): %s", arn, err)
  }

  if err := d.Set("tags", tags.IgnoreAws().IgnoreConfig(ignoreTagsConfig).Map()); err != nil {
    return fmt.Errorf("error setting tags: %w", err)
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

### Resource Tagging Acceptance Testing Implementation

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
    role_arn = aws_iam_role.test.arn

    tags = {
      %[2]q = %[3]q
    }

    vpc_config {
      subnet_ids = aws_subnet.test[*].id
    }

    depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
  }
  `, rName, tagKey1, tagValue1)
  }

  func testAccAWSEksClusterConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
    return testAccAWSEksClusterConfig_Base(rName) + fmt.Sprintf(`
  resource "aws_eks_cluster" "test" {
    name     = %[1]q
    role_arn = aws_iam_role.test.arn

    tags = {
      %[2]q = %[3]q
      %[4]q = %[5]q
    }

    vpc_config {
      subnet_ids = aws_subnet.test[*].id
    }

    depends_on = [aws_iam_role_policy_attachment.test-AmazonEKSClusterPolicy]
  }
  `, rName, tagKey1, tagValue1, tagKey2, tagValue2)
  }
  ```

- Verify all acceptance testing passes for the resource (e.g. `make testacc TESTARGS='-run=TestAccAWSEksCluster_'`)

### Resource Tagging Documentation Implementation

- In the resource documentation (e.g. `website/docs/r/eks_cluster.html.markdown`), add the following to the arguments reference:

  ```markdown
  * `tags` - (Optional) Key-value mapping of resource tags
  ```

## New Resource

_Before submitting this type of contribution, it is highly recommended to read and understand the other pages of the [Contributing Guide](../CONTRIBUTING.md)._

Implementing a new resource is a good way to learn more about how Terraform
interacts with upstream APIs. There are plenty of examples to draw from in the
existing resources, but you still get to implement something completely new.

In addition to the below checklist, please see the [Common Review
Items](pullrequest-submission-and-lifecycle.md#common-review-items) sections for more specific coding and testing
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

- [ ] __Arguments_and_Attributes__: The HCL for arguments and attributes should mimic the types and structs presented by the AWS API. API arguments should be converted from `CamelCase` to `camel_case`. The resource logic for handling these should follow the recommended implementations in the [Data Handling and Conversion](data-handling-and-conversion.md) documentation.
- [ ] __Documentation__: Each data source and resource gets a page in the Terraform
   documentation, which lives at `website/docs/d/<service>_<name>.html.markdown` and
   `website/docs/r/<service>_<name>.html.markdown` respectively.
- [ ] __Well-formed Code__: Do your best to follow existing conventions you
   see in the codebase, and ensure your code is formatted with `go fmt`.
   The PR reviewers can help out on this front, and may provide comments with
   suggestions on how to improve the code.
- [ ] __Dependency updates__: Create a separate PR if you are updating dependencies.
   This is to avoid conflicts as version updates tend to be fast-
   moving targets. We will plan to merge the PR with this change first.

## New Service

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
    - In `infrastructure/repository/labels-service.tf`: Add the new service to create a repository label.
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
        "aws/internal/service/quicksight/**/*",
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
  ```

- [ ] __Initial Resource__: Some services can be big and it can be
  difficult for both reviewer & author to go through long feedback cycles
  on a big PR with many resources. Often feedback items in one resource
  will also need to be applied in other resources. We prefer you to submit
  the necessary minimum in a single PR, ideally **just the first resource**
  of the service.

The initial resource and changes afterwards should follow the other sections
of this guide as appropriate.

## New Region

While region validation is automatically added with SDK updates, new regions
are generally limited in which services they support. Below are some
manually sourced values from documentation.

- [ ] Check [Elastic Load Balancing endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elb.html#elb_region) and add Route53 Hosted Zone ID if available to `aws/data_source_aws_elb_hosted_zone_id.go`
- [ ] Check [Amazon Simple Storage Service endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/s3.html#s3_region) and add Route53 Hosted Zone ID if available to `aws/hosted_zones.go`
- [ ] Check [CloudTrail Supported Regions docs](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html#cloudtrail-supported-regions) and add AWS Account ID if available to `aws/data_source_aws_cloudtrail_service_account.go`
- [ ] Check [Elastic Load Balancing Access Logs docs](https://docs.aws.amazon.com/elasticloadbalancing/latest/classic/enable-access-logs.html#attach-bucket-policy) and add Elastic Load Balancing Account ID if available to `aws/data_source_aws_elb_service_account.go`
- [ ] Check [Redshift Database Audit Logging docs](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) and add AWS Account ID if available to `aws/data_source_aws_redshift_service_account.go`
- [ ] Check [AWS Elastic Beanstalk endpoints and quotas](https://docs.aws.amazon.com/general/latest/gr/elasticbeanstalk.html#elasticbeanstalk_region) and add Route53 Hosted Zone ID if available to `aws/data_source_aws_elastic_beanstalk_hosted_zone.go`
