package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directoryservice"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDirectoryServiceLogSubscription_basic(t *testing.T) {
	resourceName := "aws_directory_service_log_subscription.subscription"
	logGroupName := "ad-service-log-subscription-test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSDirectoryService(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDirectoryServiceLogSubscriptionDestroy,
		Steps: []resource.TestStep{
			// test create
			{
				Config: testAccDirectoryServiceLogSubscriptionConfig(logGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDirectoryServiceLogSubscriptionExists(
						resourceName,
						logGroupName,
					),
				),
			},
			// test import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsDirectoryServiceLogSubscriptionDestroy(s *terraform.State) error {
	dsconn := testAccProvider.Meta().(*AWSClient).dsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_directory_service_log_subscription" {
			continue
		}

		res, err := dsconn.ListLogSubscriptions(&directoryservice.ListLogSubscriptionsInput{
			DirectoryId: aws.String(rs.Primary.ID),
		})

		if isAWSErr(err, directoryservice.ErrCodeEntityDoesNotExistException, "") {
			continue
		}

		if err != nil {
			return err
		}

		if len(res.LogSubscriptions) > 0 {
			return fmt.Errorf("Expected AWS Directory Service Log Subscription to be gone, but was still found")
		}
	}

	return nil
}

func testAccCheckAwsDirectoryServiceLogSubscriptionExists(name string, logGroupName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		dsconn := testAccProvider.Meta().(*AWSClient).dsconn

		res, err := dsconn.ListLogSubscriptions(&directoryservice.ListLogSubscriptionsInput{
			DirectoryId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return err
		}

		if len(res.LogSubscriptions) == 0 {
			return fmt.Errorf("No Log subscription found")
		}

		if *(res.LogSubscriptions[0].LogGroupName) != logGroupName {
			return fmt.Errorf("Expected Log subscription not found")
		}

		return nil
	}
}

func testAccDirectoryServiceLogSubscriptionConfig(logGroupName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_directory_service_directory" "bar" {
  name     = "corp.notexample.com"
  password = "SuperSecretPassw0rd"
  type     = "MicrosoftAD"
  edition  = "Standard"

  vpc_settings {
    vpc_id     = aws_vpc.main.id
    subnet_ids = [aws_subnet.foo.id, aws_subnet.bar.id]
  }

  tags = {
    Name = "terraform-testacc-directory-service-log-subscription"
  }
}

resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-directory-service-log-subscription"
  }
}

resource "aws_subnet" "foo" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.1.0/24"

  tags = {
    Name = "terraform-testacc-directory-service-log-subscription"
  }
}

resource "aws_subnet" "bar" {
  vpc_id            = aws_vpc.main.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "10.0.2.0/24"

  tags = {
    Name = "terraform-testacc-directory-service-log-subscription"
  }
}

resource "aws_cloudwatch_log_group" "logs" {
  name              = "%s"
  retention_in_days = 1
}

data "aws_iam_policy_document" "ad-log-policy" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    principals {
      identifiers = ["ds.amazonaws.com"]
      type        = "Service"
    }

    resources = ["${aws_cloudwatch_log_group.logs.arn}:*"]

    effect = "Allow"
  }
}

resource "aws_cloudwatch_log_resource_policy" "ad-log-policy" {
  policy_document = data.aws_iam_policy_document.ad-log-policy.json
  policy_name     = "ad-log-policy"
}

resource "aws_directory_service_log_subscription" "subscription" {
  directory_id   = aws_directory_service_directory.bar.id
  log_group_name = aws_cloudwatch_log_group.logs.name
}
`, logGroupName)
}
