package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSOutpostsOutpostsDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_outposts_outposts.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOutpostsOutpostsDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostsOutpostsAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostsOutpostsAttributes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["arns.#"]; v == "0" {
			return fmt.Errorf("expected at least one arns result, got none")
		}

		if v := rs.Primary.Attributes["ids.#"]; v == "0" {
			return fmt.Errorf("expected at least one ids result, got none")
		}

		return nil
	}
}

func testAccPreCheckAWSOutpostsOutposts(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).outpostsconn

	input := &outposts.ListOutpostsInput{}

	output, err := conn.ListOutposts(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Ensure there is at least one Outpost
	if output == nil || len(output.Outposts) == 0 {
		t.Skip("skipping since no Outposts found")
	}
}

func testAccAWSOutpostsOutpostsDataSourceConfig() string {
	return `
data "aws_outposts_outposts" "test" {}
`
}
