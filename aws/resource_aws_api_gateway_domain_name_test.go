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
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	var domainName apigateway.DomainName
	acmCertificateResourceName := "aws_acm_certificate.test"
	resourceName := "aws_api_gateway_domain_name.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccPreCheckApigatewayEdgeDomainName(t) },
		ProviderFactories: testAccProviderFactories,
		CheckDestroy:      testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_CertificateArn(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &domainName),
					testAccCheckResourceAttrRegionalARNApigatewayEdgeDomainName(resourceName, "arn", "apigateway", domain),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestMatchResourceAttr(resourceName, "cloudfront_domain_name", regexp.MustCompile(`[a-z0-9]+.cloudfront.net`)),
					resource.TestCheckResourceAttr(resourceName, "cloudfront_zone_id", "Z2FDTNDATAQYW2"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
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
					testAccCheckResourceAttrRegionalARNApigatewayRegionalDomainName(resourceName, "arn", "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					testAccMatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexp.MustCompile(`d-[a-z0-9]+`)),
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
					testAccCheckResourceAttrRegionalARNApigatewayRegionalDomainName(resourceName, "arn", "apigateway", rName),
					resource.TestCheckResourceAttr(resourceName, "certificate_body", certificate),
					resource.TestCheckResourceAttr(resourceName, "certificate_chain", caCertificate),
					resource.TestCheckResourceAttr(resourceName, "certificate_name", "tf-acc-apigateway-domain-name"),
					resource.TestCheckResourceAttr(resourceName, "certificate_private_key", key),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_upload_date"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rName),
					testAccMatchResourceAttrRegionalHostname(resourceName, "regional_domain_name", "execute-api", regexp.MustCompile(`d-[a-z0-9]+`)),
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

func TestAccAWSAPIGatewayDomainName_MutualTlsAuthentication(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	var v apigateway.DomainName
	resourceName := "aws_api_gateway_domain_name.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	s3BucketObjectResourceName := "aws_s3_bucket_object.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayDomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_MutualTlsAuthentication(rootDomain, domain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/+.`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3BucketObjectResourceName, "version_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test disabling mutual TLS authentication.
			{
				Config: testAccAWSAPIGatewayDomainNameConfig_MutualTlsAuthenticationMissing(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayDomainNameExists(resourceName, &v),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
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

func testAccAWSAPIGatewayDomainNameConfigPublicCert(rootDomain, domain string) string {
	return fmt.Sprintf(`
data "aws_route53_zone" "test" {
  name         = %[1]q
  private_zone = false
}

resource "aws_acm_certificate" "test" {
  domain_name       = %[2]q
  validation_method = "DNS"
}

#
# for_each acceptance testing requires:
# https://github.com/hashicorp/terraform-plugin-sdk/issues/536
#
# resource "aws_route53_record" "test" {
#   for_each = {
#     for dvo in aws_acm_certificate.test.domain_validation_options: dvo.domain_name => {
#       name   = dvo.resource_record_name
#       record = dvo.resource_record_value
#       type   = dvo.resource_record_type
#     }
#   }
#   allow_overwrite = true
#   name            = each.value.name
#   records         = [each.value.record]
#   ttl             = 60
#   type            = each.value.type
#   zone_id         = data.aws_route53_zone.test.zone_id
# }

resource "aws_route53_record" "test" {
  allow_overwrite = true
  name            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_name
  records         = [tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_value]
  ttl             = 60
  type            = tolist(aws_acm_certificate.test.domain_validation_options)[0].resource_record_type
  zone_id         = data.aws_route53_zone.test.zone_id
}

resource "aws_acm_certificate_validation" "test" {
  certificate_arn         = aws_acm_certificate.test.arn
  validation_record_fqdns = [aws_route53_record.test.fqdn]
}
`, rootDomain, domain)
}

func testAccAWSAPIGatewayDomainNameConfig_CertificateArn(rootDomain string, domain string) string {
	return composeConfig(
		testAccApigatewayEdgeDomainNameRegionProviderConfig(),
		testAccAWSAPIGatewayDomainNameConfigPublicCert(rootDomain, domain),
		`
resource "aws_api_gateway_domain_name" "test" {
  domain_name     = aws_acm_certificate.test.domain_name
  certificate_arn = aws_acm_certificate_validation.test.certificate_arn

  endpoint_configuration {
    types = ["EDGE"]
  }
}
`)
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

func testAccAWSAPIGatewayDomainNameConfig_MutualTlsAuthentication(rootDomain, domain, rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayDomainNameConfigPublicCert(rootDomain, domain),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true

  versioning {
    enabled = true
  }
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = %[1]q
  source = "test-fixtures/apigateway-domain-name-truststore-1.pem"
}

resource "aws_api_gateway_domain_name" "test" {
  domain_name              = aws_acm_certificate.test.domain_name
  regional_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }

  mutual_tls_authentication {
    truststore_uri     = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
    truststore_version = aws_s3_bucket_object.test.version_id
  }
}
`, rName))
}

func testAccAWSAPIGatewayDomainNameConfig_MutualTlsAuthenticationMissing(rootDomain, domain string) string {
	return composeConfig(
		testAccAWSAPIGatewayDomainNameConfigPublicCert(rootDomain, domain),
		`
resource "aws_api_gateway_domain_name" "test" {
  domain_name              = aws_acm_certificate.test.domain_name
  regional_certificate_arn = aws_acm_certificate_validation.test.certificate_arn
  security_policy          = "TLS_1_2"

  endpoint_configuration {
    types = ["REGIONAL"]
  }
}
`)
}
