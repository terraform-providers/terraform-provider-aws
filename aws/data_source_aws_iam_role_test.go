package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMRole_basic(t *testing.T) {
	roleName := fmt.Sprintf("test-role-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIAMRoleConfig(roleName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_iam_role.test", "unique_id"),
					resource.TestCheckResourceAttrSet("data.aws_iam_role.test", "assume_role_policy"),
					resource.TestCheckResourceAttr("data.aws_iam_role.test", "path", "/testpath/"),
					resource.TestCheckResourceAttr("data.aws_iam_role.test", "permissions_boundary", ""),
					resource.TestCheckResourceAttr("data.aws_iam_role.test", "name", roleName),
					resource.TestCheckResourceAttrSet("data.aws_iam_role.test", "create_date"),
					resource.TestMatchResourceAttr("data.aws_iam_role.test", "arn",
						regexp.MustCompile(`^arn:[\w-]+:([a-zA-Z0-9\-])+:([a-z]{2}-(gov-)?[a-z]+-\d{1})?:(\d{12})?:(.*)$`)),
				),
			},
		},
	})
}

func testAccAwsIAMRoleConfig(roleName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test_role" {
  name = "%s"

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

  path = "/testpath/"
}

data "aws_iam_role" "test" {
  name = "${aws_iam_role.test_role.name}"
}
`, roleName)
}
