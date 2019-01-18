package aws

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
data "aws_availability_zones" "availability_zones" { }
`

const testAccCheckAwsAvailabilityZonesStateConfig = `
data "aws_availability_zones" "state_filter" {
	state = "available"
}
`
