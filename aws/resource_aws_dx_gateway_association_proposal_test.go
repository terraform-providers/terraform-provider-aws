package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/directconnect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_dx_gateway_association_proposal", &resource.Sweeper{
		Name: "aws_dx_gateway_association_proposal",
		F:    testSweepDirectConnectGatewayAssociationProposals,
	})
}

func testSweepDirectConnectGatewayAssociationProposals(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).dxconn
	input := &directconnect.DescribeDirectConnectGatewayAssociationProposalsInput{}

	for {
		output, err := conn.DescribeDirectConnectGatewayAssociationProposals(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Direct Connect Gateway sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving Direct Connect Gateway Association Proposals: %s", err)
		}

		for _, gatewayAssociationProposal := range output.DirectConnectGatewayAssociationProposals {
			proposalID := aws.StringValue(gatewayAssociationProposal.ProposalId)

			if aws.StringValue(gatewayAssociationProposal.AssociatedGateway.Region) != region {
				log.Printf("[INFO] Skipping Direct Connect Gateway Association Proposal (%s) in different home region: %s", proposalID, aws.StringValue(gatewayAssociationProposal.AssociatedGateway.Region))
				continue
			}

			if aws.StringValue(gatewayAssociationProposal.ProposalState) != directconnect.GatewayAssociationProposalStateAccepted {
				log.Printf("[INFO] Skipping Direct Connect Gateway Association Proposal (%s) in non-accepted (%s) state", proposalID, aws.StringValue(gatewayAssociationProposal.ProposalState))
				continue
			}

			input := &directconnect.DeleteDirectConnectGatewayAssociationProposalInput{
				ProposalId: gatewayAssociationProposal.ProposalId,
			}

			log.Printf("[INFO] Deleting Direct Connect Gateway Association Proposal: %s", proposalID)
			_, err := conn.DeleteDirectConnectGatewayAssociationProposal(input)

			if err != nil {
				return fmt.Errorf("error deleting Direct Connect Gateway Association Proposal (%s): %s", proposalID, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAwsDxGatewayAssociationProposal_basicVpnGateway(t *testing.T) {
	var proposal1 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameVgw := "aws_vpn_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal1),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameVgw, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "virtualPrivateGateway"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_basicTransitGateway(t *testing.T) {
	var proposal1 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"
	resourceNameDxGw := "aws_dx_gateway.test"
	resourceNameTgw := "aws_ec2_transit_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_basicTransitGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal1),
					resource.TestCheckResourceAttrPair(resourceName, "dx_gateway_id", resourceNameDxGw, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "associated_gateway_id", resourceNameTgw, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "associated_gateway_owner_account_id"),
					resource.TestCheckResourceAttr(resourceName, "associated_gateway_type", "transitGateway"),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.0/30"),
					resource.TestCheckTypeSetElemAttr(resourceName, "allowed_prefixes.*", "10.255.255.8/30"),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_disappears(t *testing.T) {
	var proposal1 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal1),
					testAccCheckAwsDxGatewayAssociationProposalDisappears(&proposal1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsDxGatewayAssociationProposal_AllowedPrefixes(t *testing.T) {
	var proposal1, proposal2 directconnect.GatewayAssociationProposal
	var providers []*schema.Provider
	rBgpAsn := acctest.RandIntRange(64512, 65534)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_dx_gateway_association_proposal.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAwsDxGatewayAssociationProposalDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDxGatewayAssociationProposalConfigAllowedPrefixes1(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal1),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "1"),
				),
			},
			{
				Config:            testAccDxGatewayAssociationProposalConfigAllowedPrefixes1(rName, rBgpAsn),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccDxGatewayAssociationProposalConfigAllowedPrefixes2(rName, rBgpAsn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsDxGatewayAssociationProposalExists(resourceName, &proposal2),
					testAccCheckAwsDxGatewayAssociationProposalRecreated(&proposal1, &proposal2),
					resource.TestCheckResourceAttr(resourceName, "allowed_prefixes.#", "2"),
				),
			},
		},
	})
}

func testAccCheckAwsDxGatewayAssociationProposalDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).dxconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dx_gateway_association_proposal" {
			continue
		}

		proposal, err := describeDirectConnectGatewayAssociationProposal(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if proposal == nil {
			continue
		}

		return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsDxGatewayAssociationProposalExists(resourceName string, gatewayAssociationProposal *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).dxconn

		proposal, err := describeDirectConnectGatewayAssociationProposal(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if proposal == nil {
			return fmt.Errorf("Direct Connect Gateway Association Proposal (%s) not found", rs.Primary.ID)
		}

		*gatewayAssociationProposal = *proposal

		return nil
	}
}

func testAccCheckAwsDxGatewayAssociationProposalDisappears(proposal *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).dxconn

		input := &directconnect.DeleteDirectConnectGatewayAssociationProposalInput{
			ProposalId: proposal.ProposalId,
		}

		_, err := conn.DeleteDirectConnectGatewayAssociationProposal(input)

		return err
	}
}

func testAccCheckAwsDxGatewayAssociationProposalRecreated(i, j *directconnect.GatewayAssociationProposal) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if aws.StringValue(i.ProposalId) == aws.StringValue(j.ProposalId) {
			return fmt.Errorf("Direct Connect Gateway Association Proposal not recreated")
		}

		return nil
	}
}

func testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName string, rBgpAsn int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationProposalConfig_basicVpnGateway(rName string, rBgpAsn int) string {
	return testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn) + `
resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`
}

func testAccDxGatewayAssociationProposalConfig_basicTransitGateway(rName string, rBgpAsn int) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_dx_gateway" "test" {
  provider = "awsalternate"

  amazon_side_asn = %[2]d
  name            = %[1]q
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_dx_gateway_association_proposal" "test" {
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_ec2_transit_gateway.test.id

  allowed_prefixes = [
    "10.255.255.0/30",
    "10.255.255.8/30",
  ]
}
`, rName, rBgpAsn)
}

func testAccDxGatewayAssociationProposalConfigAllowedPrefixes1(rName string, rBgpAsn int) string {
	return testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn) + `
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/16"]
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`
}

func testAccDxGatewayAssociationProposalConfigAllowedPrefixes2(rName string, rBgpAsn int) string {
	return testAccDxGatewayAssociationProposalConfigBase_vpnGateway(rName, rBgpAsn) + `
resource "aws_dx_gateway_association_proposal" "test" {
  allowed_prefixes            = ["10.0.0.0/24", "10.0.1.0/24"]
  dx_gateway_id               = aws_dx_gateway.test.id
  dx_gateway_owner_account_id = aws_dx_gateway.test.owner_account_id
  associated_gateway_id       = aws_vpn_gateway.test.id
}
`
}
