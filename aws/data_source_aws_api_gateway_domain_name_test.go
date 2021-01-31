package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsApiGatewayDomainName_basic(t *testing.T) {
	resourceName := "aws_api_gateway_domain_name.test"
	dataSourceName := "data.aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayDomainNameConfig_RegionalCertificateArn(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", dataSourceName, "certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_name", dataSourceName, "certificate_name"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_upload_date", dataSourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_domain_name", dataSourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudfront_zone_id", dataSourceName, "cloudfront_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", dataSourceName, "domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint_configuration.#", dataSourceName, "endpoint_configuration.#"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_certificate_arn", dataSourceName, "regional_certificate_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_certificate_name", dataSourceName, "regional_certificate_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_domain_name", dataSourceName, "regional_domain_name"),
					resource.TestCheckResourceAttrPair(resourceName, "regional_zone_id", dataSourceName, "regional_zone_id"),
					resource.TestCheckResourceAttrPair(resourceName, "security_policy", dataSourceName, "security_policy"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayDomainNameConfig_RegionalCertificateArn(domainName, key, certificate string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}

data "aws_api_gateway_domain_name" "test" {
  domain_name = aws_api_gateway_domain_name.test.domain_name
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}
