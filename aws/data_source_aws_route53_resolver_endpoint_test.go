package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSRoute53ResolverEndpointDataSource_Basic(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	rInt := acctest.RandInt()
	direction := "INBOUND"
	resourceName := "aws_route53_resolver_endpoint.foo"
	datasourceName := "data.aws_route53_resolver_endpoint.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheckSkipRoute53(t),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsRoute53ResolverEndpointConfig_NonExistent,
				ExpectError: regexp.MustCompile("The ID provided could not be found"),
			},
			{
				Config: testAccDataSourceRoute53ResolverEndpointConfig_initial(rInt, direction, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttr(datasourceName, "ip_addresses.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSRoute53ResolverEndpointDataSource_Filter(t *testing.T) {
	name := acctest.RandomWithPrefix("tf-acc-test")
	rInt := acctest.RandInt()
	direction := "OUTBOUND"
	resourceName := "aws_route53_resolver_endpoint.foo"
	datasourceName := "data.aws_route53_resolver_endpoint.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheckSkipRoute53(t),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsRoute53ResolverEndpointConfig_NonExistentFilter,
				ExpectError: regexp.MustCompile("Your query returned no results. Please change your search criteria and try again"),
			},
			{
				Config: testAccDataSourceRoute53ResolverEndpointConfig_filter(rInt, direction, name),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(datasourceName, "resolver_endpoint_id", resourceName, "id"),
					resource.TestCheckResourceAttr(datasourceName, "ip_addresses.#", "2"),
				),
			},
		},
	})
}

func testAccDataSourceRoute53ResolverEndpointConfig_base(rInt int) string {
	return testAccAvailableAZsNoOptInConfig() + fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-r53-resolver-vpc-%[1]d"
  }
}

resource "aws_subnet" "sn1" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = "tf-acc-r53-resolver-sn1-%[1]d"
  }
}

resource "aws_subnet" "sn2" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = "tf-acc-r53-resolver-sn2-%[1]d"
  }
}

resource "aws_subnet" "sn3" {
  vpc_id            = aws_vpc.foo.id
  cidr_block        = cidrsubnet(aws_vpc.foo.cidr_block, 2, 2)
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = "tf-acc-r53-resolver-sn3-%[1]d"
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = aws_vpc.foo.id
  name   = "tf-acc-r53-resolver-sg1-%[1]d"

  tags = {
    Name = "tf-acc-r53-resolver-sg1-%[1]d"
  }
}

resource "aws_security_group" "sg2" {
  vpc_id = aws_vpc.foo.id
  name   = "tf-acc-r53-resolver-sg2-%[1]d"

  tags = {
    Name = "tf-acc-r53-resolver-sg2-%[1]d"
  }
}
`, rInt)
}

func testAccDataSourceRoute53ResolverEndpointConfig_initial(rInt int, direction, name string) string {
	return composeConfig(testAccDataSourceRoute53ResolverEndpointConfig_base(rInt), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "foo" {
  direction = "%s"
  name      = "%s"

  security_group_ids = [
    aws_security_group.sg1.id,
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
    ip        = cidrhost(aws_subnet.sn2.cidr_block, 8)
  }

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}

data "aws_route53_resolver_endpoint" "foo" {
  resolver_endpoint_id = aws_route53_resolver_endpoint.foo.id
}
`, direction, name))
}

func testAccDataSourceRoute53ResolverEndpointConfig_filter(rInt int, direction, name string) string {
	return composeConfig(testAccDataSourceRoute53ResolverEndpointConfig_base(rInt), fmt.Sprintf(`
resource "aws_route53_resolver_endpoint" "foo" {
  direction = "%s"
  name      = "%s"

  security_group_ids = [
    aws_security_group.sg1.id,
    aws_security_group.sg2.id,
  ]

  ip_address {
    subnet_id = aws_subnet.sn1.id
  }

  ip_address {
    subnet_id = aws_subnet.sn2.id
    ip        = cidrhost(aws_subnet.sn2.cidr_block, 8)
  }

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}

data "aws_route53_resolver_endpoint" "foo" {
  filter {
    name   = "Name"
    values = [aws_route53_resolver_endpoint.foo.name]
  }

  filter {
    name   = "SecurityGroupIds"
    values = [aws_security_group.sg1.id, aws_security_group.sg2.id]
  }
}
`, direction, name))
}

const testAccDataSourceAwsRoute53ResolverEndpointConfig_NonExistent = `
data "aws_route53_resolver_endpoint" "foo" {
  resolver_endpoint_id = "rslvr-in-8g85830108dd4c82b"
}
`

const testAccDataSourceAwsRoute53ResolverEndpointConfig_NonExistentFilter = `
data "aws_route53_resolver_endpoint" "foo" {
  filter {
    name   = "Name"
    values = ["None-Existent-Resource"]
  }
}
`
