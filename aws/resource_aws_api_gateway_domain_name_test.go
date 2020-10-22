package aws

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSAPIGatewayDomainName_CertificateArn(t *testing.T) {
	certificateArn := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_ARN")
	if certificateArn == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_ARN is not set. " +
				"This environment variable must be set to the ARN of " +
				"an ISSUED ACM certificate in us-east-1 to enable this test.")
	}

	// This test must always run in us-east-1
	// BadRequestException: Invalid certificate ARN: arn:aws:acm:us-west-2:123456789012:certificate/xxxxx. Certificate must be in 'us-east-1'.
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_CertificateArn(rName, certificateArn),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDomainName_CertificateName(t *testing.T) {
	certificateBody := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY")
	if certificateBody == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_BODY is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a publicly trusted certificate body to enable the test.")
	}

	certificateChain := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN")
	if certificateChain == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_CHAIN is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a chain certificate acceptable for the certificate to enable the test.")
	}

	certificatePrivateKey := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY")
	if certificatePrivateKey == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_CERTIFICATE_PRIVATE_KEY is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a private key of a publicly trusted certificate to enable the test.")
	}

	domainName := os.Getenv("AWS_API_GATEWAY_DOMAIN_NAME_DOMAIN_NAME")
	if domainName == "" {
		t.Skip(
			"Environment variable AWS_API_GATEWAY_DOMAIN_NAME_DOMAIN_NAME is not set. " +
				"This environment variable must be set to any non-empty value " +
				"with a domain name acceptable for the certificate to enable the test.")
	}

	var conf apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_CertificateName(domainName, certificatePrivateKey, certificateBody, certificateChain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &conf),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestCheckResourceAttrSet(resourceName, "cloudfront_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"certificate_body", "certificate_chain", "certificate_private_key"},
			},
		},
	})
}

func TestAccAWSAPIGatewayDomainName_RegionalCertificateArn(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateArn(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
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

	caKey := tlsRsaPrivateKeyPem(2048)
	caCertificate := tlsRsaX509SelfSignedCaCertificatePem(caKey)
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509LocallySignedCertificatePem(caKey, caCertificate, key, "*.terraformtest.com")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateName(rName, key, certificate, caCertificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_body", certificate),
					resource.TestCheckResourceAttr(resourceName, "certificate_chain", caCertificate),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestCheckResourceAttr(resourceName, "certificate_private_key", key),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					resource.TestMatchResourceAttr(resourceName, "regional_domain_name", regexp.MustCompile(`.*\.execute-api\..*`)),
					resource.TestMatchResourceAttr(resourceName, "regional_zone_id", regexp.MustCompile(`^Z`)),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayDomainName_SecurityPolicy(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_SecurityPolicy(rName, key, certificate, apigateway.SecurityPolicyTls12),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttr(resourceName, "security_policy", apigateway.SecurityPolicyTls12),
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

func TestAccAWSAPIGatewayDomainName_Tags(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfigTags1(rName, key, certificate, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDomainNameConfigTags2(rName, key, certificate, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSAPIGatewayDomainNameConfigTags1(rName, key, certificate, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
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

func TestAccAWSAPIGatewayDomainName_disappears(t *testing.T) {
	var domainName apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	rName := fmt.Sprintf("tf-acc-%s.terraformtest.com", acctest.RandString(8))

	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, rName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateArn(rName, key, certificate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayDomainName(), resourceName),
				),
				ExpectNonEmptyPlan: true,
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

		conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

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
	conn := testAccProvider.Meta().(*AWSClient).apigatewayconn

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

func testAccAWSAPIGatewayDomainNameConfig_CertificateName(domainName, key, certificate, chainCertificate string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  domain_name             = "%[1]s"
  certificate_body        = "%[2]s"
  certificate_chain       = "%[3]s"
  certificate_name        = "tf-acc-apigateway-domain-name"
  certificate_private_key = "%[4]s"
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(chainCertificate), tlsPemEscapeNewlines(key))
}

func testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateArn(domainName, key, certificate string) string {
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
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAWSAPIGatewayDomainNameConfig_RegionalCertificateName(domainName, key, certificate, chainCertificate string) string {
	return fmt.Sprintf(`
resource "aws_api_gateway_domain_name" "test" {
  certificate_body          = "%[2]s"
  certificate_chain         = "%[3]s"
  certificate_private_key   = "%[4]s"
  domain_name               = "%[1]s"
  regional_certificate_name = "tf-acc-apigateway-domain-name"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(chainCertificate), tlsPemEscapeNewlines(key))
}

func testAccAWSAPIGatewayDomainNameConfig_SecurityPolicy(domainName, key, certificate, securityPolicy string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[2]s"
  private_key      = "%[3]s"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = %[1]q
  regional_certificate_arn = aws_acm_certificate.test.arn
  security_policy          = %[4]q

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key), securityPolicy)
}

func testAccAWSAPIGatewayDomainNameConfigTags1(domainName, key, certificate, tagKey1, tagValue1 string) string {
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

  tags = {
    %[4]q = %[5]q
  }
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key), tagKey1, tagValue1)
}

func testAccAWSAPIGatewayDomainNameConfigTags2(domainName, key, certificate, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
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

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, domainName, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key), tagKey1, tagValue1, tagKey2, tagValue2)
}
