package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsAmiIds_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAmiIdsConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami_ids.ubuntu"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAmiIds_sorted(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAmiIdsConfig_sorted(false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEbsSnapshotDataSourceID("data.aws_ami_ids.test"),
					resource.TestCheckResourceAttr("data.aws_ami_ids.test", "ids.#", "2"),
					resource.TestCheckResourceAttrPair(
						"data.aws_ami_ids.test", "ids.0",
						"data.aws_ami.amzn_linux_2018_03", "id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_ami_ids.test", "ids.1",
						"data.aws_ami.amzn_linux_2016_09_0", "id"),
				),
			},
			// Make sure when sort_ascending is set, they're sorted in the inverse order
			// it uses the same config / dataset as above so no need to verify the other
			// bits
			{
				Config: testAccDataSourceAwsAmiIdsConfig_sorted(true),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(
						"data.aws_ami_ids.test", "ids.0",
						"data.aws_ami.amzn_linux_2016_09_0", "id"),
					resource.TestCheckResourceAttrPair(
						"data.aws_ami_ids.test", "ids.1",
						"data.aws_ami.amzn_linux_2018_03", "id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAmiIds_empty(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAmiIdsConfig_empty,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAmiDataSourceID("data.aws_ami_ids.empty"),
					resource.TestCheckResourceAttr("data.aws_ami_ids.empty", "ids.#", "0"),
				),
			},
		},
	})
}

const testAccDataSourceAwsAmiIdsConfig_basic = `
data "aws_ami_ids" "ubuntu" {
    owners = ["099720109477"]

    filter {
        name   = "name"
        values = ["ubuntu/images/ubuntu-*-*-amd64-server-*"]
    }
}
`

func testAccDataSourceAwsAmiIdsConfig_sorted(sort_ascending bool) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn_linux_2016_09_0" {
  owners = ["amazon"]
  filter {
    name   = "name"
    values = ["amzn-ami-hvm-2016.09.0.20161028-x86_64-gp2"]
  }
}

data "aws_ami" "amzn_linux_2018_03" {
  owners = ["amazon"]
  filter {
    name   = "name"
    values = ["amzn-ami-hvm-2018.03.0.20180811-x86_64-gp2"]
  }
}

data "aws_ami_ids" "test" {
  owners     = ["amazon"]
  filter {
    name   = "name"
    values = ["amzn-ami-hvm-2018.03.0.20180811-x86_64-gp2", "amzn-ami-hvm-2016.09.0.20161028-x86_64-gp2"]
  }
  sort_ascending = "%t"
}
`, sort_ascending)
}

const testAccDataSourceAwsAmiIdsConfig_empty = `
data "aws_ami_ids" "empty" {
  filter {
    name   = "name"
    values = []
  }
}
`
