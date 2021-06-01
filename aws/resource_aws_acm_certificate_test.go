package aws

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acm"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_acm_certificate", &resource.Sweeper{
		Name: "aws_acm_certificate",
		F:    testSweepAcmCertificates,
	})
}

func testSweepAcmCertificates(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).acmconn
	var sweeperErrs *multierror.Error

	err = conn.ListCertificatesPages(&acm.ListCertificatesInput{}, func(page *acm.ListCertificatesOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, certificate := range page.CertificateSummaryList {
			arn := aws.StringValue(certificate.CertificateArn)

			output, err := conn.DescribeCertificate(&acm.DescribeCertificateInput{
				CertificateArn: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error describing ACM certificate (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			if len(output.Certificate.InUseBy) > 0 {
				log.Printf("[INFO] ACM certificate (%s) is in-use, skipping", arn)
				continue
			}

			log.Printf("[INFO] Deleting ACM certificate: %s", arn)
			_, err = conn.DeleteCertificate(&acm.DeleteCertificateInput{
				CertificateArn: aws.String(arn),
			})
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting ACM certificate (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})

	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ACM certificate sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}
		return fmt.Errorf("error retrieving ACM certificates: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func testAccAwsAcmCertificateDomainFromEnv(t *testing.T) string {
	rootDomain := os.Getenv("ACM_CERTIFICATE_ROOT_DOMAIN")

	if rootDomain == "" {
		t.Skip(
			"Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is not set. " +
				"For DNS validation requests, this domain must be publicly " +
				"accessible and configurable via Route53 during the testing. " +
				"For email validation requests, you must have access to one of " +
				"the five standard email addresses used (admin|administrator|" +
				"hostmaster|postmaster|webmaster)@domain or one of the WHOIS " +
				"contact addresses.")
	}

	if len(rootDomain) >= 56 {
		t.Skip(
			"Environment variable ACM_CERTIFICATE_ROOT_DOMAIN is too long. " +
				"The domain must be shorter than 56 characters to allow for " +
				"subdomain randomization in the testing.")
	}

	return rootDomain
}

// ACM domain names cannot be longer than 64 characters
// Other resources, e.g. Cognito User Pool Domains, limit this to 63
func testAccAwsAcmCertificateRandomSubDomain(rootDomain string) string {
	// Max length (63)
	// Subtract "tf-acc-" prefix (7)
	// Subtract "." between prefix and root domain (1)
	// Subtract length of root domain
	return fmt.Sprintf("tf-acc-%s.%s", acctest.RandString(55-len(rootDomain)), rootDomain)
}

func TestAccAWSAcmCertificate_emailValidation(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodEmail),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestMatchResourceAttr(resourceName, "validation_emails.0", regexp.MustCompile(`^[^@]+@.+$`)),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodEmail),
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

func TestAccAWSAcmCertificate_dnsValidation(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_root(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(rootDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_privateCert(t *testing.T) {
	certificateAuthorityResourceName := "aws_acmpca_certificate_authority.test"
	resourceName := "aws_acm_certificate.cert"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_privateCert(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", fmt.Sprintf("%s.terraformtesting.com", rName)),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusFailed), // FailureReason: PCA_INVALID_STATE (PCA State: PENDING_CERTIFICATE)
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", "NONE"),
					resource.TestCheckResourceAttrPair(resourceName, "certificate_authority_arn", certificateAuthorityResourceName, "arn"),
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

// TestAccAWSAcmCertificate_root_TrailingPeriod updated in 3.0 to account for domain_name plan-time validation
// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/13510
func TestAccAWSAcmCertificate_root_TrailingPeriod(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := fmt.Sprintf("%s.", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				ExpectError: regexp.MustCompile(`invalid value for domain_name \(cannot end with a period\)`),
			},
		},
	})
}

func TestAccAWSAcmCertificate_rootAndWildcardSan(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(rootDomain, strconv.Quote(wildcardDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_SubjectAlternativeNames_EmptyString(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAcmCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(""), acm.ValidationMethodDns),
				ExpectError: regexp.MustCompile(`expected length`),
			},
		},
	})
}

func TestAccAWSAcmCertificate_san_single(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)
	sanDomain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_san_multiple(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)
	sanDomain1 := testAccAwsAcmCertificateRandomSubDomain(rootDomain)
	sanDomain2 := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, fmt.Sprintf("%q, %q", sanDomain1, sanDomain2), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "3"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain1,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          sanDomain2,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain1),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", sanDomain2),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_san_TrailingPeriod(t *testing.T) {
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)
	sanDomain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)
	resourceName := "aws_acm_certificate.cert"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(domain, strconv.Quote(sanDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile(`certificate/.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", domain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          domain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          strings.TrimSuffix(sanDomain, "."),
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", strings.TrimSuffix(sanDomain, ".")),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_wildcard(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(wildcardDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_wildcardAndRootSan(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	wildcardDomain := fmt.Sprintf("*.%s", rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_subjectAlternativeNames(wildcardDomain, strconv.Quote(rootDomain), acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", wildcardDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          wildcardDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "subject_alternative_names.*", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
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

func TestAccAWSAcmCertificate_disableCTLogging(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig_disableCTLogging(rootDomain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm", regexp.MustCompile("certificate/.+$")),
					resource.TestCheckResourceAttr(resourceName, "domain_name", rootDomain),
					resource.TestCheckResourceAttr(resourceName, "domain_validation_options.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "domain_validation_options.*", map[string]string{
						"domain_name":          rootDomain,
						"resource_record_type": "CNAME",
					}),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusPendingValidation),
					resource.TestCheckResourceAttr(resourceName, "validation_emails.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "validation_method", acm.ValidationMethodDns),
					resource.TestCheckResourceAttr(resourceName, "options.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "options.0.certificate_transparency_logging_preference", acm.CertificateTransparencyLoggingPreferenceDisabled),
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

func TestAccAWSAcmCertificate_tags(t *testing.T) {
	resourceName := "aws_acm_certificate.cert"
	rootDomain := testAccAwsAcmCertificateDomainFromEnv(t)
	domain := testAccAwsAcmCertificateRandomSubDomain(rootDomain)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfig(domain, acm.ValidationMethodDns),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				Config: testAccAcmCertificateConfig_twoTags(domain, acm.ValidationMethodDns, "Hello", "World", "Foo", "Bar"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Hello", "World"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
				),
			},
			{
				Config: testAccAcmCertificateConfig_twoTags(domain, acm.ValidationMethodDns, "Hello", "World", "Foo", "Baz"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Hello", "World"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Baz"),
				),
			},
			{
				Config: testAccAcmCertificateConfig_oneTag(domain, acm.ValidationMethodDns, "Environment", "Test"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Test"),
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

//lintignore:AT002
func TestAccAWSAcmCertificate_imported_DomainName(t *testing.T) {
	resourceName := "aws_acm_certificate.test"

	commonName := "example.com"
	caKey := tlsRsaPrivateKeyPem(2048)
	caCertificate := tlsRsaX509SelfSignedCaCertificatePem(caKey)
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509LocallySignedCertificatePem(caKey, caCertificate, key, commonName)

	newCaKey := tlsRsaPrivateKeyPem(2048)
	newCaCertificate := tlsRsaX509SelfSignedCaCertificatePem(newCaKey)
	newCertificate := tlsRsaX509LocallySignedCertificatePem(newCaKey, newCaCertificate, key, commonName)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfigPrivateKey(certificate, key, caCertificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", commonName),
				),
			},
			{
				Config: testAccAcmCertificateConfigPrivateKey(newCertificate, key, newCaCertificate),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", commonName),
				),
			},
			{
				Config: testAccAcmCertificateConfigPrivateKeyWithoutChain("example2.com"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "domain_name", "example2.com"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body", "certificate_chain"},
			},
		},
	})
}

//lintignore:AT002
func TestAccAWSAcmCertificate_imported_IpAddress(t *testing.T) { // Reference: https://github.com/hashicorp/terraform-provider-aws/issues/7103
	resourceName := "aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfigPrivateKeyWithoutChain("1.2.3.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "domain_name", ""),
					resource.TestCheckResourceAttr(resourceName, "status", acm.CertificateStatusIssued),
					resource.TestCheckResourceAttr(resourceName, "subject_alternative_names.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				// These are not returned by the API
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body"},
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/15055
func TestAccAWSAcmCertificate_PrivateKey_Tags(t *testing.T) {
	resourceName := "aws_acm_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAcmCertificateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAcmCertificateConfigPrivateKeyTags("1.2.3.4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"private_key", "certificate_body"},
			},
			{
				Config: testAccAcmCertificateConfigPrivateKeyTags("5.6.7.8"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
		},
	})
}

func testAccAcmCertificateConfig(domainName, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"
}
`, domainName, validationMethod)

}

func testAccAcmCertificateConfig_privateCert(rName string) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acm_certificate" "cert" {
  domain_name               = "%s.terraformtesting.com"
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}
`, rName)
}

func testAccAcmCertificateConfig_subjectAlternativeNames(domainName, subjectAlternativeNames, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name               = "%s"
  subject_alternative_names = [%s]
  validation_method         = "%s"
}
`, domainName, subjectAlternativeNames, validationMethod)
}

