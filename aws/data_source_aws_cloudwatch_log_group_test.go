package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSCloudwatchLogGroupDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudwatchLogGroupDataSource_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupDataSourceConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", ""),
				),
			},
		},
	})
}

func TestAccAWSCloudwatchLogGroupDataSource_kms(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupDataSourceConfigKMS(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudwatchLogGroupDataSource_retention(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchLogGroupDataSourceConfigRetention(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "creation_time"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
		},
	})
}

func testAccCheckAWSCloudwatchLogGroupDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "%s"
}

data aws_cloudwatch_log_group "test" {
  name = aws_cloudwatch_log_group.test.name
}
`, rName)
}

func testAccCheckAWSCloudwatchLogGroupDataSourceConfigTags(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name = "%s"

  tags = {
    Environment = "Production"
    Foo         = "Bar"
    Empty       = ""
  }
}

data aws_cloudwatch_log_group "test" {
  name = aws_cloudwatch_log_group.test.name
}
`, rName)
}

func testAccCheckAWSCloudwatchLogGroupDataSourceConfigKMS(rName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description             = "Terraform acc test %s"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
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

resource aws_cloudwatch_log_group "test" {
  name       = "%s"
  kms_key_id = aws_kms_key.foo.arn
}

data aws_cloudwatch_log_group "test" {
  name = aws_cloudwatch_log_group.test.name
}
`, rName, rName)
}

func testAccCheckAWSCloudwatchLogGroupDataSourceConfigRetention(rName string) string {
	return fmt.Sprintf(`
resource aws_cloudwatch_log_group "test" {
  name              = "%s"
  retention_in_days = 365
}

data aws_cloudwatch_log_group "test" {
  name = aws_cloudwatch_log_group.test.name
}
`, rName)
}
