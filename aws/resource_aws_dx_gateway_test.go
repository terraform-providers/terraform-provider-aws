package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dx_gateway", &resource.Sweeper{
		Name: "aws_dx_gateway",
		F:    testSweepDirectConnectGateways,
		Dependencies: []string{
			"aws_dx_gateway_association",
		},
	})
}

func testSweepDirectConnectGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).dxconn
	input := &directconnect.DescribeDirectConnectGatewaysInput{}

	for {
		output, err := conn.DescribeDirectConnectGateways(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Direct Connect Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving Direct Connect Gateways: %s", err)
		}

		for _, gateway := range output.DirectConnectGateways {
			id := aws.StringValue(gateway.DirectConnectGatewayId)

			if aws.StringValue(gateway.DirectConnectGatewayState) != directconnect.GatewayStateAvailable {
				log.Printf("[INFO] Skipping Direct Connect Gateway in non-available (%s) state: %s", aws.StringValue(gateway.DirectConnectGatewayState), id)
				continue
			}

			var associations bool
			associationInput := &directconnect.DescribeDirectConnectGatewayAssociationsInput{
				DirectConnectGatewayId: gateway.DirectConnectGatewayId,
			}

			for {
				associationOutput, err := conn.DescribeDirectConnectGatewayAssociations(associationInput)

				if err != nil {
					return fmt.Errorf("error retrieving Direct Connect Gateway (%s) Associations: %s", id, err)
				}

				// If associations still remain, its likely that our region is not the home
				// region of those associations and the previous sweepers skipped them.
				// When we hit this condition, we skip trying to delete the gateway as it
				// will go from deleting -> available after a few minutes and timeout.
				if len(associationOutput.DirectConnectGatewayAssociations) > 0 {
					associations = true
					break
				}

				if aws.StringValue(associationOutput.NextToken) == "" {
					break
				}

				associationInput.NextToken = associationOutput.NextToken
			}

			if associations {
				log.Printf("[INFO] Skipping Direct Connect Gateway with remaining associations: %s", id)
				continue
			}

			input := &directconnect.DeleteDirectConnectGatewayInput{
				DirectConnectGatewayId: aws.String(id),
			}

			log.Printf("[INFO] Deleting Direct Connect Gateway: %s", id)
			_, err := conn.DeleteDirectConnectGateway(input)

			if isAWSErr(err, directconnect.ErrCodeClientException, "does not exist") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting Direct Connect Gateway (%s): %s", id, err)
			}

			if err := waitForDirectConnectGatewayDeletion(conn, id, 20*time.Minute); err != nil {
				return fmt.Errorf("error waiting for Direct Connect Gateway (%s) to be deleted: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAwsDxGateway_basic(t *testing.T) {
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayConfig(acctest.RandString(5), acctest.RandIntRange(64512, 65534)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_account_id"),
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

func TestAccAwsDxGateway_complex(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	resourceName := "aws_dx_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDxGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationConfig_multiVpnGatewaysSingleAccount(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayExists(resourceName),
					testAccCheckResourceAttrAccountID(resourceName, "owner_account_id"),
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

func testAccCheckAwsDxGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway" {
			continue
		}

		input := &directconnect.DescribeDirectConnectGatewaysInput{
			DirectConnectGatewayId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeDirectConnectGateways(input)
		if err != nil {
			return err
		}
		for _, v := range resp.DirectConnectGateways {
			if *v.DirectConnectGatewayId == rs.Primary.ID && !(*v.DirectConnectGatewayState == directconnect.GatewayStateDeleted) {
				return fmt.Errorf("[DESTROY ERROR] DX Gateway (%s) not deleted", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccCheckAwsDxGatewayExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccDxGatewayConfig(rName string, rBgpAsn int) string {
	return fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  name            = "terraform-testacc-dxgw-%s"
  amazon_side_asn = "%d"
}
`, rName, rBgpAsn)
}
