package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDbEventCategories_basic(t *testing.T) {
	dataSourceName := "data.aws_db_event_categories.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbEventCategoriesConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// These checks are not meant to be exhaustive, as regions have different support.
					// Instead these are generally to indicate that filtering works as expected.
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "availability"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "backup"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "configuration change"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "deletion"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "failover"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "failure"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "low storage"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "maintenance"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "notification"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "recovery"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "restoration"),
				),
			},
		},
	})
}

func TestAccAWSDbEventCategories_SourceType(t *testing.T) {
	dataSourceName := "data.aws_db_event_categories.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsDbEventCategoriesConfigSourceType("db-snapshot"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// These checks are not meant to be exhaustive, as regions have different support.
					// Instead these are generally to indicate that filtering works as expected.
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "creation"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "deletion"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "notification"),
					resource.TestCheckTypeSetElemAttr(dataSourceName, "event_categories.*", "restoration"),
				),
			},
		},
	})
}

func testAccCheckAwsDbEventCategoriesConfig() string {
	return `
data "aws_db_event_categories" "test" {}
`
}

func testAccCheckAwsDbEventCategoriesConfigSourceType(sourceType string) string {
	return fmt.Sprintf(`
data "aws_db_event_categories" "test" {
  source_type = %[1]q
}
`, sourceType)
}
