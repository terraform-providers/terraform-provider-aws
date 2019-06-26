package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccAWSSsmParametersDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_parameters.test"
	name := "/path/parameter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParametersDataSourceConfig(name),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "parameters.0.name", "/path/parameter/value1"),
					resource.TestCheckResourceAttr(resourceName, "parameters.0.value", "TestValue1"),

					resource.TestCheckResourceAttr(resourceName, "parameters.1.name", "/path/parameter/value2"),
					resource.TestCheckResourceAttr(resourceName, "parameters.1.value", "TestValue2"),
				),
			},
		},
	})
}

func testAccCheckAwsSsmParametersDataSourceConfig(name string) string {
	return fmt.Sprintf(`
locals {
	prefix = "%s"
}

resource "aws_ssm_parameter" "value1" {
	name = "${local.prefix}/value1"
	type = "String"
	value = "TestValue1"
}

resource "aws_ssm_parameter" "value2" {
	name = "${local.prefix}/value2"
	type = "String"
	value = "TestValue2"
}

data "aws_ssm_parameters" "test" {
	path = "${local.prefix}"
	depends_on = ["aws_ssm_parameter.value1", "aws_ssm_parameter.value2"]
}
`, name)
}
