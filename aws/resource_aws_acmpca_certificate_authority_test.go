package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_acmpca_certificate_authority", &resource.Sweeper{
		Name: "aws_acmpca_certificate_authority",
		F:    testSweepAcmpcaCertificateAuthorities,
	})
}

func testSweepAcmpcaCertificateAuthorities(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).acmpcaconn

	certificateAuthorities, err := listAcmpcaCertificateAuthorities(conn)
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping ACMPCA Certificate Authorities sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving ACMPCA Certificate Authorities: %w", err)
	}
	if len(certificateAuthorities) == 0 {
		log.Print("[DEBUG] No ACMPCA Certificate Authorities to sweep")
		return nil
	}

	var sweeperErrs *multierror.Error

	for _, certificateAuthority := range certificateAuthorities {
		arn := aws.StringValue(certificateAuthority.Arn)

		if aws.StringValue(certificateAuthority.Status) == acmpca.CertificateAuthorityStatusActive {
			log.Printf("[INFO] Disabling ACMPCA Certificate Authority: %s", arn)
			_, err := conn.UpdateCertificateAuthority(&acmpca.UpdateCertificateAuthorityInput{
				CertificateAuthorityArn: aws.String(arn),
				Status:                  aws.String(acmpca.CertificateAuthorityStatusDisabled),
			})
			if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error disabling ACMPCA Certificate Authority (%s): %w", arn, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		log.Printf("[INFO] Deleting ACMPCA Certificate Authority: %s", arn)
		_, err := conn.DeleteCertificateAuthority(&acmpca.DeleteCertificateAuthorityInput{
			CertificateAuthorityArn:     aws.String(arn),
			PermanentDeletionTimeInDays: aws.Int64(int64(7)),
		})
		if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			sweeperErr := fmt.Errorf("error deleting ACMPCA Certificate Authority (%s): %w", arn, err)
			log.Printf("[ERROR] %s", sweeperErr)
			sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			continue
		}
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsAcmpcaCertificateAuthority_basic(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Required,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "acm-pca", regexp.MustCompile(`certificate-authority/.+`)),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.key_algorithm", "RSA_4096"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.signing_algorithm", "SHA512WITHRSA"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "certificate_authority_configuration.0.subject.0.common_name", "terraformtesting.com"),
					resource.TestCheckResourceAttr(resourceName, "certificate", ""),
					resource.TestCheckResourceAttr(resourceName, "certificate_chain", ""),
					resource.TestCheckResourceAttrSet(resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					testAccCheckResourceAttrRfc3339(resourceName, "not_after"),
					testAccCheckResourceAttrRfc3339(resourceName, "not_before"),
					resource.TestCheckResourceAttr(resourceName, "permanent_deletion_time_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "serial", ""),
					resource.TestCheckResourceAttr(resourceName, "status", "PENDING_CERTIFICATE"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "type", "SUBORDINATE"),
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

func TestAccAwsAcmpcaCertificateAuthority_disappears(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Required,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsAcmpcaCertificateAuthority(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsAcmpcaCertificateAuthority_Enabled(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Enabled(rName, acmpca.CertificateAuthorityTypeRoot, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "type", acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", acmpca.CertificateAuthorityStatusPendingCertificate),
					testAccCheckAwsAcmpcaCertificateAuthorityActivateCA(&certificateAuthority),
				),
			},
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Enabled(rName, acmpca.CertificateAuthorityTypeRoot, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "type", acmpca.CertificateAuthorityTypeRoot),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "status", acmpca.CertificateAuthorityStatusActive),
				),
			},
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Enabled(rName, acmpca.CertificateAuthorityTypeRoot, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "status", acmpca.CertificateAuthorityStatusDisabled),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
		},
	})
}

