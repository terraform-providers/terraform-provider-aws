package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsApiGatewayApiKey(t *testing.T) {
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
    id = "${aws_api_gateway_api_key.example_key.id}"
}
`, r)
}
