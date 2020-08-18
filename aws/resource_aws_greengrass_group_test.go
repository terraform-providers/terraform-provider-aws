package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGreengrassGroup_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassGroupConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("group_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
					resource.TestCheckResourceAttr("aws_greengrass_group.test", "tags.tagKey", "tagValue"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSGreengrassGroup_GroupVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassGroupConfig_groupVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("group_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "group_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSGreengrassGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_group" {
			continue
		}

		params := &greengrass.ListGroupsInput{}

		out, err := conn.ListGroups(params)
		if err != nil {
			return err
		}
		for _, groupInfo := range out.Groups {
			if *groupInfo.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Group to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassGroupConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_group" "test" {
  name = "group_%s"

  tags = {
	"tagKey" = "tagValue"
  } 
}
`, rString)
}

func testAccAWSGreengrassGroupConfig_groupVersion(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_connector_definition" "test" {
	name = "connector_definition_%[1]s"
	connector_definition_version {
		connector {
			connector_arn = "arn:aws:greengrass:eu-west-1::/connectors/RaspberryPiGPIO/versions/5"
			id = "connector_id"
			parameters = {
				"key" = "value",
			}
		}
	}
}

resource "aws_greengrass_logger_definition" "test" {
	name = "logger_definition_%[1]s"
	logger_definition_version {
		logger {
			component = "GreengrassSystem"
			id = "test_id"
			type = "FileSystem"
			level = "DEBUG"
			space = 3	
		}
	}
}

resource "aws_iot_thing" "test" {
	name = "%[1]s"
}
resource "aws_iot_certificate" "foo_cert" {
	csr = "${file("test-fixtures/iot-csr.pem")}"
	active = true
}
resource "aws_greengrass_core_definition" "test" {
	name = "core_definition_%[1]s"
	core_definition_version {
		core {
			certificate_arn = "${aws_iot_certificate.foo_cert.arn}"
			id = "core_id"
			sync_shadow = false
			thing_arn = "${aws_iot_thing.test.arn}"
		}
	}
}

resource "aws_greengrass_device_definition" "test" {
	name = "device_definition_%[1]s"
	device_definition_version {
		device {
			certificate_arn = "${aws_iot_certificate.foo_cert.arn}"
			id = "device_id"
			sync_shadow = false
			thing_arn = "${aws_iot_thing.test.arn}"
		}
	}
}


data "aws_caller_identity" "current" {}

resource "aws_greengrass_function_definition" "test" {
	name = "function_definition_%[1]s"
	function_definition_version {
		default_config {
			isolation_mode = "GreengrassContainer"
			run_as {
				gid = 1
				uid = 1
			}
		}
		function {
			function_arn = "arn:aws:lambda:us-west-2:${data.aws_caller_identity.current.account_id}:function:test_lambda_wv8l0glb:test"
			id = "test_id"
			function_configuration {}
		}
	}
}

resource "aws_greengrass_subscription_definition" "test" {
	name = "subscription_definition_%[1]s"
	subscription_definition_version {
		subscription {
			id = "test_id"
			subject = "test_subject"
			source = "arn:aws:iot:eu-west-1:111111111111:thing/Source"
			target = "arn:aws:iot:eu-west-1:222222222222:thing/Target"	
		}
	}
}

resource "aws_greengrass_resource_definition" "test" {
	name = "resource_definition_%[1]s"
	resource_definition_version {
		resource {
			id = "test_id"
			name = "test_name"
			data_container {
				local_device_resource_data {
					source_path = "/dev/source"
					group_owner_setting {
						auto_add_group_owner = false
						group_owner = "user"
					}
				}
			}
		}
	}
}

resource "aws_greengrass_group" "test" {
  name = "group_%[1]s"

  group_version {
	core_definition_version_arn = "${aws_greengrass_core_definition.test.latest_definition_version_arn}" 
	connector_definition_version_arn = "${aws_greengrass_connector_definition.test.latest_definition_version_arn}"
	function_definition_version_arn = "${aws_greengrass_function_definition.test.latest_definition_version_arn}"
	subscription_definition_version_arn = "${aws_greengrass_subscription_definition.test.latest_definition_version_arn}"
	logger_definition_version_arn = "${aws_greengrass_logger_definition.test.latest_definition_version_arn}"
	device_definition_version_arn = "${aws_greengrass_device_definition.test.latest_definition_version_arn}"
	resource_definition_version_arn = "${aws_greengrass_resource_definition.test.latest_definition_version_arn}"
  }
}
`, rString)
}
