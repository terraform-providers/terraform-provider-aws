package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsDxConnection_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_connection.test"
	datasourceName := "data.aws_dx_connection.test"

	args := map[string]string{
		"Bandwidth": "1Gbps",
		"Location":  "EqDC2",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsDxConnectionConfig_NonExistent,
				ExpectError: regexp.MustCompile(`no DirectConnect connection matched; change the search criteria and try again`),
			},
			{
				Config: testAccDataSourceAwsDxConnectionConfig(rName, args["Bandwidth"], args["Location"]),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "aws_device", resourceName, "aws_device"),
					resource.TestCheckResourceAttrPair(datasourceName, "bandwidth", resourceName, "bandwidth"),
					resource.TestCheckResourceAttrPair(datasourceName, "has_logical_redundancy", resourceName, "has_logical_redundancy"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "jumbo_frame_capable", resourceName, "jumbo_frame_capable"),
					resource.TestCheckResourceAttr(datasourceName, "lag_id", "0"),
					resource.TestCheckResourceAttrPair(datasourceName, "location", resourceName, "location"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					testAccCheckResourceAttrAccountID(datasourceName, "owner_account_id"),
					resource.TestCheckResourceAttr(datasourceName, "partner_name", ""),
					resource.TestCheckResourceAttr(datasourceName, "provider_name", ""),
					resource.TestCheckResourceAttr(datasourceName, "state", "requested"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key1", resourceName, "tags.Key1"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.Key2", resourceName, "tags.Key2"),
					resource.TestCheckResourceAttr(datasourceName, "vlan", "0"),
				),
			},
		},
	})
}

const testAccDataSourceAwsDxConnectionConfig_NonExistent = `
data "aws_dx_connection" "test" {
  connection_id = "dxcon-12345678"
}
`

func testAccDataSourceAwsDxConnectionConfig(rName, rBandwidth, rLocation string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "wrong" {
  name      = "%[1]s-wrong"
  bandwidth = "%[2]s"
  location  = "%[3]s"
}

resource "aws_dx_connection" "test" {
  name      = "%[1]s"
  bandwidth = "%[2]s"
  location  = "%[3]s"

  tags = {
    Key1 = "Value1h"
    Key2 = "Value2h"
  }
}

data "aws_dx_connection" "test" {
  connection_id = aws_dx_connection.test.id
}
`, rName, rBandwidth, rLocation)
}
