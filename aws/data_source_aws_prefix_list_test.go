package aws

import (
	"fmt"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsPrefixList_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPrefixListConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_id"),
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_name"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsPrefixList_filter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsPrefixListConfigFilter,
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_id"),
					testAccDataSourceAwsPrefixListCheck("data.aws_prefix_list.s3_by_name"),
				),
			},
		},
	})
}

func testAccDataSourceAwsPrefixListCheck(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		attr := rs.Primary.Attributes

		if attr["name"] != "com.amazonaws.us-west-2.s3" {
			return fmt.Errorf("bad name %s", attr["name"])
		}
		if attr["id"] != "pl-68a54001" {
			return fmt.Errorf("bad id %s", attr["id"])
		}

		var (
			cidrBlockSize int
			err           error
		)

		if cidrBlockSize, err = strconv.Atoi(attr["cidr_blocks.#"]); err != nil {
			return err
		}
		if cidrBlockSize < 1 {
			return fmt.Errorf("cidr_blocks seem suspiciously low: %d", cidrBlockSize)
		}

		return nil
	}
}

const testAccDataSourceAwsPrefixListConfig = `
data "aws_prefix_list" "s3_by_id" {
  prefix_list_id = "pl-68a54001"
}

data "aws_prefix_list" "s3_by_name" {
  name = "com.amazonaws.us-west-2.s3"
}
`

const testAccDataSourceAwsPrefixListConfigFilter = `
data "aws_prefix_list" "s3_by_name" {
  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.us-west-2.s3"]
  }
}

data "aws_prefix_list" "s3_by_id" {
  filter {
    name   = "prefix-list-id"
    values = ["pl-68a54001"]
  }
}
`
