package aws

import (
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"testing"
)

func TestAccAWSIotIndexingConfig_empty(t *testing.T) {
	// Note: These tests cannot be parallelized because they access a shared resource
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotIndexingConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotIndexingConfig_empty,
			},
			{
				ResourceName:      "aws_iot_indexing_config.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIotIndexingConfig_basic(t *testing.T) {
	// Note: These tests cannot be parallelized because they access a shared resource
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotIndexingConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotIndexingConfig,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_indexing_config.test", "thing_group_indexing_enabled", "true"),
					resource.TestCheckResourceAttr("aws_iot_indexing_config.test", "thing_connectivity_indexing_enabled", "true"),
					resource.TestCheckResourceAttr("aws_iot_indexing_config.test", "thing_indexing_mode", "REGISTRY_AND_SHADOW"),
				),
			},
		},
	})
}

func testAccCheckAWSIotIndexingConfigDestroy(s *terraform.State) error {
	// Intentionally noop
	// as there is no API method for deleting or resetting IoT indexing configuration
	return nil
}

const testAccAWSIotIndexingConfig_empty = `
resource "aws_iot_indexing_config" "test" {
}
`

const testAccAWSIotIndexingConfig = `
resource "aws_iot_indexing_config" "test" {
  thing_group_indexing_enabled = true
  thing_connectivity_indexing_enabled = true
  thing_indexing_mode = "REGISTRY_AND_SHADOW"
}
`
