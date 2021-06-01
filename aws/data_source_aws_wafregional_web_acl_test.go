package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/wafregional"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWafRegionalWebAcl_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_web_acl.web_acl"
	datasourceName := "data.aws_wafregional_web_acl.web_acl"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(wafregional.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalWebAclConfig_NonExistent,
				ExpectError: regexp.MustCompile(`web ACLs not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalWebAclConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRegionalWebAclConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_web_acl" "web_acl" {
  name        = %[1]q
  metric_name = "tfWebACL"

  default_action {
    type = "ALLOW"
  }
}

data "aws_wafregional_web_acl" "web_acl" {
  name = aws_wafregional_web_acl.web_acl.name
}
`, name)
}

const testAccDataSourceAwsWafRegionalWebAclConfig_NonExistent = `
data "aws_wafregional_web_acl" "web_acl" {
  name = "tf-acc-test-does-not-exist"
}
`
