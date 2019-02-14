package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSRedshiftServiceAccount_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsRedshiftServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_redshift_service_account.main", "id", "902366379725"),
					resource.TestCheckResourceAttr("data.aws_redshift_service_account.main", "arn", "arn:aws:iam::902366379725:user/logs"),
				),
			},
			{
				Config: testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_redshift_service_account.regional", "id", "307160386991"),
					resource.TestCheckResourceAttr("data.aws_redshift_service_account.regional", "arn", "arn:aws:iam::307160386991:user/logs"),
				),
			},
		},
	})
}

const testAccCheckAwsRedshiftServiceAccountConfig = `
data "aws_redshift_service_account" "main" { }
`

const testAccCheckAwsRedshiftServiceAccountExplicitRegionConfig = `
data "aws_redshift_service_account" "regional" {
	region = "eu-west-2"
}
`
