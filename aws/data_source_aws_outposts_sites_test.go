package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/outposts"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSOutpostsSitesDataSource_basic(t *testing.T) {
	dataSourceName := "data.aws_outposts_sites.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsSites(t) },
		Providers:    testAccProviders,
		CheckDestroy: nil,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOutpostsSitesDataSourceConfig(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckOutpostsSitesAttributes(dataSourceName),
				),
			},
		},
	})
}

func testAccCheckOutpostsSitesAttributes(dataSourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[dataSourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", dataSourceName)
		}

		if v := rs.Primary.Attributes["ids.#"]; v == "0" {
			return fmt.Errorf("expected at least one ids result, got none")
		}

		return nil
	}
}

func testAccPreCheckAWSOutpostsSites(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).outpostsconn

	input := &outposts.ListSitesInput{}

	output, err := conn.ListSites(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}

	// Ensure there is at least one Site
	if output == nil || len(output.Sites) == 0 {
		t.Skip("skipping since no Sites Outpost found")
	}
}

func testAccAWSOutpostsSitesDataSourceConfig() string {
	return `
data "aws_outposts_sites" "test" {}
`
}
