package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudformation"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudFormationStack_importBasic(t *testing.T) {
	stackName := fmt.Sprintf("tf-acc-test-basic-%s", acctest.RandString(10))

	resourceName := "aws_cloudformation_stack.network"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig(stackName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCloudFormationStack_basic(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-basic-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.network", &stack),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_yaml(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-yaml-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_yaml(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.yaml", &stack),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_defaultParams(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-default-params-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_defaultParams(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.asg-demo", &stack),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_allAttributes(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-all-attributes-%s", acctest.RandString(10))

	expectedPolicyBody := "{\"Statement\":[{\"Action\":\"Update:*\",\"Effect\":\"Deny\",\"Principal\":\"*\",\"Resource\":\"LogicalResourceId/StaticVPC\"},{\"Action\":\"Update:*\",\"Effect\":\"Allow\",\"Principal\":\"*\",\"Resource\":\"*\"}]}"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_allAttributesWithBodies(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.full", &stack),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "name", stackName),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "capabilities.1328347040", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "disable_rollback", "false"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "notification_arns.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "parameters.%", "1"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "tags.First", "Mickey"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "timeout_in_minutes", "10"),
				),
			},
			{
				Config: testAccAWSCloudFormationStackConfig_allAttributesWithBodies_modified(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.full", &stack),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "name", stackName),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "capabilities.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "capabilities.1328347040", "CAPABILITY_IAM"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "disable_rollback", "false"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "notification_arns.#", "1"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "parameters.%", "1"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "parameters.VpcCIDR", "10.0.0.0/16"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "policy_body", expectedPolicyBody),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "tags.First", "Mickey"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "tags.Second", "Mouse"),
					resource.TestCheckResourceAttr("aws_cloudformation_stack.full", "timeout_in_minutes", "10"),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4332
func TestAccAWSCloudFormationStack_withParams(t *testing.T) {
	var stack cloudformation.Stack
	stackName := fmt.Sprintf("tf-acc-test-with-params-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_withParams(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with_params", &stack),
				),
			},
			{
				Config: testAccAWSCloudFormationStackConfig_withParams_modified(stackName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with_params", &stack),
				),
			},
		},
	})
}

// Regression for https://github.com/hashicorp/terraform/issues/4534
func TestAccAWSCloudFormationStack_withUrl_withParams(t *testing.T) {
	var stack cloudformation.Stack
	rName := fmt.Sprintf("tf-acc-test-with-url-and-params-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with-url-and-params", &stack),
				),
			},
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack.json", "13.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with-url-and-params", &stack),
				),
			},
		},
	})
}

func TestAccAWSCloudFormationStack_withUrl_withParams_withYaml(t *testing.T) {
	var stack cloudformation.Stack
	rName := fmt.Sprintf("tf-acc-test-with-params-and-yaml-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams_withYaml(rName, "tf-cf-stack.yaml", "13.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with-url-and-params-and-yaml", &stack),
				),
			},
		},
	})
}

// Test for https://github.com/hashicorp/terraform/issues/5653
func TestAccAWSCloudFormationStack_withUrl_withParams_noUpdate(t *testing.T) {
	var stack cloudformation.Stack
	rName := fmt.Sprintf("tf-acc-test-with-params-no-update-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudFormationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack-1.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with-url-and-params", &stack),
				),
			},
			{
				Config: testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, "tf-cf-stack-2.json", "11.0.0.0/16"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFormationStackExists("aws_cloudformation_stack.with-url-and-params", &stack),
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
		if len(resp.Stacks) == 0 {
			return fmt.Errorf("CloudFormation stack not found")
		}

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
			if *s.StackId == rs.Primary.ID && *s.StackStatus != "DELETE_COMPLETE" {
				return fmt.Errorf("CloudFormation stack still exists: %q", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAWSCloudFormationStackConfig(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "network" {
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
}`, stackName)
}

func testAccAWSCloudFormationStackConfig_yaml(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "yaml" {
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
}`, stackName)
}

func testAccAWSCloudFormationStackConfig_defaultParams(stackName string) string {
	return fmt.Sprintf(`
resource "aws_cloudformation_stack" "asg-demo" {
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
resource "aws_cloudformation_stack" "full" {
  name = "%[1]s"
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
            "Principal": { "Service": "ec2.amazonaws.com" },
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

  policy_body = <<POLICY
%[2]s
POLICY
  capabilities = ["CAPABILITY_IAM"]
  notification_arns = ["${aws_sns_topic.cf-updates.arn}"]
  on_failure = "DELETE"
  timeout_in_minutes = 10
  tags = {
    First = "Mickey"
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

var tpl_testAccAWSCloudFormationStackConfig_withParams = `
resource "aws_cloudformation_stack" "with_params" {
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

  on_failure = "DELETE"
  timeout_in_minutes = 1
}
`

func testAccAWSCloudFormationStackConfig_withParams(stackName string) string {
	return fmt.Sprintf(
		tpl_testAccAWSCloudFormationStackConfig_withParams,
		stackName,
		"10.0.0.0/16")
}

func testAccAWSCloudFormationStackConfig_withParams_modified(stackName string) string {
	return fmt.Sprintf(
		tpl_testAccAWSCloudFormationStackConfig_withParams,
		stackName,
		"12.0.0.0/16")
}

func testAccAWSCloudFormationStackConfig_templateUrl_withParams(rName, bucketKey, vpcCidr string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "b" {
  bucket = "%[1]s"
  acl = "public-read"
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
      "Resource": "arn:aws:s3:::%[1]s/*"
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
  bucket = "${aws_s3_bucket.b.id}"
  key = "%[2]s"
  source = "test-fixtures/cloudformation-template.json"
}

resource "aws_cloudformation_stack" "with-url-and-params" {
  name = "%[1]s"
  parameters = {
    VpcCIDR = "%[3]s"
  }
  template_url = "https://${aws_s3_bucket.b.id}.s3-us-west-2.amazonaws.com/${aws_s3_bucket_object.object.key}"
  on_failure = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr)
}

func testAccAWSCloudFormationStackConfig_templateUrl_withParams_withYaml(rName, bucketKey, vpcCidr string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "b" {
  bucket = "%[1]s"
  acl = "public-read"
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
      "Resource": "arn:aws:s3:::%[1]s/*"
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
  bucket = "${aws_s3_bucket.b.id}"
  key = "%[2]s"
  source = "test-fixtures/cloudformation-template.yaml"
}

resource "aws_cloudformation_stack" "with-url-and-params-and-yaml" {
  name = "%[1]s"
  parameters = {
    VpcCIDR = "%[3]s"
  }
  template_url = "https://${aws_s3_bucket.b.id}.s3-us-west-2.amazonaws.com/${aws_s3_bucket_object.object.key}"
  on_failure = "DELETE"
  timeout_in_minutes = 1
}
`, rName, bucketKey, vpcCidr)
}
