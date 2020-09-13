package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsApiGatewayApiKey_basic(t *testing.T) {
	rName := acctest.RandString(8)
	resourceName1 := "aws_api_gateway_api_key.example_key"
	dataSourceName1 := "data.aws_api_gateway_api_key.test_key"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayApiKeyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, "id", dataSourceName1, "id"),
					resource.TestCheckResourceAttrPair(resourceName1, "name", dataSourceName1, "name"),
					resource.TestCheckResourceAttrPair(resourceName1, "value", dataSourceName1, "value"),
					resource.TestCheckResourceAttrPair(resourceName1, "enabled", dataSourceName1, "enabled"),
					resource.TestCheckResourceAttrPair(resourceName1, "description", dataSourceName1, "description"),
					resource.TestCheckResourceAttrSet(dataSourceName1, "last_updated_date"),
					resource.TestCheckResourceAttrSet(dataSourceName1, "created_date"),
					resource.TestCheckResourceAttr(dataSourceName1, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayApiKeyConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_api_key" "example_key" {
  name = "%s"
}

data "aws_api_gateway_api_key" "test_key" {
  id = aws_api_gateway_api_key.example_key.id
}
`, r)
}
