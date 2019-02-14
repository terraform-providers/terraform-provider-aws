package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsApiGatewayResource(t *testing.T) {
	rName := acctest.RandString(8)
	resourceName1 := "aws_api_gateway_resource.example_v1"
	dataSourceName1 := "data.aws_api_gateway_resource.example_v1"
	resourceName2 := "aws_api_gateway_resource.example_v1_endpoint"
	dataSourceName2 := "data.aws_api_gateway_resource.example_v1_endpoint"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayResourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName1, "id", dataSourceName1, "id"),
					resource.TestCheckResourceAttrPair(resourceName1, "parent_id", dataSourceName1, "parent_id"),
					resource.TestCheckResourceAttrPair(resourceName1, "path_part", dataSourceName1, "path_part"),
					resource.TestCheckResourceAttrPair(resourceName2, "id", dataSourceName2, "id"),
					resource.TestCheckResourceAttrPair(resourceName2, "parent_id", dataSourceName2, "parent_id"),
					resource.TestCheckResourceAttrPair(resourceName2, "path_part", dataSourceName2, "path_part"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayResourceConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_rest_api" "example" {
	name = "%s_example"
}

resource "aws_api_gateway_resource" "example_v1" {
	rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    parent_id   = "${aws_api_gateway_rest_api.example.root_resource_id}"
	path_part   = "v1"
}

resource "aws_api_gateway_resource" "example_v1_endpoint" {
	rest_api_id = "${aws_api_gateway_rest_api.example.id}"
    parent_id   = "${aws_api_gateway_resource.example_v1.id}"
	path_part   = "endpoint"
}

data "aws_api_gateway_resource" "example_v1" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  path        = "/${aws_api_gateway_resource.example_v1.path_part}"
}

data "aws_api_gateway_resource" "example_v1_endpoint" {
  rest_api_id = "${aws_api_gateway_rest_api.example.id}"
  path        = "/${aws_api_gateway_resource.example_v1.path_part}/${aws_api_gateway_resource.example_v1_endpoint.path_part}"
}
`, r)
}
