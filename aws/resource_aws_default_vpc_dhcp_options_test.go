package aws

import (
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDefaultVpcDhcpOptions_basic(t *testing.T) {
	var d ec2.DhcpOptions
	resourceName := "aws_default_vpc_dhcp_options.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDefaultVpcDhcpOptionsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultVpcDhcpOptionsConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDHCPOptionsExists(resourceName, &d),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`dhcp-options/dopt-.+`)),
					resource.TestCheckResourceAttr(resourceName, "domain_name", resourceAwsEc2RegionalPrivateDnsSuffix(testAccGetRegion())),
					resource.TestCheckResourceAttr(resourceName, "domain_name_servers", "AmazonProvidedDNS"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "Default DHCP Option Set"),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
		},
	})
}

func testAccCheckAWSDefaultVpcDhcpOptionsDestroy(s *terraform.State) error {
	// We expect DHCP Options Set to still exist
	return nil
}

const testAccAWSDefaultVpcDhcpOptionsConfigBasic = `
resource "aws_default_vpc_dhcp_options" "foo" {
  tags = {
    Name = "Default DHCP Option Set"
  }
}
`
