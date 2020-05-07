package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSSsmParametersByPathDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_parameters_by_path.test"
	prefix := "path"
	otherPrefix := "dummy"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParametersByPathDataSourceConfig(prefix, otherPrefix, "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "arns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "names.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "types.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "values.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsSsmParametersByPathDataSourceConfig(prefix string, otherPrefix string, withDecryption string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test_a" {
	name = "/%s/param-a"
	type = "String"
	value = "TestValueA"
}

resource "aws_ssm_parameter" "test_b" {
	name = "/%s/param-b"
	type = "String"
	value = "TestValueB"
}

resource "aws_ssm_parameter" "test_c" {
	name = "/%s/param-c"
	type = "String"
	value = "TestValueC"
}

data "aws_ssm_parameters_by_path" "test" {
	path = "/${element(split("/", aws_ssm_parameter.test_a.name), 1)}"
	with_decryption = "%s"
}
`, prefix, prefix, otherPrefix, withDecryption)
}
