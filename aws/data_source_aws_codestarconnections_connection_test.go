package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/codestarconnections"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsCodeStarConnectionsConnection_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codestarconnections_connection.test"
	resourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t) },
		ErrorCheck: testAccErrorCheck(t, codestarconnections.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSCodeStarConnectionsConnectionConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "provider_type", dataSourceName, "provider_type"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "connection_status", dataSourceName, "connection_status"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsCodeStarConnectionsConnection_tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_codestarconnections_connection.test"
	resourceName := "aws_codestarconnections_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(codestarconnections.EndpointsID, t) },
		ErrorCheck: testAccErrorCheck(t, codestarconnections.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSCodeStarConnectionsConnectionConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccDataSourceAWSCodeStarConnectionsConnectionConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"
}

data "aws_codestarconnections_connection" "test" {
  arn = aws_codestarconnections_connection.test.arn
}
`, rName)
}

func testAccDataSourceAWSCodeStarConnectionsConnectionConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_codestarconnections_connection" "test" {
  name          = %[1]q
  provider_type = "Bitbucket"

  tags = {
    "key1" = "value1"
    "key2" = "value2"
  }
}

data "aws_codestarconnections_connection" "test" {
  arn = aws_codestarconnections_connection.test.arn
}
`, rName)
}
