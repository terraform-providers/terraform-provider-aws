package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSOutpostsOutpostInstanceTypesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_outposts_outpost_instance_types.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOutpostsOutpostInstanceTypesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostsOutpostInstanceTypesAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostsOutpostInstanceTypesAttributes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["instance_types.#"]; v == "0" {
			return fmt.Errorf("expected at least one instance_types result, got none")
		}

		return nil
	}
}

func testAccAWSOutpostsOutpostInstanceTypesDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost_instance_types" "test" {
  arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}
`
}
