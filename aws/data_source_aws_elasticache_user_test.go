package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSElasticacheUser_basic(t *testing.T) {
	resourceName := "aws_elasticache_user.test-basic"
	dataSourceName := "data.aws_elasticache_user.test-basic"
	rName := fmt.Sprintf("a-user-test-tf-basic")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSElastiCacheUserConfigWithDataSource(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_id", resourceName, "user_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "user_name", resourceName, "user_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "access_string", resourceName, "access_string"),
				),
			},
		},
	})
}

// Basic Resource
func testAccAWSElastiCacheUserConfigWithDataSource(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_user" "test-basic" {
  user_id              = %[1]q
  user_name            = %[1]q
  access_string        = "on ~* +@all"
  engine               = "redis"
  no_password_required = true
}

data "aws_elasticache_user" "test-basic" {
  user_id = aws_elasticache_user.test-basic.user_id
}
`, rName)
}
