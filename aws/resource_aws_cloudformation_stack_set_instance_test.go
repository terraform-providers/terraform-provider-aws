package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudformation/waiter"
)

func init() {
	resource.AddTestSweepers("aws_cloudformation_stack_set_instance", &resource.Sweeper{
		Name: "aws_cloudformation_stack_set_instance",
		F:    testSweepCloudformationStackSetInstances,
	})
}

func testSweepCloudformationStackSetInstances(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).cfconn
	stackSets, err := listCloudFormationStackSets(conn)

	if testSweepSkipSweepError(err) || isAWSErr(err, "ValidationError", "AWS CloudFormation StackSets is not supported") {
		log.Printf("[WARN] Skipping CloudFormation StackSet Instance sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation StackSets: %w", err)
	}

	var sweeperErrs *multierror.Error

	for _, stackSet := range stackSets {
		stackSetName := aws.StringValue(stackSet.StackSetName)
		instances, err := listCloudFormationStackSetInstances(conn, stackSetName)

		if err != nil {
			sweeperErr := fmt.Errorf("error listing CloudFormation StackSet (%s) Instances: %w", stackSetName, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}

		for _, instance := range instances {
			accountID := aws.StringValue(instance.Account)
			region := aws.StringValue(instance.Region)
			id := fmt.Sprintf("%s / %s / %s", stackSetName, accountID, region)

			input := &cloudformation.DeleteStackInstancesInput{
				Accounts:     aws.StringSlice([]string{accountID}),
				OperationId:  aws.String(resource.UniqueId()),
				Regions:      aws.StringSlice([]string{region}),
				RetainStacks: aws.Bool(false),
				StackSetName: aws.String(stackSetName),
			}

			log.Printf("[INFO] Deleting CloudFormation StackSet Instance: %s", id)
			output, err := conn.DeleteStackInstances(input)

			if isAWSErr(err, cloudformation.ErrCodeStackInstanceNotFoundException, "") {
				continue
			}

			if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudFormation StackSet Instance (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if err := waiter.StackSetOperationSucceeded(conn, stackSetName, aws.StringValue(output.OperationId), waiter.StackSetInstanceDeletedDefaultTimeout); err != nil {
				sweeperErr := fmt.Errorf("error waiting for CloudFormation StackSet Instance (%s) deletion: %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudFormationStackSetInstance_basic(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	cloudformationStackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					testAccCheckResourceAttrAccountID(resourceName, "account_id"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "stack_id"),
					resource.TestCheckResourceAttrPair(resourceName, "stack_set_name", cloudformationStackSetResourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
		},
	})
}

func TestAccAWSCloudFormationStackSetInstance_disappears(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationStackSetInstance(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFormationStackSetInstance_disappears_StackSet(t *testing.T) {
	var stackInstance1 cloudformation.StackInstance
	var stackSet1 cloudformation.StackSet
	rName := acctest.RandomWithPrefix("tf-acc-test")
	stackSetResourceName := "aws_cloudformation_stack_set.test"
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetExists(stackSetResourceName, &stackSet1),
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationStackSetInstance(), resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationStackSet(), stackSetResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFormationStackSetInstance_ParameterOverrides(t *testing.T) {
	var stackInstance1, stackInstance2, stackInstance3, stackInstance4 cloudformation.StackInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides1(rName, "overridevalue1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides2(rName, "overridevalue1updated", "overridevalue2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance2),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter2", "overridevalue2"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides1(rName, "overridevalue1updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance3),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.Parameter1", "overridevalue1updated"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance4),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance3, &stackInstance4),
					resource.TestCheckResourceAttr(resourceName, "parameter_overrides.%", "0"),
				),
			},
		},
	})
}

// TestAccAWSCloudFrontDistribution_RetainStack verifies retain_stack = true
// This acceptance test performs the following steps:
//  * Trigger a Terraform destroy of the resource, which should only remove the instance from the StackSet
//  * Check it still exists outside Terraform
//  * Destroy for real outside Terraform
func TestAccAWSCloudFormationStackSetInstance_RetainStack(t *testing.T) {
	var stack1 cloudformation.Stack
	var stackInstance1, stackInstance2, stackInstance3 cloudformation.StackInstance
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack_set_instance.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSCloudFormationStackSet(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationStackSetInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance1),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"retain_stack",
				},
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance2),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance1, &stackInstance2),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "false"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceExists(resourceName, &stackInstance3),
					testAccCheckCloudFormationStackSetInstanceNotRecreated(&stackInstance2, &stackInstance3),
					resource.TestCheckResourceAttr(resourceName, "retain_stack", "true"),
				),
			},
			{
				Config:  testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName, true),
				Destroy: true,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackSetInstanceStackExists(&stackInstance3, &stack1),
					testAccCheckCloudFormationStackDisappears(&stack1),
				),
			},
		},
	})
}

