package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewayv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_apigatewayv2_domain_name", &resource.Sweeper{
		Name: "aws_apigatewayv2_domain_name",
		F:    testSweepAPIGatewayV2DomainNames,
	})
}

func testSweepAPIGatewayV2DomainNames(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).apigatewayv2conn
	input := &apigatewayv2.GetDomainNamesInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.GetDomainNames(input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway v2 domain names sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving API Gateway v2 domain names: %s", err)
		}

		for _, domainName := range output.Items {
			log.Printf("[INFO] Deleting API Gateway v2 domain name: %s", aws.StringValue(domainName.DomainName))
			_, err := conn.DeleteDomainName(&apigatewayv2.DeleteDomainNameInput{
				DomainName: domainName.DomainName,
			})
			if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting API Gateway v2 domain name (%s): %s", aws.StringValue(domainName.DomainName), err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}
		input.NextToken = output.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSAPIGatewayV2DomainName_basic(t *testing.T) {
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	certResourceName := "aws_acm_certificate.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, domainName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSAPIGatewayV2DomainName_disappears(t *testing.T) {
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, domainName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsApiGatewayV2DomainName(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSAPIGatewayV2DomainName_Tags(t *testing.T) {
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	certResourceName := "aws_acm_certificate.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, domainName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfig_tags(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key, 1, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayV2DomainName_UpdateCertificate(t *testing.T) {
	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	certResourceName0 := "aws_acm_certificate.test.0"
	certResourceName1 := "aws_acm_certificate.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	key := tlsRsaPrivateKeyPem(2048)
	domainName := fmt.Sprintf("%s.example.com", rName)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, domainName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName0, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key, 2, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName1, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfig_tags(rName, certificate, key, 2, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domainName),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", certResourceName0, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
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

func TestAccAWSAPIGatewayV2DomainName_MutualTlsAuthentication(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	var v apigatewayv2.GetDomainNameOutput
	resourceName := "aws_apigatewayv2_domain_name.test"
	acmCertificateResourceName := "aws_acm_certificate.test"
	s3BucketObjectResourceName := "aws_s3_bucket_object.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSAPIGatewayV2DomainNameDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfigMututalTlsAuthentication(rootDomain, domain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s.1", rName, rName)),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_version", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfigMututalTlsAuthenticationUpdated(rootDomain, domain, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.0.truststore_uri", fmt.Sprintf("s3://%s/%s.2", rName, rName)),
					resource.TestCheckResourceAttrPair(resourceName, "mutual_tls_authentication.0.truststore_version", s3BucketObjectResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test disabling mutual TLS authentication.
			{
				Config: testAccAWSAPIGatewayV2DomainNameConfigMututalTlsAuthenticationMissing(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSAPIGatewayV2DomainNameExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARNNoAccount(resourceName, "arn", "apigateway", regexp.MustCompile(`/domainnames/.+`)),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name", acmCertificateResourceName, "domain_name"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "domain_name_configuration.0.certificate_arn", acmCertificateResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.endpoint_type", "REGIONAL"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.hosted_zone_id"),
					resource.TestCheckResourceAttr(resourceName, "domain_name_configuration.0.security_policy", "TLS_1_2"),
					resource.TestCheckResourceAttrSet(resourceName, "domain_name_configuration.0.target_domain_name"),
					resource.TestCheckResourceAttr(resourceName, "mutual_tls_authentication.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSAPIGatewayV2DomainNameDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_apigatewayv2_domain_name" {
			continue
		}

		_, err := conn.GetDomainName(&apigatewayv2.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		})
		if isAWSErr(err, apigatewayv2.ErrCodeNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("API Gateway v2 domain name %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSAPIGatewayV2DomainNameExists(n string, v *apigatewayv2.GetDomainNameOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No API Gateway v2 domain name ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).apigatewayv2conn

		resp, err := conn.GetDomainName(&apigatewayv2.GetDomainNameInput{
			DomainName: aws.String(rs.Primary.ID),
		})
		if err != nil {
			return err
		}

		*v = *resp

		return nil
	}
}

func testAccAWSAPIGatewayV2DomainNameConfigImportedCerts(rName, certificate, key string, count int) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  count = %[4]d

  certificate_body = %[2]q
  private_key      = %[3]q

  tags = {
    Name = %[1]q
  }
}
`, rName, certificate, key, count)
}

func testAccAWSAPIGatewayV2DomainNameConfigPublicCert(rootDomain, domain string) string {
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

func testAccAWSAPIGatewayV2DomainNameConfig_basic(rName, certificate, key string, count, index int) string {
	return composeConfig(
		testAccAWSAPIGatewayV2DomainNameConfigImportedCerts(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test[%[2]d].arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
`, rName, index))
}

func testAccAWSAPIGatewayV2DomainNameConfig_tags(rName, certificate, key string, count, index int) string {
	return composeConfig(
		testAccAWSAPIGatewayV2DomainNameConfigImportedCerts(rName, certificate, key, count),
		fmt.Sprintf(`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = "%[1]s.example.com"

  domain_name_configuration {
    certificate_arn = aws_acm_certificate.test[%[2]d].arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
  }
}
`, rName, index))
}

func testAccAWSAPIGatewayV2DomainNameConfigMututalTlsAuthentication(rootDomain, domain, rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2DomainNameConfigPublicCert(rootDomain, domain),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q

  force_destroy = true
}

resource "aws_s3_bucket_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "%[1]s.1"
  source = "test-fixtures/apigateway-domain-name-truststore-1.pem"
}

resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_acm_certificate.test.domain_name

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.test.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  mutual_tls_authentication {
    truststore_uri = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2DomainNameConfigMututalTlsAuthenticationUpdated(rootDomain, domain, rName string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2DomainNameConfigPublicCert(rootDomain, domain),
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
  key    = "%[1]s.2"
  source = "test-fixtures/apigateway-domain-name-truststore-2.pem"
}

resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_acm_certificate.test.domain_name

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.test.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }

  mutual_tls_authentication {
    truststore_uri     = "s3://${aws_s3_bucket_object.test.bucket}/${aws_s3_bucket_object.test.key}"
    truststore_version = aws_s3_bucket_object.test.version_id
  }
}
`, rName))
}

func testAccAWSAPIGatewayV2DomainNameConfigMututalTlsAuthenticationMissing(rootDomain, domain string) string {
	return composeConfig(
		testAccAWSAPIGatewayV2DomainNameConfigPublicCert(rootDomain, domain),
		`
resource "aws_apigatewayv2_domain_name" "test" {
  domain_name = aws_acm_certificate.test.domain_name

  domain_name_configuration {
    certificate_arn = aws_acm_certificate_validation.test.certificate_arn
    endpoint_type   = "REGIONAL"
    security_policy = "TLS_1_2"
  }
}
`)
}
