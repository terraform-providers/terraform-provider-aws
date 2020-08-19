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

func TestAccAWSIotThingGroup_base(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckNoResourceAttr(resourceName, "parent_group_name"),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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

func TestAccAWSIotThingGroup_full(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	parentThingGroupName := thingGroupName + "_parent"
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{ // BASE
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckNoResourceAttr(resourceName, "parent_group_name"),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // UPDATE full
				Config: testAccAWSIotThingGroupConfig_full(thingGroupName, parentThingGroupName, "7", "this is my thing group", "myTag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", parentThingGroupName),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Answer", "7"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "this is my thing group"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "myTag"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{ // DELETE full
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", ""),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttrSet(resourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
		},
	})
}

func TestAccAWSIotThingGroup_name(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{ // CREATE
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // UPDATE
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName + "_updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "name", thingGroupName+"_updated"),
				),
			},
		},
	})
}

func TestAccAWSIotThingGroup_tags(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{ // BASE
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // CREATE Tags
				Config: testAccAWSIotThingGroupConfig_withTags(thingGroupName, "myTag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "myTag"),
				),
			},
			{ // UPDATE Tags
				Config: testAccAWSIotThingGroupConfig_withTags(thingGroupName, "myUpdatedTag"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.tagKey", "myUpdatedTag"),
				),
			},
			{ // DELETE Tags
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSIotThingGroup_propsAttr(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{ // BASE
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // CREATE Properties
				Config: testAccAWSIotThingGroupConfig_withPropAttr(thingGroupName, "42"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Answer", "42"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", ""),
				),
			},
			{ // UPDATE Properties
				Config: testAccAWSIotThingGroupConfig_withPropAttr(thingGroupName, "7"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Answer", "7"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", ""),
				),
			},
			{ // DELETE Properties
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
				),
			},
		},
	})
}

func TestAccAWSIotThingGroup_propsDesc(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{ // BASE
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // CREATE Properties
				Config: testAccAWSIotThingGroupConfig_withPropDesc(thingGroupName, "this is my thing group"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties.0.attributes"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "this is my thing group"),
				),
			},
			{ // UPDATE Properties
				Config: testAccAWSIotThingGroupConfig_withPropDesc(thingGroupName, "this is my updated thing group"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties.0.attributes"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "this is my updated thing group"),
				),
			},
			{ // DELETE Properties
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
				),
			},
		},
	})
}

func TestAccAWSIotThingGroup_propsAll(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{ // BASE
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // CREATE Properties
				Config: testAccAWSIotThingGroupConfig_withPropAll(thingGroupName, "42", "this is my thing group"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Answer", "42"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "this is my thing group"),
				),
			},
			{ // UPDATE Properties
				Config: testAccAWSIotThingGroupConfig_withPropAll(thingGroupName, "7", "this is my updated thing group"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attributes.Answer", "7"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "this is my updated thing group"),
				),
			},
			{ // DELETE Properties
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "properties"),
				),
			},
		},
	})
}

func TestAccAWSIotThingGroup_parent(t *testing.T) {
	var thingGroup iot.DescribeThingGroupOutput
	rString := acctest.RandString(8)
	thingGroupName := fmt.Sprintf("tf_acc_thing_group_%s", rString)
	parentThingGroupName := thingGroupName + "_parent"
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckNoResourceAttr(resourceName, "parent_group_name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // CREATE parent_group_name
				Config: testAccAWSIotThingGroupConfig_withParent(thingGroupName, parentThingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", parentThingGroupName),
				),
			},
			{ // UPDATE parent_group_name
				Config: testAccAWSIotThingGroupConfig_withParent(thingGroupName, parentThingGroupName+"_updated"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", parentThingGroupName+"_updated"),
				),
			},
			{ // DELETE parent_group_name
				Config: testAccAWSIotThingGroupConfig_base(thingGroupName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotThingGroupExists(resourceName, &thingGroup),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", ""),
				),
			},
		},
	})
}

func testAccCheckIotThingGroupExists(n string, thing *iot.DescribeThingGroupOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no IoT Thing Group ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		input := &iot.DescribeThingGroupInput{
			ThingGroupName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeThingGroup(input)
		if err != nil {
			return err
		}

		*thing = *resp

		return nil
	}
}

func testAccCheckAWSIotThingGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_group" {
			continue
		}

		input := &iot.DescribeThingGroupInput{
			ThingGroupName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThingGroup(input)
		if err != nil {
			if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("expected IoT Thing Group to be destroyed, %s found", rs.Primary.ID)
	}
	return nil
}

func testAccAWSIotThingGroupConfig_base(thingGroupName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = "%s"
}
`, thingGroupName)
}

func testAccAWSIotThingGroupConfig_full(thingGroupName, parentThingGroupName, answer, description, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "parent" {
  name = "%s"
}

resource "aws_iot_thing_group" "test" {
  name = "%s"
  parent_group_name = "${aws_iot_thing_group.parent.name}"

  properties {
    attributes = {
      One    = "11111"
      Two    = "TwoTwo"
      Answer = "%s"
    }
    description = "%s"
  }

  tags = {
    tagKey = "%s"
  }
}
`, parentThingGroupName, thingGroupName, answer, description, tagValue)
}

func testAccAWSIotThingGroupConfig_withTags(thingGroupName, tagValue string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = "%s"

  tags = {
    tagKey = "%s"
  }
}
`, thingGroupName, tagValue)
}

func testAccAWSIotThingGroupConfig_withPropAttr(thingGroupName, answer string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = "%s"

  properties {
    attributes = {
      One    = "11111"
      Two    = "TwoTwo"
      Answer = "%s"
    }
  }

}
`, thingGroupName, answer)
}

func testAccAWSIotThingGroupConfig_withPropDesc(thingGroupName, description string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = "%s"

  properties {
    description = "%s"
  }

}
`, thingGroupName, description)
}

func testAccAWSIotThingGroupConfig_withPropAll(thingGroupName, answer, description string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = "%s"

  properties {
    attributes = {
      One    = "11111"
      Two    = "TwoTwo"
      Answer = "%s"
    }
    description = "%s"
  }

}
`, thingGroupName, answer, description)
}

func testAccAWSIotThingGroupConfig_withParent(thingGroupName, parentThingGroupName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "parent" {
  name = "%s"
}

resource "aws_iot_thing_group" "test" {
  name = "%s"
  parent_group_name = "${aws_iot_thing_group.parent.name}"
}
`, parentThingGroupName, thingGroupName)
}
