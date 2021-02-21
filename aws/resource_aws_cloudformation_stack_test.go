package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudformation_stack", &resource.Sweeper{
		Name: "aws_cloudformation_stack",
		Dependencies: []string{
			"aws_cloudformation_stack_set_instance",
		},
		F: testSweepCloudformationStacks,
	})
}

func testSweepCloudformationStacks(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).cfconn
	input := &cloudformation.ListStacksInput{
		StackStatusFilter: aws.StringSlice([]string{
			cloudformation.StackStatusCreateComplete,
			cloudformation.StackStatusImportComplete,
			cloudformation.StackStatusRollbackComplete,
			cloudformation.StackStatusUpdateComplete,
		}),
	}
	var sweeperErrs *multierror.Error

	err = conn.ListStacksPages(input, func(page *cloudformation.ListStacksOutput, lastPage bool) bool {
		for _, stack := range page.StackSummaries {
			input := &cloudformation.DeleteStackInput{
				StackName: stack.StackName,
			}
			name := aws.StringValue(stack.StackName)

			log.Printf("[INFO] Deleting CloudFormation Stack: %s", name)
			_, err := conn.DeleteStack(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudFormation Stack (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudFormation Stack sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error listing CloudFormation Stacks: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudFormationStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test-basic")
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckNoResourceAttr(resourceName, "on_failure"),
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

func TestAccAWSCloudFormationStack_CreationFailure_DoNothing(t *testing.T) {
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudFormationStackConfigCreationFailure(stackName, cloudformation.OnFailureDoNothing),
				ExpectError: regexp.MustCompile(`failed to create CloudFormation stack \(CREATE_FAILED\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_CreationFailure_Delete(t *testing.T) {
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudFormationStackConfigCreationFailure(stackName, cloudformation.OnFailureDelete),
				ExpectError: regexp.MustCompile(`failed to create CloudFormation stack, delete requested \(DELETE_COMPLETE\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_CreationFailure_Rollback(t *testing.T) {
	stackName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudFormationStackConfigCreationFailure(stackName, cloudformation.OnFailureRollback),
				ExpectError: regexp.MustCompile(`failed to create CloudFormation stack, rollback requested \(ROLLBACK_COMPLETE\).*The following resource\(s\) failed to create.*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_UpdateFailure(t *testing.T) {
	var stack cloudformation.Stack
	stackName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_cloudformation_stack.test"

	vpcCidrInitial := "10.0.0.0/16"
	vpcCidrInvalid := "1000.0.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_withParams(stackName, vpcCidrInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
				),
			},
			{
				Config:      testAccAWSCloudFormationStackConfig_withParams(stackName, vpcCidrInvalid),
				ExpectError: regexp.MustCompile(`failed to update CloudFormation stack \(UPDATE_ROLLBACK_COMPLETE\).*This is not a valid CIDR block`),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_disappears(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-basic-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFormationStack(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFormationStack_yaml(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-yaml-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_yaml(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
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

func TestAccAWSCloudFormationStack_defaultParams(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-default-params-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_defaultParams(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
		},
	})
}

func TestAccAWSCloudFormationStack_allAttributes(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-all-attributes-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"
	expectedPolicyBody := "{\"Statement\":[{\"Action\":\"Update:*\",\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"LogicalResourceId/StaticVPC\"},{\"Action\":\"Update:*\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"*\"}]}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_allAttributesWithBodies(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.First", "Mickey"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", "10"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "policy_body"},
			},
			{
				Config: testAccAWSCloudFormationStackConfig_allAttributesWithBodies_modified(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "name", stackName),
					resource.TestCheckResourceAttr(resourceName, "capabilities.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "capabilities.*", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr(resourceName, "disable_rollback", "false"),
					resource.TestCheckResourceAttr(resourceName, "notification_arns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr(resourceName, "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.First", "Mickey"),
					resource.TestCheckResourceAttr(resourceName, "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr(resourceName, "timeout_in_minutes", "10"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4332
func TestAccAWSCloudFormationStack_withParams(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-with-params-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"

	vpcCidrInitial := "10.0.0.0/16"
	vpcCidrUpdated := "12.0.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_withParams(stackName, vpcCidrInitial),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", vpcCidrInitial),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters"},
			},
			{
				Config: testAccAWSCloudFormationStackConfig_withParams(stackName, vpcCidrUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
					resource.TestCheckResourceAttr(resourceName, "parameters.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.VpcCIDR", vpcCidrUpdated),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4534
func TestAccAWSCloudFormationStack_withUrl_withParams(t *testing.T) {
	var stack cloudformation.Stack
	rName := fmt.Sprintf("tf-acc-test-with-url-and-params-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "template_url"},
			},
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack.json", "13.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_withUrl_withParams_withYaml(t *testing.T) {
	var stack cloudformation.Stack
	rName := fmt.Sprintf("tf-acc-test-with-params-and-yaml-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams_withYaml(rName, "tf-cf-stack.test", "13.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "template_url"},
			},
		},
	})
}

// Test for https://github.com/hashicorp/terraform/issues/5653
func TestAccAWSCloudFormationStack_withUrl_withParams_noUpdate(t *testing.T) {
	var stack cloudformation.Stack
	rName := fmt.Sprintf("tf-acc-test-with-params-no-update-%s", acctest.RandString(10))
	resourceName := "aws_cloudformation_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack-1.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"on_failure", "parameters", "template_url"},
			},
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack-2.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists(resourceName, &stack),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_withTransform(t *testing.T) {
	var stack cloudformation.Stack
	rName := fmt.Sprintf("tf-acc-test-with-transform-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_withTransform(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with-transform", &stack),
				),
			},
			{
				PlanOnly: true,
				Config:   testAccAWSCloudFormationStackConfig_withTransform(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with-transform", &stack),
				),
			},
		},
	})
}

func testAccCheckCloudFormationStackExists(n string, stack *cloudformation.Stack) resource.TestCheckFunc {
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
		if len(resp.Stacks) == 0 || resp.Stacks[0] == nil {
			return fmt.Errorf("CloudFormation stack not found")
		}

		*stack = *resp.Stacks[0]

		return nil
	}
}

func testAccCheckAWSCloudFormationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cfconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudformation_stack" {
			continue
		}

		params := cloudformation.DescribeStacksInput{
			StackName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeStacks(&params)

		if err != nil {
			return err
		}

		for _, s := range resp.Stacks {
			if aws.StringValue(s.StackId) == rs.Primary.ID && aws.StringValue(s.StackStatus) != cloudformation.StackStatusDeleteComplete {
				return fmt.Errorf("CloudFormation stack still exists: %q", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckCloudFormationStackNotRecreated(i, j *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.StackId) != aws.StringValue(j.StackId) {
			return fmt.Errorf("CloudFormation stack recreated")
		}

		return nil
	}
}

func testAccCheckCloudFormationStackDisappears(stack *cloudformation.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cfconn

		input := &cloudformation.DeleteStackInput{
			StackName: stack.StackName,
		}

		_, err := conn.DeleteStack(input)

		if err != nil {
			return err
		}

		// Use the AWS Go SDK waiter until the resource is refactored
		describeStacksInput := &cloudformation.DescribeStacksInput{
			StackName: stack.StackName,
		}

		return conn.WaitUntilStackDeleteComplete(describeStacksInput)
	}
}

func testAccAWSCloudFormationStackConfig(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = "%[1]s"

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "10.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
}
`, stackName)
}

func testAccAWSCloudFormationStackConfigCreationFailure(stackName, onFailure string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name       = %[1]q
  on_failure = %[2]q

  template_body = <<STACK
{
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : "1000.0.0.0/16",
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  },
  "Outputs" : {
    "DefaultSgId" : {
      "Description": "The ID of default security group",
      "Value" : { "Fn::GetAtt" : [ "MyVPC", "DefaultSecurityGroup" ]}
    },
    "VpcID" : {
      "Description": "The VPC ID",
      "Value" : { "Ref" : "MyVPC" }
    }
  }
}
STACK
}
`, stackName, onFailure)
}

func testAccAWSCloudFormationStackConfig_yaml(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = "%[1]s"

  template_body = <<STACK
Resources:
  MyVPC:
    Type: AWS::EC2::VPC
    Properties:
      CidrBlock: 10.0.0.0/16
      Tags:
        -
          Key: Name
          Value: Primary_CF_VPC

Outputs:
  DefaultSgId:
    Description: The ID of default security group
    Value: !GetAtt MyVPC.DefaultSecurityGroup
  VpcID:
    Description: The VPC ID
    Value: !Ref MyVPC
STACK
}
`, stackName)
}

func testAccAWSCloudFormationStackConfig_defaultParams(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = "%[1]s"

  template_body = <<BODY
{
    "Parameters": {
        "TopicName": {
            "Type": "String"
        },
        "VPCCIDR": {
            "Type": "String",
            "Default": "10.10.0.0/16"
        }
    },
    "Resources": {
        "NotificationTopic": {
            "Type": "AWS::SNS::Topic",
            "Properties": {
                "TopicName": {
                    "Ref": "TopicName"
                }
            }
        },
        "MyVPC": {
            "Type": "AWS::EC2::VPC",
            "Properties": {
                "CidrBlock": {
                    "Ref": "VPCCIDR"
                },
                "Tags": [
                    {
                        "Key": "Name",
                        "Value": "Primary_CF_VPC"
                    }
                ]
            }
        }
    },
    "Outputs": {
        "VPCCIDR": {
            "Value": {
                "Ref": "VPCCIDR"
            }
        }
    }
}
BODY


  parameters = {
    TopicName = "%[1]s"
  }
}
`, stackName)
}

var testAccAWSCloudFormationStackConfig_allAttributesWithBodies_tpl = `
data "aws_partition" "current" {}

resource "aws_cloudformation_stack" "test" {
  name          = "%[1]s"
  template_body = <<STACK
{
  "Parameters" : {
    "VpcCIDR" : {
      "Description" : "CIDR to be used for the VPC",
      "Type" : "String"
    }
  },
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": "%[1]s"}
        ]
      }
    },
    "StaticVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": "%[1]s-2"}
        ]
      }
    },
    "InstanceRole" : {
      "Type" : "AWS::IAM::Role",
      "Properties" : {
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [ {
            "Effect": "Allow",
            "Principal": { "Service": "ec2.${data.aws_partition.current.dns_suffix}" },
            "Action": "sts:AssumeRole"
          } ]
        },
        "Path" : "/",
        "Policies" : [ {
          "PolicyName": "terraformtest",
          "PolicyDocument": {
            "Version": "2012-10-17",
            "Statement": [ {
              "Effect": "Allow",
              "Action": [ "ec2:DescribeSnapshots" ],
              "Resource": [ "*" ]
            } ]
          }
        } ]
      }
    }
  }
}
STACK
  parameters = {
    VpcCIDR = "10.0.0.0/16"
  }

  policy_body        = <<POLICY
