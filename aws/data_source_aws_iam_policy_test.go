package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMPolicy_basic(t *testing.T) {
	policyName := fmt.Sprintf("test-policy-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIamPolicyConfig(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_iam_policy.test", "name", policyName),
					resource.TestCheckResourceAttr("data.aws_iam_policy.test", "description", "My test policy"),
					resource.TestCheckResourceAttr("data.aws_iam_policy.test", "path", "/"),
					resource.TestCheckResourceAttrSet("data.aws_iam_policy.test", "policy"),
					resource.TestMatchResourceAttr("data.aws_iam_policy.test", "arn",
						regexp.MustCompile(`^arn:[\w-]+:([a-zA-Z0-9\-])+:([a-z]{2}-(gov-)?[a-z]+-\d{1})?:(\d{12})?:(.*)$`)),
				),
			},
		},
	})

}

func testAccAwsDataSourceIamPolicyConfig(policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test_policy" {
    name = "%s"
    path = "/"
    description = "My test policy"
    policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": [
        "ec2:Describe*"
      ],
      "Effect": "Allow",
      "Resource": "*"
    }
  ]
}
EOF
}

data "aws_iam_policy" "test" {
	arn = "${aws_iam_policy.test_policy.arn}"
}
`, policyName)
}
