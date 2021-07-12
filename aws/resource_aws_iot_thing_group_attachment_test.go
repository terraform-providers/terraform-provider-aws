package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIotThingGroupAttachment_basic(t *testing.T) {
	rString := acctest.RandString(8)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupAttachmentDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingGroupAttachmentConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_thing_group_attachment.test_attachment", "thing_name", fmt.Sprintf("test_thing_%s", rString)),
					resource.TestCheckResourceAttr("aws_iot_thing_group_attachment.test_attachment", "thing_group_name", fmt.Sprintf("test_group_%s", rString)),
					resource.TestCheckResourceAttr("aws_iot_thing_group_attachment.test_attachment", "override_dynamics_group", "false"),
					testAccAWSIotThingGroupAttachmentExists_basic(rString),
				),
			},
			{
				ResourceName:      "aws_iot_thing_group_attachment.test_attachment",
				ImportStateIdFunc: testAccAWSIotThingGroupAttachmentImportStateIdFunc("aws_iot_thing_group_attachment.test_attachment"),
				ImportState:       true,
				// We do not have a way to align IDs since the Create function uses resource.PrefixedUniqueId()
				// Failed state verification, resource with ID ROLE-POLICYARN not found
				// ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSIotThingGroupAttachmentExists_basic(rString string) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		conn := testAccProvider.Meta().(*AWSClient).iotconn

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing_group_attachment" {
				continue
			}

			thingName := rs.Primary.Attributes["thing_name"]
			thingGroupName := rs.Primary.Attributes["thing_group_name"]
			hasThingGroup, err := iotThingHasThingGroup(conn, thingName, thingGroupName, "")

			if err != nil {
				return err
			}

			if !hasThingGroup {
				return fmt.Errorf("IoT Thing (%s) is not in IoT Thing Group (%s)", thingName, thingGroupName)
			}

			return nil
		}
		return nil
	}
}

func testAccCheckAWSIotThingGroupAttachmentDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_group_attachment" {
			continue
		}

		thingName := rs.Primary.Attributes["thing_name"]
		thingGroupName := rs.Primary.Attributes["thing_group_name"]

		hasThingGroup, err := iotThingHasThingGroup(conn, thingName, thingGroupName, "")

		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			return nil
		}

		if err != nil {
			return err
		}

		if hasThingGroup {
			return fmt.Errorf("IoT Thing (%s) still in IoT Thing Group (%s)", thingName, thingGroupName)
		}
	}
	return nil
}

func testAccAWSIotThingGroupAttachmentImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["thing_name"], rs.Primary.Attributes["thing_group_name"]), nil
	}
}

func testAccAWSIotThingGroupAttachmentConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test_thing" {
	name = "test_thing_%s"
}

resource "aws_iot_thing_group" "test_thing_group" {
	name = "test_group_%[1]s"
	properties {
		attributes = {
			"attr1": "val1",
		}
		merge = false
	}
}

resource "aws_iot_thing_group_attachment" "test_attachment" {
	thing_name = "${aws_iot_thing.test_thing.name}"
	thing_group_name = "${aws_iot_thing_group.test_thing_group.name}"
	override_dynamics_group = false
}
`, rString)
}
