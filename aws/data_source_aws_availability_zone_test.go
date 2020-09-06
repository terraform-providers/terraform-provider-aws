package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAwsAvailabilityZone_AllAvailabilityZones(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAvailabilityZoneConfigAllAvailabilityZones(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", "opt-in-not-required"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAvailabilityZone_Filter(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAvailabilityZoneConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", "opt-in-not-required"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAvailabilityZone_Name(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAvailabilityZoneConfigName(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", "opt-in-not-required"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAvailabilityZone_ZoneId(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAvailabilityZoneConfigZoneId(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "group_name", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z]$`)),
					resource.TestCheckResourceAttr(dataSourceName, "network_border_group", testAccGetRegion()),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", "opt-in-not-required"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
				),
			},
		},
	})
}

func testAccDataSourceAwsAvailabilityZoneConfigAllAvailabilityZones() string {
	return testAccAvailableAZsNoOptInConfig() + `
data "aws_availability_zone" "test" {
  all_availability_zones = true
  name                   = data.aws_availability_zones.available.names[0]
}
`
}

func testAccDataSourceAwsAvailabilityZoneConfigFilter() string {
	return testAccAvailableAZsNoOptInConfig() + `
data "aws_availability_zone" "test" {
  filter {
    name   = "zone-name"
    values = [data.aws_availability_zones.available.names[0]]
  }
}
`
}

func testAccDataSourceAwsAvailabilityZoneConfigName() string {
	return testAccAvailableAZsNoOptInConfig() + `
data "aws_availability_zone" "test" {
  name = data.aws_availability_zones.available.names[0]
}
`
}

func testAccDataSourceAwsAvailabilityZoneConfigZoneId() string {
	return testAccAvailableAZsNoOptInConfig() + `
data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}
`
}
