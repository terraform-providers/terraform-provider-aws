package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
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
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
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
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAvailabilityZone_LocalZone(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSLocalZoneAvailable(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAvailabilityZoneConfigZoneType("local-zone"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z0-9][a-z0-9-]+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_name"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "local-zone"),
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
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsAvailabilityZone_WavelengthZone(t *testing.T) {
	availabilityZonesDataSourceName := "data.aws_availability_zones.available"
	dataSourceName := "data.aws_availability_zone.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSWavelengthZoneAvailable(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsAvailabilityZoneConfigZoneType("wavelength-zone"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(dataSourceName, "group_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", availabilityZonesDataSourceName, "names.0"),
					resource.TestMatchResourceAttr(dataSourceName, "name_suffix", regexp.MustCompile(`^[a-z0-9][a-z0-9-]+$`)),
					resource.TestCheckResourceAttrSet(dataSourceName, "network_border_group"),
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptedIn),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "parent_zone_name"),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "wavelength-zone"),
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
					resource.TestCheckResourceAttr(dataSourceName, "opt_in_status", ec2.AvailabilityZoneOptInStatusOptInNotRequired),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_id", ""),
					resource.TestCheckResourceAttr(dataSourceName, "parent_zone_name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "region", testAccGetRegion()),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", availabilityZonesDataSourceName, "zone_ids.0"),
					resource.TestCheckResourceAttr(dataSourceName, "zone_type", "availability-zone"),
				),
			},
		},
	})
}

func testAccPreCheckAWSLocalZoneAvailable(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeAvailabilityZonesInput{
		Filters: buildEC2AttributeFilterList(map[string]string{
			"zone-type":     "local-zone",
			"opt-in-status": "opted-in",
		}),
	}

	output, err := conn.DescribeAvailabilityZones(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	if output == nil || len(output.AvailabilityZones) == 0 {
		t.Skip("skipping since no Local Zones are available")
	}
}

func testAccDataSourceAwsAvailabilityZoneConfigAllAvailabilityZones() string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		`
data "aws_availability_zone" "test" {
  all_availability_zones = true
  name                   = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccDataSourceAwsAvailabilityZoneConfigFilter() string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		`
data "aws_availability_zone" "test" {
  filter {
    name   = "zone-name"
    values = [data.aws_availability_zones.available.names[0]]
  }
}
`)
}

func testAccDataSourceAwsAvailabilityZoneConfigName() string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		`
data "aws_availability_zone" "test" {
  name = data.aws_availability_zones.available.names[0]
}
`)
}

func testAccDataSourceAwsAvailabilityZoneConfigZoneId() string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		`
data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}
`)
}

func testAccDataSourceAwsAvailabilityZoneConfigZoneType(zoneType string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "zone-type"
    values = [%[1]q]
  }

  filter {
    name   = "opt-in-status"
    values = ["opted-in"]
  }
}

data "aws_availability_zone" "test" {
  zone_id = data.aws_availability_zones.available.zone_ids[0]
}
`, zoneType)
}
