package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSElasticBeanstalkSolutionStackDataSource(t *testing.T) {
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsElasticBeanstalkSolutionStackDataSourceConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsElasticBeanstalkSolutionStackDataSourceID("data.aws_elastic_beanstalk_solution_stack.multi_docker"),
					resource.TestMatchResourceAttr("data.aws_elastic_beanstalk_solution_stack.multi_docker", "name", regexp.MustCompile("^64bit Amazon Linux (.*) Multi-container Docker (.*)$")),
				),
			},
		},
	})
}

func testAccCheckAwsElasticBeanstalkSolutionStackDataSourceID(n string) resource.TestCheckFunc {
	// Wait for solution stacks
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Can't find solution stack data source: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Solution stack data source ID not set")
		}
		return nil
	}
}

const testAccCheckAwsElasticBeanstalkSolutionStackDataSourceConfig = `
data "aws_elastic_beanstalk_solution_stack" "multi_docker" {
	most_recent = true
	name_regex  = "^64bit Amazon Linux (.*) Multi-container Docker (.*)$"
}
`
