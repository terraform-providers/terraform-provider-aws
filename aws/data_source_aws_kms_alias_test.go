package aws

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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
					resource.TestMatchResourceAttr(resourceName, "arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:kms:[^:]+:[^:]+:%s$", name))),
					resource.TestCheckResourceAttr(resourceName, "name", name),
					resource.TestMatchResourceAttr(resourceName, "target_key_arn", regexp.MustCompile(fmt.Sprintf("^arn:[^:]+:kms:[^:]+:[^:]+:key/[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}$"))),
					resource.TestMatchResourceAttr(resourceName, "target_key_id", regexp.MustCompile(fmt.Sprintf("^[a-z0-9]{8}-([a-z0-9]{4}-){3}[a-z0-9]{12}$"))),
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
					testAccDataSourceAwsKmsAliasCheckCMKAttributes(aliasResourceName, datasourceAliasResourceName),
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

func testAccDataSourceAwsKmsAliasCheckCMKAttributes(aliasResourceName, datasourceAliasResourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[datasourceAliasResourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", datasourceAliasResourceName)
		}

		kmsKeyRs, ok := s.RootModule().Resources[aliasResourceName]
		if !ok {
			return fmt.Errorf("can't find %s in state", aliasResourceName)
		}

		attr := rs.Primary.Attributes

		if attr["arn"] != kmsKeyRs.Primary.Attributes["arn"] {
			return fmt.Errorf(
				"arn is %s; want %s",
				attr["arn"],
				kmsKeyRs.Primary.Attributes["arn"],
			)
		}

		expectedTargetKeyArnSuffix := fmt.Sprintf("key/%s", kmsKeyRs.Primary.Attributes["target_key_id"])
		if !strings.HasSuffix(attr["target_key_arn"], expectedTargetKeyArnSuffix) {
			return fmt.Errorf(
				"target_key_arn is %s; want suffix %s",
				attr["target_key_arn"],
				expectedTargetKeyArnSuffix,
			)
		}

		if attr["target_key_id"] != kmsKeyRs.Primary.Attributes["target_key_id"] {
			return fmt.Errorf(
				"target_key_id is %s; want %s",
				attr["target_key_id"],
				kmsKeyRs.Primary.Attributes["target_key_id"],
			)
		}

		return nil
	}
}

func testAccDataSourceAwsKmsAlias_name(name string) string {
	return fmt.Sprintf(`
data "aws_kms_alias" "test" {
  name = "%s"
}`, name)
}

func testAccDataSourceAwsKmsAlias_CMK(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
    description = "Terraform acc test"
    deletion_window_in_days = 7
}

resource "aws_kms_alias" "test" {
    name = "alias/tf-acc-key-alias-%d"
    target_key_id = "${aws_kms_key.test.key_id}"
}

%s
`, rInt, testAccDataSourceAwsKmsAlias_name("${aws_kms_alias.test.name}"))
}
