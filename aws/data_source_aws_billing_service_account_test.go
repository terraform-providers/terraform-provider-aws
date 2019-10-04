package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSBillingServiceAccount_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsBillingServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_billing_service_account.main", "id", "386209384616"),
					resource.TestCheckResourceAttr("data.aws_billing_service_account.main", "arn", "arn:aws:iam::386209384616:root"),
				),
			},
		},
	})
}

const testAccCheckAwsBillingServiceAccountConfig = `
data "aws_billing_service_account" "main" { }
`
