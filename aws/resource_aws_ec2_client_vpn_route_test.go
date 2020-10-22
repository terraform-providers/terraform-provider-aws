package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfec2 "github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
)

func testAccAwsEc2ClientVpnRoute_basic(t *testing.T) {
	var v ec2.ClientVpnRoute
	rStr := acctest.RandString(5)

	resourceName := "aws_ec2_client_vpn_route.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnRouteExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_vpc_subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", "0.0.0.0/0"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", "add-route"),
					resource.TestCheckResourceAttr(resourceName, "type", "Nat"),
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

func testAccAwsEc2ClientVpnRoute_description(t *testing.T) {
	var v ec2.ClientVpnRoute
	rStr := acctest.RandString(5)

	resourceName := "aws_ec2_client_vpn_route.test"
	endpointResourceName := "aws_ec2_client_vpn_endpoint.test"
	subnetResourceName := "aws_subnet.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfigDescription(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnRouteExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "client_vpn_endpoint_id", endpointResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "target_vpc_subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "description", "test client VPN route"),
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

func testAccAwsEc2ClientVpnRoute_disappears(t *testing.T) {
	var v ec2.ClientVpnRoute
	rStr := acctest.RandString(5)

	resourceName := "aws_ec2_client_vpn_route.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheckClientVPNSyncronize(t); testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2ClientVpnRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2ClientVpnRouteConfigBasic(rStr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2ClientVpnRouteExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2ClientVpnRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsEc2ClientVpnRouteDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_client_vpn_route" {
			continue
		}

		_, err := finder.ClientVpnRouteByID(conn, rs.Primary.ID)
		if err == nil {
			return fmt.Errorf("Client VPN route (%s) still exists", rs.Primary.ID)
		}
		if isAWSErr(err, tfec2.ErrCodeClientVpnRouteNotFound, "") {
			continue
		}
	}

	return nil
}

func testAccCheckAwsEc2ClientVpnRouteExists(name string, route *ec2.ClientVpnRoute) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := finder.ClientVpnRouteByID(conn, rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("Error reading Client VPN route (%s): %w", rs.Primary.ID, err)
		}

		if resp != nil || len(resp.Routes) == 1 || resp.Routes[0] != nil {
			*route = *resp.Routes[0]
			return nil
		}

		return fmt.Errorf("Client VPN route (%s) not found", rs.Primary.ID)
	}
}

func testAccEc2ClientVpnRouteConfigBasic(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnRouteVpcBase(rName, 1),
		testAccEc2ClientVpnRouteAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_subnet.test[0].id

  depends_on = [
    aws_ec2_client_vpn_network_association.test,
  ]
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName))
}

func testAccEc2ClientVpnRouteConfigDescription(rName string) string {
	return composeConfig(
		testAccEc2ClientVpnRouteVpcBase(rName, 1),
		testAccEc2ClientVpnRouteAcmCertificateBase(),
		fmt.Sprintf(`
resource "aws_ec2_client_vpn_route" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  destination_cidr_block = "0.0.0.0/0"
  target_vpc_subnet_id   = aws_subnet.test[0].id
  description            = "test client VPN route"

  depends_on = [
    aws_ec2_client_vpn_network_association.test,
  ]
}

resource "aws_ec2_client_vpn_network_association" "test" {
  client_vpn_endpoint_id = aws_ec2_client_vpn_endpoint.test.id
  subnet_id              = aws_subnet.test[0].id
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-testacc-clientvpn-%s"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled = false
  }
}
`, rName))
}

func testAccEc2ClientVpnRouteVpcBase(rName string, subnetCount int) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  # InvalidParameterValue: AZ us-west-2d is not currently supported. Please choose another az in this region
  exclude_zone_ids = ["usw2-az4"]
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-subnet-%[1]s"
  }
}

resource "aws_subnet" "test" {
  count                   = %[2]d
  availability_zone       = data.aws_availability_zones.available.names[count.index]
  cidr_block              = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id                  = aws_vpc.test.id
  map_public_ip_on_launch = true

  tags = {
    Name = "tf-acc-subnet-%[1]s"
  }
}
`, rName, subnetCount)
}

func testAccEc2ClientVpnRouteAcmCertificateBase() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}
