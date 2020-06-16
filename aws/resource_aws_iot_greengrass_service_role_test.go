package aws

import (
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsIotGreengrassServiceRole_basic(t *testing.T) {
	resourceName := "aws_iot_greengrass_service_role.test"
	roleResourceName := "aws_iam_role.greengrass_service_role"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsIotGreengrassServiceRoleDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotGreengrassServiceRoleConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsIotGreengrassServiceRoleExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "role_arn", roleResourceName, "arn"),
				),
			},
		},
	})
}

func testAccCheckAwsIotGreengrassServiceRoleDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn
	input := &greengrass.GetServiceRoleForAccountInput{}

	_, err := conn.GetServiceRoleForAccount(input)

	if err != nil {
		if isAWSErrRequestFailureStatusCode(err, 404) {
			//No greengrass service role is set for this account
			return nil
		}
		return err
	}
	return errors.New("greengrass service role was not reset")
}

func testAccCheckAwsIotGreengrassServiceRoleExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSIotGreengrassServiceRoleConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "greengrass_service_role" {
  name = "greengrass_service_role_test_%[1]d"
  assume_role_policy = <<EOF
{
"Version": "2012-10-17",
"Statement": [
	{
	"Effect": "Allow",
	"Principal": {
		"Service": "greengrass.amazonaws.com"
	},
	"Action": "sts:AssumeRole"
	}
]
}
EOF
}

resource "aws_iot_greengrass_service_role" "test" {
	role_arn    = aws_iam_role.greengrass_service_role.arn
}
`, rInt)
}
