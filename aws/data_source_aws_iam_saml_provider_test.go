package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDataSourceSAMLProvider_basic(t *testing.T) {
	providerName := acctest.RandomWithPrefix("tf-saml-provider-test")
	dataSourceName := "data.aws_iam_saml_provider.test"
	resourceName := "aws_iam_saml_provider.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccIAMSamlProviderDataConfig(providerName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "saml_metadata_document", resourceName, "saml_metadata_document"),
					resource.TestCheckResourceAttrPair(dataSourceName, "valid_until", resourceName, "valid_until"),
				),
			},
		},
	})
}

func testAccIAMSamlProviderDataConfig(providerName string) string {
	return fmt.Sprintf(`
resource "aws_iam_saml_provider" "test" {
  name                   = %q
  saml_metadata_document = "${file("./test-fixtures/saml-metadata.xml")}"
}

data "aws_iam_saml_provider" "test" {
  arn                    = "${aws_iam_saml_provider.test.arn}"
}
`, providerName)
}