%[2]s
POLICY
  capabilities       = ["CAPABILITY_IAM"]
  notification_arns  = ["${aws_sns_topic.cf-updates.arn}"]
  on_failure         = "DELETE"
  timeout_in_minutes = 10
  tags = {
    First  = "Mickey"
    Second = "Mouse"
  }
}

resource "aws_sns_topic" "cf-updates" {
  name = "tf-cf-notifications"
}
`

var policyBody = `
{
  "Statement" : [
    {
      "Effect" : "Deny",
      "Action" : "Update:*",
      "Principal": "*",
      "Resource" : "LogicalResourceId/StaticVPC"
    },
    {
      "Effect" : "Allow",
      "Action" : "Update:*",
      "Principal": "*",
      "Resource" : "*"
    }
  ]
}
`

func testAccAWSCloudFormationStackConfig_allAttributesWithBodies(stackName string) string {
	return fmt.Sprintf(
		testAccAWSCloudFormationStackConfig_allAttributesWithBodies_tpl,
		stackName,
		policyBody)
}

func testAccAWSCloudFormationStackConfig_allAttributesWithBodies_modified(stackName string) string {
	return fmt.Sprintf(
		testAccAWSCloudFormationStackConfig_allAttributesWithBodies_tpl,
		stackName,
		policyBody)
}

func testAccAWSCloudFormationStackConfig_withParams(stackName, cidr string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "test" {
  name = "%[1]s"
  parameters = {
    VpcCIDR = "%[2]s"
  }
  template_body = <<STACK
{
  "Parameters" : {
    "VpcCIDR" : {
      "Description" : "CIDR to be used for the VPC",
      "Type" : "String"
    }
  },
  "Resources" : {
    "MyVPC": {
      "Type" : "AWS::EC2::VPC",
      "Properties" : {
        "CidrBlock" : {"Ref": "VpcCIDR"},
        "Tags" : [
          {"Key": "Name", "Value": "Primary_CF_VPC"}
        ]
      }
    }
  }
}
STACK

  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, stackName, cidr)
}

func testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, bucketKey, vpcCidr string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_s3_bucket" "b" {
  bucket = "%[1]s"
  acl    = "public-read"

  policy = <<POLICY
{
  "Version":"2008-10-17",
  "Statement": [
    {
      "Sid":"AllowPublicRead",
      "Effect":"Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
    }
  ]
}
POLICY


  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.b.id
  key    = "%[2]s"
  source = "test-fixtures/cloudformation-template.json"
}

resource "aws_cloudformation_stack" "test" {
  name = "%[1]s"

  parameters = {
    VpcCIDR = "%[3]s"
  }

  template_url       = "https://${aws_s3_bucket.b.id}.s3-${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket_object.object.key}"
  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr)
}

func testAccAWSCloudFormationStackConfig_templateUrl_withParams_withYaml(rName, bucketKey, vpcCidr string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_s3_bucket" "b" {
  bucket = "%[1]s"
  acl    = "public-read"

  policy = <<POLICY
{
  "Version":"2008-10-17",
  "Statement": [
    {
      "Sid":"AllowPublicRead",
      "Effect":"Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "s3:GetObject",
      "Resource": "arn:${data.aws_partition.current.partition}:s3:::%[1]s/*"
    }
  ]
}
POLICY


  website {
    index_document = "index.html"
    error_document = "error.html"
  }
}

resource "aws_s3_bucket_object" "object" {
  bucket = aws_s3_bucket.b.id
  key    = "%[2]s"
  source = "test-fixtures/cloudformation-template.yaml"
}

resource "aws_cloudformation_stack" "test" {
  name = "%[1]s"

  parameters = {
    VpcCIDR = "%[3]s"
  }

  template_url       = "https://${aws_s3_bucket.b.id}.s3-${data.aws_region.current.name}.${data.aws_partition.current.dns_suffix}/${aws_s3_bucket_object.object.key}"
  on_failure         = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr)
}

func testAccAWSCloudFormationStackConfig_withTransform(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "with-transform" {
  name = "%[1]s"

  template_body = <<STACK
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Transform": "AWS::Serverless-2016-10-31",
  "Resources": {
    "Api": {
      "Type": "AWS::Serverless::Api",
      "Properties": {
        "StageName": "Prod",
        "EndpointConfiguration": "REGIONAL",
        "DefinitionBody": {
          "swagger": "2.0",
          "paths": {
            "/": {
              "get": {
                "consumes": ["application/json"],
                "produces": ["application/json"],
                "responses": {
                  "200": {
                    "description": "200 response"
                  }
                },
                "x-amazon-apigateway-integration": {
                  "responses": {
                    "default": {
                      "statusCode": "200"
                    }
                  },
                  "requestTemplates": {
                    "application/json": "{\"statusCode\": 200}"
                  },
                  "passthroughBehavior": "when_no_match",
                  "type": "mock"
                }
              }
            }
          }
        }
      }
    }
  }
}
STACK

  capabilities       = ["CAPABILITY_AUTO_EXPAND"]
  on_failure         = "DELETE"
  timeout_in_minutes = 10
}
`, rName)
}
