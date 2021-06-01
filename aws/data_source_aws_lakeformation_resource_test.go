package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/lakeformation"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSLakeFormationResourceDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_lakeformation_resource.test"
	resourceName := "aws_lakeformation_resource.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(lakeformation.EndpointsID, t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSLakeFormationResourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLakeFormationResourceDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "role_arn", resourceName, "role_arn"),
				),
			},
		},
	})
}

func testAccAWSLakeFormationResourceDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_iam_role" "test" {
  name = %[1]q
  path = "/test/"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "s3.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

data "aws_partition" "current" {}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:GetBucketLocation",
        "s3:ListAllMyBuckets",
        "s3:GetObjectVersion",
        "s3:GetBucketAcl",
        "s3:GetObject",
        "s3:GetObjectACL",
        "s3:PutObject",
        "s3:PutObjectAcl"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}/*",
        "arn:${data.aws_partition.current.partition}:s3:::${aws_s3_bucket.test.id}"
      ]
    }
  ]
}
EOF
}

resource "aws_lakeformation_resource" "test" {
  arn      = aws_s3_bucket.test.arn
  role_arn = aws_iam_role.test.arn
}

data "aws_lakeformation_resource" "test" {
  arn = aws_lakeformation_resource.test.arn
}
`, rName)
}
