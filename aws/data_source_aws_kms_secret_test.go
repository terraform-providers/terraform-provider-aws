package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/kms"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSKmsSecretDataSource_removed(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, kms.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsKmsSecretDataSourceConfig,
				ExpectError: regexp.MustCompile(dataSourceAwsKmsSecretRemovedMessage),
			},
		},
	})
}

const testAccAwsKmsSecretDataSourceConfig = `
data "aws_kms_secret" "testing" {
  secret {
    name    = "secret_name"
    payload = "data-source-removed"

    context = {
      name = "value"
    }
  }
}
`
