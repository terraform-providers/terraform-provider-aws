package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIotThingGroup_basic(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iot_thing_group.group"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingGroup_basic(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingGroupExists_basic("aws_iot_thing_group.group"),
					resource.TestCheckResourceAttr("aws_iot_thing_group.group", "name", fmt.Sprintf("test_group_%s", rString)),
					resource.TestCheckResourceAttr("aws_iot_thing_group.group", "tags.tagKey", "tagValue"),
					resource.TestCheckResourceAttrSet("aws_iot_thing_group.group", "arn"),
					resource.TestCheckResourceAttrSet("aws_iot_thing_group.group", "version"),
					testAccCheckAWSIotThingGroup_basic,
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

func testAccCheckAWSIotThingGroup_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_group" {
			continue
		}

		params := &iot.DescribeThingGroupInput{
			ThingGroupName: aws.String(rs.Primary.ID),
		}

		out, err := conn.DescribeThingGroup(params)

		if err != nil {
			return err
		}

		properties := out.ThingGroupProperties

		if properties == nil {
			return fmt.Errorf("properties is equal nil")
		}

		if properties.ThingGroupDescription == nil {
			return fmt.Errorf("ThingGroupDescription is equal nil")
		}
		expectedDescription := "test description"
		if *properties.ThingGroupDescription != expectedDescription {
			return fmt.Errorf("ThingGroupDescription %s is not equal expected %s", *properties.ThingGroupDescription, expectedDescription)
		}

		attributePayload := properties.AttributePayload
		if attributePayload == nil {
			return fmt.Errorf("attributePayload is equal nil")
		}
		expectedMerge := false
		if *attributePayload.Merge != expectedMerge {
			return fmt.Errorf("Merge %t is not equal expected %t", *attributePayload.Merge, expectedMerge)
		}
		expectedLen := 2
		if len(attributePayload.Attributes) != expectedLen {
			return fmt.Errorf("len of Attributes %d is not equal expected %d", len(attributePayload.Attributes), expectedLen)
		}

		expectedAttributes := map[string]string{
			"attr1": "val1",
			"attr2": "val2",
		}
		if _, ok := attributePayload.Attributes["attr1"]; !ok {
			return fmt.Errorf("No such key `attr1` in Attributes")
		}
		if *attributePayload.Attributes["attr1"] != expectedAttributes["attr1"] {
			return fmt.Errorf("`attr1` value %s is not equal expected %s", *attributePayload.Attributes["attr1"], expectedAttributes["attr1"])
		}

		if _, ok := attributePayload.Attributes["attr2"]; !ok {
			return fmt.Errorf("No such key `attr2` in Attributes")
		}
		if *attributePayload.Attributes["attr2"] != expectedAttributes["attr2"] {
			return fmt.Errorf("`attr2` value %s is not equal expected %s", *attributePayload.Attributes["attr2"], expectedAttributes["attr2"])
		}

	}

	return nil
}

func TestAccAWSIotThingGroup_parentGroup(t *testing.T) {
	rString := acctest.RandString(5)
	resourceName := "aws_iot_thing_group.group"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingGroup_parentGroup(rString),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSIotThingGroupExists_basic("aws_iot_thing_group.group"),
					resource.TestCheckResourceAttr("aws_iot_thing_group.group", "name", fmt.Sprintf("test_group_%s", rString)),
					resource.TestCheckResourceAttr("aws_iot_thing_group.parent_group", "name", fmt.Sprintf("test_parent_group_%s", rString)),
					resource.TestCheckResourceAttrSet("aws_iot_thing_group.group", "arn"),
					resource.TestCheckResourceAttrSet("aws_iot_thing_group.group", "version"),
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

func testAccCheckAWSIotThingGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_group" {
			continue
		}

		params := &iot.DescribeThingGroupInput{
			ThingGroupName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThingGroup(params)
		if err != nil {
			if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Expected IoT Thing Group to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAWSIotThingGroupExists_basic(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAWSIotThingGroup_basic(rString string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "group" {
	name = "test_group_%[1]s"

	tags = {
		"tagKey" = "tagValue",
	}

	properties {
		description = "test description"
		attributes = {
			"attr1": "val1",
			"attr2": "val2",
		}
		merge = false 
	}
}
`, rString)
}

func testAccAWSIotThingGroup_parentGroup(rString string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "parent_group" {
	name = "test_parent_group_%[1]s"

	properties {
		description = "test description"
		attributes = {
			"attr1": "val1",
			"attr2": "val2",
		}
		merge = false 
	}
}

resource "aws_iot_thing_group" "group" {
	name = "test_group_%[1]s"
	parent_group_name = "${aws_iot_thing_group.parent_group.name}"

	properties {
		description = "test description"
		attributes = {
			"attr1": "val1",
			"attr2": "val2",
		}
		merge = false 
	}
}
`, rString)
}
