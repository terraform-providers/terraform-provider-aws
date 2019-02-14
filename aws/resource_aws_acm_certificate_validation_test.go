package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSAcmCertificateValidation_basic(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that validation succeeds
			{
				Config: testAccAcmCertificateValidation_basic(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateValidation_timeout(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAcmCertificateValidation_timeout(domain),
				ExpectError: regexp.MustCompile("Expected certificate to be issued but was in state PENDING_VALIDATION"),
			},
		},
	})
}

func TestAccAWSAcmCertificateValidation_validationRecordFqdns(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			// Test that validation fails if given validation_fqdns don't match
			{
				Config:      testAccAcmCertificateValidation_validationRecordFqdnsWrongFqdn(domain),
				ExpectError: regexp.MustCompile("missing .+ DNS validation record: .+"),
			},
			// Test that validation fails if not DNS validation
			{
				Config:      testAccAcmCertificateValidation_validationRecordFqdnsEmailValidation(domain),
				ExpectError: regexp.MustCompile("validation_record_fqdns is only valid for DNS validation"),
			},
			// Test that validation succeeds with validation
			{
				Config: testAccAcmCertificateValidation_validationRecordFqdnsOneRoute53Record(rootDomain, domain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateValidation_validationRecordFqdnsRoot(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateValidation_validationRecordFqdnsOneRoute53Record(rootDomain, rootDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateValidation_validationRecordFqdnsRootAndWildcard(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateValidation_validationRecordFqdnsTwoRoute53Records(rootDomain, rootDomain, strconv.Quote(wildcardDomain)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateValidation_validationRecordFqdnsSan(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	rInt1 := acctest.RandInt()

	domain := fmt.Sprintf("tf-acc-%d.%s", rInt1, rootDomain)
	sanDomain := fmt.Sprintf("tf-acc-%d-san.%s", rInt1, rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateValidation_validationRecordFqdnsTwoRoute53Records(rootDomain, domain, strconv.Quote(sanDomain)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateValidation_validationRecordFqdnsWildcard(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateValidation_validationRecordFqdnsOneRoute53Record(rootDomain, wildcardDomain),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
		},
	})
}

func TestAccAWSAcmCertificateValidation_validationRecordFqdnsWildcardAndRoot(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateValidation_validationRecordFqdnsTwoRoute53Records(rootDomain, wildcardDomain, strconv.Quote(rootDomain)),
				Check: resource.ComposeTestCheckFunc(
					resource.TestMatchResourceAttr("aws_acm_certificate_validation.cert", "certificate_arn", certificateArnRegex),
				),
			},
		},
	})
}

func testAccAcmCertificateValidation_basic(rootZoneDomain, domainName string) string {
	return fmt.Sprintf(`
%s

data "aws_route53_zone" "zone" {
  name = "%s."
  private_zone = false
}

resource "aws_route53_record" "cert_validation" {
  name = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl = 60
}

resource "aws_acm_certificate_validation" "cert" {
  depends_on = ["aws_route53_record.cert_validation"]

  certificate_arn = "${aws_acm_certificate.cert.arn}"
}
`, testAccAcmCertificateConfig(domainName, acm.ValidationMethodDns), rootZoneDomain)
}

func testAccAcmCertificateValidation_timeout(domainName string) string {
	return fmt.Sprintf(`
%s

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  timeouts {
    create = "5s"
  }
}
`, testAccAcmCertificateConfig(domainName, acm.ValidationMethodDns))
}

func testAccAcmCertificateValidation_validationRecordFqdnsEmailValidation(domainName string) string {
	return fmt.Sprintf(`
%s

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = ["wrong-validation-fqdn.example.com"]
}
`, testAccAcmCertificateConfig(domainName, acm.ValidationMethodEmail))
}

func testAccAcmCertificateValidation_validationRecordFqdnsOneRoute53Record(rootZoneDomain, domainName string) string {
	return fmt.Sprintf(`
%s

data "aws_route53_zone" "zone" {
  name = "%s."
  private_zone = false
}

resource "aws_route53_record" "cert_validation" {
  name = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = ["${aws_route53_record.cert_validation.fqdn}"]
}
`, testAccAcmCertificateConfig(domainName, acm.ValidationMethodDns), rootZoneDomain)
}

func testAccAcmCertificateValidation_validationRecordFqdnsTwoRoute53Records(rootZoneDomain, domainName, subjectAlternativeNames string) string {
	return fmt.Sprintf(`
%s

data "aws_route53_zone" "zone" {
  name = "%s."
  private_zone = false
}

resource "aws_route53_record" "cert_validation" {
  name = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.0.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.0.resource_record_value}"]
  ttl = 60
}

resource "aws_route53_record" "cert_validation_san" {
  name = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_name}"
  type = "${aws_acm_certificate.cert.domain_validation_options.1.resource_record_type}"
  zone_id = "${data.aws_route53_zone.zone.id}"
  records = ["${aws_acm_certificate.cert.domain_validation_options.1.resource_record_value}"]
  ttl = 60
}

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = [
    "${aws_route53_record.cert_validation.fqdn}",
    "${aws_route53_record.cert_validation_san.fqdn}"
  ]
}
`, testAccAcmCertificateConfig_subjectAlternativeNames(domainName, subjectAlternativeNames, acm.ValidationMethodDns), rootZoneDomain)
}

func testAccAcmCertificateValidation_validationRecordFqdnsWrongFqdn(domainName string) string {
	return fmt.Sprintf(`
%s

resource "aws_acm_certificate_validation" "cert" {
  certificate_arn = "${aws_acm_certificate.cert.arn}"
  validation_record_fqdns = ["wrong-validation-fqdn.example.com"]
}
`, testAccAcmCertificateConfig(domainName, acm.ValidationMethodDns))
}
