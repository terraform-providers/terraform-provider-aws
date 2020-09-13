package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccDataSourceAwsKmsAlias_AwsService(t *testing.T) {
	name := "alias/aws/s3"
	resourceName := "data.aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsKmsAlias_name(name),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsKmsAliasCheckExists(resourceName),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kms", name),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					testAccMatchResourceAttrRegionalARN(resourceName, "target_key_arn", "kms", regexp.MustCompile(`key/[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}`)),
					resource.TestMatchResourceAttr(resourceName, "target_key_id", regexp.MustCompile("^[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}$")),
				),
			},
		},
	})
}

func TestAccDataSourceAwsKmsAlias_CMK(t *testing.T) {
	rInt := acctest.RandInt()
	aliasResourceName := "aws_kms_alias.test"
	datasourceAliasResourceName := "data.aws_kms_alias.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsKmsAlias_CMK(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccDataSourceAwsKmsAliasCheckExists(datasourceAliasResourceName),
					resource.TestCheckResourceAttrPair(datasourceAliasResourceName, "arn", aliasResourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceAliasResourceName, "target_key_arn", aliasResourceName, "target_key_arn"),
					resource.TestCheckResourceAttrPair(datasourceAliasResourceName, "target_key_id", aliasResourceName, "target_key_id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsKmsAliasCheckExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", name)
		}

		return nil
	}
}

func testAccDataSourceAwsKmsAlias_name(name string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "test" {
  name = "%s"
}
`, name)
}

func testAccDataSourceAwsKmsAlias_CMK(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "Terraform acc test"
  deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
  name          = "alias/tf-acc-key-alias-%d"
  target_key_id = aws_kms_key.test.key_id
}

%s
`, rInt, testAccDataSourceAwsKmsAlias_name("${aws_kms_alias.test.name}"))
}
