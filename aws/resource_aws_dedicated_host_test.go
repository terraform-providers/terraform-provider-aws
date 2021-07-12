package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

const testAccAWSDefaultDedicatedHostConfigBasic = `
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
`
const testAccAWSDefaultDedicatedHostConfigBasicUpdate = `
resource "aws_dedicated_host" "test" {
   instance_type = "c5.xlarge"
   availability_zone = "${data.aws_availability_zones.available.names[0]}"
   host_recovery = "on"
   auto_placement = "on"
	tags = {
    tag1 = "test-value1-%$#!."
    tag2 = "test-value2-changed"
	tag3 = "test-value3"
  }
}
`

func init() {
	resource.AddTestSweepers("aws_dedicated_host", &resource.Sweeper{
		Name: "aws_dedicated_host",
		F:    testSweepDedicatedHosts,
	})
}
func testSweepDedicatedHosts(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeHostsInput{}
	var sweeperErrs *multierror.Error

	err = conn.DescribeHostsPages(input, func(page *ec2.DescribeHostsOutput, lastPage bool) bool {
		for _, host := range page.Hosts {
			if host == nil {
				continue
			}
			id := aws.StringValue(host.HostId)
			input := &ec2.ReleaseHostsInput{
				HostIds: []*string{host.HostId},
			}

			log.Printf("[INFO] Deleting EC2 dedicated host: %s", id)

			// Handle EC2 eventual consistency
			err := resource.Retry(1*time.Minute, func() *resource.RetryError {
				_, err := conn.ReleaseHosts(input)
				if isAWSErr(err, "DependencyViolation", "") {
					return resource.RetryableError(err)
				}
				if err != nil {
					return resource.NonRetryableError(err)
				}
				return nil
			})

			if isResourceTimeoutError(err) {
				_, err = conn.ReleaseHosts(input)
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error releasing EC2 dedicated host (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Host sweep for %s: %s", region, err)
		return nil
	}

	if err != nil {
		return fmt.Errorf("Error describing hosts: %s", err)
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSDedicatedHost_basic(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_dedicated_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultDedicatedHostConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "host_recovery", "on"),
					resource.TestCheckResourceAttr(resourceName, "auto_placement", "on"),
					resource.TestCheckResourceAttr(resourceName, "instance_type", "c5.xlarge"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckHostExists(n string, host *ec2.Host) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Host ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		DescribeHostOpts := &ec2.DescribeHostsInput{
			HostIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeHosts(DescribeHostOpts)
		if err != nil {
			return err
		}
		if len(resp.Hosts) == 0 || resp.Hosts[0] == nil {
			return fmt.Errorf("Host not found")
		}

		*host = *resp.Hosts[0]

		return nil
	}
}
func TestAccAWSDedicatedHost_tags(t *testing.T) {
	var host ec2.Host
	resourceName := "aws_dedicated_host.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckHostDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultDedicatedHostConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "test-value1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-value2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSDefaultDedicatedHostConfigBasicUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckHostExists(resourceName, &host),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag1", "test-value1-%$#!."),
					resource.TestCheckResourceAttr(resourceName, "tags.tag2", "test-value2-changed"),
					resource.TestCheckResourceAttr(resourceName, "tags.tag3", "test-value3"),
				),
			},
		},
	})
}
