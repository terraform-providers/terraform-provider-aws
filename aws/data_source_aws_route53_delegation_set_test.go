package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceRoute53DelegationSet_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_delegation_set.dset"
	resourceName := "aws_route53_delegation_set.dset"

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDataSourceAWSRoute53DelegationSetConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						dataSourceName, "name_servers.#",
						resourceName, "name_servers.#",
					),
					resource.TestMatchResourceAttr(
						"data.aws_route53_delegation_set.dset",
						"caller_reference",
						regexp.MustCompile("DynDNS(.*)"),
					),
				),
			},
		},
	})
}

const testAccAWSDataSourceAWSRoute53DelegationSetConfig_basic = `
resource "aws_route53_delegation_set" "dset" {
  reference_name = "DynDNS"
}

resource "aws_route53_zone" "primary" {
  name              = "example.xyz"
  delegation_set_id = "${aws_route53_delegation_set.dset.id}"
}

data "aws_route53_delegation_set" "dset" {
  id = "${aws_route53_delegation_set.dset.id}"
}
`
