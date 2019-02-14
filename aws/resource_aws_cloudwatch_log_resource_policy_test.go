package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudWatchLogResourcePolicy_Basic(t *testing.T) {
	name := acctest.RandString(5)
	var resourcePolicy cloudwatchlogs.ResourcePolicy
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudWatchLogResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogResourcePolicy("aws_cloudwatch_log_resource_policy.test", &resourcePolicy),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_resource_policy.test", "policy_name", name),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_resource_policy.test", "policy_document", "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"route53.amazonaws.com\"},\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Resource\":\"arn:aws:logs:*:*:log-group:/aws/route53/*\"}]}"),
				),
			},
			{
				Config: testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogResourcePolicy("aws_cloudwatch_log_resource_policy.test", &resourcePolicy),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_resource_policy.test", "policy_name", name),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_resource_policy.test", "policy_document", "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"route53.amazonaws.com\"},\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Resource\":\"arn:aws:logs:*:*:log-group:/aws/route53/example.com\"}]}"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchLogResourcePolicy_Import(t *testing.T) {
	resourceName := "aws_cloudwatch_log_resource_policy.test"

	name := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudWatchLogResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic1(name),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckCloudWatchLogResourcePolicy(pr string, resourcePolicy *cloudwatchlogs.ResourcePolicy) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		rs, ok := s.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		policy, exists, err := lookupCloudWatchLogResourcePolicy(conn, rs.Primary.ID, nil)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Resource policy does not exist: %q", rs.Primary.ID)
		}

		*resourcePolicy = *policy

		return nil
	}
}

func testAccCheckCloudWatchLogResourcePolicyDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_resource_policy" {
			continue
		}

		_, exists, err := lookupCloudWatchLogResourcePolicy(conn, rs.Primary.ID, nil)
		if err != nil {
			return nil
		}

		if exists {
			return fmt.Errorf("Resource policy exists: %q", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic1(name string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:log-group:/aws/route53/*"]

    principals {
      identifiers = ["route53.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = "%s"
  policy_document = "${data.aws_iam_policy_document.test.json}"
}
`, name)
}

func testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic2(name string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:aws:logs:*:*:log-group:/aws/route53/example.com"]

    principals {
      identifiers = ["route53.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name = "%s"
  policy_document = "${data.aws_iam_policy_document.test.json}"
}
`, name)
}
