package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/glue"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsGlueConnection_basic(t *testing.T) {
	resourceName := "aws_glue_connection.test"
	datasourceName := "data.aws_glue_connection.test"
	rName := fmt.Sprintf("tf-testacc-glue-connection-%s", acctest.RandString(13))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, glue.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsGlueConnectionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsGlueConnectionCheck(datasourceName),
					resource.TestCheckResourceAttrPair(datasourceName, "catalog_id", resourceName, "catalog_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_type", resourceName, "connection_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_properties", resourceName, "connection_properties"),
					resource.TestCheckResourceAttrPair(datasourceName, "physical_connection_requirements", resourceName, "physical_connection_requirements"),
					resource.TestCheckResourceAttrPair(datasourceName, "match_criteria", resourceName, "match_criteria"),
				),
			},
		},
	})
}

func testAccDataSourceAwsGlueConnectionCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccDataSourceAwsGlueConnectionConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_glue_connection" "test" {
  connection_properties = {
    JDBC_CONNECTION_URL = "jdbc:mysql://terraformacctesting.com/testdatabase"
    PASSWORD            = "testpassword"
    USERNAME            = "testusername"
  }

  name = "%s"
}

data "aws_glue_connection" "test" {
  id = aws_glue_connection.test.id
}
`, rName)
}