func testAccAcmCertificateConfig_oneTag(domainName, validationMethod, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"

  tags = {
    "%s" = "%s"
  }
}
`, domainName, validationMethod, tag1Key, tag1Value)
}

func testAccAcmCertificateConfig_twoTags(domainName, validationMethod, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"

  tags = {
    "%s" = "%s"
    "%s" = "%s"
  }
}
`, domainName, validationMethod, tag1Key, tag1Value, tag2Key, tag2Value)
}

func testAccAcmCertificateConfigPrivateKeyWithoutChain(commonName string) string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, commonName)

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAcmCertificateConfigPrivateKeyTags(commonName string) string {
	key := tlsRsaPrivateKeyPem(2048)
	certificate := tlsRsaX509SelfSignedCertificatePem(key, commonName)

	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body = "%[1]s"
  private_key      = "%[2]s"

  tags = {
    key1 = "value1"
  }
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(key))
}

func testAccAcmCertificateConfigPrivateKey(certificate, privateKey, chain string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "test" {
  certificate_body  = "%[1]s"
  private_key       = "%[2]s"
  certificate_chain = "%[3]s"
}
`, tlsPemEscapeNewlines(certificate), tlsPemEscapeNewlines(privateKey), tlsPemEscapeNewlines(chain))
}

func testAccAcmCertificateConfig_disableCTLogging(domainName, validationMethod string) string {
	return fmt.Sprintf(`
resource "aws_acm_certificate" "cert" {
  domain_name       = "%s"
  validation_method = "%s"
  options {
    certificate_transparency_logging_preference = "DISABLED"
  }
}
`, domainName, validationMethod)

}

func testAccCheckAcmCertificateDestroy(s *terraform.State) error {
	acmconn := testAccProvider.Meta().(*AWSClient).acmconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acm_certificate" {
			continue
		}
		_, err := acmconn.DescribeCertificate(&acm.DescribeCertificateInput{
			CertificateArn: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Certificate still exists.")
		}

		// Verify the error is what we want
		if !isAWSErr(err, acm.ErrCodeResourceNotFoundException, "") {
			return err
		}
	}

	return nil
}
