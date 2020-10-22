package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMAssociation_basic(t *testing.T) {
	name := fmt.Sprintf("tf-acc-ssm-association-%s", acctest.RandString(10))
	resourceName := "aws_ssm_association.test"

	deleteSsmAssociaton := func() {
		ec2conn := testAccProvider.Meta().(*AWSClient).ec2conn
		ssmconn := testAccProvider.Meta().(*AWSClient).ssmconn

		ins, err := resourceAwsInstanceFind(ec2conn, &ec2.DescribeInstancesInput{
			Filters: []*ec2.Filter{
				{
					Name:   aws.String("tag:Name"),
					Values: []*string{aws.String(name)},
				},
			},
		})
		if err != nil {
			t.Fatalf("Error getting instance with tag:Name %s: %s", name, err)
		}
		if len(ins) == 0 {
			t.Fatalf("No instance exists with tag:Name %s", name)
		}
		instanceID := ins[0].InstanceId

		_, err = ssmconn.DeleteAssociation(&ssm.DeleteAssociationInput{
			Name:       aws.String(name),
			InstanceId: instanceID,
		})
		if err != nil {
			t.Fatalf("Error deleting ssm association %s/%s: %s", name, aws.StringValue(instanceID), err)
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: deleteSsmAssociaton,
				Config:    testAccAWSSSMAssociationBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withTargets(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_association.test"
	oneTarget := `

targets {
  key    = "tag:Name"
  values = ["acceptanceTest"]
}
`

	twoTargets := `

targets {
  key    = "tag:Name"
  values = ["acceptanceTest"]
}

targets {
  key    = "tag:ExtraName"
  values = ["acceptanceTest"]
}
`

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(name, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "targets.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(name, twoTargets),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "targets.#", "2"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.values.0", "acceptanceTest"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.1.key", "tag:ExtraName"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.1.values.0", "acceptanceTest"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(name, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "targets.#", "1"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(
						resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withParameters(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParameters(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParametersUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAssociationName(t *testing.T) {
	assocName1 := acctest.RandString(10)
	assocName2 := acctest.RandString(10)
	rName := acctest.RandString(5)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, assocName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, assocName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName2),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAssociationNameAndScheduleExpression(t *testing.T) {
	assocName := acctest.RandString(10)
	rName := acctest.RandString(5)
	resourceName := "aws_ssm_association.test"
	scheduleExpression1 := "cron(0 16 ? * TUE *)"
	scheduleExpression2 := "cron(0 16 ? * WED *)"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, assocName, scheduleExpression1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, assocName, scheduleExpression2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression2),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withDocumentVersion(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithDocumentVersion(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "document_version", "1"),
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

func TestAccAWSSSMAssociation_withOutputLocation(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocation(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-%s", name)),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateBucketName(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-updated-%s", name)),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateKeyPrefix(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_bucket_name", fmt.Sprintf("tf-acc-test-ssmoutput-updated-%s", name)),
					resource.TestCheckResourceAttr(
						resourceName, "output_location.0.s3_key_prefix", "UpdatedAssociation"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAutomationTargetParamName(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAutomationTargetParamName(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParametersUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withScheduleExpression(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithScheduleExpression(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "schedule_expression", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithScheduleExpressionUpdated(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "schedule_expression", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withComplianceSeverity(t *testing.T) {
	assocName := acctest.RandString(10)
	rName := acctest.RandString(10)
	compSeverity1 := "HIGH"
	compSeverity2 := "LOW"
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithComplianceSeverity(compSeverity1, rName, assocName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(
						resourceName, "compliance_severity", compSeverity1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithComplianceSeverity(compSeverity2, rName, assocName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "association_name", assocName),
					resource.TestCheckResourceAttr(
						resourceName, "compliance_severity", compSeverity2),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_rateControl(t *testing.T) {
	name := acctest.RandString(10)
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationRateControlConfig(name, "10%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "max_concurrency", "10%"),
					resource.TestCheckResourceAttr(
						resourceName, "max_errors", "10%"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationRateControlConfig(name, "20%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(
						resourceName, "max_concurrency", "20%"),
					resource.TestCheckResourceAttr(
						resourceName, "max_errors", "20%"),
				),
			},
		},
	})
}

func testAccCheckAWSSSMAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No SSM Assosciation ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ssmconn

		_, err := conn.DescribeAssociation(&ssm.DescribeAssociationInput{
			AssociationId: aws.String(rs.Primary.Attributes["association_id"]),
		})

		if err != nil {
			if isAWSErr(err, ssm.ErrCodeAssociationDoesNotExist, "") {
				return nil
			}
			return err
		}

		return nil
	}
}

func testAccCheckAWSSSMAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ssmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ssm_association" {
			continue
		}

		out, err := conn.DescribeAssociation(&ssm.DescribeAssociationInput{
			AssociationId: aws.String(rs.Primary.Attributes["association_id"]),
		})

		if err != nil {
			if isAWSErr(err, ssm.ErrCodeAssociationDoesNotExist, "") {
				return nil
			}
			return err
		}

		if out != nil {
			return fmt.Errorf("Expected AWS SSM Association to be gone, but was still found")
		}
	}

	return fmt.Errorf("Default error in SSM Association Test")
}

func testAccAWSSSMAssociationBasicConfigWithAutomationTargetParamName(rName string) string {
	return fmt.Sprintf(`
data "aws_ami" "ssm_ami" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  filter {
    name   = "name"
    values = ["*hvm-ssd/ubuntu-trusty-14.04*"]
  }
}

resource "aws_iam_instance_profile" "ssm_profile" {
  name = "ssm_profile-%[1]s"
  role = aws_iam_role.ssm_role.name
}

resource "aws_iam_role" "ssm_role" {
  name = "ssm_role-%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_ssm_document" "foo" {
  name          = "test_document-%[1]s"
  document_type = "Automation"

  content = <<DOC
{
  "description": "Systems Manager Automation Demo",
  "schemaVersion": "0.3",
  "assumeRole": "${aws_iam_role.ssm_role.arn}",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "mainSteps": [
    {
      "name": "startInstances",
      "action": "aws:runInstances",
      "timeoutSeconds": 1200,
      "maxAttempts": 1,
      "onFailure": "Abort",
      "inputs": {
        "ImageId": "${data.aws_ami.ssm_ami.id}",
        "InstanceType": "t2.small",
        "MinInstanceCount": 1,
        "MaxInstanceCount": 1,
        "IamInstanceProfileName": "${aws_iam_instance_profile.ssm_profile.name}"
      }
    },
    {
      "name": "stopInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "stopped"
      }
    },
    {
      "name": "terminateInstance",
      "action": "aws:changeInstanceState",
      "maxAttempts": 1,
      "onFailure": "Continue",
      "inputs": {
        "InstanceIds": [
          "{{ startInstances.InstanceIds }}"
        ],
        "DesiredState": "terminated"
      }
    }
  ]
}
DOC

}

resource "aws_ssm_association" "test" {
  name                             = aws_ssm_document.foo.name
  automation_target_parameter_name = "Directory"

  parameters = {
    AutomationAssumeRole = aws_iam_role.ssm_role.id
    Directory            = "myWorkSpace"
  }

  targets {
    key    = "tag:myTagName"
    values = ["myTagValue"]
  }

  schedule_expression = "rate(60 minutes)"
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithParametersUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<-DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
  DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  parameters = {
    Directory = "myWorkSpaceUpdated"
  }

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithParameters(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<-DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {
    "Directory": {
      "description": "(Optional) The path to the working directory on your instance.",
      "default": "",
      "type": "String",
      "maxChars": 4096
    }
  },
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
  DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  parameters = {
    Directory = "myWorkSpace"
  }

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithTargets(rName, targetsStr string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name
  %s
}
`, rName, targetsStr)
}

func testAccAWSSSMAssociationBasicConfig(rName string) string {
	return fmt.Sprintf(`
variable "name" {
  default = "%s"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

data "aws_ami" "amzn" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn2-ami-hvm-*-x86_64-gp2"]
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.name
  }
}

resource "aws_subnet" "first" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_security_group" "test" {
  name        = var.name
  description = "foo"
  vpc_id      = aws_vpc.main.id

  ingress {
    protocol    = "icmp"
    from_port   = -1
    to_port     = -1
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_instance" "test" {
  ami                    = data.aws_ami.amzn.image_id
  availability_zone      = data.aws_availability_zones.available.names[0]
  instance_type          = "t2.micro"
  vpc_security_group_ids = [aws_security_group.test.id]
  subnet_id              = aws_subnet.first.id

  tags = {
    Name = var.name
  }
}

resource "aws_ssm_document" "test" {
  name          = var.name
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name        = var.name
  instance_id = aws_instance.test.id
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithDocumentVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name             = "test_document_association-%s"
  document_version = aws_ssm_document.test.latest_version

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithScheduleExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  schedule_expression = "cron(0 16 ? * TUE *)"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithScheduleExpressionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  schedule_expression = "cron(0 16 ? * WED *)"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocation(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location.id
    s3_key_prefix  = "SSMAssociation"
  }
}
`, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateBucketName(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket        = "tf-acc-test-ssmoutput-updated-%s"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location_updated.id
    s3_key_prefix  = "SSMAssociation"
  }
}
`, rName, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateKeyPrefix(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = "tf-acc-test-ssmoutput-%s"
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket        = "tf-acc-test-ssmoutput-updated-%s"
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name = aws_ssm_document.test.name

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }

  output_location {
    s3_bucket_name = aws_s3_bucket.output_location_updated.id
    s3_key_prefix  = "UpdatedAssociation"
  }
}
`, rName, rName, rName)
}

func testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name             = aws_ssm_document.test.name
  association_name = "%s"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName)
}

func testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, associationName, scheduleExpression string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  association_name    = %q
  name                = aws_ssm_document.test.name
  schedule_expression = %q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, associationName, scheduleExpression)
}

func testAccAWSSSMAssociationBasicConfigWithComplianceSeverity(compSeverity, rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "test_document_association-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name                = aws_ssm_document.test.name
  association_name    = "%s"
  compliance_severity = "%s"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName, compSeverity)
}

func testAccAWSSSMAssociationRateControlConfig(rName, rate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = "tf-test-ssm-document-%s"
  document_type = "Command"

  content = <<DOC
{
  "schemaVersion": "1.2",
  "description": "Check ip configuration of a Linux instance.",
  "parameters": {},
  "runtimeConfig": {
    "aws:runShellScript": {
      "properties": [
        {
          "id": "0.aws:runShellScript",
          "runCommand": [
            "ifconfig"
          ]
        }
      ]
    }
  }
}
DOC

}

resource "aws_ssm_association" "test" {
  name            = aws_ssm_document.test.name
  max_concurrency = "%s"
  max_errors      = "%s"

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, rate, rate)
}
