package aws

import (
	"fmt"
	"sort"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotevents"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSIotEventsInput_basic(t *testing.T) {
	var input iotevents.DescribeInputOutput
	rString := acctest.RandString(8)
	inputName := fmt.Sprintf("tf_acc_input_%s", rString)
	inputDescription := "Test Description"
	inputAttributes := make([]string, 2)
	inputAttributes[0] = "attribute_1"
	inputAttributes[1] = "attribute_2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotEventsInputDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotEventsInputConfig_basic(inputName, inputDescription, inputAttributes),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotEventsInputExists("aws_iotevents_input.test", &input),
					resource.TestCheckResourceAttr("aws_iotevents_input.test", "name", inputName),
					resource.TestCheckResourceAttr("aws_iotevents_input.test", "description", "Test Description"),
					resource.TestCheckResourceAttr("aws_iotevents_input.test", "tags.tagKey", "tagValue"),
					testAccCheckIotEventsInputAttribute("aws_iotevents_input.test", inputAttributes),
				),
			},
		},
	})
}

func TestAccAWSIotEventsInput_importBasic(t *testing.T) {
	resourceName := "aws_iotevents_input.test"
	rString := acctest.RandString(8)
	inputName := fmt.Sprintf("tf_acc_input_%s", rString)
	inputDescription := "Test Description"
	inputAttributes := make([]string, 2)
	inputAttributes[0] = "attribute_1"
	inputAttributes[1] = "attribute_2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIotEventsInputDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotEventsInputConfig_basic(inputName, inputDescription, inputAttributes),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIotEventsInputExists(n string, input *iotevents.DescribeInputOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoTEvents Input ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ioteventsconn
		params := &iotevents.DescribeInputInput{
			InputName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeInput(params)
		if err != nil {
			return err
		}

		*input = *resp

		return nil
	}
}

// Write function that will describe input
// Then get its attributes and check if they are equal to attributes you define.

func SliceEqual(slice_1, slice_2 []string) bool {
	if len(slice_1) != len(slice_2) {
		return false
	}
	for i, v := range slice_1 {
		if v != slice_2[i] {
			return false
		}
	}
	return true
}

func testAccCheckIotEventsInputAttribute(n string, testAttributes []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoTEvents Input ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ioteventsconn
		params := &iotevents.DescribeInputInput{
			InputName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeInput(params)
		if err != nil {
			return err
		}

		var inputAttributes []string
		for _, attr := range resp.Input.InputDefinition.Attributes {
			inputAttributes = append(inputAttributes, *attr.JsonPath)
		}

		sort.Strings(testAttributes)
		sort.Strings(inputAttributes)

		if !SliceEqual(inputAttributes, testAttributes) {
			return fmt.Errorf("Attributes of created Input(%v) diiferentiane from input data(%v)", inputAttributes, testAttributes)
		}

		return nil
	}
}

func testAccCheckAWSIotEventsInputDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ioteventsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotevents_input" {
			continue
		}

		params := &iotevents.DescribeInputInput{
			InputName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeInput(params)
		if err != nil {
			if isAWSErr(err, iotevents.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Expected IoTEvents Input to be destroyed, %s found", rs.Primary.ID)

	}

	return nil
}

func testAccAWSIotEventsInputConfig_basic(inputName string, inputDescription string, attributes []string) string {
	return fmt.Sprintf(`
resource "aws_iotevents_input" "test" {
  name = "%s"
  description = "%s"
  
  definition {
	  attribute {
		json_path = "%s"
	  }

	  attribute {
		json_path = "%s"
	  }
  }

  tags = {
	  "tagKey" = "tagValue",
  }
}
`, inputName, inputDescription, attributes[0], attributes[1])
}
