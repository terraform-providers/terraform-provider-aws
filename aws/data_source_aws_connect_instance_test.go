package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsConnectInstanceDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("datasource-test-terraform")
	resourceName := "aws_connect_instance.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config:      testAccAwsConnectInstanceDataSourceConfig_nonExistentId,
				ExpectError: regexp.MustCompile(`error getting Connect Instance by instance_id`),
			},
			{
				Config:      testAccAwsConnectInstanceDataSourceConfig_nonExistentAlias,
				ExpectError: regexp.MustCompile(`error finding Connect Instance by instance_alias`),
			},
			{
				Config: testAccAwsConnectInstanceDataSourceConfigBasic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "connect", regexp.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "identity_management_type"),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexp.MustCompile(rName)),
					resource.TestCheckResourceAttrSet(resourceName, "inbound_calls_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "outbound_calls_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_logs_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_lens_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "auto_resolve_best_voices"),
					resource.TestCheckResourceAttrSet(resourceName, "use_custom_tts_voices"),
					resource.TestCheckResourceAttrSet(resourceName, "early_media_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
				),
			},
		},
	})
}

func TestAccAwsConnectInstanceDataSource_alias(t *testing.T) {
	rName := acctest.RandomWithPrefix("datasource-test-terraform")
	resourceName := "aws_connect_instance.foo"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectInstanceDataSourceConfigAlias(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "connect", regexp.MustCompile(`instance/.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "created_time"),
					resource.TestCheckResourceAttrSet(resourceName, "identity_management_type"),
					resource.TestMatchResourceAttr(resourceName, "instance_alias", regexp.MustCompile(rName)),
					resource.TestCheckResourceAttrSet(resourceName, "inbound_calls_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "outbound_calls_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_flow_logs_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "contact_lens_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "auto_resolve_best_voices"),
					resource.TestCheckResourceAttrSet(resourceName, "use_custom_tts_voices"),
					resource.TestCheckResourceAttrSet(resourceName, "early_media_enabled"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "service_role"),
				),
			},
		},
	})
}

const testAccAwsConnectInstanceDataSourceConfig_nonExistentId = `
data "aws_connect_instance" "foo" {
  instance_id = "97afc98d-101a-ba98-ab97-ae114fc115ec"
}
`

const testAccAwsConnectInstanceDataSourceConfig_nonExistentAlias = `
data "aws_connect_instance" "foo" {
  instance_alias = "tf-acc-test-does-not-exist"
}
`

func testAccAwsConnectInstanceDataSourceConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  instance_alias = %[1]q
}

data "aws_connect_instance" "foo" {
  instance_id = aws_connect_instance.foo.id
}
`, rName)
}

func testAccAwsConnectInstanceDataSourceConfigAlias(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "foo" {
  instance_alias = %[1]q
}

data "aws_connect_instance" "foo" {
  instance_alias = aws_connect_instance.foo.instance_alias
}
`, rName)
}
