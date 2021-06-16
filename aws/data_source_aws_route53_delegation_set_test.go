package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRoute53DelegationSetDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_route53_delegation_set.dset"
	resourceName := "aws_route53_delegation_set.dset"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, route53.EndpointsID),
		Providers:  testAccProviders,
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
  delegation_set_id = aws_route53_delegation_set.dset.id
}

data "aws_route53_delegation_set" "dset" {
  id = aws_route53_delegation_set.dset.id
}
`
