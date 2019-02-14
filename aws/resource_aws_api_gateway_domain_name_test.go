package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayDomainName_CertificateArn(t *testing.T) {
	// This test must always run in us-east-1
	// BadRequestException: Invalid certificate ARN: arn:aws:acm:us-west-2:123456789012:certificate/xxxxx. Certificate must be in 'us-east-1'.
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	// For now, use an environment variable to limit running this test
	certificateArn := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"an ISSUED ACM certificate in us-east-1 to enable this test.")
	}

	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_CertificateArn(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDomainName_CertificateName(t *testing.T) {
	var conf apigateway.DomainName

	rString := acctest.RandString(8)
	name := fmt.Sprintf("tf-acc-%s.terraformtest.com", rString)
	nameModified := fmt.Sprintf("tf-acc-%s-modified.terraformtest.com", rString)
	commonName := "*.terraformtest.com"
	certRe := regexp.MustCompile("^-----BEGIN CERTIFICATE-----\n")
	keyRe := regexp.MustCompile("^-----BEGIN RSA PRIVATE KEY-----\n")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_CertificateName(name, commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists("aws_api_gateway_domain_name.test", &conf),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_body", certRe),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_chain", certRe),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_private_key", keyRe),
					resource.TestCheckResourceAttrSet("aws_api_gateway_domain_name.test", "cloudfront_domain_name"),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "domain_name", name),
					resource.TestCheckResourceAttrSet("aws_api_gateway_domain_name.test", "certificate_upload_date"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_CertificateName(nameModified, commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists("aws_api_gateway_domain_name.test", &conf),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_body", certRe),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_chain", certRe),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestMatchResourceAttr("aws_api_gateway_domain_name.test", "certificate_private_key", keyRe),
					resource.TestCheckResourceAttrSet("aws_api_gateway_domain_name.test", "cloudfront_domain_name"),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr("aws_api_gateway_domain_name.test", "domain_name", nameModified),
					resource.TestCheckResourceAttrSet("aws_api_gateway_domain_name.test", "certificate_upload_date"),
				),
			},
			{
				ResourceName:            "aws_api_gateway_domain_name.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_body", "certificate_chain", "certificate_private_key"},
			},
		},
	})
}

func TestAccAWSAPIGatewayDomainName_RegionalCertificateArn(t *testing.T) {
	// For now, use an environment variable to limit running this test
	regionalCertificateArn := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_ARN")
	if regionalCertificateArn == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"an ISSUED ACM certificate in the region where this test " +
				"is running to enable the test.")
	}

	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateArn(rName, regionalCertificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestMatchResourceAttr(resourceName, "regional_domain_name", regexp.MustCompile(`.*\.execute-api\..*`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexp.MustCompile(`^Z`)),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDomainName_RegionalCertificateName(t *testing.T) {
	// For now, use an environment variable to limit running this test
	// BadRequestException: Uploading certificates is not supported for REGIONAL.
	// See Remarks section of https://docs.aws.amazon.com/apigateway/api-reference/link-relation/domainname-create/
	// which suggests this configuration should be possible somewhere, e.g. AWS China?
	regionalCertificateArn := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED")
	if regionalCertificateArn == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_REGIONAL_CERTIFICATE_NAME_ENABLED is not set. " +
				"This environment variable must be set to any non-empty value " +
				"in a region where uploading REGIONAL certificates is allowed " +
				"to enable the test.")
	}

	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"

	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))
	commonName := "*.terraformtest.com"
	certRe := regexp.MustCompile("^-----BEGIN CERTIFICATE-----\n")
	keyRe := regexp.MustCompile("^-----BEGIN RSA PRIVATE KEY-----\n")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProvidersWithTLS,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateName(rName, commonName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					resource.TestMatchResourceAttr(resourceName, "certificate_body", certRe),
					resource.TestMatchResourceAttr(resourceName, "certificate_chain", certRe),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestMatchResourceAttr(resourceName, "certificate_private_key", keyRe),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestMatchResourceAttr(resourceName, "regional_domain_name", regexp.MustCompile(`.*\.execute-api\..*`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexp.MustCompile(`^Z`)),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayDomainNameExists(n string, res *apigateway.DomainName) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway DomainName ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		req := &apigateway.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		}
		describe, err := conn.GetDomainName(req)
		if err != nil {
			return err
		}

		if *describe.DomainName != rs.Primary.ID {
			return fmt.Errorf("APIGateway DomainName not found")
		}

		*res = *describe

		return nil
	}
}

func testAccCheckAWSAPIGatewayDomainNameDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_domain_name" {
			continue
		}

		_, err := conn.GetDomainName(&apigateway.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		})

		if err != nil {
			if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("API Gateway Domain Name still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccAWSAPIGatewayCerts(commonName string) string {
	return fmt.Sprintf(`
resource "tls_private_key" "test" {
  algorithm = "RSA"
}

resource "tls_self_signed_cert" "ca" {
  key_algorithm         = "RSA"
  private_key_pem       = "${tls_private_key.test.private_key_pem}"
  is_ca_certificate     = true
  validity_period_hours = 12

  subject {
    common_name  = "ACME Root CA"
    organization = "ACME Example Holdings"
  }

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}

resource "tls_cert_request" "test" {
  key_algorithm   = "RSA"
  private_key_pem = "${tls_private_key.test.private_key_pem}"

  subject {
    common_name  = "%s"
    organization = "ACME Example Holdings, Inc"
  }
}

resource "tls_locally_signed_cert" "leaf" {
  cert_request_pem      = "${tls_cert_request.test.cert_request_pem}"
  ca_key_algorithm      = "RSA"
  ca_private_key_pem    = "${tls_private_key.test.private_key_pem}"
  ca_cert_pem           = "${tls_self_signed_cert.ca.cert_pem}"
  validity_period_hours = 12

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
  ]
}
`, commonName)
}

func testAccAWSAPIGatewayDomainNameConfig_CertificateArn(domainName, certificateArn string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name     = "%s"
  certificate_arn = "%s"

  endpoint_configuration {
    types = ["EDGE"]
  }
}
`, domainName, certificateArn)
}

func testAccAWSAPIGatewayDomainNameConfig_CertificateName(domainName, commonName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name = "%s"
  certificate_body = "${tls_locally_signed_cert.leaf.cert_pem}"
  certificate_chain = "${tls_self_signed_cert.ca.cert_pem}"
  certificate_name = "tf-acc-apigateway-domain-name"
  certificate_private_key = "${tls_private_key.test.private_key_pem}"
}
%s
`, domainName, testAccAWSAPIGatewayCerts(commonName))
}

func testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateArn(domainName, regionalCertificateArn string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name              = "%s"
  regional_certificate_arn = "%s"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, regionalCertificateArn)
}

func testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateName(domainName, commonName string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  certificate_body          = "${tls_locally_signed_cert.leaf.cert_pem}"
  certificate_chain         = "${tls_self_signed_cert.ca.cert_pem}"
  certificate_private_key   = "${tls_private_key.test.private_key_pem}"
  domain_name               = "%s"
  regional_certificate_name = "tf-acc-apigateway-domain-name"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
%s
`, domainName, testAccAWSAPIGatewayCerts(commonName))
}
