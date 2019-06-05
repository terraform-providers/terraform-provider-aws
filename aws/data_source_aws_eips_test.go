package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"testing"
)

func TestAccDataSourceAwsEips_Tags(t *testing.T) {
	var conf ec2.Address

	dataSourceName := "data.aws_eips.test"
	resourceName := "aws_eip.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipsConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &conf),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ips.0", resourceName, "public_ip"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "public_ips.#", "1"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsEipsClassic_Tags(t *testing.T) {
	var conf ec2.Address

	dataSourceName := "data.aws_eips.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEIPDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsEipsClassicConfigTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSEIPExists("aws_eip.test", &conf),
					resource.TestCheckResourceAttrPair(dataSourceName, "ids.0", "aws_eips.test", "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceName, "public_ips.0", "aws_eips.test", "public_ip"),
					resource.TestCheckResourceAttr(dataSourceName, "ids.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "public_ips.#", "1"),
				),
			},
		},
	})
}

func testAccDataSourceAwsEipsConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %q
  }
}

data "aws_eips" "test" {
  tags = {
    Name = "${aws_eip.test.tags["Name"]}"
  }
}
`, rName)
}

func testAccDataSourceAwsEipsClassicConfigTags(rName string) string {
	return fmt.Sprintf(`
resource "aws_eip" "test" {
  vpc = false

  tags = {
    Name = %q
  }
}

data "aws_eips" "test" {
  tags = {
    Name = "${aws_eip.test.tags["Name"]}"
  }
}
`, rName)
}
