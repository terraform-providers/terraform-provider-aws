package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsAcmpcaCertificateAuthority_basic(t *testing.T) {
	resourceName := "aws_acmpca_certificate_authority.test"
	datasourceName := "data.aws_acmpca_certificate_authority.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_NonExistent,
				ExpectError: regexp.MustCompile(`(AccessDeniedException|ResourceNotFoundException)`),
			},
			{
				Config: testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_ARN,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate", resourceName, "certificate"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_chain", resourceName, "certificate_chain"),
					resource.TestCheckResourceAttrPair(datasourceName, "certificate_signing_request", resourceName, "certificate_signing_request"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_after", resourceName, "not_after"),
					resource.TestCheckResourceAttrPair(datasourceName, "not_before", resourceName, "not_before"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.#", resourceName, "revocation_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.#", resourceName, "revocation_configuration.0.crl_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "revocation_configuration.0.crl_configuration.0.enabled", resourceName, "revocation_configuration.0.crl_configuration.0.enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, "serial", resourceName, "serial"),
					resource.TestCheckResourceAttrPair(datasourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "type", resourceName, "type"),
				),
			},
		},
	})
}

const testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_ARN = `
resource "aws_acmpca_certificate_authority" "wrong" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "terraformtesting.com"
    }
  }
}

data "aws_acmpca_certificate_authority" "test" {
  arn = aws_acmpca_certificate_authority.test.arn
}
`

//lintignore:AWSAT003,AWSAT005
const testAccDataSourceAwsAcmpcaCertificateAuthorityConfig_NonExistent = `
data "aws_acmpca_certificate_authority" "test" {
  arn = "arn:aws:acm-pca:us-east-1:123456789012:certificate-authority/tf-acc-test-does-not-exist"
}
`
