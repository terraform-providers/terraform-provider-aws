package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsAcmpcaCertificateAuthority_Basic(t *testing.T) {
	resourceName := "aws_acmpca_certificate_authority.test"
	datasourceName := "data.aws_acmpca_certificate_authority.test"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_NonExistent,
				ExpectError: regexp.MustCompile(`ResourceNotFoundException`),
			},
			{
				Config: testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_ARN,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsAcmpcaCertificateAuthorityCheck(datasourceName, resourceName),
				),
			},
		},
	})
}

func testAccDataSourceAwsAcmpcaCertificateAuthorityCheck(datasourceName, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		resource, ok := s.RootModule().Resources[datasourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceName)
		}

		dataSource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		attrNames := []string{
			"arn",
			"certificate",
			"certificate_chain",
			"certificate_signing_request",
			"not_after",
			"not_before",
			"revocation_configuration.#",
			"revocation_configuration.0.crl_configuration.#",
			"revocation_configuration.0.crl_configuration.0.enabled",
			"serial",
			"status",
			"tags.%",
			"type",
		}

		for _, attrName := range attrNames {
			if resource.Primary.Attributes[attrName] != dataSource.Primary.Attributes[attrName] {
				return fmt.Errorf(
					"%s is %s; want %s",
					attrName,
					resource.Primary.Attributes[attrName],
					dataSource.Primary.Attributes[attrName],
				)
			}
		}

		return nil
	}
}

const testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_ARN = `
resource "aws_acmpca_certificate_authority" "wrong" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

data "aws_acmpca_certificate_authority" "test" {
  arn = "${aws_acmpca_certificate_authority.test.arn}"
}
`

const testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_NonExistent = `
data "aws_acmpca_certificate_authority" "test" {
  arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/tf-acc-test-does-not-exist"
}
`
