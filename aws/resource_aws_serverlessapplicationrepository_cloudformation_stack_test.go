package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

// Since aws_serverlessapplicationrepository_cloudformation_stack creates CloudFormation stacks,
// the aws_cloudformation_stack sweeper will clean these up as well.

func TestAccAwsServerlessApplicationRepositoryCloudFormationStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessApplicationRepositoryCloudFormationStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "application_id", "serverlessrepo", "applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttrSet(resourceName, "semantic_version"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", fmt.Sprintf("func-%s", stackName)),
					testAccCheckResourceAttrRegionalHostnameService(resourceName, "parameters.endpoint", "secretsmanager"),
					resource.TestCheckResourceAttr(resourceName, "outputs.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "outputs.RotationLambdaARN"),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsServerlessApplicationRepositoryCloudFormationStackNameImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAwsServerlessApplicationRepositoryCloudFormationStackNameNoPrefixImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsServerlessApplicationRepositoryCloudFormationStack_disappears(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAmiDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessApplicationRepositoryCloudFormationStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsServerlessApplicationRepositoryCloudFormationStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsServerlessApplicationRepositoryCloudFormationStack_versioned(t *testing.T) {
	var stack1, stack2, stack3 cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	const (
		version1 = "1.0.13"
		version2 = "1.1.36"
	)

	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_versioned(stackName, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack1),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_versioned2(stackName, version2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack2),
					testAccCheckCloudFormationStackNotRecreated(&stack1, &stack2),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version2),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
			{
				// Confirm removal of "CAPABILITY_RESOURCE_POLICY" is handled properly
				Config: testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_versioned(stackName, version1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack3),
					testAccCheckCloudFormationStackNotRecreated(&stack2, &stack3),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version1),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
				),
			},
		},
	})
}

func TestAccAwsServerlessApplicationRepositoryCloudFormationStack_paired(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	const version = "1.1.36"

	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_versionedPaired(stackName, version),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "semantic_version", version),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_RESOURCE_POLICY"),
				),
			},
		},
	})
}

func TestAccAwsServerlessApplicationRepositoryCloudFormationStack_Tags(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsServerlessApplicationRepositoryCloudFormationStackConfigTags1(stackName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
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
				Config: testAccAwsServerlessApplicationRepositoryCloudFormationStackConfigTags2(stackName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsServerlessApplicationRepositoryCloudFormationStackConfigTags1(stackName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAwsServerlessApplicationRepositoryCloudFormationStack_update(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	initialName := acctest.RandomWithPrefix("FuncName1")
	updatedName := acctest.RandomWithPrefix("FuncName2")

	resourceName := "aws_serverlessapplicationrepository_cloudformation_stack.postgres-rotator"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_updateInitial(stackName, initialName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
					testAccCheckResourceAttrRegionalARNIgnoreRegionAndAccount(resourceName, "application_id", "serverlessrepo", "applications/SecretsManagerRDSPostgreSQLRotationSingleUser"),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", initialName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
			{
				Config: testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_updateUpdated(stackName, updatedName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.functionName", updatedName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key", "value"),
				),
			},
		},
	})
}

func testAccCheckServerlessApplicationRepositoryCloudFormationStackExists(n string, stack *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cfconn
		params := &cloudformation.DescribeStacksInput{
			StackName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeStacks(params)
		if err != nil {
			return err
		}
		if len(resp.Stacks) == 0 {
			return fmt.Errorf("CloudFormation stack (%s) not found", rs.Primary.ID)
		}

		*stack = *resp.Stacks[0]

		return nil
	}
}

func testAccAwsServerlessApplicationRepositoryCloudFormationStackNameImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s%s", serverlessApplicationRepositoryCloudFormationStackNamePrefix, rs.Primary.Attributes["name"]), nil
	}
}

func testAccAwsServerlessApplicationRepositoryCloudFormationStackNameNoPrefixImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["name"], nil
	}
}

func testAccAwsServerlessApplicationRepositoryCloudFormationStackConfig(stackName string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = local.postgres_single_user_rotator_arn
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_region" "current" {}
`, stackName))
}

func testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_updateInitial(stackName, functionName string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = local.postgres_single_user_rotator_arn
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "%[2]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    key = "value"
  }
}

data "aws_region" "current" {}
`, stackName, functionName))
}

func testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_updateUpdated(stackName, functionName string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = local.postgres_single_user_rotator_arn
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "%[2]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    key = "value"
  }
}

data "aws_region" "current" {}
`, stackName, functionName))
}

func testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_versioned(stackName, version string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name             = "%[1]s"
  application_id   = local.postgres_single_user_rotator_arn
  semantic_version = "%[2]s"
  capabilities = [
    "CAPABILITY_IAM",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_region" "current" {}
`, stackName, version))
}

func testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_versioned2(stackName, version string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = local.postgres_single_user_rotator_arn
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  semantic_version = "%[2]s"
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_region" "current" {}
`, stackName, version))
}

func testAccAWSServerlessApplicationRepositoryCloudFormationStackConfig_versionedPaired(stackName, version string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name             = "%[1]s"
  application_id   = data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator.application_id
  semantic_version = data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator.semantic_version
  capabilities     = data.aws_serverlessapplicationrepository_application.secrets_manager_postgres_single_user_rotator.required_capabilities
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
}

data "aws_serverlessapplicationrepository_application" "secrets_manager_postgres_single_user_rotator" {
  application_id   = local.postgres_single_user_rotator_arn
  semantic_version = "%[2]s"
}

data "aws_region" "current" {}
`, stackName, version))
}

func testAccAwsServerlessApplicationRepositoryCloudFormationStackConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = local.postgres_single_user_rotator_arn
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    %[2]q = %[3]q
  }
}

data "aws_region" "current" {}
`, rName, tagKey1, tagValue1))
}

func testAccAwsServerlessApplicationRepositoryCloudFormationStackConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication,
		fmt.Sprintf(`
resource "aws_serverlessapplicationrepository_cloudformation_stack" "postgres-rotator" {
  name           = "%[1]s"
  application_id = local.postgres_single_user_rotator_arn
  capabilities = [
    "CAPABILITY_IAM",
    "CAPABILITY_RESOURCE_POLICY",
  ]
  parameters = {
    functionName = "func-%[1]s"
    endpoint     = "secretsmanager.${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}"
  }
  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

data "aws_region" "current" {}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

const testAccCheckAwsServerlessApplicationRepositoryPostgresSingleUserRotatorApplication = `
data "aws_partition" "current" {}

locals {
  postgres_single_user_rotator_arn = "arn:${data.aws_partition.current.partition}:serverlessrepo:${local.application_region}:${local.application_account}:applications/SecretsManagerRDSPostgreSQLRotationSingleUser"

  application_region  = local.security_manager_regions[data.aws_partition.current.partition]
  application_account = local.security_manager_accounts[data.aws_partition.current.partition]

  security_manager_regions = {
    "aws"        = "us-east-1",
    "aws-us-gov" = "us-gov-west-1",
  }
  security_manager_accounts = {
    "aws"        = "297356227824",
    "aws-us-gov" = "023102451235",
  }
}
`
