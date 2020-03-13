package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSInspectorTemplate_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInspectorTemplateAssessment(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists("aws_inspector_assessment_template.foo"),
				),
			},
			{
				Config: testAccCheckAWSInspectorTemplatetModified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists("aws_inspector_assessment_template.foo"),
				),
			},
		},
	})
}

func testAccCheckAWSInspectorTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).inspectorconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_inspector_assessment_template" {
			continue
		}

		resp, err := conn.DescribeAssessmentTemplates(&inspector.DescribeAssessmentTemplatesInput{
			AssessmentTemplateArns: []*string{
				aws.String(rs.Primary.ID),
			},
		})

		if err != nil {
			if inspectorerr, ok := err.(awserr.Error); ok && inspectorerr.Code() == "InvalidInputException" {
				return nil
			} else {
				return fmt.Errorf("Error finding Inspector Assessment Template: %s", err)
			}
		}

		if len(resp.AssessmentTemplates) > 0 {
			return fmt.Errorf("Found Template, expected none: %s", resp)
		}
	}

	return nil
}

func testAccCheckAWSInspectorTemplateExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("not found: %s", name)
		}

		return nil
	}
}

func testAccCheckIsSubscribedToEvent(templateName string, eventName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		is, err := primaryInstanceState(s, templateName)
		if err != nil {
			return err
		}

		for k, v := range is.Attributes {
			if strings.HasPrefix(k, "subscribe_to_event.") && strings.HasSuffix(k, ".event") {
				if v == eventName {
					return nil
				}
			}
		}
		return fmt.Errorf("could not find an event named: %s", eventName)
	}
}

func testAccAWSInspectorTemplateAssessment(rInt int) string {
	return fmt.Sprintf(`
data "aws_inspector_rules_packages" "rules" {}

resource "aws_inspector_resource_group" "foo" {
  tags = {
    Name = "tf-acc-test-%d"
  }
}

resource "aws_inspector_assessment_target" "foo" {
  name               = "tf-acc-test-basic-%d"
  resource_group_arn = "${aws_inspector_resource_group.foo.arn}"
}

resource "aws_inspector_assessment_template" "foo" {
  name       = "tf-acc-test-basic-tpl-%d"
  target_arn = "${aws_inspector_assessment_target.foo.arn}"
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.rules.arns
}
`, rInt, rInt, rInt)
}

func testAccCheckAWSInspectorTemplatetModified(rInt int) string {
	return fmt.Sprintf(`
data "aws_inspector_rules_packages" "rules" {}

resource "aws_inspector_resource_group" "foo" {
  tags = {
    Name = "tf-acc-test-%d"
  }
}

resource "aws_inspector_assessment_target" "foo" {
  name               = "tf-acc-test-basic-%d"
  resource_group_arn = "${aws_inspector_resource_group.foo.arn}"
}

resource "aws_inspector_assessment_template" "foo" {
  name       = "tf-acc-test-basic-tpl-%d"
  target_arn = "${aws_inspector_assessment_target.foo.arn}"
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.rules.arns
}
`, rInt, rInt, rInt)
}
