package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIoTEventConfigurations_basic(t *testing.T) {
	resourceName := "aws_iot_event_configurations.event_configurations"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTEventConfigurationsDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTEventConfigurations_basic(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.THING", "true"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.THING_GROUP", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.THING_TYPE", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.THING_GROUP_MEMBERSHIP", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.THING_GROUP_HIERARCHY", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.THING_TYPE_ASSOCIATION", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.JOB", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.JOB_EXECUTION", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.POLICY", "false"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.CERTIFICATE", "true"),
					resource.TestCheckResourceAttr("aws_iot_event_configurations.event_configurations", "values.CA_CERTIFICATE", "false"),
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

func testAccCheckAWSIoTEventConfigurationsDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_event_configurations" {
			continue
		}

		params := &iot.DescribeEventConfigurationsInput{}
		out, err := conn.DescribeEventConfigurations(params)

		if err != nil {
			return err
		}

		for key, value := range out.EventConfigurations {
			if *value.Enabled != false {
				return fmt.Errorf("Event configuration under key %s not equals false", key)
			}
		}

	}

	return nil
}

func testAccAWSIoTEventConfigurations_basic() string {
	return fmt.Sprintf(`
resource "aws_iot_event_configurations" "event_configurations" {
	values = {
		"THING" = true,
		"THING_GROUP" = false,
		"THING_TYPE" = false,
		"THING_GROUP_MEMBERSHIP" = false,
		"THING_GROUP_HIERARCHY" = false,
		"THING_TYPE_ASSOCIATION" = false,
		"JOB" = false,
		"JOB_EXECUTION" = false,
		"POLICY" = false,
		"CERTIFICATE" = true,
		"CA_CERTIFICATE" = false,
	}
}
`)
}
