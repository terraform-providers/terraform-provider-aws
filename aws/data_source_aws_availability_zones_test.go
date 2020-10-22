package aws

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAvailabilityZonesSort(t *testing.T) {
	azs := []*ec2.AvailabilityZone{
		{
			ZoneName: aws.String("name_YYY"),
			ZoneId:   aws.String("id_YYY"),
		},
		{
			ZoneName: aws.String("name_AAA"),
			ZoneId:   aws.String("id_AAA"),
		},
		{
			ZoneName: aws.String("name_ZZZ"),
			ZoneId:   aws.String("id_ZZZ"),
		},
		{
			ZoneName: aws.String("name_BBB"),
			ZoneId:   aws.String("id_BBB"),
		},
	}
	sort.Slice(azs, func(i, j int) bool {
		return aws.StringValue(azs[i].ZoneName) < aws.StringValue(azs[j].ZoneName)
	})

	cases := []struct {
		Index    int
		ZoneName string
		ZoneId   string
	}{
		{
			Index:    0,
			ZoneName: "name_AAA",
			ZoneId:   "id_AAA",
		},
		{
			Index:    1,
			ZoneName: "name_BBB",
			ZoneId:   "id_BBB",
		},
		{
			Index:    2,
			ZoneName: "name_YYY",
			ZoneId:   "id_YYY",
		},
		{
			Index:    3,
			ZoneName: "name_ZZZ",
			ZoneId:   "id_ZZZ",
		},
	}
	for _, tc := range cases {
		az := azs[tc.Index]
		if aws.StringValue(az.ZoneName) != tc.ZoneName {
			t.Fatalf("AvailabilityZones index %d got zone name %s, expected %s", tc.Index, aws.StringValue(az.ZoneName), tc.ZoneName)
		}
		if aws.StringValue(az.ZoneId) != tc.ZoneId {
			t.Fatalf("AvailabilityZones index %d got zone ID %s, expected %s", tc.Index, aws.StringValue(az.ZoneId), tc.ZoneId)
		}
	}
}

func TestAccAWSAvailabilityZones_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAvailabilityZonesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAvailabilityZonesMeta("data.aws_availability_zones.availability_zones"),
				),
			},
		},
	})
}

func TestAccAWSAvailabilityZones_AllAvailabilityZones(t *testing.T) {
	dataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAvailabilityZonesConfigAllAvailabilityZones(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAvailabilityZonesMeta(dataSourceName),
				),
			},
		},
	})
}

func TestAccAWSAvailabilityZones_Filter(t *testing.T) {
	dataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAvailabilityZonesConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAvailabilityZonesMeta(dataSourceName),
				),
			},
		},
	})
}

func TestAccAWSAvailabilityZones_ExcludeNames(t *testing.T) {
	allDataSourceName := "data.aws_availability_zones.all"
	excludeDataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAvailabilityZonesConfigExcludeNames(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName),
				),
			},
		},
	})
}

func TestAccAWSAvailabilityZones_ExcludeZoneIds(t *testing.T) {
	allDataSourceName := "data.aws_availability_zones.all"
	excludeDataSourceName := "data.aws_availability_zones.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAvailabilityZonesConfigExcludeZoneIds(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName),
				),
			},
		},
	})
}

func TestAccAWSAvailabilityZones_stateFilter(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsAvailabilityZonesStateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAvailabilityZoneState("data.aws_availability_zones.state_filter"),
				),
			},
		},
	})
}

func testAccCheckAwsAvailabilityZonesMeta(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AZ resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AZ resource ID not set.")
		}

		actual, err := testAccCheckAwsAvailabilityZonesBuildAvailable(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		expected := actual
		sort.Strings(expected)
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("AZs not sorted - expected %v, got %v", expected, actual)
		}
		return nil
	}
}

