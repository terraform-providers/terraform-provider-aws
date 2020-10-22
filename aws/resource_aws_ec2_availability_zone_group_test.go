package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEc2AvailabilityZoneGroup_OptInStatus(t *testing.T) {
	resourceName := "aws_ec2_availability_zone_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2AvailabilityZoneGroup(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(ec2.AvailabilityZoneOptInStatusOptedIn),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// InvalidOptInStatus: Opting out of Local Zones is not currently supported. Contact AWS Support for additional assistance.
			/*
				{
					Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(ec2.AvailabilityZoneOptInStatusNotOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusNotOptedIn),
					),
				},
				{
					Config: testAccEc2AvailabilityZoneGroupConfigOptInStatus(ec2.AvailabilityZoneOptInStatusOptedIn),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(resourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					),
				},
			*/
		},
	})
}

func testAccPreCheckAWSEc2AvailabilityZoneGroup(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeAvailabilityZonesInput{
		AllAvailabilityZones: aws.Bool(true),
		Filters: []*ec2.Filter{
			{
				Name: aws.String("opt-in-status"),
				Values: aws.StringSlice([]string{
					ec2.AvailabilityZoneOptInStatusNotOptedIn,
					ec2.AvailabilityZoneOptInStatusOptedIn,
				}),
			},
		},
	}

	output, err := conn.DescribeAvailabilityZones(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if output == nil || len(output.AvailabilityZones) == 0 || output.AvailabilityZones[0] == nil {
		t.Skipf("skipping acceptance testing: no opt-in EC2 Availability Zone Groups found")
	}
}

func testAccEc2AvailabilityZoneGroupConfigOptInStatus(optInStatus string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "test" {
  all_availability_zones = true

  # Filter to one Availability Zone Group per Region as Local Zones become available
  # e.g. ensure there are not two us-west-2-XXX when adding to this list
  filter {
    name = "group-name"
    values = [
      "us-west-2-lax-1",
    ]
  }
}

resource "aws_ec2_availability_zone_group" "test" {
  # The above group-name filter should ensure one Availability Zone Group per Region
  group_name    = tolist(data.aws_availability_zones.test.group_names)[0]
  opt_in_status = %[1]q
}
`, optInStatus)
}
