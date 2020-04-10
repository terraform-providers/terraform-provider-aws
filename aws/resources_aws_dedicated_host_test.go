package aws

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDedicatedHostDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedHostDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_dedicated_host.test_data", "instance_type", "c5.18xlarge"),
					resource.TestCheckResourceAttr("data.aws_dedicated_host.test_data", "availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr("data.aws_dedicated_host.test_data", "host_recovery", "on"),
					resource.TestCheckResourceAttr("data.aws_dedicated_host.test_data", "auto_placement", "on"),
				),
			},
		},
	})
}

const testAccDedicatedHostDataSourceConfig = `
resource "aws_dedicated_host" "test" {
   #us-west-2
   instance_type = "c5.18xlarge"
   availability_zone = "us-west-2a"
   host_recovery = "on"
   auto_placement = "on"
}

data "aws_dedicated_host" "test_data" {
   host_id="${aws_dedicated_host.test.id}"
   instance_type = "c5.18xlarge"
   availability_zone = "us-west-2a"
   host_recovery = "on"
   auto_placement = "on"
}

`
