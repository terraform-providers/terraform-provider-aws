package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsWafRegionalIPSet_basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_wafregional_ipset.ipset"
	datasourceName := "data.aws_wafregional_ipset.ipset"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsWafRegionalIPSet_NonExistent,
				ExpectError: regexp.MustCompile(`WAF Regional IP Set not found`),
			},
			{
				Config: testAccDataSourceAwsWafRegionalIPSet_Name(name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsWafRegionalIPSet_Name(name string) string {
	return fmt.Sprintf(`
resource "aws_wafregional_ipset" "ipset" {
  name = %[1]q
}

data "aws_wafregional_ipset" "ipset" {
  name = aws_wafregional_ipset.ipset.name
}
`, name)
}

const testAccDataSourceAwsWafRegionalIPSet_NonExistent = `
data "aws_wafregional_ipset" "ipset" {
  name = "tf-acc-test-does-not-exist"
}
`