func testAccCheckAwsAvailabilityZonesExcluded(allDataSourceName, excludeDataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		allResourceState, ok := s.RootModule().Resources[allDataSourceName]
		if !ok {
			return fmt.Errorf("Resource does not exist: %s", allDataSourceName)
		}

		excludeResourceState, ok := s.RootModule().Resources[excludeDataSourceName]
		if !ok {
			return fmt.Errorf("Resource does not exist: %s", excludeDataSourceName)
		}

		for _, attribute := range []string{"names.#", "zone_ids.#"} {
			allValue, ok := allResourceState.Primary.Attributes[attribute]

			if !ok {
				return fmt.Errorf("cannot find %s in %s resource state attributes: %+v", attribute, allDataSourceName, allResourceState.Primary.Attributes)
			}

			excludeValue, ok := excludeResourceState.Primary.Attributes[attribute]

			if !ok {
				return fmt.Errorf("cannot find %s in %s resource state attributes: %+v", attribute, excludeDataSourceName, excludeResourceState.Primary.Attributes)
			}

			if allValue == excludeValue {
				return fmt.Errorf("expected %s attribute value difference, got: %s", attribute, allValue)
			}
		}

		return nil
	}
}

func testAccCheckAwsAvailabilityZoneState(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find AZ resource: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("AZ resource ID not set.")
		}

		if _, ok := rs.Primary.Attributes["state"]; !ok {
			return fmt.Errorf("AZs state filter is missing, should be set.")
		}

		_, err := testAccCheckAwsAvailabilityZonesBuildAvailable(rs.Primary.Attributes)
		return err
	}
}

func testAccCheckAwsAvailabilityZonesBuildAvailable(attrs map[string]string) ([]string, error) {
	groupNames, groupNamesOk := attrs["group_names.#"]

	if !groupNamesOk {
		return nil, fmt.Errorf("Availability Zone Group names list is missing.")
	}

	groupNamesQty, err := strconv.Atoi(groupNames)

	if err != nil {
		return nil, err
	}

	if groupNamesQty < 1 {
		return nil, fmt.Errorf("No Availability Zone Groups found in region, this is probably a bug.")
	}

	v, ok := attrs["names.#"]
	if !ok {
		return nil, fmt.Errorf("Available AZ name list is missing.")
	}
	qty, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	if qty < 1 {
		return nil, fmt.Errorf("No AZs found in region, this is probably a bug.")
	}
	_, ok = attrs["zone_ids.#"]
	if !ok {
		return nil, fmt.Errorf("Available AZ ID list is missing.")
	}
	zones := make([]string, qty)
	for n := range zones {
		zone, ok := attrs["names."+strconv.Itoa(n)]
		if !ok {
			return nil, fmt.Errorf("AZ list corrupt, this is definitely a bug.")
		}
		zones[n] = zone
	}
	return zones, nil
}

const testAccCheckAwsAvailabilityZonesConfig = `
data "aws_availability_zones" "availability_zones" {}
`

func testAccCheckAwsAvailabilityZonesConfigAllAvailabilityZones() string {
	return `
data "aws_availability_zones" "test" {
  all_availability_zones = true
}
`
}

func testAccCheckAwsAvailabilityZonesConfigFilter() string {
	return `
data "aws_availability_zones" "test" {
  filter {
    name   = "state"
    values = ["available"]
  }
}
`
}

func testAccCheckAwsAvailabilityZonesConfigExcludeNames() string {
	return `
data "aws_availability_zones" "all" {}

data "aws_availability_zones" "test" {
  exclude_names = [data.aws_availability_zones.all.names[0]]
}
`
}

func testAccCheckAwsAvailabilityZonesConfigExcludeZoneIds() string {
	return `
data "aws_availability_zones" "all" {}

data "aws_availability_zones" "test" {
  exclude_zone_ids = [data.aws_availability_zones.all.zone_ids[0]]
}
`
}

const testAccCheckAwsAvailabilityZonesStateConfig = `
data "aws_availability_zones" "state_filter" {
  state = "available"
}
`
