package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsCognitoUserPoolClient_basic(t *testing.T) {
	rPoolName := fmt.Sprintf("tf_acc_ds_cognito_user_pool_%s", acctest.RandString(7))
	rClientName := fmt.Sprintf("tf_acc_ds_cognito_user_pool_client_%s", acctest.RandString(7))

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsCognitoUserPoolClientConfig_basic(rPoolName, rClientName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCognitoUserPoolClientHasExpectedValues("data.aws_cognito_user_pool_client.selected"),
				),
			},
			{
				Config:      testAccDataSourceAwsCognitoUserPoolClientConfig_notFound(rPoolName, rClientName),
				ExpectError: regexp.MustCompile(`No cognito user pool client found with name:`),
			},
		},
	})
}

func testAccDataSourceAwsCognitoUserPoolClientConfig_basic(rPoolName string, rClientName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
	name  = "%s"
}

resource "aws_cognito_user_pool_client" "main" {
	name         = "%s"
  user_pool_id = "${aws_cognito_user_pool.main.id}"
}

data "aws_cognito_user_pool_client" "selected" {
	user_pool_id = "${aws_cognito_user_pool.main.id}"
	name         = "%s"

	depends_on   = ["aws_cognito_user_pool_client.main"]
}
`, rPoolName, rClientName, rClientName)
}

func testAccDataSourceAwsCognitoUserPoolClientConfig_notFound(rPoolName string, rClientName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_cognito_user_pool" "main" {
        name  = "%s"
}

data "aws_cognito_user_pool_client" "selected" {
	name         = "%s-not-found"
	user_pool_id = "${aws_cognito_user_pool.main.id}"
}
`, rPoolName, rClientName)
}

func testAccDataSourceCognitoUserPoolClientHasExpectedValues(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Can't find resource: %s", name)
		}

		if rs.Primary.Attributes["client_id"] == "" {
			return fmt.Errorf("Missing Client ID")
		}
		if rs.Primary.Attributes["client_secret"] == "" {
			return fmt.Errorf("Missing Client Secret")
		}

		return nil
	}
}
