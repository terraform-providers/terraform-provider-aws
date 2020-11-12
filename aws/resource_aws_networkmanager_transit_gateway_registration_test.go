package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkmanager"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_networkmanager_transit_gateway_registration", &resource.Sweeper{
		Name: "aws_networkmanager_transit_gateway_registration",
		F:    testSweepNetworkManagerTransitGatewayRegistration,
	})
}
func testSweepNetworkManagerTransitGatewayRegistration(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).networkmanagerconn
	var sweeperErrs *multierror.Error
	err = conn.GetTransitGatewayRegistrationsPages(&networkmanager.GetTransitGatewayRegistrationsInput{},
		func(page *networkmanager.GetTransitGatewayRegistrationsOutput, lastPage bool) bool {
			for _, transitGatewayRegistration := range page.TransitGatewayRegistrations {
				input := &networkmanager.DeregisterTransitGatewayInput{
					GlobalNetworkId:   transitGatewayRegistration.GlobalNetworkId,
					TransitGatewayArn: transitGatewayRegistration.TransitGatewayArn,
				}
				transitGatewayArn := aws.StringValue(transitGatewayRegistration.TransitGatewayArn)
				globalNetworkID := aws.StringValue(transitGatewayRegistration.GlobalNetworkId)
				log.Printf("[INFO] Deleting Network Manager Transit Gateway Registration: %s", transitGatewayArn)
				req, _ := conn.DeregisterTransitGatewayRequest(input)
				err = req.Send()
				if isAWSErr(err, "InvalidTransitGatewayRegistrationArn.NotFound", "") {
					continue
				}
				if err != nil {
					sweeperErr := fmt.Errorf("failed to delete Network Manager Transit Gateway Registration %s: %s", transitGatewayArn, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
				if err := waitForNetworkManagerTransitGatewayRegistrationDeletion(conn, globalNetworkID, transitGatewayArn); err != nil {
					sweeperErr := fmt.Errorf("error waiting for Network Manager Transit Gateway Registration (%s) deletion: %s", transitGatewayArn, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
			return !lastPage
		})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Network Manager Transit Gateway Registration sweep for %s: %s", region, err)
		return nil
	}
	if err != nil {
		return fmt.Errorf("Error retrieving Network Manager Transit Gateway Registrations: %s", err)
	}
	return sweeperErrs.ErrorOrNil()
}
func TestAccAWSNetworkManagerTransitGatewayRegistration_basic(t *testing.T) {
	resourceName := "aws_networkmanager_transit_gateway_registration.test"
	gloablNetworkResourceName := "aws_networkmanager_global_network.test"
	gloablNetwork2ResourceName := "aws_networkmanager_global_network.test2"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkManagerTransitGatewayRegistrationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkManagerTransitGatewayRegistrationConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerTransitGatewayRegistrationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetworkResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_arn", transitGatewayResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSNetworkManagerTransitGatewayRegistrationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccNetworkManagerTransitGatewayRegistrationConfig_Update(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkManagerTransitGatewayRegistrationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "global_network_id", gloablNetwork2ResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_arn", transitGatewayResourceName, "arn"),
				),
			},
		},
	})
}
func testAccCheckAwsNetworkManagerTransitGatewayRegistrationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkmanager_transit_gateway_registration" {
			continue
		}
		transitGatewayArn, err := networkmanagerDescribeTransitGatewayRegistration(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.Attributes["transit_gateway_arn"])
		if err != nil {
			if isAWSErr(err, networkmanager.ErrCodeValidationException, "") {
				return nil
			}
			return err
		}
		if transitGatewayArn == nil {
			continue
		}
		return fmt.Errorf("Expected Transit Gateway Registration to be destroyed, %s found", rs.Primary.ID)
	}
	return nil
}
func testAccCheckAwsNetworkManagerTransitGatewayRegistrationExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		conn := testAccProvider.Meta().(*AWSClient).networkmanagerconn
		transitGatewayArn, err := networkmanagerDescribeTransitGatewayRegistration(conn, rs.Primary.Attributes["global_network_id"], rs.Primary.Attributes["transit_gateway_arn"])
		if err != nil {
			return err
		}
		if transitGatewayArn == nil {
			return fmt.Errorf("Network Manager Transit Gateway Registration not found")
		}
		if aws.StringValue(transitGatewayArn.State.Code) != networkmanager.TransitGatewayRegistrationStateAvailable && aws.StringValue(transitGatewayArn.State.Code) != networkmanager.TransitGatewayRegistrationStatePending {
			return fmt.Errorf("Network Manager Transit Gateway Registration (%s) exists in (%s) state", rs.Primary.Attributes["transit_gateway_arn"], aws.StringValue(transitGatewayArn.State.Code))
		}
		return err
	}
}
func testAccNetworkManagerTransitGatewayRegistrationConfig() string {
	return `
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}
resource "aws_ec2_transit_gateway" "test" {}
resource "aws_networkmanager_transit_gateway_registration" "test" {
 global_network_id   = aws_networkmanager_global_network.test.id
 transit_gateway_arn = aws_ec2_transit_gateway.test.arn
}
`
}
func testAccNetworkManagerTransitGatewayRegistrationConfig_Update() string {
	return `
resource "aws_networkmanager_global_network" "test" {
 description = "test"
}
resource "aws_networkmanager_global_network" "test2" {
 description = "test2"
}
resource "aws_ec2_transit_gateway" "test" {}
resource "aws_networkmanager_transit_gateway_registration" "test" {
 global_network_id   = aws_networkmanager_global_network.test2.id
 transit_gateway_arn = aws_ec2_transit_gateway.test.arn
}
`
}
func testAccAWSNetworkManagerTransitGatewayRegistrationImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}
		return fmt.Sprintf("%s,%s", rs.Primary.Attributes["transit_gateway_arn"], rs.Primary.Attributes["global_network_id"]), nil
	}
}
