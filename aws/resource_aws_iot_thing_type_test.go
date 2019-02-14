package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIotThingType_importBasic(t *testing.T) {
	resourceName := "aws_iot_thing_type.foo"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingTypeConfig_basic(rInt),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSIotThingType_basic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingTypeConfig_basic(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "name", fmt.Sprintf("tf_acc_iot_thing_type_%d", rInt)),
				),
			},
		},
	})
}

func TestAccAWSIotThingType_full(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotThingTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotThingTypeConfig_full(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("aws_iot_thing_type.foo", "arn"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.description", "MyDescription"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "properties.0.searchable_attributes.#", "3"),
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "true"),
				),
			},
			{
				Config: testAccAWSIotThingTypeConfig_fullUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_thing_type.foo", "deprecated", "false"),
				),
			},
		},
	})
}

func testAccCheckAWSIotThingTypeDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_thing_type" {
			continue
		}

		params := &iot.DescribeThingTypeInput{
			ThingTypeName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeThingType(params)
		if err == nil {
			return fmt.Errorf("Expected IoT Thing Type to be destroyed, %s found", rs.Primary.ID)
		}

	}

	return nil
}

func testAccAWSIotThingTypeConfig_basic(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name = "tf_acc_iot_thing_type_%d"
}
`, rName)
}

func testAccAWSIotThingTypeConfig_full(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = true

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }
}
`, rName)
}

func testAccAWSIotThingTypeConfig_fullUpdated(rName int) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "foo" {
  name       = "tf_acc_iot_thing_type_%d"
  deprecated = false

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }
}
`, rName)
}
