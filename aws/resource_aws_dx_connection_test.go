package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDxConnection_basic(t *testing.T) {
	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfig(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttr(resourceName, "location", "EqSe2-EQ"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDxConnection_tags(t *testing.T) {
	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfig_tags(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDxConnectionConfig_tagsChanged(connectionName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
		},
	})
}

func testAccCheckAwsDxConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_connection" {
			continue
		}

		input := &directconnect.DescribeConnectionsInput{
			ConnectionId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeConnections(input)
		if err != nil {
			return err
		}
		for _, v := range resp.Connections {
			if *v.ConnectionId == rs.Primary.ID && !(*v.ConnectionState == directconnect.ConnectionStateDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Dx Connection (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsDxConnectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxConnectionConfig(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "test" {
  name      = "%s"
  bandwidth = "1Gbps"
  location  = "EqSe2-EQ"
}
`, n)
}

func testAccDxConnectionConfig_tags(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "test" {
  name      = "%s"
  bandwidth = "1Gbps"
  location  = "EqSe2-EQ"

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, n)
}

func testAccDxConnectionConfig_tagsChanged(n string) string {
	return fmt.Sprintf(`
resource "aws_dx_connection" "test" {
  name      = "%s"
  bandwidth = "1Gbps"
  location  = "EqSe2-EQ"

  tags = {
    Usage = "changed"
  }
}
`, n)
}
