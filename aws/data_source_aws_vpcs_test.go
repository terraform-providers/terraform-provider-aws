package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceAwsVpcs_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcsConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcsDataSourceExists("data.aws_vpcs.all"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcs_tags(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcsConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcsDataSourceExists("data.aws_vpcs.selected"),
					resource.TestCheckResourceAttr("data.aws_vpcs.selected", "ids.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcs_filters(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsVpcsConfig_filters(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsVpcsDataSourceExists("data.aws_vpcs.selected"),
					testCheckResourceAttrGreaterThanValue("data.aws_vpcs.selected", "ids.#", "0"),
				),
			},
		},
	})
}

func testCheckResourceAttrGreaterThanValue(name, key, value string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		ms := s.RootModule()
		rs, ok := ms.Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s in %s", name, ms.Path)
		}

		is := rs.Primary
		if is == nil {
			return fmt.Errorf("No primary instance: %s in %s", name, ms.Path)
		}

		if v, ok := is.Attributes[key]; !ok || !(v > value) {
			if !ok {
				return fmt.Errorf("%s: Attribute '%s' not found", name, key)
			}

			return fmt.Errorf(
				"%s: Attribute '%s' is not greater than %#v, got %#v",
				name,
				key,
				value,
				v)
		}
		return nil

	}
}

func testAccCheckAwsVpcsDataSourceExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find aws_vpcs data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("aws_vpcs data source ID not set")
		}
		return nil
	}
}

func testAccDataSourceAwsVpcsConfig() string {
	return fmt.Sprintf(`
	resource "aws_vpc" "test-vpc" {
  		cidr_block = "10.0.0.0/24"
	}

	data "aws_vpcs" "all" {}
	`)
}

func testAccDataSourceAwsVpcsConfig_tags(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "test-vpc" {
  		cidr_block = "10.0.0.0/24"

  		tags = {
  			Name = "testacc-vpc-%s"
  			Service = "testacc-test"
  		}
	}

	data "aws_vpcs" "selected" {
	tags = {
			Name = "testacc-vpc-%s"
			Service = "${aws_vpc.test-vpc.tags["Service"]}"
		}
	}
	`, rName, rName)
}

func testAccDataSourceAwsVpcsConfig_filters(rName string) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "test-vpc" {
  		cidr_block = "192.168.0.0/25"
  		tags = {
  			Name = "testacc-vpc-%s"
  		}
	}

	data "aws_vpcs" "selected" {
		filter {
			name = "cidr"
    		values = ["${aws_vpc.test-vpc.cidr_block}"]
		}
	}
	`, rName)
}
