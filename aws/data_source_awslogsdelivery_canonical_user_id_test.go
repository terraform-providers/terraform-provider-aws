package aws

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"testing"
)

func TestAwslogsdeliveryCanonicalUserId(t *testing.T) {
	dataSourceName := "data.aws_awslogsdelivery_canonical_user_id.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAwslogsdeliveryCanonicalIdConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwslogsdeliveryCanonicalUserIdCheckExists(dataSourceName),
				),
			},
		},
	})
}

func testAccDataSourceAwslogsdeliveryCanonicalUserIdCheckExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find awslogsdelivery canonical user ID resource: %s", name)
		}

		if rs.Primary.Attributes["id"] != "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0" {
			return fmt.Errorf("invalid awslogsdelivery canonical user id")
		}

		return nil
	}
}

const testAwslogsdeliveryCanonicalIdConfig = `
data "aws_awslogsdelivery_canonical_user_id" "main" {}
`
