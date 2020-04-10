package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAwsEc2ClientVpnNetworkAssociation_basic(t *testing.T) {
	var assoc1 ec2.TargetNetwork
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists("aws_ec2_client_vpn_network_association.test", &assoc1),
				),
			},
		},
	})
}

func TestAccAwsEc2ClientVpnNetworkAssociation_disappears(t *testing.T) {
	var assoc1 ec2.TargetNetwork
	rStr := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnNetworkAssociationConfig(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnNetworkAssociationExists("aws_ec2_client_vpn_network_association.test", &assoc1),
					testAccCheckAwsEc2ClientVpnNetworkAssociationDisappears(&assoc1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_network_association" {
			continue
		}

		resp, _ := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		for _, v := range resp.ClientVpnTargetNetworks {
			if *v.AssociationId == rs.Primary.ID && !(*v.Status.Code == "Disassociated") {
				return fmt.Errorf("[DESTROY ERROR] Client VPN network association (%s) not deleted", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationDisappears(targetNetwork *ec2.TargetNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DisassociateClientVpnTargetNetwork(&ec2.DisassociateClientVpnTargetNetworkInput{
			AssociationId:       targetNetwork.AssociationId,
			ClientVpnEndpointId: targetNetwork.ClientVpnEndpointId,
		})

		if err != nil {
			return err
		}

		stateConf := &resource.StateChangeConf{
			Pending: []string{ec2.AssociationStatusCodeDisassociating},
			Target:  []string{ec2.AssociationStatusCodeDisassociated},
			Refresh: clientVpnNetworkAssociationRefreshFunc(conn, aws.StringValue(targetNetwork.AssociationId), aws.StringValue(targetNetwork.ClientVpnEndpointId)),
			Timeout: 10 * time.Minute,
		}

		_, err = stateConf.WaitForState()

		return err
	}
}

func testAccCheckAwsEc2ClientVpnNetworkAssociationExists(name string, assoc *ec2.TargetNetwork) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeClientVpnTargetNetworks(&ec2.DescribeClientVpnTargetNetworksInput{
			ClientVpnEndpointId: aws.String(rs.Primary.Attributes["client_vpn_endpoint_id"]),
			AssociationIds:      []*string{aws.String(rs.Primary.ID)},
		})

		if err != nil {
			return fmt.Errorf("Error reading Client VPN network association (%s): %s", rs.Primary.ID, err)
		}

		for _, a := range resp.ClientVpnTargetNetworks {
			if *a.AssociationId == rs.Primary.ID && !(*a.Status.Code == "Disassociated") {
				*assoc = *a
				return nil
			}
		}

		return fmt.Errorf("Client VPN network association (%s) not found", rs.Primary.ID)
	}
}

func testAccEc2ClientVpnNetworkAssociationConfigAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccEc2ClientVpnNetworkAssociationConfig(rName string) string {
	return testAccEc2ClientVpnNetworkAssociationConfigAcmCertificateBase() + fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # InvalidParameterValue: AZ us-west-2d is not currently supported. Please choose another az in this region
  blacklisted_zone_ids = ["usw2-az4"]
  state                = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%s"
  }
}

resource "aws_subnet" "test" {
  availability_zone       = data.aws_availability_zones.available.names[0]
  cidr_block              = "10.1.1.0/24"
  vpc_id                  = "${aws_vpc.test.id}"
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%s"
  }
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = "${aws_acm_certificate.test.arn}"
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = "${aws_acm_certificate.test.arn}"
  }

  connection_log_options {
    enabled = false
  }
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = "${aws_ec2_client_vpn_endpoint.test.id}"
  subnet_id              = "${aws_subnet.test.id}"
}
`, rName, rName, rName)
}
