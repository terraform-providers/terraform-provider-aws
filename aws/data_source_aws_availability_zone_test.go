package aws

import (
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsAvailabilityZone(t *testing.T) {
	ds1ResourceName := "data.aws_availability_zone.by_name"
	ds2ResourceName := "data.aws_availability_zone.by_zone_id"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAvailabilityZoneConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(ds1ResourceName, "name", "us-west-2a"),
					resource.TestCheckResourceAttr(ds1ResourceName, "name_suffix", "a"),
					resource.TestCheckResourceAttr(ds1ResourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttrSet(ds1ResourceName, "zone_id"),

					resource.TestCheckResourceAttr(ds2ResourceName, "name", "us-west-2a"),
					resource.TestCheckResourceAttr(ds2ResourceName, "name_suffix", "a"),
					resource.TestCheckResourceAttr(ds2ResourceName, "region", "us-west-2"),
					resource.TestCheckResourceAttrPair(ds2ResourceName, "zone_id", ds1ResourceName, "zone_id"),
				),
			},
		},
	})
}

const testAccDataSourceAwsAvailabilityZoneConfig = `
provider "aws" {
  region = "us-west-2"
}

data "aws_availability_zone" "by_name" {
  name = "us-west-2a"
}

data "aws_availability_zone" "by_zone_id" {
  zone_id = "${data.aws_availability_zone.by_name.zone_id}"
}
`
