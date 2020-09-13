package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEcrAuthorizationTokenDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_ecr_authorization_token.repo"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsEcrAuthorizationTokenDataSourceBasicConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "proxy_endpoint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_name"),
					resource.TestMatchResourceAttr(dataSourceName, "user_name", regexp.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "password"),
				),
			},
			{
				Config: testAccCheckAwsEcrAuthorizationTokenDataSourceRepositoryConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "registry_id", "aws_ecr_repository.repo", "registry_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "authorization_token"),
					resource.TestCheckResourceAttrSet(dataSourceName, "proxy_endpoint"),
					resource.TestCheckResourceAttrSet(dataSourceName, "expires_at"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_name"),
					resource.TestMatchResourceAttr(dataSourceName, "user_name", regexp.MustCompile(`AWS`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "password"),
				),
			},
		},
	})
}

var testAccCheckAwsEcrAuthorizationTokenDataSourceBasicConfig = `
data "aws_ecr_authorization_token" "repo" {}
`

func testAccCheckAwsEcrAuthorizationTokenDataSourceRepositoryConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ecr_repository" "repo" {
  name = %q
}

data "aws_ecr_authorization_token" "repo" {
  registry_id = aws_ecr_repository.repo.registry_id
}
`, rName)
}
