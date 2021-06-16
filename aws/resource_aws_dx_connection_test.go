package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dx_connection", &resource.Sweeper{
		Name: "aws_dx_connection",
		F:    testSweepDxConnections,
	})
}

func testSweepDxConnections(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).dxconn

	var sweeperErrs *multierror.Error

	input := &directconnect.DescribeConnectionsInput{}

	// DescribeConnections has no pagination support
	output, err := conn.DescribeConnections(input)

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Direct Connect Connection sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErr := fmt.Errorf("error listing Direct Connect Connections for %s: %w", region, err)
		log.Printf("[ERROR] %s", sweeperErr)
		sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
		return sweeperErrs.ErrorOrNil()
	}

	if output == nil {
		log.Printf("[WARN] Skipping Direct Connect Connection sweep for %s: empty response", region)
		return sweeperErrs.ErrorOrNil()
	}

	for _, connection := range output.Connections {
		if connection == nil {
			continue
		}

		id := aws.StringValue(connection.ConnectionId)

		r := resourceAwsDxConnection()
		d := r.Data(nil)
		d.SetId(id)

		err = r.Delete(d, client)

		if err != nil {
			sweeperErr := fmt.Errorf("error deleting Direct Connect Connection (%s): %w", id, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSDxConnection_basic(t *testing.T) {
	connectionName := fmt.Sprintf("tf-dx-%s", acctest.RandString(5))
	resourceName := "aws_dx_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, directconnect.EndpointsID),
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
		ErrorCheck:   testAccErrorCheck(t, directconnect.EndpointsID),
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
