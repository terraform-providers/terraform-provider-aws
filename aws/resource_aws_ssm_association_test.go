package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSSSMAssociation_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "output_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "InstanceIds"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "targets.0.values.0", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "target_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "document_version", "$DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sync_compliance", "AUTO"),
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

func TestAccAWSSSMAssociation_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSsmAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSSMAssociation_ApplyOnlyAtCronInterval(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithApplyOnlyAtCronInterval(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithApplyOnlyAtCronInterval(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "apply_only_at_cron_interval", "false"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withTargets(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
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
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(rName, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(rName, twoTargets),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "targets.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptanceTest"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.key", "tag:ExtraName"),
					resource.TestCheckResourceAttr(resourceName, "targets.1.values.0", "acceptanceTest"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithTargets(rName, oneTarget),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "targets.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.key", "tag:Name"),
					resource.TestCheckResourceAttr(resourceName, "targets.0.values.0", "acceptanceTest"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withParameters(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParameters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAssociationName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", rNameUpdated),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAssociationNameAndScheduleExpression(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"
	scheduleExpression1 := "cron(0 16 ? * TUE *)"
	scheduleExpression2 := "cron(0 16 ? * WED *)"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, scheduleExpression1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, scheduleExpression2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", rName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", scheduleExpression2),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withDocumentVersion(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithDocumentVersion(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "document_version", "1"),
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
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rNameUpdated := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_bucket_name", rName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateBucketName(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_bucket_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_key_prefix", "SSMAssociation"),
				),
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateKeyPrefix(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_bucket_name", rNameUpdated),
					resource.TestCheckResourceAttr(resourceName, "output_location.0.s3_key_prefix", "UpdatedAssociation"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withAutomationTargetParamName(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithAutomationTargetParamName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpace"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"parameters"},
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithParametersUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "parameters.Directory", "myWorkSpaceUpdated"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withScheduleExpression(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithScheduleExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 16 ? * TUE *)"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithScheduleExpressionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "schedule_expression", "cron(0 16 ? * WED *)"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_withComplianceSeverity(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	compSeverity1 := "HIGH"
	compSeverity2 := "LOW"
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationBasicConfigWithComplianceSeverity(compSeverity1, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compliance_severity", compSeverity1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationBasicConfigWithComplianceSeverity(compSeverity2, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "association_name", rName),
					resource.TestCheckResourceAttr(resourceName, "compliance_severity", compSeverity2),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_rateControl(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ssm.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationRateControlConfig(rName, "10%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "10%"),
					resource.TestCheckResourceAttr(resourceName, "max_errors", "10%"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationRateControlConfig(rName, "20%"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "max_concurrency", "20%"),
					resource.TestCheckResourceAttr(resourceName, "max_errors", "20%"),
				),
			},
		},
	})
}

func TestAccAWSSSMAssociation_syncCompliance(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ssm_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSSMAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSSMAssociationSyncComplianceConfig(rName, "MANUAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sync_compliance", "MANUAL"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSSMAssociationSyncComplianceConfig(rName, "AUTO"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sync_compliance", "AUTO"),
				),
			},
			{
				Config: testAccAWSSSMAssociationSyncComplianceConfig(rName, "MANUAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSSMAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "sync_compliance", "MANUAL"),
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
				continue
			}
			return err
		}

		if out != nil {
			return fmt.Errorf("Expected AWS SSM Association to be gone, but was still found")
		}
	}

	return nil
}

func testAccAWSSSMAssociationBasicConfigWithApplyOnlyAtCronInterval(rName string, applyOnlyAtCronInterval bool) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  name                        = aws_ssm_document.test.name
  schedule_expression         = "cron(0 16 ? * TUE *)"
  apply_only_at_cron_interval = %[2]t

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, applyOnlyAtCronInterval)
}

func testAccAWSSSMAssociationBasicConfigWithAutomationTargetParamName(rName string) string {
	return composeConfig(testAccLatestAmazonLinuxHvmEbsAmiConfig(), fmt.Sprintf(`
resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.test.name
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "ec2.${data.aws_partition.current.dns_suffix}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF

}

resource "aws_ssm_document" "test" {
  name          = %[1]q
  document_type = "Automation"

  content = <<DOC
{
  "description": "Systems Manager Automation Demo",
  "schemaVersion": "0.3",
  "assumeRole": "${aws_iam_role.test.arn}",
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
        "ImageId": "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}",
        "InstanceType": "t2.small",
        "MinInstanceCount": 1,
        "MaxInstanceCount": 1,
        "IamInstanceProfileName": "${aws_iam_instance_profile.test.name}"
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
  name                             = aws_ssm_document.test.name
  automation_target_parameter_name = "Directory"

  parameters = {
    AutomationAssumeRole = aws_iam_role.test.id
    Directory            = "myWorkSpace"
  }

  targets {
    key    = "tag:myTagName"
    values = ["myTagValue"]
  }

  schedule_expression = "rate(60 minutes)"
}
`, rName))
}

func testAccAWSSSMAssociationBasicConfigWithParametersUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  name          = %[1]q
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
  name          = %[1]q
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
    Name = %[1]q
  }
}

resource "aws_subnet" "first" {
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]
}

resource "aws_security_group" "test" {
  name        = %[1]q
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
    Name = %[1]q
  }
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  name        = %[1]q
  instance_id = aws_instance.test.id
}
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithDocumentVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  name             = %[1]q
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
  name          = %[1]q
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
  name          = %[1]q
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
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
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
`, rName)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateBucketName(rName, rNameUpdated string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "output_location" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_s3_bucket" "output_location_updated" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_ssm_document" "test" {
  name          = %[1]q
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
`, rName, rNameUpdated)
}

func testAccAWSSSMAssociationBasicConfigWithOutPutLocationUpdateKeyPrefix(rName, rNameUpdated string) string {
	return fmt.Sprintf(`
  resource "aws_s3_bucket" "output_location" {
    bucket        = %[1]q
    force_destroy = true
  }
  
  resource "aws_s3_bucket" "output_location_updated" {
    bucket        = %[2]q
    force_destroy = true
  }

resource "aws_ssm_document" "test" {
  name          = %[1]q
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
`, rName, rNameUpdated)
}

func testAccAWSSSMAssociationBasicConfigWithAssociationName(rName, assocName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  association_name = %[2]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, assocName)
}

func testAccAWSSSMAssociationConfigWithAssociationNameAndScheduleExpression(rName, scheduleExpression string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  association_name    = %[1]q
  name                = aws_ssm_document.test.name
  schedule_expression = %[2]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, scheduleExpression)
}

func testAccAWSSSMAssociationBasicConfigWithComplianceSeverity(compSeverity, rName string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  association_name    = %[1]q
  compliance_severity = %[2]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, compSeverity)
}

func testAccAWSSSMAssociationRateControlConfig(rName, rate string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  max_concurrency = %[2]q
  max_errors      = %[2]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, rate)
}

func testAccAWSSSMAssociationSyncComplianceConfig(rName, mode string) string {
	return fmt.Sprintf(`
resource "aws_ssm_document" "test" {
  name          = %[1]q
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
  sync_compliance     = %[2]q

  targets {
    key    = "tag:Name"
    values = ["acceptanceTest"]
  }
}
`, rName, mode)
}
