package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSDataSourceIAMInstanceProfile_basic(t *testing.T) {
	roleName := fmt.Sprintf("tf-acc-ds-instance-profile-role-%d", acctest.RandInt())
	profileName := fmt.Sprintf("tf-acc-ds-instance-profile-%d", acctest.RandInt())

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDatasourceAwsIamInstanceProfileConfig(roleName, profileName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr(
						"data.aws_iam_instance_profile.test",
						"arn",
						regexp.MustCompile("^arn:[^:]+:iam::[0-9]{12}:instance-profile/testpath/"+profileName+"$"),
					),
					resource.TestCheckResourceAttr("data.aws_iam_instance_profile.test", "path", "/testpath/"),
					resource.TestMatchResourceAttr(
						"data.aws_iam_instance_profile.test",
						"role_arn",
						regexp.MustCompile("^arn:[^:]+:iam::[0-9]{12}:role/"+roleName+"$"),
					),
					resource.TestCheckResourceAttrSet("data.aws_iam_instance_profile.test", "role_id"),
					resource.TestCheckResourceAttr("data.aws_iam_instance_profile.test", "role_name", roleName),
				),
			},
		},
	})
}

func testAccDatasourceAwsIamInstanceProfileConfig(roleName, profileName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
	name = "%s"
	assume_role_policy = "{\"Version\":\"2012-10-17\",\"Statement\":[{\"Effect\":\"Allow\",\"Principal\":{\"Service\":[\"ec2.amazonaws.com\"]},\"Action\":[\"sts:AssumeRole\"]}]}"
}

resource "aws_iam_instance_profile" "test" {
	name = "%s"
	role = "${aws_iam_role.test.name}"
	path = "/testpath/"
}

data "aws_iam_instance_profile" "test" {
	name = "${aws_iam_instance_profile.test.name}"
}
`, roleName, profileName)
}