func TestAccAwsAcmpcaCertificateAuthority_RevocationConfiguration_CrlConfiguration_CustomCname(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCname(rName, "crl.terraformtesting.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", "crl.terraformtesting.com"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCname(rName, "crl2.terraformtesting.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", "crl2.terraformtesting.com"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing custom cname on resource update
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test adding custom cname on resource update
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCname(rName, "crl.terraformtesting.com"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", "crl.terraformtesting.com"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Required,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsAcmpcaCertificateAuthority_RevocationConfiguration_CrlConfiguration_Enabled(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test disabling revocation configuration
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
			// Test enabling revocation configuration
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Required,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsAcmpcaCertificateAuthority_RevocationConfiguration_CrlConfiguration_ExpirationInDays(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			// Test creating revocation configuration on resource creation
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_ExpirationInDays(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.custom_cname", ""),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test importing revocation configuration
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
			// Test updating revocation configuration
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_ExpirationInDays(rName, 2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.expiration_in_days", "2"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.s3_bucket_name", rName),
				),
			},
			// Test removing revocation configuration on resource update
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Required,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "revocation_configuration.0.crl_configuration.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccAwsAcmpcaCertificateAuthority_Tags(t *testing.T) {
	var certificateAuthority acmpca.CertificateAuthority
	resourceName := "aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAcmpcaCertificateAuthorityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Tags_Single,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Tags_SingleUpdated,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value-updated"),
				),
			},
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Tags_Multiple,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "tag2value"),
				),
			},
			{
				Config: testAccAwsAcmpcaCertificateAuthorityConfig_Tags_Single,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName, &certificateAuthority),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "tag1value"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"permanent_deletion_time_in_days",
				},
			},
		},
	})
}

func testAccCheckAwsAcmpcaCertificateAuthorityDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).acmpcaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_acmpca_certificate_authority" {
			continue
		}

		input := &acmpca.DescribeCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeCertificateAuthority(input)

		if err != nil {
			if isAWSErr(err, acmpca.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}

		if output != nil && output.CertificateAuthority != nil && aws.StringValue(output.CertificateAuthority.Arn) == rs.Primary.ID && aws.StringValue(output.CertificateAuthority.Status) != acmpca.CertificateAuthorityStatusDeleted {
			return fmt.Errorf("ACMPCA Certificate Authority %q still exists in non-DELETED state: %s", rs.Primary.ID, aws.StringValue(output.CertificateAuthority.Status))
		}
	}

	return nil

}

func testAccCheckAwsAcmpcaCertificateAuthorityExists(resourceName string, certificateAuthority *acmpca.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).acmpcaconn
		input := &acmpca.DescribeCertificateAuthorityInput{
			CertificateAuthorityArn: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeCertificateAuthority(input)

		if err != nil {
			return err
		}

		if output == nil || output.CertificateAuthority == nil {
			return fmt.Errorf("ACMPCA Certificate Authority %q does not exist", rs.Primary.ID)
		}

		*certificateAuthority = *output.CertificateAuthority

		return nil
	}
}

func testAccCheckAwsAcmpcaCertificateAuthorityActivateCA(certificateAuthority *acmpca.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).acmpcaconn

		arn := aws.StringValue(certificateAuthority.Arn)

		getCsrResp, err := conn.GetCertificateAuthorityCsr(&acmpca.GetCertificateAuthorityCsrInput{
			CertificateAuthorityArn: aws.String(arn),
		})
		if err != nil {
			return fmt.Errorf("error getting ACMPCA Certificate Authority (%s) CSR: %s", arn, err)
		}

		issueCertResp, err := conn.IssueCertificate(&acmpca.IssueCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			Csr:                     []byte(aws.StringValue(getCsrResp.Csr)),
			IdempotencyToken:        aws.String(resource.UniqueId()),
			SigningAlgorithm:        certificateAuthority.CertificateAuthorityConfiguration.SigningAlgorithm,
			TemplateArn:             aws.String(fmt.Sprintf("arn:%s:acm-pca:::template/RootCACertificate/V1", testAccGetPartition())),
			Validity: &acmpca.Validity{
				Type:  aws.String(acmpca.ValidityPeriodTypeYears),
				Value: aws.Int64(10),
			},
		})
		if err != nil {
			return fmt.Errorf("error issuing ACMPCA Certificate Authority (%s) Root CA certificate from CSR: %s", arn, err)
		}

		// Wait for certificate status to become ISSUED.
		err = conn.WaitUntilCertificateIssued(&acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			CertificateArn:          issueCertResp.CertificateArn,
		})
		if err != nil {
			return fmt.Errorf("error waiting for ACMPCA Certificate Authority (%s) Root CA certificate to become ISSUED: %s", arn, err)
		}

		getCertResp, err := conn.GetCertificate(&acmpca.GetCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			CertificateArn:          issueCertResp.CertificateArn,
		})
		if err != nil {
			return fmt.Errorf("error getting ACMPCA Certificate Authority (%s) issued Root CA certificate: %s", arn, err)
		}

		_, err = conn.ImportCertificateAuthorityCertificate(&acmpca.ImportCertificateAuthorityCertificateInput{
			CertificateAuthorityArn: aws.String(arn),
			Certificate:             []byte(aws.StringValue(getCertResp.Certificate)),
		})
		if err != nil {
			return fmt.Errorf("error importing ACMPCA Certificate Authority (%s) Root CA certificate: %s", arn, err)
		}

		return err
	}
}