func testAccCheckCloudFormationStackSetInstanceExists(resourceName string, stackInstance *cloudformation.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]

		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).cfconn

		stackSetName, accountID, region, err := resourceAwsCloudFormationStackSetInstanceParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &cloudformation.DescribeStackInstanceInput{
			StackInstanceAccount: aws.String(accountID),
			StackInstanceRegion:  aws.String(region),
			StackSetName:         aws.String(stackSetName),
		}

		output, err := conn.DescribeStackInstance(input)

		if err != nil {
			return err
		}

		if output == nil || output.StackInstance == nil {
			return fmt.Errorf("CloudFormation StackSet Instance (%s) not found", rs.Primary.ID)
		}

		*stackInstance = *output.StackInstance

		return nil
	}
}

func testAccCheckCloudFormationStackSetInstanceStackExists(stackInstance *cloudformation.StackInstance, stack *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cfconn

		input := &cloudformation.DescribeStacksInput{
			StackName: stackInstance.StackId,
		}

		output, err := conn.DescribeStacks(input)

		if err != nil {
			return err
		}

		if len(output.Stacks) == 0 || output.Stacks[0] == nil {
			return fmt.Errorf("CloudFormation Stack (%s) not found", aws.StringValue(stackInstance.StackId))
		}

		*stack = *output.Stacks[0]

		return nil
	}
}

func testAccCheckAWSCloudFormationStackSetInstanceDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack_set_instance" {
			continue
		}

		stackSetName, accountID, region, err := resourceAwsCloudFormationStackSetInstanceParseId(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &cloudformation.DescribeStackInstanceInput{
			StackInstanceAccount: aws.String(accountID),
			StackInstanceRegion:  aws.String(region),
			StackSetName:         aws.String(stackSetName),
		}

		output, err := conn.DescribeStackInstance(input)

		if isAWSErr(err, cloudformation.ErrCodeStackInstanceNotFoundException, "") {
			return nil
		}

		if isAWSErr(err, cloudformation.ErrCodeStackSetNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		if output != nil && output.StackInstance != nil {
			return fmt.Errorf("CloudFormation StackSet Instance (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudFormationStackSetInstanceNotRecreated(i, j *cloudformation.StackInstance) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackId) != aws.StringValue(j.StackId) {
			return fmt.Errorf("CloudFormation StackSet Instance (%s,%s,%s) recreated", aws.StringValue(i.StackSetId), aws.StringValue(i.Account), aws.StringValue(i.Region))
		}

		return nil
	}
}

func testAccAWSCloudFormationStackSetInstanceConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "Administration" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "cloudformation.amazonaws.com"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Administration"
}

resource "aws_iam_role_policy" "Administration" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Administration.name
}

resource "aws_iam_role" "Execution" {
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "AWS": [
          "${aws_iam_role.Administration.arn}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF

  name = "%[1]s-Execution"
}

resource "aws_iam_role_policy" "Execution" {
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Resource": [
        "*"
      ],
      "Action": [
        "*"
      ]
    }
  ]
}
EOF

  role = aws_iam_role.Execution.name
}

resource "aws_cloudformation_stack_set" "test" {
  depends_on = [aws_iam_role_policy.Execution]

  administration_role_arn = aws_iam_role.Administration.arn
  execution_role_name     = aws_iam_role.Execution.name
  name                    = %[1]q

  parameters = {
    Parameter1 = "stacksetvalue1"
    Parameter2 = "stacksetvalue2"
  }

  template_body = <<TEMPLATE
Parameters:
  Parameter1:
    Type: String
  Parameter2:
    Type: String
Resources:
  TestVpc:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        - Key: Name
          Value: %[1]q
Outputs:
  Parameter1Value:
    Value: !Ref Parameter1
  Parameter2Value:
    Value: !Ref Parameter2
  Region:
    Value: !Ref "AWS::Region"
  TestVpcID:
    Value: !Ref TestVpc
TEMPLATE
}
`, rName)
}

func testAccAWSCloudFormationStackSetInstanceConfig(rName string) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + `
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`
}

func testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides1(rName, value1 string) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1)
}

func testAccAWSCloudFormationStackSetInstanceConfigParameterOverrides2(rName, value1, value2 string) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  parameter_overrides = {
    Parameter1 = %[1]q
    Parameter2 = %[2]q
  }

  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, value1, value2)
}

func testAccAWSCloudFormationStackSetInstanceConfigRetainStack(rName string, retainStack bool) string {
	return testAccAWSCloudFormationStackSetInstanceConfigBase(rName) + fmt.Sprintf(`
resource "aws_cloudformation_stack_set_instance" "test" {
  depends_on = [aws_iam_role_policy.Administration, aws_iam_role_policy.Execution]

  retain_stack   = %[1]t
  stack_set_name = aws_cloudformation_stack_set.test.name
}
`, retainStack)
}
