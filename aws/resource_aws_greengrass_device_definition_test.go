package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSGreengrassDeviceDefinition_basic(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_device_definition.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassDeviceDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassDeviceDefinitionConfig_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("device_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttr("aws_greengrass_device_definition.test", "tags.tagKey", "tagValue"),
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

func TestAccAWSGreengrassDeviceDefinition_DefinitionVersion(t *testing.T) {
	rString := acctest.RandString(8)
	resourceName := "aws_greengrass_device_definition.test"

	device := map[string]interface{}{
		"sync_shadow": false,
		"id":          "device_id",
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSGreengrassDeviceDefinitionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSGreengrassDeviceDefinitionConfig_definitionVersion(rString),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("device_definition_%s", rString)),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					testAccCheckGreengrassDevice_checkDevice(resourceName, device),
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

func testAccCheckGreengrassDevice_checkDevice(n string, expectedDevice map[string]interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Greengrass Device Definition ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).greengrassconn

		getDeviceInput := &greengrass.GetDeviceDefinitionInput{
			DeviceDefinitionId: aws.String(rs.Primary.ID),
		}
		definitionOut, err := conn.GetDeviceDefinition(getDeviceInput)

		if err != nil {
			return err
		}

		getVersionInput := &greengrass.GetDeviceDefinitionVersionInput{
			DeviceDefinitionId:        aws.String(rs.Primary.ID),
			DeviceDefinitionVersionId: definitionOut.LatestVersion,
		}
		versionOut, err := conn.GetDeviceDefinitionVersion(getVersionInput)
		if err != nil {
			return err
		}

		device := versionOut.Definition.Devices[0]
		expectedSyncShadow := expectedDevice["sync_shadow"].(bool)
		if *device.SyncShadow != expectedSyncShadow {
			return fmt.Errorf("Sync Shadow %t is not equal expected %t", *device.SyncShadow, expectedSyncShadow)
		}

		expectedDeviceId := expectedDevice["id"].(string)
		if *device.Id != expectedDeviceId {
			return fmt.Errorf("Device Id %s is not equal expected %s", *device.Id, expectedDeviceId)
		}

		expectedCertArn, err := getAttrFromResourceState("aws_iot_certificate.foo_cert", "arn", s)
		if err != nil {
			return err
		}
		if *device.CertificateArn != expectedCertArn {
			return fmt.Errorf("Device Certificate Arn %s is not equal expected %s", *device.CertificateArn, expectedCertArn)
		}

		expectedThingArn, err := getAttrFromResourceState("aws_iot_thing.test", "arn", s)
		if err != nil {
			return err
		}
		if *device.ThingArn != expectedThingArn {
			return fmt.Errorf("Device Thing Arn %s is not equal expected %s", *device.ThingArn, expectedThingArn)
		}

		return nil
	}
}

func testAccCheckAWSGreengrassDeviceDefinitionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).greengrassconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_greengrass_device_definition" {
			continue
		}

		params := &greengrass.ListDeviceDefinitionsInput{
			MaxResults: aws.String("20"),
		}

		out, err := conn.ListDeviceDefinitions(params)
		if err != nil {
			return err
		}
		for _, definition := range out.Definitions {
			if *definition.Id == rs.Primary.ID {
				return fmt.Errorf("Expected Greengrass Device Definition to be destroyed, %s found", rs.Primary.ID)
			}
		}
	}
	return nil
}

func testAccAWSGreengrassDeviceDefinitionConfig_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_greengrass_device_definition" "test" {
  name = "device_definition_%s"

  tags = {
	"tagKey" = "tagValue"
  } 
}
`, rString)
}

func testAccAWSGreengrassDeviceDefinitionConfig_definitionVersion(rString string) string {
	return fmt.Sprintf(`

resource "aws_iot_thing" "test" {
	name = "%[1]s"
}

resource "aws_iot_certificate" "foo_cert" {
	csr = "${file("test-fixtures/iot-csr.pem")}"
	active = true
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
`, rString)
}
