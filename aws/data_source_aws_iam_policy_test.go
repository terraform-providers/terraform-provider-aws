package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDataSourceIAMPolicy_basic(t *testing.T) {
	resourceName := "data.aws_iam_policy.test"
	policyName := fmt.Sprintf("test-policy-%s", acctest.RandString(10))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDataSourceIamPolicyConfig(policyName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", policyName),
					resource.TestCheckResourceAttr(resourceName, "description", "My test policy"),
					resource.TestCheckResourceAttr(resourceName, "path", "/"),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_iam_policy.test_policy", "arn"),
				),
			},
		},
	})

}

func testAccAwsDataSourceIamPolicyConfig(policyName string) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test_policy" {
  name        = "%s"
  path        = "/"
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
  arn = aws_iam_policy.test_policy.arn
}
`, policyName)
}
