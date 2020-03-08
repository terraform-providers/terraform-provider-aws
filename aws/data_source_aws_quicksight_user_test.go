package aws

import (
	"fmt"
	"os"
	"regexp"

	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDataSourceQuickSightUser_basic(t *testing.T) {
	email := "vainamoinen@tuone.la"
	namespace := "default"
	region := os.Getenv("AWS_DEFAULT_REGION")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceQuickSightUserConfig(email, namespace),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_quicksight_user.test", "user_id"),
					// https://docs.aws.amazon.com/sdk-for-go/api/service/quicksight/#User
					// The active status of user. When you create an Amazon QuickSight user that’s
					// not an IAM user or an Active Directory user, that user is inactive until
					// they sign in and provide a password.
					// therefore testing the value doesn't make that much sense and in fact the data source does not define the
					// the value either from now on if identity type is missing from the *User
					// resource.TestCheckResourceAttr("data.aws_quicksight_user.test", "identity_type", ""),
					resource.TestCheckResourceAttr("data.aws_quicksight_user.test", "user_role", "READER"),
					resource.TestCheckResourceAttr("data.aws_quicksight_user.test", "email", email),
					resource.TestCheckResourceAttr("data.aws_quicksight_user.test", "user_name", email),
					resource.TestMatchResourceAttr("data.aws_quicksight_user.test", "arn", regexp.MustCompile("^arn:[^:]+:quicksight:"+region+":[0-9]{12}:user/"+namespace+"/"+email)),
				),
			},
		},
	})
}

func testAccAWSDataSourceQuickSightUserConfig(email, namespace string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_quicksight_user" "test" {
  aws_account_id = "${data.aws_caller_identity.current.account_id}"
  user_name      = %[1]q
  email          = %[1]q
  identity_type  = "QUICKSIGHT"
  user_role      = "READER"
  namespace		 = %[2]q
}

data "aws_quicksight_user" "test" {
  user_name 	 = "${aws_quicksight_user.test.user_name}"
  aws_account_id = "${data.aws_caller_identity.current.account_id}"
  namespace		 = %[2]q
}

`, email, namespace)
}
