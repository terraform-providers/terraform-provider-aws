package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDedicatedHostDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_dedicated_host.test_data"
	resourceName := "aws_dedicated_host.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDedicatedHostDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_type", resourceName, "instance_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone", resourceName, "availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "instance_family", resourceName, "instance_family"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cores", resourceName, "cores"),
					resource.TestCheckResourceAttrPair(dataSourceName, "total_vcpus", resourceName, "total_vcpus"),
					resource.TestCheckResourceAttrPair(dataSourceName, "sockets", resourceName, "sockets"),
					resource.TestCheckResourceAttrPair(dataSourceName, "host_recovery", resourceName, "host_recovery"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_placement", resourceName, "auto_placement"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.tag2", "test-value2"),
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
	tags = {
    tag1 = "test-value1"
    tag2 = "test-value2"
  }
}


data "aws_dedicated_host" "test_data" {
  host_id = "${aws_dedicated_host.test.id}"
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