func testAccCheckAwsAcmpcaCertificateAuthorityDisableCA(certificateAuthority *acmpca.CertificateAuthority) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).acmpcaconn

		_, err := conn.UpdateCertificateAuthority(&acmpca.UpdateCertificateAuthorityInput{
			CertificateAuthorityArn: certificateAuthority.Arn,
			Status:                  aws.String(acmpca.CertificateAuthorityStatusDisabled),
		})

		return err
	}
}

func listAcmpcaCertificateAuthorities(conn *acmpca.ACMPCA) ([]*acmpca.CertificateAuthority, error) {
	certificateAuthorities := []*acmpca.CertificateAuthority{}
	input := &acmpca.ListCertificateAuthoritiesInput{}

	for {
		output, err := conn.ListCertificateAuthorities(input)
		if err != nil {
			return certificateAuthorities, err
		}
		certificateAuthorities = append(certificateAuthorities, output.CertificateAuthorities...)
		if output.NextToken == nil {
			break
		}
		input.NextToken = output.NextToken
	}

	return certificateAuthorities, nil
}

func testAccAwsAcmpcaCertificateAuthorityConfig_Enabled(rName, certificateAuthorityType string, enabled bool) string {
	return fmt.Sprintf(`
resource "aws_acmpca_certificate_authority" "test" {
  enabled                         = %[1]t
  permanent_deletion_time_in_days = 7
  type                            = %[2]q

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "%[3]s.com"
    }
  }
}
`, enabled, certificateAuthorityType, rName)
}

const testAccAwsAcmpcaCertificateAuthorityConfig_Required = `
resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}
`

func testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_CustomCname(rName, customCname string) string {
	return fmt.Sprintf(`
%s

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }

  revocation_configuration {
    crl_configuration {
      custom_cname       = "%s"
      enabled            = true
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }

  depends_on = [aws_s3_bucket_policy.test]
}
`, testAccAwsAcmpcaCertificateAuthorityConfig_S3Bucket(rName), customCname)
}

func testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_Enabled(rName string, enabled bool) string {
	return fmt.Sprintf(`
%s

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = %t
      expiration_in_days = 1
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }
}
`, testAccAwsAcmpcaCertificateAuthorityConfig_S3Bucket(rName), enabled)
}

func testAccAwsAcmpcaCertificateAuthorityConfig_RevocationConfiguration_CrlConfiguration_ExpirationInDays(rName string, expirationInDays int) string {
	return fmt.Sprintf(`
%s

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }

  revocation_configuration {
    crl_configuration {
      enabled            = true
      expiration_in_days = %d
      s3_bucket_name     = aws_s3_bucket.test.id
    }
  }
}
`, testAccAwsAcmpcaCertificateAuthorityConfig_S3Bucket(rName), expirationInDays)
}

func testAccAwsAcmpcaCertificateAuthorityConfig_S3Bucket(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = "%s"
  force_destroy = true
}

data "aws_iam_policy_document" "acmpca_bucket_access" {
  statement {
    actions = [
      "s3:GetBucketAcl",
      "s3:GetBucketLocation",
      "s3:PutObject",
      "s3:PutObjectAcl",
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*",
    ]

    principals {
      identifiers = ["acm-pca.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.acmpca_bucket_access.json
}
`, rName)
}

const testAccAwsAcmpcaCertificateAuthorityConfig_Tags_Single = `
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }

  tags = {
    tag1 = "tag1value"
  }
}
`

const testAccAwsAcmpcaCertificateAuthorityConfig_Tags_SingleUpdated = `
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }

  tags = {
    tag1 = "tag1value-updated"
  }
}
`

const testAccAwsAcmpcaCertificateAuthorityConfig_Tags_Multiple = `
resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }

  tags = {
    tag1 = "tag1value"
    tag2 = "tag2value"
  }
}
`
