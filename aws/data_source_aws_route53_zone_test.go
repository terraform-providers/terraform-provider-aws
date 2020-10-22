package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsRoute53Zone_id(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigId(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers", dataSourceName, "name_servers"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRoute53Zone_name(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigName(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers", dataSourceName, "name_servers"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRoute53Zone_tags(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigTagsPrivate(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers", dataSourceName, "name_servers"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRoute53Zone_vpc(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_route53_zone.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigVpc(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "id", dataSourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "name_servers", dataSourceName, "name_servers"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsRoute53Zone_serviceDiscovery(t *testing.T) {
	rInt := acctest.RandInt()
	resourceName := "aws_service_discovery_private_dns_namespace.test"
	dataSourceName := "data.aws_route53_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53ZoneDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsRoute53ZoneConfigServiceDiscovery(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttr(dataSourceName, "linked_service_principal", "servicediscovery.amazonaws.com"),
					resource.TestCheckResourceAttrPair(dataSourceName, "linked_service_description", resourceName, "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAwsRoute53ZoneConfigId(rInt int) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "terraformtestacchz-%[1]d.com."
}

data "aws_route53_zone" "test" {
  zone_id = aws_route53_zone.test.zone_id
}
`, rInt)
}

func testAccDataSourceAwsRoute53ZoneConfigName(rInt int) string {
	return fmt.Sprintf(`
resource "aws_route53_zone" "test" {
  name = "terraformtestacchz-%[1]d.com."
}

data "aws_route53_zone" "test" {
  name = aws_route53_zone.test.name
}
`, rInt)
}

func testAccDataSourceAwsRoute53ZoneConfigTagsPrivate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-r53-zone-data-source-%[1]d"
  }
}

resource "aws_route53_zone" "test" {
  name = "terraformtestacchz-%[1]d.com."

  vpc {
    vpc_id = aws_vpc.test.id
  }

  tags = {
    Environment = "tf-acc-test-%[1]d"
    Name        = "tf-acc-test-%[1]d"
  }
}

data "aws_route53_zone" "test" {
  name         = aws_route53_zone.test.name
  private_zone = true
  vpc_id       = aws_vpc.test.id

  tags = {
    Environment = "tf-acc-test-%[1]d"
  }
}
`, rInt)
}

func testAccDataSourceAwsRoute53ZoneConfigVpc(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-r53-zone-data-source-%[1]d"
  }
}

resource "aws_route53_zone" "test" {
  name = "test.acc-%[1]d."

  vpc {
    vpc_id = aws_vpc.test.id
  }

  tags = {
    Environment = "dev-%[1]d"
  }
}

data "aws_route53_zone" "test" {
  name         = aws_route53_zone.test.name
  private_zone = true
  vpc_id       = aws_vpc.test.id
}
`, rInt)
}

func testAccDataSourceAwsRoute53ZoneConfigServiceDiscovery(rInt int) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-r53-zone-data-source-%[1]d"
  }
}

resource "aws_service_discovery_private_dns_namespace" "test" {
  name = "test.acc-sd-%[1]d"
  vpc  = aws_vpc.test.id
}

data "aws_route53_zone" "test" {
  name   = aws_service_discovery_private_dns_namespace.test.name
  vpc_id = aws_vpc.test.id
}
`, rInt)
}
