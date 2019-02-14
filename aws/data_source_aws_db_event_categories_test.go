package aws

import (
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strconv"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSDbEventCategories_basic(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbEventCategoriesConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccAwsDbEventCategoriesAttrCheck("data.aws_db_event_categories.example",
						completeEventCategoriesList),
				),
			},
		},
	})
}

func TestAccAWSDbEventCategories_sourceType(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbEventCategoriesConfig_sourceType,
				Check: resource.ComposeTestCheckFunc(
					testAccAwsDbEventCategoriesAttrCheck("data.aws_db_event_categories.example",
						DbSnapshotEventCategoriesList),
				),
			},
		},
	})
}

func testAccAwsDbEventCategoriesAttrCheck(n string, expected []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find DB Event Categories: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("DB Event Categories resource ID not set.")
		}

		actual, err := testAccCheckAwsDbEventCategoriesBuild(rs.Primary.Attributes)
		if err != nil {
			return err
		}

		sort.Strings(actual)
		sort.Strings(expected)
		if !reflect.DeepEqual(expected, actual) {
			return fmt.Errorf("DB Event Categories not matched: expected %v, got %v", expected, actual)
		}

		return nil
	}
}

func testAccCheckAwsDbEventCategoriesBuild(attrs map[string]string) ([]string, error) {
	v, ok := attrs["event_categories.#"]
	if !ok {
		return nil, fmt.Errorf("DB Event Categories list is missing.")
	}

	qty, err := strconv.Atoi(v)
	if err != nil {
		return nil, err
	}
	if qty < 1 {
		return nil, fmt.Errorf("No DB Event Categories found.")
	}

	var eventCategories []string
	for k, v := range attrs {
		matched, _ := regexp.MatchString("event_categories.[0-9]+", k)
		if matched {
			eventCategories = append(eventCategories, v)
		}
	}

	return eventCategories, nil
}

var testAccCheckAwsDbEventCategoriesConfig = `
data "aws_db_event_categories" "example" {}
`

var completeEventCategoriesList = []string{
	"notification",
	"deletion",
	"failover",
	"maintenance",
	"availability",
	"read replica",
	"failure",
	"configuration change",
	"recovery",
	"low storage",
	"backup",
	"creation",
	"backtrack",
	"restoration",
}

var testAccCheckAwsDbEventCategoriesConfig_sourceType = `
data "aws_db_event_categories" "example" {
	source_type = "db-snapshot"
}
`

var DbSnapshotEventCategoriesList = []string{
	"notification",
	"deletion",
	"creation",
	"restoration",
}
