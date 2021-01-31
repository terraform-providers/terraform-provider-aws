package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_log_resource_policy", &resource.Sweeper{
		Name: "aws_cloudwatch_log_resource_policy",
		F:    testSweepCloudWatchLogResourcePolicies,
	})
}

func testSweepCloudWatchLogResourcePolicies(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudwatchlogsconn

	input := &cloudwatchlogs.DescribeResourcePoliciesInput{}

	for {
		output, err := conn.DescribeResourcePolicies(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudWatchLog Resource Policy sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error describing CloudWatchLog Resource Policy: %s", err)
		}

		for _, resourcePolicy := range output.ResourcePolicies {
			policyName := aws.StringValue(resourcePolicy.PolicyName)
			deleteInput := &cloudwatchlogs.DeleteResourcePolicyInput{
				PolicyName: resourcePolicy.PolicyName,
			}

			log.Printf("[INFO] Deleting CloudWatch Log Resource Policy: %s", policyName)

			if _, err := conn.DeleteResourcePolicy(deleteInput); err != nil {
				return fmt.Errorf("error deleting CloudWatch log resource policy (%s): %s", policyName, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSCloudWatchLogResourcePolicy_basic(t *testing.T) {
	name := acctest.RandString(5)
	resourceName := "aws_cloudwatch_log_resource_policy.test"
	var resourcePolicy cloudwatchlogs.ResourcePolicy

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudWatchLogResourcePolicyDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic1(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", name),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/*\"}]}", testAccGetPartitionDNSSuffix(), testAccGetPartition())),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic2(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogResourcePolicy(resourceName, &resourcePolicy),
					resource.TestCheckResourceAttr(resourceName, "policy_name", name),
					resource.TestCheckResourceAttr(resourceName, "policy_document", fmt.Sprintf("{\"Version\":\"2012-10-17\",\"Statement\":[{\"Sid\":\"\",\"Effect\":\"Allow\",\"Principal\":{\"Service\":\"rds.%s\"},\"Action\":[\"logs:PutLogEvents\",\"logs:CreateLogStream\"],\"Resource\":\"arn:%s:logs:*:*:log-group:/aws/rds/example.com\"}]}", testAccGetPartitionDNSSuffix(), testAccGetPartition())),
				),
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
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/*"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = "%s"
  policy_document = data.aws_iam_policy_document.test.json
}
`, name)
}

func testAccCheckAWSCloudWatchLogResourcePolicyResourceConfigBasic2(name string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_iam_policy_document" "test" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents",
    ]

    resources = ["arn:${data.aws_partition.current.partition}:logs:*:*:log-group:/aws/rds/example.com"]

    principals {
      identifiers = ["rds.${data.aws_partition.current.dns_suffix}"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "test" {
  policy_name     = "%s"
  policy_document = data.aws_iam_policy_document.test.json
}
`, name)
}
