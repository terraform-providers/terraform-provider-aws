package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSDedicatedHostDataSource_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedHostDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_dedicated_host.test_data", "instance_type", "aws_dedicated_host.test", "instance_type"),
					resource.TestCheckResourceAttrPair("data.aws_dedicated_host.test_data", "availability_zone", "aws_dedicated_host.test", "availability_zone"),
					resource.TestCheckResourceAttr("data.aws_dedicated_host.test_data", "host_recovery", "on"),
					resource.TestCheckResourceAttr("data.aws_dedicated_host.test_data", "auto_placement", "on"),
				),
			},
		},
	})
}

const testAccDedicatedHostDataSourceConfig = `
resource "aws_dedicated_host" "test" {
   instance_type = "c5.xlarge"
   availability_zone = "${data.aws_availability_zones.available.names[0]}"
   host_recovery = "on"
   auto_placement = "on"
}


data "aws_dedicated_host" "test_data" {
  host_id = "${aws_dedicated_host.test.id}"
  instance_type = "${aws_dedicated_host.test.instance_type}"
  host_recovery = "${aws_dedicated_host.test.host_recovery}"
  auto_placement = "${aws_dedicated_host.test.auto_placement}"
}

data "aws_availability_zones" "available" {
	state = "available"
  
	filter {
	  name   = "opt-in-status"
	  values = ["opt-in-not-required"]
	}
  }
  
	

`

func testAccCheckHostDestroy(s *terraform.State) error {
	return testAccCheckHostDestroyWithProvider(s, testAccProvider)
}
func testAccCheckHostDestroyWithProvider(s *terraform.State, provider *schema.Provider) error {
	conn := provider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_dedicated_host" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeHosts(&ec2.DescribeHostsInput{
			HostIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			for _, r := range resp.Hosts {
				if r.State != nil && *r.State != "released" {
					return fmt.Errorf("Found unterminated host: %s", r)
				}

			}
		}

		// Verify the error is what we want
		if isAWSErr(err, "InvalidID.NotFound", "") {
			continue
		}

		return err
	}

	return nil
}
