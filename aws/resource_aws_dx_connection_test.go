package aws

import (
	"fmt"
	"log"
	"regexp"
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

func TestAccAwsDxConnection_basic(t *testing.T) {
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := fmt.Sprintf("tf-testacc-dxconn-%s", acctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxConnectionExists(resourceName, &connection),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxConnection_Tags(t *testing.T) {
	var connection directconnect.Connection
	resourceName := "aws_dx_connection.test"
	rName := fmt.Sprintf("tf-testacc-dxconn-%s", acctest.RandString(14))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, directconnect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxConnectionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxConnectionConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxConnectionExists(resourceName, &connection),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2a"),
				),
			},
			{
				Config: testAccDxConnectionConfig_tagsUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxConnectionExists(resourceName, &connection),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "directconnect", regexp.MustCompile(`dxcon/.+`)),
					resource.TestCheckResourceAttr(resourceName, "bandwidth", "1Gbps"),
					resource.TestCheckResourceAttrSet(resourceName, "location"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2b"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
				),
			},
			// Test import.
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeConnections(&directconnect.DescribeConnectionsInput{
			ConnectionId: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, directconnect.ErrCodeClientException, "does not exist") {
			continue
		}
		if err != nil {
			return err
		}

		for _, v := range resp.Connections {
			if aws.StringValue(v.ConnectionId) == rs.Primary.ID && aws.StringValue(v.ConnectionState) != directconnect.ConnectionStateDeleted {
				return fmt.Errorf("[DESTROY ERROR] Direct Connect connection (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsDxConnectionExists(name string, connection *directconnect.Connection) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).dxconn

		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		resp, err := conn.DescribeConnections(&directconnect.DescribeConnectionsInput{
			ConnectionId: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		for _, c := range resp.Connections {
			if aws.StringValue(c.ConnectionId) == rs.Primary.ID {
				*connection = *c

				return nil
			}
		}

		return fmt.Errorf("Direct Connect connection (%s) not found", rs.Primary.ID)
	}
}

func testAccDxConnectionConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]
}
`, rName)
}

func testAccDxConnectionConfig_tags(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]

  tags = {
    Name = %[1]q
    Key1 = "Value1"
    Key2 = "Value2a"
  }
}
`, rName)
}

func testAccDxConnectionConfig_tagsUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_dx_locations" "test" {}

resource "aws_dx_connection" "test" {
  name      = %[1]q
  bandwidth = "1Gbps"
  location  = tolist(data.aws_dx_locations.test.location_codes)[0]

  tags = {
    Name = %[1]q
    Key2 = "Value2b"
    Key3 = "Value3"
  }
}
`, rName)
}
