# Running and Writing Acceptance Tests

- [Acceptance Tests Often Cost Money to Run](#acceptance-tests-often-cost-money-to-run)
- [Running an Acceptance Test](#running-an-acceptance-test)
    - [Running Cross-Account Tests](#running-cross-account-tests)
    - [Running Cross-Region Tests](#running-cross-region-tests)
- [Writing an Acceptance Test](#writing-an-acceptance-test)
    - [Anatomy of an Acceptance Test](#anatomy-of-an-acceptance-test)
    - [Resource Acceptance Testing](#resource-acceptance-testing)
        - [Test Configurations](#test-configurations)
        - [Combining Test Configurations](#combining-test-configurations)
            - [Base Test Configurations](#base-test-configurations)
            - [Available Common Test Configurations](#available-common-test-configurations)
        - [Randomized Naming](#randomized-naming)
        - [Other Recommended Variables](#other-recommended-variables)
        - [Basic Acceptance Tests](#basic-acceptance-tests)
        - [PreChecks](#prechecks)
            - [Standard Provider PreChecks](#standard-provider-prechecks)
            - [Custom PreChecks](#custom-prechecks)
        - [Disappears Acceptance Tests](#disappears-acceptance-tests)
        - [Per Attribute Acceptance Tests](#per-attribute-acceptance-tests)
        - [Cross-Account Acceptance Tests](#cross-account-acceptance-tests)
        - [Cross-Region Acceptance Tests](#cross-region-acceptance-tests)
        - [Service-Specific Region Acceptance Tests](#service-specific-region-acceptance-tests)
    - [Data Source Acceptance Testing](#data-source-acceptance-testing)
- [Acceptance Test Sweepers](#acceptance-test-sweepers)
    - [Running Test Sweepers](#running-test-sweepers)
    - [Writing Test Sweepers](#writing-test-sweepers)
- [Acceptance Test Checklists](#acceptance-test-checklists)
    - [Basic Acceptance Test Design](#basic-acceptance-test-design)
    - [Test Implementation](#test-implementation)
    - [Avoid Hard Coding](#avoid-hard-coding)
        - [Hardcoded Account IDs](#hardcoded-account-ids)
        - [Hardcoded AMI IDs](#hardcoded-ami-ids)
        - [Hardcoded Availability Zones](#hardcoded-availability-zones)
        - [Hardcoded Database Versions](#hardcoded-database-versions)
        - [Hardcoded Instance Types](#hardcoded-instance-types)
        - [Hardcoded Partition DNS Suffix](#hardcoded-partition-dns-suffix)
        - [Hardcoded Partition in ARN](#hardcoded-partition-in-arn)
        - [Hardcoded Region](#hardcoded-region)
        - [Hardcoded Spot Price](#hardcoded-spot-price)

Terraform includes an acceptance test harness that does most of the repetitive
work involved in testing a resource. For additional information about testing
Terraform Providers, see the [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/index.html).

## Acceptance Tests Often Cost Money to Run

Because acceptance tests create real resources, they often cost money to run.
Because the resources only exist for a short period of time, the total amount
of money required is usually a relatively small. Nevertheless, we don't want
financial limitations to be a barrier to contribution, so if you are unable to
pay to run acceptance tests for your contribution, mention this in your
pull request. We will happily accept "best effort" implementations of
acceptance tests and run them for you on our side. This might mean that your PR
takes a bit longer to merge, but it most definitely is not a blocker for
contributions.

## Running an Acceptance Test

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
overridden via the `AWS_DEFAULT_REGION` environment variable, if necessary. This
is especially important for testing AWS GovCloud (US), which requires:

```sh
export AWS_DEFAULT_REGION=us-gov-west-1
```

Tests can then be run by specifying a regular expression defining the tests to run:

```sh
$ make testacc TESTARGS='-run=TestAccAWSCloudWatchDashboard_update'
==> Checking that code complies with gofmt requirements...
TF_ACC=1 go test ./aws -v -run=TestAccAWSCloudWatchDashboard_update -timeout 120m
=== RUN   TestAccAWSCloudWatchDashboard_update
--- PASS: TestAccAWSCloudWatchDashboard_update (26.56s)
PASS
ok  	github.com/hashicorp/terraform-provider-aws/aws	26.607s
```

Entire resource test suites can be targeted by using the naming convention to
write the regular expression. For example, to run all tests of the
`aws_cloudwatch_dashboard` resource rather than just the update test, you can start
testing like this:

```sh
$ make testacc TESTARGS='-run=TestAccAWSCloudWatchDashboard'
==> Checking that code complies with gofmt requirements...
TF_ACC=1 go test ./aws -v -run=TestAccAWSCloudWatchDashboard -timeout 120m
=== RUN   TestAccAWSCloudWatchDashboard_importBasic
--- PASS: TestAccAWSCloudWatchDashboard_importBasic (15.06s)
=== RUN   TestAccAWSCloudWatchDashboard_basic
--- PASS: TestAccAWSCloudWatchDashboard_basic (12.70s)
=== RUN   TestAccAWSCloudWatchDashboard_update
--- PASS: TestAccAWSCloudWatchDashboard_update (27.81s)
PASS
ok  	github.com/hashicorp/terraform-provider-aws/aws	55.619s
```

Running acceptance tests requires version 0.12.26 or higher of the Terraform CLI to be installed.

For advanced developers, the acceptance testing framework accepts some additional environment variables that can be used to control Terraform CLI binary selection, logging, and other behaviors. See the [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html#environment-variables) for more information.

Please Note: On macOS 10.14 and later (and some Linux distributions), the default user open file limit is 256. This may cause unexpected issues when running the acceptance testing since this can prevent various operations from occurring such as opening network connections to AWS. To view this limit, the `ulimit -n` command can be run. To update this limit, run `ulimit -n 1024`  (or higher).

### Running Cross-Account Tests

Certain testing requires multiple AWS accounts. This additional setup is not typically required and the testing will return an error (shown below) if your current setup does not have the secondary AWS configuration:

```console
$ make testacc TEST=./aws TESTARGS='-run=TestAccAWSDBInstance_DbSubnetGroupName_RamShared'
=== RUN   TestAccAWSDBInstance_DbSubnetGroupName_RamShared
=== PAUSE TestAccAWSDBInstance_DbSubnetGroupName_RamShared
=== CONT  TestAccAWSDBInstance_DbSubnetGroupName_RamShared
    TestAccAWSDBInstance_DbSubnetGroupName_RamShared: provider_test.go:386: AWS_ALTERNATE_ACCESS_KEY_ID or AWS_ALTERNATE_PROFILE must be set for acceptance tests
--- FAIL: TestAccAWSDBInstance_DbSubnetGroupName_RamShared (2.22s)
FAIL
FAIL	github.com/hashicorp/terraform-provider-aws/aws	4.305s
FAIL
```

Running these acceptance tests is the same as before, except the following additional AWS credential information is required:

```sh
# Using a profile
export AWS_ALTERNATE_PROFILE=...
# Otherwise
export AWS_ALTERNATE_ACCESS_KEY_ID=...
export AWS_ALTERNATE_SECRET_ACCESS_KEY=...
```

### Running Cross-Region Tests

Certain testing requires multiple AWS regions. Additional setup is not typically required because the testing defaults the second AWS region to `us-east-1` and the third AWS region to `us-east-2`.

Running these acceptance tests is the same as before, but if you wish to override the second and third regions:

```sh
export AWS_ALTERNATE_REGION=...
export AWS_THIRD_REGION=...
```

## Writing an Acceptance Test

Terraform has a framework for writing acceptance tests which minimises the
amount of boilerplate code necessary to use common testing patterns. This guide is meant to augment the general [Extending Terraform documentation](https://www.terraform.io/docs/extend/testing/acceptance-tests/index.html) with Terraform AWS Provider specific conventions and helpers.

### Anatomy of an Acceptance Test

This section describes in detail how the Terraform acceptance testing framework operates with respect to the Terraform AWS Provider. We recommend those unfamiliar with this provider, or Terraform resource testing in general, take a look here first to generally understand how we interact with AWS and the resource code to verify functionality.

The entry point to the framework is the `resource.ParallelTest()` function. This wraps our testing to work with the standard Go testing framework, while also preventing unexpected usage of AWS by requiring the `TF_ACC=1` environment variable. This function accepts a `TestCase` parameter, which has all the details about the test itself. For example, this includes the test steps (`TestSteps`) and how to verify resource deletion in the API after all steps have been run (`CheckDestroy`).

Each `TestStep` proceeds by applying some
Terraform configuration using the provider under test, and then verifying that
results are as expected by making assertions using the provider API. It is
common for a single test function to exercise both the creation of and updates
to a single resource. Most tests follow a similar structure.

1. Pre-flight checks are made to ensure that sufficient provider configuration
   is available to be able to proceed - for example in an acceptance test
   targeting AWS, `AWS_ACCESS_KEY_ID` and `AWS_SECRET_ACCESS_KEY` must be set prior
   to running acceptance tests. This is common to all tests exercising a single
   provider.

Most assertion
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

1. The resources created by the test are destroyed. This step happens
   automatically, and is the equivalent of calling `terraform destroy`.

1. Assertions are made against the provider API to verify that the resources
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

### Resource Acceptance Testing

Most resources that implement standard Create, Read, Update, and Delete functionality should follow the pattern below. Each test type has a section that describes them in more detail:

- **basic**: This represents the bare minimum verification that the resource can be created, read, deleted, and optionally imported.
- **disappears**: A test that verifies Terraform will offer to recreate a resource if it is deleted outside of Terraform (e.g. via the Console) instead of returning an error that it cannot be found.
- **Per Attribute**: A test that verifies the resource with a single additional argument can be created, read, optionally updated (or force resource recreation), deleted, and optionally imported.

The leading sections below highlight additional recommended patterns.

#### Test Configurations

Most of the existing test configurations you will find in the Terraform AWS Provider are written in the following function-based style:

```go
func TestAccAwsExampleThing_basic(t *testing.T) {
  // ... omitted for brevity ...

  resource.ParallelTest(t, resource.TestCase{
    // ... omitted for brevity ...
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleThingConfig(),
        // ... omitted for brevity ...
      },
    },
  })
}

func testAccAwsExampleThingConfig() string {
  return `
resource "aws_example_thing" "test" {
  # ... omitted for brevity ...
}
`
}
```

Even when no values need to be passed in to the test configuration, we have found this setup to be the most flexible for allowing that to be easily implemented. Any configurable values are handled via `fmt.Sprintf()`. Using `text/template` or other templating styles is explicitly forbidden.

For consistency, resources in the test configuration should be named `resource "..." "test"` unless multiple of that resource are necessary.

We discourage re-using test configurations across test files (except for some common configuration helpers we provide) as it is much harder to discover potential testing regressions.

Please also note that the newline on the first line of the configuration (before `resource`) and the newline after the last line of configuration (after `}`) are important to allow test configurations to be easily combined without generating Terraform configuration language syntax errors.

#### Combining Test Configurations

We include a helper function, `composeConfig()` for iteratively building and chaining test configurations together. It accepts any number of configurations to combine them. This simplifies a single resource's testing by allowing the creation of a "base" test configuration for all the other test configurations (if necessary) and also allows the maintainers to curate common configurations. Each of these is described in more detail in below sections.

Please note that we do discourage _excessive_ chaining of configurations such as implementing multiple layers of "base" configurations. Usually these configurations are harder for maintainers and other future readers to understand due to the multiple levels of indirection.

##### Base Test Configurations

If a resource requires the same Terraform configuration as a prerequisite for all test configurations, then a common pattern is implementing a "base" test configuration that is combined with each test configuration.

For example:

```go
func testAccAwsExampleThingConfigBase() string {
  return `
resource "aws_iam_role" "test" {
  # ... omitted for brevity ...
}

resource "aws_iam_role_policy" "test" {
  # ... omitted for brevity ...
}
`
}

func testAccAwsExampleThingConfig() string {
  return composeConfig(
    testAccAwsExampleThingConfigBase(),
    `
resource "aws_example_thing" "test" {
  # ... omitted for brevity ...
}
`)
}
```

##### Available Common Test Configurations

These test configurations are typical implementations we have found or allow testing to implement best practices easier, since the Terraform AWS Provider testing is expected to run against various AWS Regions and Partitions.

- `testAccAvailableEc2InstanceTypeForRegion("type1", "type2", ...)`: Typically used to replace hardcoded EC2 Instance Types. Uses `aws_ec2_instance_type_offering` data source to return an available EC2 Instance Type in preferred ordering. Reference the instance type via: `data.aws_ec2_instance_type_offering.available.instance_type`
- `testAccLatestAmazonLinuxHvmEbsAmiConfig()`: Typically used to replace hardcoded EC2 Image IDs (`ami-12345678`). Uses `aws_ami` data source to find the latest Amazon Linux image. Reference the AMI ID via: `data.aws_ami.amzn-ami-minimal-hvm-ebs.id`

#### Randomized Naming

For AWS resources that require unique naming, the tests should implement a randomized name, typically coded as a `rName` variable in the test and passed as a parameter to creating the test configuration.

For example:

```go
func TestAccAwsExampleThing_basic(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  // ... omitted for brevity ...

  resource.ParallelTest(t, resource.TestCase{
    // ... omitted for brevity ...
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleThingConfigName(rName),
        // ... omitted for brevity ...
      },
    },
  })
}

func testAccAwsExampleThingConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}
`, rName)
}
```

Typically the `rName` is always the first argument to the test configuration function, if used, for consistency.

#### Other Recommended Variables

We also typically recommend saving a `resourceName` variable in the test that contains the resource reference, e.g. `aws_example_thing.test`, which is repeatedly used in the checks.

For example:

```go
func TestAccAwsExampleThing_basic(t *testing.T) {
  // ... omitted for brevity ...
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    // ... omitted for brevity ...
    Steps: []resource.TestStep{
      {
        // ... omitted for brevity ...
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleThingExists(resourceName),
          testAccCheckResourceAttrRegionalARN(resourceName, "arn", "example", fmt.Sprintf("thing/%s", rName)),
          resource.TestCheckResourceAttr(resourceName, "description", ""),
          resource.TestCheckResourceAttr(resourceName, "name", rName),
        ),
      },
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

// below all TestAcc functions

func testAccAwsExampleThingConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}
`, rName)
}
```

#### Basic Acceptance Tests

Usually this test is implemented first. The test configuration should contain only required arguments (`Required: true` attributes) and it should check the values of all read-only attributes (`Computed: true` without `Optional: true`). If the resource supports it, it verifies import. It should _NOT_ perform other `TestStep` such as updates or verify recreation.

These are typically named `TestAccAws{SERVICE}{THING}_basic`, e.g. `TestAccAwsCloudWatchDashboard_basic`

For example:

```go
func TestAccAwsExampleThing_basic(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    Providers:    testAccProviders,
    CheckDestroy: testAccCheckAwsExampleThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleThingExists(resourceName),
          testAccCheckResourceAttrRegionalARN(resourceName, "arn", "example", fmt.Sprintf("thing/%s", rName)),
          resource.TestCheckResourceAttr(resourceName, "description", ""),
          resource.TestCheckResourceAttr(resourceName, "name", rName),
        ),
      },
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
    },
  })
}

// below all TestAcc functions

func testAccAwsExampleThingConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}
`, rName)
}
```

#### PreChecks

Acceptance test cases have a PreCheck. The PreCheck ensures that the testing environment meets certain preconditions. If the environment does not meet the preconditions, Go skips the test. Skipping a test avoids reporting a failure and wasting resources where the test cannot succeed.

Here is an example of the default PreCheck:

```go
func TestAccAwsExampleThing_basic(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    // ... additional checks follow ...
  })
}
```

Extend the default PreCheck by adding calls to functions in the anonymous PreCheck function. The functions can be existing functions in the provider or custom functions you add for new capabilities.

##### Standard Provider PreChecks

If you add a new test that has preconditions which are checked by an existing provider function, use that standard PreCheck instead of creating a new one. Some existing tests are missing standard PreChecks and you can help by adding them where appropriate.

These are some of the standard provider PreChecks:

* `testAccPartitionHasServicePreCheck(serviceId string, t *testing.T)` checks whether the current partition lists the service as part of its offerings. Note: AWS may not add new or public preview services to the service list immediately. This function will return a false positive in that case.
* `testAccOrganizationsAccountPreCheck(t *testing.T)` checks whether the current account can perform AWS Organizations tests.
* `testAccAlternateAccountPreCheck(t *testing.T)` checks whether the environment is set up for tests across accounts.
* `testAccMultipleRegionPreCheck(t *testing.T, regions int)` checks whether the environment is set up for tests across regions.

This is an example of using a standard PreCheck function. For an established service, such as WAF or FSx, use `testAccPartitionHasServicePreCheck()` and the service endpoint ID to check that a partition supports the service.

```go
func TestAccAwsExampleThing_basic(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(waf.EndpointsID, t) },
    // ... additional checks follow ...
  })
}
```

##### Custom PreChecks

In situations where standard PreChecks do not test for the required preconditions, create a custom PreCheck.

Below is an example of adding a custom PreCheck function. For a new or preview service that AWS does not include in the partition service list yet, you can verify the existence of the service with a simple read-only request (e.g., list all X service things). (For acceptance tests of established services, use `testAccPartitionHasServicePreCheck()` instead.)

```go
func TestAccAwsExampleThing_basic(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t), testAccPreCheckAwsExample(t) },
    // ... additional checks follow ...
  })
}

func testAccPreCheckAwsExample(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).exampleconn
	input := &example.ListThingsInput{}
	_, err := conn.ListThings(input)
	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}
```

#### Disappears Acceptance Tests

This test is generally implemented second. It is straightforward to setup once the basic test is passing since it can reuse that test configuration. It prevents a common bug report with Terraform resources that error when they can not be found (e.g. deleted outside Terraform).

These are typically named `TestAccAws{SERVICE}{THING}_disappears`, e.g. `TestAccAwsCloudWatchDashboard_disappears`

For example:

```go
func TestAccAwsExampleThing_disappears(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    Providers:    testAccProviders,
    CheckDestroy: testAccCheckAwsExampleThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleThingExists(resourceName, &job),
          testAccCheckResourceDisappears(testAccProvider, resourceAwsExampleThing(), resourceName),
        ),
        ExpectNonEmptyPlan: true,
      },
    },
  })
}
```

If this test does fail, the fix for this is generally adding error handling immediately after the `Read` API call that catches the error and tells Terraform to remove the resource before returning the error:

```go
output, err := conn.GetThing(input)

if isAWSErr(err, example.ErrCodeResourceNotFound, "") {
  log.Printf("[WARN] Example Thing (%s) not found, removing from state", d.Id())
  d.SetId("")
  return nil
}

if err != nil {
  return fmt.Errorf("error reading Example Thing (%s): %w", d.Id(), err)
}
```

For children resources that are encapsulated by a parent resource, it is also preferable to verify that removing the parent resource will not generate an error either. These are typically named `TestAccAws{SERVICE}{THING}_disappears_{PARENT}`, e.g. `TestAccAwsRoute53ZoneAssociation_disappears_Vpc`

```go
func TestAccAwsExampleChildThing_disappears_ParentThing(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  parentResourceName := "aws_example_parent_thing.test"
  resourceName := "aws_example_child_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    Providers:    testAccProviders,
    CheckDestroy: testAccCheckAwsExampleChildThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleThingConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleThingExists(resourceName),
          testAccCheckResourceDisappears(testAccProvider, resourceAwsExampleParentThing(), parentResourceName),
        ),
        ExpectNonEmptyPlan: true,
      },
    },
  })
}
```

#### Per Attribute Acceptance Tests

These are typically named `TestAccAws{SERVICE}{THING}_{ATTRIBUTE}`, e.g. `TestAccAwsCloudWatchDashboard_Name`

For example:

```go
func TestAccAwsExampleThing_Description(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    Providers:    testAccProviders,
    CheckDestroy: testAccCheckAwsExampleThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleThingConfigDescription(rName, "description1"),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleThingExists(resourceName),
          resource.TestCheckResourceAttr(resourceName, "description", "description1"),
        ),
      },
      {
        ResourceName:      resourceName,
        ImportState:       true,
        ImportStateVerify: true,
      },
      {
        Config: testAccAwsExampleThingConfigDescription(rName, "description2"),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleThingExists(resourceName),
          resource.TestCheckResourceAttr(resourceName, "description", "description2"),
        ),
      },
    },
  })
}

// below all TestAcc functions

func testAccAwsExampleThingConfigDescription(rName string, description string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  description = %[2]q
  name        = %[1]q
}
`, rName, description)
}
```

#### Cross-Account Acceptance Tests

When testing requires AWS infrastructure in a second AWS account, the below changes to the normal setup will allow the management or reference of resources and data sources across accounts:

- In the `PreCheck` function, include `testAccAlternateAccountPreCheck(t)` to ensure a standardized set of information is required for cross-account testing credentials
- Declare a `providers` variable at the top of the test function: `var providers []*schema.Provider`
- Switch usage of `Providers: testAccProviders` to `ProviderFactories: testAccProviderFactoriesAlternate(&providers)`
- Add `testAccAlternateAccountProviderConfig()` to the test configuration and use `provider = awsalternate` for cross-account resources. The resource that is the focus of the acceptance test should _not_ use the alternate provider identification to simplify the testing setup.
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
    ProviderFactories: testAccProviderFactoriesAlternate(&providers),
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
# The standardized provider block to use is awsalternate as seen below.
resource "aws_cross_account_example" "test" {
  provider = awsalternate

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

#### Cross-Region Acceptance Tests

When testing requires AWS infrastructure in a second or third AWS region, the below changes to the normal setup will allow the management or reference of resources and data sources across regions:

- In the `PreCheck` function, include `testAccMultipleRegionPreCheck(t, ###)` to ensure a standardized set of information is required for cross-region testing configuration. If the infrastructure in the second AWS region is also in a second AWS account also include `testAccAlternateAccountPreCheck(t)`
- Declare a `providers` variable at the top of the test function: `var providers []*schema.Provider`
- Switch usage of `Providers: testAccProviders` to `ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 2)` (where the last parameter is number of regions)
- Add `testAccMultipleRegionProviderConfig(###)` to the test configuration and use `provider = awsalternate` (and potentially `provider = awsthird`) for cross-region resources. The resource that is the focus of the acceptance test should _not_ use the alternative providers to simplify the testing setup. If the infrastructure in the second AWS region is also in a second AWS account use `testAccAlternateAccountAlternateRegionProviderConfig()` instead
- For any `TestStep` that includes `ImportState: true`, add the `Config` that matches the previous `TestStep` `Config`

An example acceptance test implementation can be seen below:

```go
func TestAccAwsExample_basic(t *testing.T) {
  var providers []*schema.Provider
  resourceName := "aws_example.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck: func() {
      testAccPreCheck(t)
      testAccMultipleRegionPreCheck(t, 2)
    },
    ProviderFactories: testAccProviderFactoriesMultipleRegion(&providers, 2),
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
  return testAccMultipleRegionProviderConfig(2) + fmt.Sprintf(`
# Cross region resources should be handled by the cross region provider.
# The standardized provider is awsalternate as seen below.
resource "aws_cross_region_example" "test" {
  provider = awsalternate

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

Searching for usage of `testAccMultipleRegionPreCheck` in the codebase will yield real world examples of this setup in action.

#### Service-Specific Region Acceptance Testing

Certain AWS service APIs are only available in specific AWS regions. For example as of this writing, the `pricing` service is available in `ap-south-1` and `us-east-1`, but no other regions or partitions. When encountering these types of services, the acceptance testing can be setup to automatically detect the correct region(s), while skipping the testing in unsupported partitions.

To prepare the shared service functionality, create a file named `aws/{SERVICE}_test.go`. A starting example with the Pricing service (`aws/pricing_test.go`):

```go
package aws

import (
  "context"
  "sync"
  "testing"

  "github.com/aws/aws-sdk-go/aws/endpoints"
  "github.com/aws/aws-sdk-go/service/pricing"
  "github.com/hashicorp/terraform-plugin-sdk/v2/diag"
  "github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
  "github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// testAccPricingRegion is the chosen Pricing testing region
//
// Cached to prevent issues should multiple regions become available.
var testAccPricingRegion string

// testAccProviderPricing is the Pricing provider instance
//
// This Provider can be used in testing code for API calls without requiring
// the use of saving and referencing specific ProviderFactories instances.
//
// testAccPreCheckPricing(t) must be called before using this provider instance.
var testAccProviderPricing *schema.Provider

// testAccProviderPricingConfigure ensures the provider is only configured once
var testAccProviderPricingConfigure sync.Once

// testAccPreCheckPricing verifies AWS credentials and that Pricing is supported
func testAccPreCheckPricing(t *testing.T) {
  testAccPartitionHasServicePreCheck(pricing.EndpointsID, t)

  // Since we are outside the scope of the Terraform configuration we must
  // call Configure() to properly initialize the provider configuration.
  testAccProviderPricingConfigure.Do(func() {
    testAccProviderPricing = Provider()

    config := map[string]interface{}{
      "region": testAccGetPricingRegion(),
    }

    diags := testAccProviderPricing.Configure(context.Background(), terraform.NewResourceConfigRaw(config))

    if diags != nil && diags.HasError() {
      for _, d := range diags {
        if d.Severity == diag.Error {
          t.Fatalf("error configuring Pricing provider: %s", d.Summary)
        }
      }
    }
  })
}

// testAccPricingRegionProviderConfig is the Terraform provider configuration for Pricing region testing
//
// Testing Pricing assumes no other provider configurations
// are necessary and overwrites the "aws" provider configuration.
func testAccPricingRegionProviderConfig() string {
  return testAccRegionalProviderConfig(testAccGetPricingRegion())
}

// testAccGetPricingRegion returns the Pricing region for testing
func testAccGetPricingRegion() string {
  if testAccPricingRegion != "" {
    return testAccPricingRegion
  }

  if rs, ok := endpoints.RegionsForService(endpoints.DefaultPartitions(), testAccGetPartition(), pricing.ServiceName); ok {
    // return available region (random if multiple)
    for regionID := range rs {
      testAccPricingRegion = regionID
      return testAccPricingRegion
    }
  }

  testAccPricingRegion = testAccGetRegion()

  return testAccPricingRegion
}
```

For the resource or data source acceptance tests, the key items to adjust are:

* Ensure `TestCase` uses `ProviderFactories: testAccProviderFactories` instead of `Providers: testAccProviders`
* Add the call for the new `PreCheck` function (keeping `testAccPreCheck(t)`), e.g. `PreCheck: func() { testAccPreCheck(t); testAccPreCheckPricing(t) },`
* If the testing is for a managed resource with a `CheckDestroy` function, ensure it uses the new provider instance, e.g. `testAccProviderPricing`, instead of `testAccProvider`.
* If the testing is for a managed resource with a `Check...Exists` function, ensure it uses the new provider instance, e.g. `testAccProviderPricing`, instead of `testAccProvider`.
* In each `TestStep` configuration, ensure the new provider configuration function is called, e.g.

```go
func testAccDataSourceAwsPricingProductConfigRedshift() string {
  return composeConfig(
    testAccPricingRegionProviderConfig(),
    `
# ... test configuration ...
`)
}
```

If the testing configurations require more than one region, reach out to the maintainers for further assistance.

### Data Source Acceptance Testing

Writing acceptance testing for data sources is similar to resources, with the biggest changes being:

- Adding `DataSource` to the test and configuration naming, such as `TestAccAwsExampleThingDataSource_Filter`
- The basic test _may_ be named after the easiest lookup attribute instead, e.g. `TestAccAwsExampleThingDataSource_Name`
- No disappears testing
- Almost all checks should be done with [`resource.TestCheckResourceAttrPair()`](https://pkg.go.dev/github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource?tab=doc#TestCheckResourceAttrPair) to compare the data source attributes to the resource attributes
- The usage of an additional `dataSourceName` variable to store a data source reference, e.g. `data.aws_example_thing.test`

Data sources testing should still utilize the `CheckDestroy` function of the resource, just to continue verifying that there are no dangling AWS resources after a test is ran.

Please note that we do not recommend re-using test configurations between resources and their associated data source as it is harder to discover testing regressions. Authors are encouraged to potentially implement similar "base" configurations though.

For example:

```go
func TestAccAwsExampleThingDataSource_Name(t *testing.T) {
  rName := acctest.RandomWithPrefix("tf-acc-test")
  dataSourceName := "data.aws_example_thing.test"
  resourceName := "aws_example_thing.test"

  resource.ParallelTest(t, resource.TestCase{
    PreCheck:     func() { testAccPreCheck(t) },
    Providers:    testAccProviders,
    CheckDestroy: testAccCheckAwsExampleThingDestroy,
    Steps: []resource.TestStep{
      {
        Config: testAccAwsExampleThingDataSourceConfigName(rName),
        Check: resource.ComposeTestCheckFunc(
          testAccCheckAwsExampleThingExists(resourceName),
          resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
          resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
          resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
        ),
      },
    },
  })
}

// below all TestAcc functions

func testAccAwsExampleThingDataSourceConfigName(rName string) string {
  return fmt.Sprintf(`
resource "aws_example_thing" "test" {
  name = %[1]q
}

data "aws_example_thing" "test" {
  name = aws_example_thing.test.name
}
`, rName)
}
```

## Acceptance Test Sweepers

When running the acceptance tests, especially when developing or troubleshooting Terraform resources, its possible for code bugs or other issues to prevent the proper destruction of AWS infrastructure. To prevent lingering resources from consuming quota or causing unexpected billing, the Terraform Plugin SDK supports the test sweeper framework to clear out an AWS region of all resources. This section is meant to augment the [Extending Terraform documentation on test sweepers](https://www.terraform.io/docs/extend/testing/acceptance-tests/sweepers.html) with Terraform AWS Provider specific details.

### Running Test Sweepers

**WARNING: Test Sweepers will destroy AWS infrastructure and backups in the target AWS account and region! These are designed to override any API deletion protection. Never run these outside a development AWS account that should be completely empty of resources.**

To run the sweepers for all resources in `us-west-2` and `us-east-1` (default testing regions):

```console
$ make sweep
```

To run a specific resource sweeper:

```console
$ SWEEPARGS=-sweep-run=aws_example_thing make sweep
```

### Writing Test Sweepers

The first step is to initialize the resource into the test sweeper framework:

```go
func init() {
  resource.AddTestSweepers("aws_example_thing", &resource.Sweeper{
    Name: "aws_example_thing",
    F:    testSweepExampleThings,
    // Optionally
    Dependencies: []string{
      "aws_other_thing",
    },
  })
}
```

Then add the actual implementation. Preferably, if a paginated SDK call is available:

```go
func testSweepExampleThings(region string) error {
  client, err := sharedClientForRegion(region)

  if err != nil {
    return fmt.Errorf("error getting client: %w", err)
  }

  conn := client.(*AWSClient).exampleconn
  input := &example.ListThingsInput{}
  var sweeperErrs *multierror.Error

  err = conn.ListThingsPages(input, func(page *example.ListThingsOutput, isLast bool) bool {
    if page == nil {
      return !isLast
    }

    for _, thing := range page.Things {
      id := aws.StringValue(thing.Id)
      input := &example.DeleteThingInput{
        Id: thing.Id,
      }

      log.Printf("[INFO] Deleting Example Thing: %s", id)
      _, err := conn.DeleteThing(input)

      if err != nil {
        sweeperErr := fmt.Errorf("error deleting Example Thing (%s): %w", id, err)
        log.Printf("[ERROR] %s", sweeperErr)
        sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
        continue
      }
    }

    return !isLast
  })

  if testSweepSkipSweepError(err) {
    log.Printf("[WARN] Skipping Example Thing sweep for %s: %s", region, err)
    return sweeperErrs.ErrorOrNil()
  }

  if err != nil {
    sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Example Things: %w", err))
  }

  return sweeperErrs.ErrorOrNil()
}
```

Otherwise, if no paginated SDK call is available:

```go
func testSweepExampleThings(region string) error {
  client, err := sharedClientForRegion(region)

  if err != nil {
    return fmt.Errorf("error getting client: %w", err)
  }

  conn := client.(*AWSClient).exampleconn
  input := &example.ListThingsInput{}
  var sweeperErrs *multierror.Error

  for {
    output, err := conn.ListThings(input)

    if testSweepSkipSweepError(err) {
      log.Printf("[WARN] Skipping Example Thing sweep for %s: %s", region, err)
      return sweeperErrs.ErrorOrNil()
    }

    if err != nil {
      sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Example Thing: %w", err))
      return sweeperErrs
    }

    for _, thing := range output.Things {
      id := aws.StringValue(thing.Id)
      input := &example.DeleteThingInput{
        Id: thing.Id,
      }

      log.Printf("[INFO] Deleting Example Thing: %s", id)
      _, err := conn.DeleteThing(input)

      if err != nil {
        sweeperErr := fmt.Errorf("error deleting Example Thing (%s): %w", id, err)
        log.Printf("[ERROR] %s", sweeperErr)
        sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
        continue
      }
    }

    if aws.StringValue(output.NextToken) == "" {
      break
    }

    input.NextToken = output.NextToken
  }

  return sweeperErrs.ErrorOrNil()
}
```

## Acceptance Test Checklists

There are several aspects to writing good acceptance tests. These checklists will help ensure effective testing from the design stage through to implementation details.

### Basic Acceptance Test Design

These are basic principles to help guide the creation of acceptance tests.

- [ ] __Covers Changes__: Every line of resource or data source code added or changed should be covered by one or more tests. For example, if a resource has two ways of functioning, tests should cover both possible paths. Nearly every codebase change needs test coverage to ensure functionality and prevent future regressions. If a bug or other problem prompted a fix, a test should be added that previously would have failed, especially if the report included a configuration.
- [ ] __Follows the Single Responsibility Principle__: Every test should have a single responsibility and effectively test that responsibility. This may include individual tests for verifying basic functionality of the resource (Create, Read, Delete), separately verifying using and updating a single attribute in a resource, or separately changing between two attributes to verify two "modes"/"types" possible with a resource configuration. In following this principle, test configurations should be as simple as possible. For example, not including extra configuration unless it is necessary for the specific test.

### Test Implementation

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

### Avoid Hard Coding

Avoid hard coding values in acceptance test checks and configurations for consistency and testing flexibility. Resource testing is expected to pass across multiple AWS environments supported by the Terraform AWS Provider (e.g. AWS Standard and AWS GovCloud (US)). Contributors are not expected or required to perform testing outside of AWS Standard, e.g. running only in the `us-west-2` region is perfectly acceptable. However, contributors are expected to avoid hard coding with these guidelines.

#### Hardcoded Account IDs

- [ ] __Uses Account Data Sources__: Any hardcoded account numbers in configuration, e.g., `137112412989`, should be replaced with a data source. Depending on the situation, there are several data sources for account IDs including:
    - [`aws_caller_identity` data source](https://www.terraform.io/docs/providers/aws/d/caller_identity.html),
    - [`aws_canonical_user_id` data source](https://www.terraform.io/docs/providers/aws/d/canonical_user_id.html),
    - [`aws_billing_service_account` data source](https://www.terraform.io/docs/providers/aws/d/billing_service_account.html), and
    - [`aws_sagemaker_prebuilt_ecr_image` data source](https://www.terraform.io/docs/providers/aws/d/sagemaker_prebuilt_ecr_image.html).
- [ ] __Uses Account Test Checks__: Any check required to verify an AWS Account ID of the current testing account or another account should use one of the following available helper functions over the usage of `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`:
    - `testAccCheckResourceAttrAccountID()`: Validates the state value equals the AWS Account ID of the current account running the test. This is the most common implementation.
    - `testAccMatchResourceAttrAccountID()`: Validates the state value matches any AWS Account ID (e.g. a 12 digit number). This is typically only used in data source testing of AWS managed components.

Here's an example of using `aws_caller_identity`:

```hcl
data "aws_caller_identity" "current" {}

resource "aws_backup_selection" "test" {
  plan_id      = aws_backup_plan.test.id
  name         = "tf_acc_test_backup_selection_%[1]d"
  iam_role_arn = "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:role/service-role/AWSBackupDefaultServiceRole"
}
```

#### Hardcoded AMI IDs

- [ ] __Uses aws_ami Data Source__: Any hardcoded AMI ID configuration, e.g. `ami-12345678`, should be replaced with the [`aws_ami` data source](https://www.terraform.io/docs/providers/aws/d/ami.html) pointing to an Amazon Linux image. The codebase includes test configuration helper functions to simplify these lookups:
    - `testAccLatestAmazonLinuxHvmEbsAmiConfig()`: The recommended AMI for most situations, using Amazon Linux, HVM virtualization, and EBS storage. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-hvm-ebs.id`.
    - `testAccLatestAmazonLinuxHvmInstanceStoreAmiConfig()`: AMI lookup using Amazon Linux, HVM virtualization, and Instance Store storage. Should only be used in testing that requires Instance Store storage rather than EBS. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-hvm-instance-store.id`.
    - `testAccLatestAmazonLinuxPvEbsAmiConfig()`: AMI lookup using Amazon Linux, Paravirtual virtualization, and EBS storage. Should only be used in testing that requires Paravirtual over Hardware Virtual Machine (HVM) virtualization. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-pv-ebs.id`.
    - `testAccLatestAmazonLinuxPvInstanceStoreAmiConfig`: AMI lookup using Amazon Linux, Paravirtual virtualization, and Instance Store storage. Should only be used in testing that requires Paravirtual virtualization over HVM and Instance Store storage over EBS. To reference the AMI ID in the test configuration: `data.aws_ami.amzn-ami-minimal-pv-instance-store.id`.
    - `testAccLatestWindowsServer2016CoreAmiConfig()`: AMI lookup using Windows Server 2016 Core, HVM virtualization, and EBS storage. Should only be used in testing that requires Windows. To reference the AMI ID in the test configuration: `data.aws_ami.win2016core-ami.id`.

Here's an example of using `testAccLatestAmazonLinuxHvmEbsAmiConfig()` and `data.aws_ami.amzn-ami-minimal-hvm-ebs.id`:

```go
func testAccLaunchConfigurationDataSourceConfig_basic(rName string) string {
	return composeConfig(testAccLatestAmazonLinuxHvmEbsAmiConfig(), fmt.Sprintf(`
resource "aws_launch_configuration" "test" {
  name          = %[1]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = "m1.small"
}
`, rName))
}
```

#### Hardcoded Availability Zones

- [ ] __Uses aws_availability_zones Data Source__: Any hardcoded AWS Availability Zone configuration, e.g. `us-west-2a`, should be replaced with the [`aws_availability_zones` data source](https://www.terraform.io/docs/providers/aws/d/availability_zones.html). Use the convenience function called `testAccAvailableAZsNoOptInConfig()` (defined in `resource_aws_instance_test.go`) to declare `data "aws_availability_zones" "available" {...}`. You can then reference the data source via `data.aws_availability_zones.available.names[0]` or `data.aws_availability_zones.available.names[count.index]` in resources utilizing `count`.

Here's an example of using `testAccAvailableAZsNoOptInConfig()` and `data.aws_availability_zones.available.names[0]`:

```go
func testAccAwsInstanceVpcConfigBasic(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id
}
`, rName))
}
```

#### Hardcoded Versions

- [ ] __Uses Version Data Sources__: Hardcoded versions, e.g. RDS MySQL Engine Version `5.7.42`, should be removed (which means the AWS-defined default version will be used) or replaced with a list of preferred versions using a data source. Because versions change over times and version offerings vary from region to region and partition to partition, using the default version or providing a list of preferences ensures a version will be available. Depending on the situation, there are several data sources for versions, including:
    - [`aws_rds_engine_version` data source](https://www.terraform.io/docs/providers/aws/d/rds_engine_version.html),
    - [`aws_docdb_engine_version` data source](https://www.terraform.io/docs/providers/aws/d/docdb_engine_version.html), and
    - [`aws_neptune_engine_version` data source](https://www.terraform.io/docs/providers/aws/d/neptune_engine_version.html).

Here's an example of using `aws_rds_engine_version` and `data.aws_rds_engine_version.default.version`:

```hcl
data "aws_rds_engine_version" "default" {
  engine = "mysql"
}

data "aws_rds_orderable_db_instance" "test" {
  engine                     = data.aws_rds_engine_version.default.engine
  engine_version             = data.aws_rds_engine_version.default.version
  preferred_instance_classes = ["db.t3.small", "db.t2.small", "db.t2.medium"]
}

resource "aws_db_instance" "bar" {
  engine               = data.aws_rds_engine_version.default.engine
  engine_version       = data.aws_rds_engine_version.default.version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot  = true
  parameter_group_name = "default.${data.aws_rds_engine_version.default.parameter_group_family}"
}
```

#### Hardcoded Instance Types

- [ ] __Uses Instance Type Data Source__: Singular hardcoded instance types and classes, e.g., `t2.micro` and `db.t2.micro`, should be replaced with a list of preferences using a data source. Because offerings vary from region to region and partition to partition, providing a list of preferences dramatically improves the likelihood that one of the options will be available. Depending on the situation, there are several data sources for instance types and classes, including:
    - [`aws_ec2_instance_type_offering` data source](https://www.terraform.io/docs/providers/aws/d/ec2_instance_type_offering.html) - Convenience functions declare configurations that are referenced with `data.aws_ec2_instance_type_offering.available` including:
        - The `testAccAvailableEc2InstanceTypeForAvailabilityZone()` function for test configurations using an EC2 Subnet which is inherently within a single Availability Zone
        - The `testAccAvailableEc2InstanceTypeForRegion()` function for test configurations that do not include specific Availability Zones
    - [`aws_rds_orderable_db_instance` data source](https://www.terraform.io/docs/providers/aws/d/rds_orderable_db_instance.html),
    - [`aws_neptune_orderable_db_instance` data source](https://www.terraform.io/docs/providers/aws/d/neptune_orderable_db_instance.html), and
    - [`aws_docdb_orderable_db_instance` data source](https://www.terraform.io/docs/providers/aws/d/docdb_orderable_db_instance.html).

Here's an example of using `testAccAvailableEc2InstanceTypeForRegion()` and `data.aws_ec2_instance_type_offering.available.instance_type`:

```go
func testAccAWSSpotInstanceRequestConfig(rInt int) string {
	return composeConfig(testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_spot_instance_request" "test" {
  instance_type        = data.aws_ec2_instance_type_offering.available.instance_type
  spot_price           = "0.05"
  wait_for_fulfillment = true
}
`, rInt))
}
```

Here's an example of using `aws_rds_orderable_db_instance` and `data.aws_rds_orderable_db_instance.test.instance_class`:

```hcl
data "aws_rds_orderable_db_instance" "test" {
  engine                     = "mysql"
  engine_version             = "5.7.31"
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t3.small"]
}

resource "aws_db_instance" "test" {
  engine              = data.aws_rds_orderable_db_instance.test.engine
  engine_version      = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class      = data.aws_rds_orderable_db_instance.test.instance_class
  skip_final_snapshot = true
  username            = "test"
}
```

#### Hardcoded Partition DNS Suffix

- [ ] __Uses aws_partition Data Source__: Any hardcoded DNS suffix configuration, e.g. the `amazonaws.com` in a `ec2.amazonaws.com` service principal, should be replaced with the [`aws_partition` data source](https://www.terraform.io/docs/providers/aws/d/partition.html). A common pattern is declaring `data "aws_partition" "current" {}` and referencing it via `data.aws_partition.current.dns_suffix`.

Here's an example of using `aws_partition` and `data.aws_partition.current.dns_suffix`:

```hcl
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "cloudtrail.${data.aws_partition.current.dns_suffix}"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}
```

#### Hardcoded Partition in ARN

- [ ] __Uses aws_partition Data Source__: Any hardcoded AWS Partition configuration, e.g. the `aws` in a `arn:aws:SERVICE:REGION:ACCOUNT:RESOURCE` ARN, should be replaced with the [`aws_partition` data source](https://www.terraform.io/docs/providers/aws/d/partition.html). A common pattern is declaring `data "aws_partition" "current" {}` and referencing it via `data.aws_partition.current.partition`.
- [ ] __Uses Builtin ARN Check Functions__: Tests should utilize available ARN check functions to validate ARN attribute values in the Terraform state over `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`:

    - `testAccCheckResourceAttrRegionalARN()` verifies an ARN matches the account ID and region of the test execution and an exact resource value
    - `testAccMatchResourceAttrRegionalARN()` verifies an ARN matches the account ID and region of the test execution and a regular expression of the resource value
    - `testAccCheckResourceAttrGlobalARN()` verifies an ARN matches the account ID of the test execution and an exact resource value
    - `testAccMatchResourceAttrGlobalARN()` verifies an ARN matches the account ID of the test execution and a regular expression of the resource value

Here's an example of using `aws_partition` and `data.aws_partition.current.partition`:

```hcl
data "aws_partition" "current" {}

resource "aws_iam_role_policy_attachment" "test" {
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/service-role/AmazonRDSEnhancedMonitoringRole"
  role       = aws_iam_role.test.name
}
```

#### Hardcoded Region

- [ ] __Uses aws_region Data Source__: Any hardcoded AWS Region configuration, e.g. `us-west-2`, should be replaced with the [`aws_region` data source](https://www.terraform.io/docs/providers/aws/d/region.html). A common pattern is declaring `data "aws_region" "current" {}` and referencing it via `data.aws_region.current.name`

Here's an example of using `aws_region` and `data.aws_region.current.name`:

```hcl
data "aws_region" "current" {}

resource "aws_route53_zone" "test" {
  vpc {
    vpc_id     = aws_vpc.test.id
    vpc_region = data.aws_region.current.name
  }
}
```

#### Hardcoded Spot Price

- [ ] __Uses aws_ec2_spot_price Data Source__: Any hardcoded spot prices, e.g. `0.05`, should be replaced with the [`aws_ec2_spot_price` data source](https://www.terraform.io/docs/providers/aws/d/ec2_spot_price.html). A common pattern is declaring `data "aws_ec2_spot_price" "current" {}` and referencing it via `data.aws_ec2_spot_price.current.spot_price`.

Here's an example of using `aws_ec2_spot_price` and `data.aws_ec2_spot_price.current.spot_price`:

```hcl
data "aws_ec2_spot_price" "current" {
  instance_type = "t3.medium"

  filter {
    name   = "product-description"
    values = ["Linux/UNIX"]
  }
}

resource "aws_spot_fleet_request" "test" {
  spot_price      = data.aws_ec2_spot_price.current.spot_price
  target_capacity = 2
}
```
