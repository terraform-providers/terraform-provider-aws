package aws

import (
	"errors"
	"fmt"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

type testAccDxHostedConnectionEnv struct {
	ConnectionId   string
	OwnerAccountId string
}

func TestAccAWSDxHostedConnection_basic(t *testing.T) {
	env, err := testAccCheckAwsDxHostedConnectionEnv()
	if err != nil {
		TestAccSkip(t, err.Error())
	}

	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_hosted_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxHostedConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedConnectionConfig(connectionName, env.ConnectionId, env.OwnerAccountId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "connection_id", env.ConnectionId),
					resource.TestCheckResourceAttr(resourceName, "owner_account_id", env.OwnerAccountId),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "100Mbps"),
					resource.TestCheckResourceAttr(resourceName, "vlan", "4094"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSDxHostedConnection_tags(t *testing.T) {
	env, err := testAccCheckAwsDxHostedConnectionEnv()
	if err != nil {
		TestAccSkip(t, err.Error())
	}

	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_hosted_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxHostedConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxHostedConnectionConfig_tags(connectionName, env.ConnectionId, env.OwnerAccountId),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxHostedConnectionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "name", connectionName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Origin", "Acceptance Test"),
				),
			},
		},
	})
}

func testAccCheckAwsDxHostedConnectionEnv() (*testAccDxHostedConnectionEnv, error) {
	result := &testAccDxHostedConnectionEnv{
		ConnectionId:   os.Getenv("TEST_AWS_DX_CONNECTION_ID"),
		OwnerAccountId: os.Getenv("TEST_AWS_DX_OWNER_ACCOUNT_ID"),
	}

	if result.ConnectionId == "" || result.OwnerAccountId == "" {
		return nil, errors.New("TEST_AWS_DX_CONNECTION_ID and TEST_AWS_DX_OWNER_ACCOUNT_ID must be set for aws_dx_hosted_connection tests")
	}

	return result, nil
}

func testAccCheckAwsDxHostedConnectionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_hosted_connection" {
			continue
		}

		input := &directconnect.DescribeHostedConnectionsInput{
			ConnectionId: aws.String(rs.Primary.Attributes["connection_id"]),
		}

		resp, err := conn.DescribeHostedConnections(input)
		if err != nil {
			return err
		}
		for _, v := range resp.Connections {
			if *v.ConnectionId == rs.Primary.ID && !(*v.ConnectionState == directconnect.ConnectionStateDeleted) {
				return fmt.Errorf("[DESTROY ERROR] Dx Hosted Connection (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsDxHostedConnectionExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxHostedConnectionConfig(name, connectionId, ownerAccountId string) string {
	return fmt.Sprintf(`
resource "aws_dx_hosted_connection" "test" {
  name             = "%s"
  connection_id    = "%s"
  owner_account_id = "%s"
  bandwidth        = "100Mbps"
  vlan             = 4094
}
`, name, connectionId, ownerAccountId)
}

func testAccDxHostedConnectionConfig_tags(name, connectionId, ownerAccountId string) string {
	return fmt.Sprintf(`
resource "aws_dx_hosted_connection" "test" {
  name             = "%s"
  connection_id    = "%s"
  owner_account_id = "%s"
  bandwidth        = "100Mbps"
  vlan             = 4093

  tags = {
    Origin = "Acceptance Test"
  }
}
`, name, connectionId, ownerAccountId)
}
