package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSCloudwatchLogGroupsDataSource(t *testing.T) {
	rInt := acctest.RandInt()

	dataSourceName := "data.aws_cloudwatch_log_groups.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupsDataSourceConfigCreateResource(rInt),
			},
			{
				Config: testAccCheckAWSCloudwatchLogGroupsDataSourceConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "names.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "arns.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "creation_times.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_in_days.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "metric_filter_counts.#", "4"),
					resource.TestCheckResourceAttr(dataSourceName, "kms_key_ids.#", "4"),
				),
			},
		},
	})
}

func testAccCheckAWSCloudwatchLogGroupsDataSourceConfigCreateResource(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "default" {
	name = "tf-acc-test-%[1]d-default"
}

resource "aws_cloudwatch_log_group" "retention_in_days" {
    name = "tf-acc-test-%[1]d-retention-in-days"
    retention_in_days = 365
}

resource "aws_cloudwatch_log_group" "name_prefix" {
	name_prefix = "tf-acc-test-%[1]d-"
}

resource "aws_kms_key" "foo" {
    description = "Terraform acc test %[1]d"
    deletion_window_in_days = 7
    policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-%[1]d",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "kms_key_id" {
  name = "tf-acc-test-%[1]d-kms-key-id"
	kms_key_id = "${aws_kms_key.foo.arn}"
}
`, rInt)
}

func testAccCheckAWSCloudwatchLogGroupsDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
%s

data "aws_cloudwatch_log_groups" "test" {
  prefix = "tf-acc-test-%[2]d-"
}
`, testAccCheckAWSCloudwatchLogGroupsDataSourceConfigCreateResource(rInt), rInt)
}
