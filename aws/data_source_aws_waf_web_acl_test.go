package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/waf"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWafWebAcl_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_waf_web_acl.web_acl"
	datasourceName := "data.aws_waf_web_acl.web_acl"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(waf.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafWebAclConfig_NonExistent,
				ExpectError: regexp.MustCompile(`web ACLs not found`),
			},
			{
				Config: testAccDataSourceAwsWafWebAclConfig_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafWebAclConfig_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_waf_web_acl" "web_acl" {
  name        = %[1]q
  metric_name = "tfWebACL"

  default_action {
    type = "ALLOW"
  }
}

data "aws_waf_web_acl" "web_acl" {
  name = aws_waf_web_acl.web_acl.name
}
`, name)
}

const testAccDataSourceAwsWafWebAclConfig_NonExistent = `
data "aws_waf_web_acl" "web_acl" {
  name = "tf-acc-test-does-not-exist"
}
`
