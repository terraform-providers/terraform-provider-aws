package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_ec2_transit_gateway_peering_attachment", &resource.Sweeper{
		Name: "aws_ec2_transit_gateway_peering_attachment",
		F:    testSweepEc2TransitGatewayPeeringAttachments,
	})
}

func testSweepEc2TransitGatewayPeeringAttachments(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeTransitGatewayPeeringAttachmentsInput{}
	var sweeperErrs *multierror.Error

	err = conn.DescribeTransitGatewayPeeringAttachmentsPages(input,
		func(page *ec2.DescribeTransitGatewayPeeringAttachmentsOutput, lastPage bool) bool {
			for _, transitGatewayPeeringAttachment := range page.TransitGatewayPeeringAttachments {
				if aws.StringValue(transitGatewayPeeringAttachment.State) == ec2.TransitGatewayAttachmentStateDeleted {
					continue
				}

				id := aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId)

				input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
					TransitGatewayAttachmentId: aws.String(id),
				}

				log.Printf("[INFO] Deleting EC2 Transit Gateway Peering Attachment: %s", id)
				_, err := conn.DeleteTransitGatewayPeeringAttachment(input)

				if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
					continue
				}

				if err != nil {
					sweeperErr := fmt.Errorf("error deleting EC2 Transit Gateway Peering Attachment (%s): %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}

				if err := waitForEc2TransitGatewayPeeringAttachmentDeletion(conn, id); err != nil {
					sweeperErr := fmt.Errorf("error waiting for EC2 Transit Gateway Peering Attachment (%s) deletion: %w", id, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}
			return !lastPage
		})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Transit Gateway Peering Attachment sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("error retrieving EC2 Transit Gateway Peering Attachments: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSEc2TransitGatewayPeeringAttachment_basic(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					testAccCheckResourceAttrAccountID(resourceName, "peer_account_id"),
					resource.TestCheckResourceAttr(resourceName, "peer_region", testAccGetAlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayPeeringAttachment_disappears(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentDisappears(&transitGatewayPeeringAttachment),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayPeeringAttachment_Tags_sameAccount(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags1_sameAccount(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags1_sameAccount(rName, "key1", "value1"),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags2_sameAccount(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayPeeringAttachment_differentAccount(t *testing.T) {
	var transitGatewayPeeringAttachment ec2.TransitGatewayPeeringAttachment
	var providers []*schema.Provider
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_transit_gateway_peering_attachment.test"
	transitGatewayResourceName := "aws_ec2_transit_gateway.test"
	transitGatewayResourceNamePeer := "aws_ec2_transit_gateway.peer"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSEc2TransitGateway(t)
			testAccMultipleRegionPreCheck(t, 2)
			testAccAlternateAccountPreCheck(t)
		},
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		CheckDestroy:      testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_differentAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName, &transitGatewayPeeringAttachment),
					// Test that the peer account ID != the primary (request) account ID
					func(s *terraform.State) error {
						if testAccCheckResourceAttrAccountID(resourceName, "peer_account_id") == nil {
							return fmt.Errorf("peer_account_id attribute incorrectly to the requester's account ID")
						}
						return nil
					},
					resource.TestCheckResourceAttr(resourceName, "peer_region", testAccGetAlternateRegion()),
					resource.TestCheckResourceAttrPair(resourceName, "peer_transit_gateway_id", transitGatewayResourceNamePeer, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", transitGatewayResourceName, "id"),
				),
			},
			{
				Config:            testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_differentAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSEc2TransitGatewayPeeringAttachmentExists(resourceName string, transitGatewayPeeringAttachment *ec2.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No EC2 Transit Gateway Peering Attachment ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		attachment, err := ec2DescribeTransitGatewayPeeringAttachment(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if attachment == nil {
			return fmt.Errorf("EC2 Transit Gateway Peering Attachment not found")
		}

		if aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStateAvailable && aws.StringValue(attachment.State) != ec2.TransitGatewayAttachmentStatePendingAcceptance {
			return fmt.Errorf("EC2 Transit Gateway Peering Attachment (%s) exists in non-available/pending acceptance (%s) state", rs.Primary.ID, aws.StringValue(attachment.State))
		}

		*transitGatewayPeeringAttachment = *attachment

		return nil
	}
}

func testAccCheckAWSEc2TransitGatewayPeeringAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_transit_gateway_peering_attachment" {
			continue
		}

		peeringAttachment, err := ec2DescribeTransitGatewayPeeringAttachment(conn, rs.Primary.ID)

		if isAWSErr(err, "InvalidTransitGatewayAttachmentID.NotFound", "") {
			continue
		}

		if err != nil {
			return err
		}

		if peeringAttachment == nil {
			continue
		}

		if aws.StringValue(peeringAttachment.State) != ec2.TransitGatewayAttachmentStateDeleted {
			return fmt.Errorf("EC2 Transit Gateway Peering Attachment (%s) still exists in non-deleted (%s) state", rs.Primary.ID, aws.StringValue(peeringAttachment.State))
		}
	}

	return nil
}

func testAccCheckAWSEc2TransitGatewayPeeringAttachmentDisappears(transitGatewayPeeringAttachment *ec2.TransitGatewayPeeringAttachment) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := &ec2.DeleteTransitGatewayPeeringAttachmentInput{
			TransitGatewayAttachmentId: transitGatewayPeeringAttachment.TransitGatewayAttachmentId,
		}

		if _, err := conn.DeleteTransitGatewayPeeringAttachment(input); err != nil {
			return err
		}

		return waitForEc2TransitGatewayPeeringAttachmentDeletion(conn, aws.StringValue(transitGatewayPeeringAttachment.TransitGatewayAttachmentId))
	}
}

func testAccAWSEc2TransitGatewayPeeringAttachmentConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "peer" {
  provider = "awsalternate"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentConfig_sameAccount_base(rName string) string {
	return testAccAlternateRegionProviderConfig() + testAccAWSEc2TransitGatewayPeeringAttachmentConfig_base(rName)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentConfig_differentAccount_base(rName string) string {
	return testAccAlternateAccountAlternateRegionProviderConfig() + testAccAWSEc2TransitGatewayPeeringAttachmentConfig_base(rName)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_sameAccount(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentConfig_sameAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
}
`, testAccGetAlternateRegion())
}

func testAccAWSEc2TransitGatewayPeeringAttachmentConfigBasic_differentAccount(rName string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentConfig_differentAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_account_id         = aws_ec2_transit_gateway.peer.owner_id
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
}
`, testAccGetAlternateRegion())
}

func testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags1_sameAccount(rName, tagKey1, tagValue1 string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentConfig_sameAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[2]q

    %[3]s = %[4]q
  }
}
`, testAccGetAlternateRegion(), rName, tagKey1, tagValue1)
}

func testAccAWSEc2TransitGatewayPeeringAttachmentConfigTags2_sameAccount(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSEc2TransitGatewayPeeringAttachmentConfig_sameAccount_base(rName) + fmt.Sprintf(`
resource "aws_ec2_transit_gateway_peering_attachment" "test" {
  peer_region             = %[1]q
  peer_transit_gateway_id = aws_ec2_transit_gateway.peer.id
  transit_gateway_id      = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[2]q

    %[3]s = %[4]q
    %[5]s = %[6]q
  }
}
`, testAccGetAlternateRegion(), rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
