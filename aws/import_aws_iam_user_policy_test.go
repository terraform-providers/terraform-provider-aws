package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func testAccAwsIamUserPolicyConfig(suffix string) string {
	return fmt.Sprintf(`
resource "aws_iam_user" "user_%[1]s" {
	name = "tf_test_user_test_%[1]s"
	path = "/"
}

resource "aws_iam_user_policy" "foo_%[1]s" {
	name = "tf_test_policy_test_%[1]s"
	user = "${aws_iam_user.user_%[1]s.name}"
	policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}
`, suffix)
}

func TestAccAWSIAMUserPolicy_importBasic(t *testing.T) {
	suffix := randomString(10)
	resourceName := fmt.Sprintf("aws_iam_user_policy.foo_%s", suffix)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckIAMUserPolicyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccAwsIamUserPolicyConfig(suffix),
			},

			resource.TestStep{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}
