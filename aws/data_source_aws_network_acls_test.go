package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsNetworkAcls_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				// Ensure at least 1 network ACL exists. We cannot use depends_on.
				Config: testAccDataSourceAwsNetworkAclsConfig_Base(rName),
			},
			{
				Config: testAccDataSourceAwsNetworkAclsConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					// At least 1
					resource.TestMatchResourceAttr("data.aws_network_acls.test", "ids.#", regexp.MustCompile(`^[1-9][0-9]*`)),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkAcls_Filter(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkAclsConfig_Filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_acls.test", "ids.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkAcls_Tags(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkAclsConfig_Tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_network_acls.test", "ids.#", "2"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsNetworkAcls_VpcID(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsNetworkAclsConfig_VpcID(rName),
				Check: resource.ComposeTestCheckFunc(
					// The VPC will have a default network ACL
					resource.TestCheckResourceAttr("data.aws_network_acls.test", "ids.#", "3"),
				),
			},
		},
	})
}

func testAccDataSourceAwsNetworkAclsConfig_Base(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "testacc-acl-%s"
  }
}

resource "aws_network_acl" "acl" {
  count = 2

  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "testacc-acl-%s"
  }
}
`, rName, rName)
}

func testAccDataSourceAwsNetworkAclsConfig_basic(rName string) string {
	return testAccDataSourceAwsNetworkAclsConfig_Base(rName) + `
data "aws_network_acls" "test" {}
`
}

func testAccDataSourceAwsNetworkAclsConfig_Filter(rName string) string {
	return testAccDataSourceAwsNetworkAclsConfig_Base(rName) + `
data "aws_network_acls" "test" {
  filter {
    name   = "network-acl-id"
    values = ["${aws_network_acl.acl.0.id}"]
  }
}
`
}

func testAccDataSourceAwsNetworkAclsConfig_Tags(rName string) string {
	return testAccDataSourceAwsNetworkAclsConfig_Base(rName) + `
data "aws_network_acls" "test" {
  tags = {
    Name = "${aws_network_acl.acl.0.tags.Name}"
  }
}
`
}

func testAccDataSourceAwsNetworkAclsConfig_VpcID(rName string) string {
	return testAccDataSourceAwsNetworkAclsConfig_Base(rName) + `
data "aws_network_acls" "test" {
  vpc_id = "${aws_network_acl.acl.0.vpc_id}"
}
`
}
