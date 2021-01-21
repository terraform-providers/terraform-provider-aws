package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSBillingServiceAccount_basic(t *testing.T) {
	dataSourceName := "data.aws_billing_service_account.main"

	billingAccountID := "386209384616"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsBillingServiceAccountConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "id", billingAccountID),
					testAccCheckResourceAttrGlobalARNAccountID(dataSourceName, "arn", billingAccountID, "iam", "root"),
				),
			},
		},
	})
}

const testAccCheckAwsBillingServiceAccountConfig = `
data "aws_billing_service_account" "main" {}
`
