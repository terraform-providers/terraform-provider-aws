package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsCognitoUserPools_basic(t *testing.T) {
	rName := fmt.Sprintf("tf_acc_ds_cognito_user_pools_%s", acctest.RandString(7))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsCognitoUserPoolsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.selected", "ids.#", "2"),
					resource.TestCheckResourceAttr("data.aws_cognito_user_pools.selected", "arns.#", "2"),
				),
			},
			{
				Config:      testAccDataSourceAwsCognitoUserPoolsConfig_notFound(rName),
				ExpectError: regexp.MustCompile(`No cognito user pool found with name:`),
			},
		},
	})
}

func testAccDataSourceAwsCognitoUserPoolsConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cognito_user_pool" "main" {
	count = 2
	name  = "%s"
}

data "aws_cognito_user_pools" "selected" {
	name = "${aws_cognito_user_pool.main.*.name[0]}"
}
`, rName)
}

func testAccDataSourceAwsCognitoUserPoolsConfig_notFound(rName string) string {
	return fmt.Sprintf(`
data "aws_cognito_user_pools" "selected" {
	name = "%s-not-found"
}
`, rName)
}
