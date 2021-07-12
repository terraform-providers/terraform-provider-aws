package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsDxConnections_Name(t *testing.T) {
	dataSource1Name := "data.aws_dx_connections.test1"
	dataSource2Name := "data.aws_dx_connections.test2"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	args := map[string]string{
		"Bandwidth":   "1Gbps",
		"Location":    "EqDC2",
		"AltLocation": "CSVA1",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDxConnectionsDataSourceConfigName(rName1, rName2, args),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsDxConnections_Tags(t *testing.T) {
	dataSource1Name := "data.aws_dx_connections.test1"
	dataSource2Name := "data.aws_dx_connections.test2"
	dataSource3Name := "data.aws_dx_connections.test3"
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")

	args := map[string]string{
		"Bandwidth":   "1Gbps",
		"Location":    "EqDC2",
		"AltLocation": "CSVA1",
	}

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDxConnectionsDataSourceConfigTags(rName1, rName2, args),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSource1Name, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSource2Name, "ids.#", "2"),
					resource.TestCheckResourceAttr(dataSource3Name, "ids.#", "0"),
				),
			},
		},
	})
}

func testAccAwsDxConnectionsDataSourceConfigBase(rName1, rName2 string, args map[string]string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "test1" {
  name      = %[1]q
  bandwidth = %[3]q
  location  = %[4]q

  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_connection" "test2" {
  name      = %[2]q
  bandwidth = %[3]q
  location  = %[4]q

  tags = {
    Name = %[2]q
  }
}

resource "aws_dx_connection" "test3" {
  name      = %[2]q
  bandwidth = %[3]q
  location  = %[5]q

  tags = {
    Name = %[2]q
  }
}
`, rName1, rName2, args["Bandwidth"], args["Location"], args["AltLocation"])
}

func testAccAwsDxConnectionsDataSourceConfigName(rName1, rName2 string, args map[string]string) string {
	return composeConfig(
		testAccAwsDxConnectionsDataSourceConfigBase(rName1, rName2, args),
		`
data "aws_dx_connections" "test1" {
  # Force dependency on resources
  name = element([aws_dx_connection.test1.name, aws_dx_connection.test2.name, aws_dx_connection.test3.name], 0)
}

data "aws_dx_connections" "test2" {
  # Force dependency on resources
  name = element([aws_dx_connection.test1.name, aws_dx_connection.test2.name, aws_dx_connection.test3.name], 1)
}
`)
}

func testAccAwsDxConnectionsDataSourceConfigTags(rName1, rName2 string, args map[string]string) string {
	return composeConfig(
		testAccAwsDxConnectionsDataSourceConfigBase(rName1, rName2, args),
		`
data "aws_dx_connections" "test1" {
  # Force dependency on resources
  tags = {
    Name = element([aws_dx_connection.test1.name, aws_dx_connection.test2.name, aws_dx_connection.test3.name], 0)
  }
}

data "aws_dx_connections" "test2" {
  # Force dependency on resources
  tags = {
    Name = element([aws_dx_connection.test1.name, aws_dx_connection.test2.name, aws_dx_connection.test3.name], 1)
  }
}

data "aws_dx_connections" "test3" {
  # Force dependency on resources
  tags = {
    Name = element([aws_dx_connection.test1.name, aws_dx_connection.test2.name, aws_dx_connection.test3.name], 2)
    Key2 = "Value2"
  }
}
`)
}
