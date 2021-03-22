package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/acmpca"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsAcmpcaCertificate_Basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_acmpca_certificate.test"
	dataSourceName := "data.aws_acmpca_certificate.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, acmpca.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsAcmpcaCertificateConfig_NonExistent,
				ExpectError: regexp.MustCompile(`ResourceNotFoundException`),
			},
			{
				Config: testAccDataSourceAwsAcmpcaCertificateConfig_ARN(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "certificate", resourceName, "certificate"),
					resource.TestCheckResourceAttrPair(dataSourceName, "certificate_chain", resourceName, "certificate_chain"),
					resource.TestCheckResourceAttrPair(dataSourceName, "certificate_authority_arn", resourceName, "certificate_authority_arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsAcmpcaCertificateConfig_ARN(rName string) string {
	return fmt.Sprintf(`
data "aws_acmpca_certificate" "test" {
  arn                       = aws_acmpca_certificate.test.arn
  certificate_authority_arn = aws_acmpca_certificate_authority.test.arn
}

resource "aws_acmpca_certificate" "test" {
  certificate_authority_arn   = aws_acmpca_certificate_authority.test.arn
  certificate_signing_request = aws_acmpca_certificate_authority.test.certificate_signing_request
  signing_algorithm           = "SHA256WITHRSA"

  template_arn = "arn:${data.aws_partition.current.partition}:acm-pca:::template/RootCACertificate/V1"

  validity {
    type  = "YEARS"
    value = 1
  }
}

resource "aws_acmpca_certificate_authority" "test" {
  permanent_deletion_time_in_days = 7
  type                            = "ROOT"

  certificate_authority_configuration {
    key_algorithm     = "RSA_4096"
    signing_algorithm = "SHA512WITHRSA"

    subject {
      common_name = "%[1]s.com"
    }
  }
}

data "aws_partition" "current" {}
`, rName)
}

const testAccDataSourceAwsAcmpcaCertificateConfig_NonExistent = `
data "aws_acmpca_certificate" "test" {
  arn                       = "arn:${data.aws_partition.current.partition}:acm-pca:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:certificate-authority/does-not-exist/certificate/does-not-exist"
  certificate_authority_arn = "arn:${data.aws_partition.current.partition}:acm-pca:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:certificate-authority/does-not-exist"
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}
`
