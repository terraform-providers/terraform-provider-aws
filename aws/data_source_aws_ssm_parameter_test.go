package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSSsmParameterDataSource_basic(t *testing.T) {
	resourceName := "data.aws_ssm_parameter.test"
	name := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfig(name, "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_ssm_parameter.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfig(name, "true"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_ssm_parameter.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "true"),
				),
			},
		},
	})
}

func TestAccAWSSsmParameterDataSource_fullPath(t *testing.T) {
	resourceName := "data.aws_ssm_parameter.test"
	name := acctest.RandomWithPrefix("/tf-acc-test/tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAwsSsmParameterDataSourceConfig(name, "false"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", "aws_ssm_parameter.test", "arn"),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestCheckResourceAttr(resourceName, "type", "String"),
					resource.TestCheckResourceAttr(resourceName, "value", "TestValue"),
					resource.TestCheckResourceAttr(resourceName, "with_decryption", "false"),
				),
			},
		},
	})
}

func testAccCheckAwsSsmParameterDataSourceConfig(name string, withDecryption string) string {
	return fmt.Sprintf(`
resource "aws_ssm_parameter" "test" {
  name  = "%s"
  type  = "String"
  value = "TestValue"
}

data "aws_ssm_parameter" "test" {
  name            = aws_ssm_parameter.test.name
  with_decryption = %s
}
`, name, withDecryption)
}
