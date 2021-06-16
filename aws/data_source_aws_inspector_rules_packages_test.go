package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/inspector"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSInspectorRulesPackages_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, inspector.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSInspectorRulesPackagesConfig,
				Check:  resource.TestCheckResourceAttrSet("data.aws_inspector_rules_packages.test", "arns.#"),
			},
		},
	})
}

const testAccCheckAWSInspectorRulesPackagesConfig = `
data "aws_inspector_rules_packages" "test" {}
`
