package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccCloudfrontCanonicalUserId_basic(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_canonical_user_id.main"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAwsCloudfrontCanonicalIdConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceCloudfrontCanonicalUserIdCheckExists(dataSourceName),
				),
			},
		},
	})
}

func testAccDataSourceCloudfrontCanonicalUserIdCheckExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find cloudfront canonical user ID resource: %s", name)
		}

		if rs.Primary.Attributes["id"] != "c4c1ede66af53448b93c283ce9448c4ba468c9432aa01d700d3878632f77d2d0" {
			return fmt.Errorf("invalid cloudfront canonical user id")
		}

		return nil
	}
}

func TestAccCloudfrontCanonicalUserIdCheck_chinaRegion(t *testing.T) {
	dataSourceName := "data.aws_cloudfront_canonical_user_id.main"

	tests := []struct {
		name   string
		config string
	}{
		{
			name: "cn-north-1",
			config: `
data "aws_cloudfront_canonical_user_id" "main" {
  region = "cn-north-1"
}
`,
		},
		{
			name: "cn-northwest-1",
			config: `
data "aws_cloudfront_canonical_user_id" "main" {
  region = "cn-northwest-1"
}
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resource.ParallelTest(t, resource.TestCase{
				PreCheck:  func() { testAccPreCheck(t) },
				Providers: testAccProviders,
				Steps: []resource.TestStep{
					{
						Config: tt.config,
						Check: resource.ComposeTestCheckFunc(
							testAccDataSourceCloudfrontCanonicalUserIdCheckChina(dataSourceName),
						),
					},
				},
			})
		})
	}
}

func testAccDataSourceCloudfrontCanonicalUserIdCheckChina(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("can't find cloudfront canonical user ID resource: %s", name)
		}

		if rs.Primary.Attributes["id"] != "a52cb28745c0c06e84ec548334e44bfa7fc2a85c54af20cd59e4969344b7af56" {
			return fmt.Errorf("invalid cloudfront canonical user id")
		}

		return nil
	}
}

const testAwsCloudfrontCanonicalIdConfig = `
data "aws_cloudfront_canonical_user_id" "main" {}
`
