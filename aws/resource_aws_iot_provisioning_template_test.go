package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSIoTProvisioningTemplate_basic(t *testing.T) {
	resourceName := "aws_iot_provisioning_template.fleet"
	rName := acctest.RandomWithPrefix("Fleet-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTProvisioningTemplateDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTProvisioningTemplateConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version_id", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "My provisioning template"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "provisioning_role_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "template_arn"),
					resource.TestCheckResourceAttr(resourceName, "template_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					testAccAWSIoTProvisioningTemplateCheckVersionExists(rName, 1),
				),
			},
			{
				Config:            testAccAWSIoTProvisioningTemplateConfigInitialState(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSIoTProvisioningTemplateConfigTemplateBodyUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "default_version_id", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "template_body"),
					testAccAWSIoTProvisioningTemplateCheckVersionExists(rName, 2),
				),
			},
		},
	})
}

func testAccCheckAWSIoTProvisioningTemplateDestroy_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_provisioning_template" {
			continue
		}

		_, err := conn.DescribeProvisioningTemplate(&iot.DescribeProvisioningTemplateInput{
			TemplateName: aws.String(rs.Primary.Attributes["template_name"]),
		})

		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			return nil
		} else if err != nil {
			return err
		}

		return fmt.Errorf("IoT Provisioning Template still exists")
	}

	return nil
}

func testAccAWSIoTProvisioningTemplateCheckVersionExists(templateName string, numVersions int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).iotconn

		resp, err := conn.ListProvisioningTemplateVersions(&iot.ListProvisioningTemplateVersionsInput{
			TemplateName: aws.String(templateName),
		})

		if err != nil {
			return err
		}

		if len(resp.Versions) != numVersions {
			return fmt.Errorf("Expected %d versions for template %s but found %d", numVersions, templateName, len(resp.Versions))
		}

		return nil
	}
}

func testAccAWSIoTProvisioningTemplateConfigInitialState(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "iot_assume_role_policy" {
  statement {
    actions = ["sts:AssumeRole"]

    principals {
      type        = "Service"
      identifiers = ["iot.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "iot_fleet_provisioning" {
  name               = "IoTProvisioningServiceRole"
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.iot_assume_role_policy.json
}

resource "aws_iam_role_policy_attachment" "iot_fleet_provisioning_registration" {
  role       = aws_iam_role.iot_fleet_provisioning.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSIoTThingsRegistration"
}

data "aws_iam_policy_document" "device_policy" {
  statement {
    actions   = ["iot:Subscribe"]
    resources = ["*"]
  }
}

resource "aws_iot_policy" "device_policy" {
  name   = "DevicePolicy"
  policy = data.aws_iam_policy_document.device_policy.json
}

resource "aws_iot_provisioning_template" "fleet" {
  template_name         = "%s"
  description           = "My provisioning template"
  provisioning_role_arn = aws_iam_role.iot_fleet_provisioning.arn

  template_body = jsonencode({
    Parameters = {
      SerialNumber = { Type = "String" }
    }

    Resources = {
      certificate = {
        Properties = {
          CertificateId = { Ref = "AWS::IoT::Certificate::Id" }
          Status        = "Active"
        }

        Type = "AWS::IoT::Certificate"
      }

      policy = {
        Properties = {
          PolicyName = aws_iot_policy.device_policy.name
        }

        Type = "AWS::IoT::Policy"
      }
    }
  })
}
`, rName)
}

func testAccAWSIoTProvisioningTemplateConfigTemplateBodyUpdate(rName string) string {
	return strings.ReplaceAll(testAccAWSIoTProvisioningTemplateConfigInitialState(rName), "Active", "Inactive")
}
