package aws

import (
	"fmt"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/worklink"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSWorkLinkWorkLinkWebsiteCertificateAuthorityAssociation_basic(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_website_certificate_authority_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkWebsiteCertificateAuthorityAssociationConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(
						resourceName, "fleet_arn",
						"aws_worklink_fleet.test", "arn"),
					resource.TestMatchResourceAttr(resourceName, "certificate", regexp.MustCompile("^-----BEGIN CERTIFICATE-----")),
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

func TestAccAWSWorkLinkWorkLinkWebsiteCertificateAuthorityAssociation_DisplayName(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_website_certificate_authority_association.test"
	displayName1 := fmt.Sprintf("tf-website-certificate-%s", acctest.RandStringFromCharSet(5, acctest.CharSetAlpha))
	displayName2 := fmt.Sprintf("tf-website-certificate-%s", acctest.RandStringFromCharSet(5, acctest.CharSetAlpha))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkWebsiteCertificateAuthorityAssociationConfigDisplayName(suffix, displayName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName1),
				),
			},
			{
				Config: testAccAWSWorkLinkWebsiteCertificateAuthorityAssociationConfigDisplayName(suffix, displayName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "display_name", displayName2),
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

func TestAccAWSWorkLinkWorkLinkWebsiteCertificateAuthorityAssociation_Disappears(t *testing.T) {
	suffix := acctest.RandStringFromCharSet(20, acctest.CharSetAlpha)
	resourceName := "aws_worklink_website_certificate_authority_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWorkLink(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSWorkLinkWebsiteCertificateAuthorityAssociationConfig(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationExists(resourceName),
					testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).worklinkconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_worklink_website_certificate_authority_association" {
			continue
		}

		_, err := conn.DescribeWebsiteCertificateAuthority(&worklink.DescribeWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(rs.Primary.Attributes["fleet_arn"]),
			WebsiteCaId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, worklink.ErrCodeResourceNotFoundException, "") {
				return nil
			}

			return err
		}
		return fmt.Errorf("Worklink Website Certificate Authority Association(%s) still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationDisappears(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No resource ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).worklinkconn
		fleetArn, websiteCaID, err := decodeWorkLinkWebsiteCertificateAuthorityAssociationResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		input := &worklink.DisassociateWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(fleetArn),
			WebsiteCaId: aws.String(websiteCaID),
		}

		if _, err := conn.DisassociateWebsiteCertificateAuthority(input); err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending:    []string{"DELETING"},
			Target:     []string{"DELETED"},
			Refresh:    worklinkWebsiteCertificateAuthorityAssociationStateRefresh(conn, websiteCaID, fleetArn),
			Timeout:    15 * time.Minute,
			Delay:      10 * time.Second,
			MinTimeout: 3 * time.Second,
		}

		_, err = stateConf.WaitForState()

		return err
	}

}

func testAccCheckAWSWorkLinkWebsiteCertificateAuthorityAssociationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Worklink Website Certificate Authority Association ID is set")
		}

		if _, ok := rs.Primary.Attributes["fleet_arn"]; !ok {
			return fmt.Errorf("WorkLink Fleet ARN is missing, should be set.")
		}

		conn := testAccProvider.Meta().(*AWSClient).worklinkconn
		fleetArn, websiteCaID, err := decodeWorkLinkWebsiteCertificateAuthorityAssociationResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		_, err = conn.DescribeWebsiteCertificateAuthority(&worklink.DescribeWebsiteCertificateAuthorityInput{
			FleetArn:    aws.String(fleetArn),
			WebsiteCaId: aws.String(websiteCaID),
		})

		return err
	}
}

func testAccAWSWorkLinkWebsiteCertificateAuthorityAssociationConfig(r string) string {
	return fmt.Sprintf(`
%s

resource "aws_worklink_website_certificate_authority_association" "test" {
  fleet_arn   = aws_worklink_fleet.test.arn
  certificate = file("test-fixtures/worklink-website-certificate-authority-association.pem")
}
`, testAccAWSWorkLinkFleetConfig(r))
}

func testAccAWSWorkLinkWebsiteCertificateAuthorityAssociationConfigDisplayName(r, displayName string) string {
	return fmt.Sprintf(`
%s

resource "aws_worklink_website_certificate_authority_association" "test" {
  fleet_arn    = aws_worklink_fleet.test.arn
  certificate  = file("test-fixtures/worklink-website-certificate-authority-association.pem")
  display_name = "%s"
}
`, testAccAWSWorkLinkFleetConfig(r), displayName)
}
