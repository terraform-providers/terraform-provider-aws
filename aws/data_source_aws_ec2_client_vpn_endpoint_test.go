package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEc2ClientVpnEndpointDataSource_basic(t *testing.T) {
	datasourceName := "data.aws_ec2_client_vpn_endpoint.test"
	resourceName := "aws_ec2_client_vpn_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsEc2ClientVpnEndpointDataSourceConfig_nonExistent,
				ExpectError: regexp.MustCompile(`Endpoint non-existent-endpoint-id does not exist`),
			},
			{
				Config: testAccAwsEc2ClientVpnEndpointDataSourceConfig_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "authentication_options", resourceName, "authentication_options"),
					resource.TestCheckResourceAttrPair(datasourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(datasourceName, "client_cidr_block", resourceName, "client_cidr_block"),
					resource.TestCheckResourceAttrPair(datasourceName, "connection_log_options", resourceName, "connection_log_options"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_servers", resourceName, "dns_servers"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "server_certificate_arn", resourceName, "server_certificate_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "split_tunnel", resourceName, "split_tunnel"),
					resource.TestCheckResourceAttrPair(datasourceName, "transport_protocol", resourceName, "transport_protocol"),
				),
			},
		},
	})
}

const testAccAwsEc2ClientVpnEndpointDataSourceConfig_nonExistent = `
data "aws_ec2_client_vpn_endpoint" "test" {
  id = "non-existent-endpoint-id"
}
`

func testAccAwsEc2ClientVpnEndpointDataSourceConfig_basic() string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, "example.com")

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}

resource "aws_ec2_client_vpn_endpoint" "test" {
  description            = "terraform-clientvpn-example"
  server_certificate_arn = aws_acm_certificate.test.arn
  client_cidr_block      = "10.0.0.0/16"
  dns_servers            = ["8.8.8.8", "8.8.1.1"]
  split_tunnel           = true

  authentication_options {
    type                       = "certificate-authentication"
    root_certificate_chain_arn = aws_acm_certificate.test.arn
  }

  connection_log_options {
    enabled               = false
  }
}

data "aws_ec2_client_vpn_endpoint" "test" {
  id = aws_ec2_client_vpn_endpoint.test.id
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}
