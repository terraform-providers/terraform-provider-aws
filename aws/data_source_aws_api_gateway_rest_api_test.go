package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsApiGatewayRestApi_basic(t *testing.T) {
	rName := acctest.RandString(8)
	dataSourceName := "data.aws_api_gateway_rest_api.test"
	resourceName := "aws_api_gateway_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccAPIGatewayTypeEDGEPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSAPIGatewayRestAPIConfig_Name(rName),
					testAccDataSourceAwsApiGatewayRestApiConfigName(),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_resource_id", resourceName, "root_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_source", resourceName, "api_key_source"),
					resource.TestCheckResourceAttrPair(dataSourceName, "minimum_compression_size", resourceName, "minimum_compression_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "binary_media_types", resourceName, "binary_media_types"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_configuration", resourceName, "endpoint_configuration"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsApiGatewayRestApi_EndpointConfiguration_VpcEndpointIds(t *testing.T) {
	rName := acctest.RandString(8)
	dataSourceName := "data.aws_api_gateway_rest_api.test"
	resourceName := "aws_api_gateway_rest_api.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: composeConfig(
					testAccAWSAPIGatewayRestAPIConfigEndpointConfigurationVpcEndpointIds1(rName),
					testAccDataSourceAwsApiGatewayRestApiConfigName(),
				),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "root_resource_id", resourceName, "root_resource_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags", resourceName, "tags"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "policy", resourceName, "policy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "api_key_source", resourceName, "api_key_source"),
					resource.TestCheckResourceAttrPair(dataSourceName, "minimum_compression_size", resourceName, "minimum_compression_size"),
					resource.TestCheckResourceAttrPair(dataSourceName, "binary_media_types", resourceName, "binary_media_types"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_configuration.#", resourceName, "endpoint_configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint_configuration.0.vpc_endpoint_ids.#", resourceName, "endpoint_configuration.0.vpc_endpoint_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "execution_arn", resourceName, "execution_arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayRestApiConfigName() string {
	return `
data "aws_api_gateway_rest_api" "test" {
  name = aws_api_gateway_rest_api.test.name
}
`
}
