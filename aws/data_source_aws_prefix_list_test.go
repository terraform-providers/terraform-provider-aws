package aws

import (
	"fmt"
	"regexp"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
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

func TestAccDataSourceAwsPrefixList_nameDoesNotOverrideFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAwsPrefixListConfig_nameDoesNotOverrideFilter,
				ExpectError: regexp.MustCompile(`no matching prefix list found`),
			},
		},
	})
}

func testAccDataSourceAwsPrefixListCheck(name string) resource.TestCheckFunc {
	getPrefixListId := func(name string) (string, error) {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		input := ec2.DescribePrefixListsInput{
			Filters: buildEC2AttributeFilterList(map[string]string{
				"prefix-list-name": name,
			}),
		}

		output, err := conn.DescribePrefixLists(&input)
		if err != nil {
			return "", err
		}

		if len(output.PrefixLists) != 1 {
			return "", fmt.Errorf("prefix list %s not found", name)
		}

		return aws.StringValue(output.PrefixLists[0].PrefixListId), nil
	}

	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		attr := rs.Primary.Attributes

		region := testAccGetRegion()
		prefixListName := fmt.Sprintf("com.amazonaws.%s.s3", region)
		prefixListId, err := getPrefixListId(prefixListName)
		if err != nil {
			return err
		}

		if attr["name"] != prefixListName {
			return fmt.Errorf("bad name %s", attr["name"])
		}
		if attr["id"] != prefixListId {
			return fmt.Errorf("bad id %s", attr["id"])
		}

		var cidrBlockSize int

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
data "aws_region" "current" {}

data "aws_prefix_list" "s3_by_id" {
  prefix_list_id = data.aws_prefix_list.s3_by_name.id
}

data "aws_prefix_list" "s3_by_name" {
  name = "com.amazonaws.${data.aws_region.current.name}.s3"
}
`

const testAccDataSourceAwsPrefixListConfigFilter = `
data "aws_region" "current" {}

data "aws_prefix_list" "s3_by_name" {
  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.${data.aws_region.current.name}.s3"]
  }
}

data "aws_prefix_list" "s3_by_id" {
  filter {
    name   = "prefix-list-id"
    values = [data.aws_prefix_list.s3_by_name.id]
  }
}
`

const testAccDataSourceAwsPrefixListConfig_nameDoesNotOverrideFilter = `
data "aws_region" "current" {}

data "aws_prefix_list" "test" {
  name = "com.amazonaws.${data.aws_region.current.name}.dynamodb"

  filter {
    name   = "prefix-list-name"
    values = ["com.amazonaws.${data.aws_region.current.name}.s3"]
  }
}
`
