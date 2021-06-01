package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSSSOAdminPermissionSet_arn(t *testing.T) {
	dataSourceName := "data.aws_ssoadmin_permission_set.test"
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSSSOPermissionSetByArnConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "relay_state", dataSourceName, "relay_state"),
					resource.TestCheckResourceAttrPair(resourceName, "session_duration", dataSourceName, "session_duration"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSSSOAdminPermissionSet_name(t *testing.T) {
	dataSourceName := "data.aws_ssoadmin_permission_set.test"
	resourceName := "aws_ssoadmin_permission_set.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSSSOPermissionSetByNameConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "arn", dataSourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "name", dataSourceName, "name"),
					resource.TestCheckResourceAttrPair(resourceName, "description", dataSourceName, "description"),
					resource.TestCheckResourceAttrPair(resourceName, "relay_state", dataSourceName, "relay_state"),
					resource.TestCheckResourceAttrPair(resourceName, "session_duration", dataSourceName, "session_duration"),
					resource.TestCheckResourceAttrPair(resourceName, "tags", dataSourceName, "tags"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSSSOAdminPermissionSet_nonExistent(t *testing.T) {

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSSSOAdminInstances(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccDataSourceAWSSSOPermissionSetByNameConfig_nonExistent,
				ExpectError: regexp.MustCompile(`not found`),
			},
		},
	})
}

func testAccDataSourceAWSSSOPermissionSetBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_ssoadmin_instances" "test" {}

resource "aws_ssoadmin_permission_set" "test" {
  name         = %[1]q
  description  = %[1]q
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  relay_state  = "https://example.com"

  tags = {
    Key1 = "Value1"
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName)
}

func testAccDataSourceAWSSSOPermissionSetByArnConfig(rName string) string {
	return composeConfig(
		testAccDataSourceAWSSSOPermissionSetBaseConfig(rName),
		`
data "aws_ssoadmin_permission_set" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  arn          = aws_ssoadmin_permission_set.test.arn
}
`)
}

func testAccDataSourceAWSSSOPermissionSetByNameConfig(rName string) string {
	return composeConfig(
		testAccDataSourceAWSSSOPermissionSetBaseConfig(rName),
		`
data "aws_ssoadmin_permission_set" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  name         = aws_ssoadmin_permission_set.test.name
}
`)
}

const testAccDataSourceAWSSSOPermissionSetByNameConfig_nonExistent = `
data "aws_ssoadmin_instances" "test" {}

data "aws_ssoadmin_permission_set" "test" {
  instance_arn = tolist(data.aws_ssoadmin_instances.test.arns)[0]
  name         = "does-not-exist"
}
`
