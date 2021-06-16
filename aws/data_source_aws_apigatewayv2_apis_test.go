package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSAPIGatewayV2ApisDataSource_Name(t *testing.T) {
	dataSource1Name := "data.aws_apigatewayv2_apis.test1"
	dataSource2Name := "data.aws_apigatewayv2_apis.test2"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, apigatewayv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApisDataSourceConfigName(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2ApisDataSource_ProtocolType(t *testing.T) {
	dataSource1Name := "data.aws_apigatewayv2_apis.test1"
	dataSource2Name := "data.aws_apigatewayv2_apis.test2"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, apigatewayv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApisDataSourceConfigProtocolType(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2ApisDataSource_Tags(t *testing.T) {
	dataSource1Name := "data.aws_apigatewayv2_apis.test1"
	dataSource2Name := "data.aws_apigatewayv2_apis.test2"
	dataSource3Name := "data.aws_apigatewayv2_apis.test3"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, apigatewayv2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2ApisDataSourceConfigTags(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSource3Name, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccAWSAPIGatewayV2ApisDataSourceConfigBase(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_apigatewayv2_api" "test1" {
  name          = %[1]q
  protocol_type = "HTTP"

  tags = {
    Name = %[1]q
  }
}

resource "aws_apigatewayv2_api" "test2" {
  name          = %[2]q
  protocol_type = "HTTP"

  tags = {
    Name = %[2]q
  }
}

resource "aws_apigatewayv2_api" "test3" {
  name                       = %[2]q
  protocol_type              = "WEBSOCKET"
  route_selection_expression = "$request.body.action"

  tags = {
    Name = %[2]q
  }
}
`, rName1, rName2)
}

func testAccAWSAPIGatewayV2ApisDataSourceConfigName(rName1, rName2 string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2ApisDataSourceConfigBase(rName1, rName2),
		`
data "aws_apigatewayv2_apis" "test1" {
  # Force dependency on resources.
  name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 0)
}

data "aws_apigatewayv2_apis" "test2" {
  # Force dependency on resources.
  name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 1)
}
`)
}

func testAccAWSAPIGatewayV2ApisDataSourceConfigProtocolType(rName1, rName2 string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2ApisDataSourceConfigBase(rName1, rName2),
		fmt.Sprintf(`
data "aws_apigatewayv2_apis" "test1" {
  name = %[1]q

  protocol_type = element([aws_apigatewayv2_api.test1.protocol_type, aws_apigatewayv2_api.test2.protocol_type, aws_apigatewayv2_api.test3.protocol_type], 0)
}

data "aws_apigatewayv2_apis" "test2" {
  name = %[2]q

  protocol_type = element([aws_apigatewayv2_api.test1.protocol_type, aws_apigatewayv2_api.test2.protocol_type, aws_apigatewayv2_api.test3.protocol_type], 3)
}
`, rName1, rName2))
}

func testAccAWSAPIGatewayV2ApisDataSourceConfigTags(rName1, rName2 string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2ApisDataSourceConfigBase(rName1, rName2),
		`
data "aws_apigatewayv2_apis" "test1" {
  # Force dependency on resources.
  tags = {
    Name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 0)
  }
}

data "aws_apigatewayv2_apis" "test2" {
  # Force dependency on resources.
  tags = {
    Name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 1)
  }
}

data "aws_apigatewayv2_apis" "test3" {
  # Force dependency on resources.
  tags = {
    Name = element([aws_apigatewayv2_api.test1.name, aws_apigatewayv2_api.test2.name, aws_apigatewayv2_api.test3.name], 2)
    Key2 = "Value2"
  }
}
`)
}
