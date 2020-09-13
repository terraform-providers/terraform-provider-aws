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
        - [Service Availability PreCheck](#service-availability-precheck)
        - [Disappears Acceptance Tests](#disappears-acceptance-tests)
        - [Per Attribute Acceptance Tests](#per-attribute-acceptance-tests)
        - [Cross-Account Acceptance Tests](#cross-account-acceptance-tests)
        - [Cross-Region Acceptance Tests](#cross-region-acceptance-tests)
    - [Data Source Acceptance Testing](#data-source-acceptance-testing)
- [Acceptance Test Sweepers](#acceptance-test-sweepers)
    - [Running Test Sweepers](#running-test-sweepers)
    - [Writing Test Sweepers](#writing-test-sweepers)
- [Acceptance Test Checklist](#acceptance-test-checklist)

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
ok  	github.com/terraform-providers/terraform-provider-aws/aws	26.607s
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
ok  	github.com/terraform-providers/terraform-provider-aws/aws	55.619s
```

Running acceptance tests requires version 0.12.26 or higher of the Terraform CLI to be installed.

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
FAIL	github.com/terraform-providers/terraform-provider-aws/aws	4.305s
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

#### Service Availability PreCheck

When new AWS services are added to the provider, a simple read-only request (e.g. list all X service things) should be implemented in an acceptance test PreCheck function that helps the acceptance testing determine if the service exists for the particular environment it runs in. Note: This was not common practice in the past so many existing tests will not include this step.

For example:

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
- Switch usage of `Providers: testAccProviders` to `ProviderFactories: testAccProviderFactories(&providers)`
- Add `testAccAlternateAccountProviderConfig()` to the test configuration and use `provider = "awsalternate"` for cross-account resources. The resource that is the focus of the acceptance test should _not_ use the alternate provider identification to simplify the testing setup.
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
# The standardized provider block to use is awsalternate as seen below.
resource "aws_cross_account_example" "test" {
  provider = "awsalternate"

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
- Switch usage of `Providers: testAccProviders` to `ProviderFactories: testAccProviderFactories(&providers)`
- Add `testAccMultipleRegionProviderConfig(###)` to the test configuration and use `provider = "awsalternate"` (and/or `provider = "awsthird"`) for cross-region resources. The resource that is the focus of the acceptance test should _not_ use the alternative providers to simplify the testing setup. If the infrastructure in the second AWS region is also in a second AWS account use `testAccAlternateAccountAlternateRegionProviderConfig()` instead
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
  return testAccMultipleRegionProviderConfig(2) + fmt.Sprintf(`
# Cross region resources should be handled by the cross region provider.
# The standardized provider is awsalternate as seen below.
resource "aws_cross_region_example" "test" {
  provider = "awsalternate"

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

#### Please Note

When adding a new provider to the codebase for the purposes of cross-account/cross-region testing, please ensure the provider name in the config matches an entry in the list of factories in `provider_test.go`

```hcl
# provider block, ensure name does not include periods '.'
provider "awsnewalternate" {
  region = "us-west-3"
}
```

```go
// provider_testo.go in init()
var providerNames = []string{"aws", "awseast", "awswest", "awsalternate", /* ... */ "awsnewalternate"}
```

Searching for usage of `testAccMultipleRegionPreCheck` in the codebase will yield real world examples of this setup in action.

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

## Acceptance Test Checklist

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

- [ ] __Uses aws_availability_zones Data Source__: Any hardcoded AWS Availability Zone configuration, e.g. `us-west-2a`, should be replaced with the [`aws_availability_zones` data source](https://www.terraform.io/docs/providers/aws/d/availability_zones.html). Use the convenience function called `testAccAvailableAZsNoOptInConfig()` (defined in `resource_aws_instance_test.go`) to declare `data "aws_availability_zones" "available" {...}`. You can then reference the data source via `data.aws_availability_zones.available.names[0]` or `data.aws_availability_zones.available.names[count.index]` in resources utilizing `count`.

Here's an example of using `testAccAvailableAZsNoOptInConfig()` and `data.aws_availability_zones.available.names[0]`:

```go
func testAccAwsInstanceVpcConfigBasic(rName string) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.0.0.0/24"
  vpc_id                  = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}
```

- [ ] __Uses aws_region Data Source__: Any hardcoded AWS Region configuration, e.g. `us-west-2`, should be replaced with the [`aws_region` data source](https://www.terraform.io/docs/providers/aws/d/region.html). A common pattern is declaring `data "aws_region" "current" {}` and referencing it via `data.aws_region.current.name`
- [ ] __Uses aws_partition Data Source__: Any hardcoded AWS Partition configuration, e.g. the `aws` in a `arn:aws:SERVICE:REGION:ACCOUNT:RESOURCE` ARN, should be replaced with the [`aws_partition` data source](https://www.terraform.io/docs/providers/aws/d/partition.html). A common pattern is declaring `data "aws_partition" "current" {}` and referencing it via `data.aws_partition.current.partition`
- [ ] __Uses Builtin ARN Check Functions__: Tests should utilize available ARN check functions, e.g. `testAccMatchResourceAttrRegionalARN()`, to validate ARN attribute values in the Terraform state over `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`
- [ ] __Uses testAccCheckResourceAttrAccountID()__: Tests should utilize the available AWS Account ID check function, `testAccCheckResourceAttrAccountID()` to validate account ID attribute values in the Terraform state over `resource.TestCheckResourceAttrSet()` and `resource.TestMatchResourceAttr()`
