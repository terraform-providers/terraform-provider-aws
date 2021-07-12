package aws

import (
	"log"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsIoTAnalyticsChannel_basic(t *testing.T) {
	var channel iotanalytics.DescribeChannelOutput
	rString := acctest.RandString(8)
	channelName := fmt.Sprintf("tf_acc_channel_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsIoTAnalyticsChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIoTAnalyticsChannelConfig_basic(channelName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIotAnalyticsChannelExists("aws_iotanalytics_channel.test", &channel),
					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "name", channelName),
					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "channel_storage.type", "service_managed_s3"),
					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "arn"),
				),
			},
		},
	})
}

//  func TestAccAwsIoTAnalyticsChannel_full(t *testing.T) {
// 	var thing iot.DescribeThingOutput
// 	rString := acctest.RandString(8)
// 	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)
// 	typeName := fmt.Sprintf("tf_acc_type_%s", rString)

// 	resource.ParallelTest(t, resource.TestCase{
// 		PreCheck:     func() { testAccPreCheck(t) },
// 		Providers:    testAccProviders,
// 		CheckDestroy: testAccCheckAwsIoTAnalyticsChannelDestroy,
// 		Steps: []resource.TestStep{
// 			{
// 				Config: testAccAwsIoTAnalyticsChannelConfig_full(thingName, typeName, "42"),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckIotAnalyticsChannelExists("aws_iotanalytics_channel.test", &thing),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "name", thingName),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "thing_type_name", typeName),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.%", "3"),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.One", "11111"),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.Two", "TwoTwo"),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.Answer", "42"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "arn"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "default_client_id"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "version"),
// 				),
// 			},
// 			{ // Update attribute
// 				Config: testAccAwsIoTAnalyticsChannelConfig_full(thingName, typeName, "differentOne"),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckIotAnalyticsChannelExists("aws_iotanalytics_channel.test", &thing),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "name", thingName),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "thing_type_name", typeName),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.%", "3"),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.One", "11111"),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.Two", "TwoTwo"),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.Answer", "differentOne"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "arn"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "default_client_id"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "version"),
// 				),
// 			},
// 			{ // Remove thing type association
// 				Config: testAccAwsIoTAnalyticsChannelConfig_basic(thingName),
// 				Check: resource.ComposeTestCheckFunc(
// 					testAccCheckIotAnalyticsChannelExists("aws_iotanalytics_channel.test", &thing),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "name", thingName),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "attributes.%", "0"),
// 					resource.TestCheckResourceAttr("aws_iotanalytics_channel.test", "thing_type_name", ""),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "arn"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "default_client_id"),
// 					resource.TestCheckResourceAttrSet("aws_iotanalytics_channel.test", "version"),
// 				),
// 			},
// 		},
// 	})
// }

func TestAccAwsIoTAnalyticsChannel_importBasic(t *testing.T) {
	resourceName := "aws_iotanalytics_channel.test"
	rString := acctest.RandString(8)
	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsIoTAnalyticsChannelDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsIoTAnalyticsChannelConfig_basic(thingName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckIotAnalyticsChannelExists(n string, thing *iotanalytics.DescribeChannelOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		log.Print("Couldn't read first byte")
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Thing ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn
		params := &iotanalytics.DescribeChannelInput{
			ChannelName: aws.String(rs.Primary.ID),
		}
		resp, err := conn.DescribeChannel(params)
		if err != nil {
			return err
		}

		*thing = *resp

		return nil
	}
}

func testAccCheckAwsIoTAnalyticsChannelDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iotanalytics_channel" {
			continue
		}

		params := &iotanalytics.DescribeChannelInput{
			ChannelName: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeChannel(params)
		if err != nil {
			if isAWSErr(err, iotanalytics.ErrCodeResourceNotFoundException, "") {
				return nil
			}
			return err
		}
		return fmt.Errorf("Expected IoT Analytics Channel to be destroyed, %s found", rs.Primary.ID)

	}

	return nil
}

func testAccAwsIoTAnalyticsChannelConfig_basic(ChannelName string) string {
	return fmt.Sprintf(`
resource "aws_iotanalytics_channel" "test" {
  name = "%s"
  channel_storage {
	  type = "service_managed_s3"
  }
}
`, ChannelName)
}

func testAccAwsIoTAnalyticsChannelConfig_full(thingName, typeName, answer string) string {
	return fmt.Sprintf(`
resource "aws_iotanalytics_channel" "test" {
  name = "%s"

  attributes = {
    One    = "11111"
    Two    = "TwoTwo"
    Answer = "%s"
  }

  thing_type_name = "${aws_iotanalytics_channel_type.test.name}"
}

resource "aws_iotanalytics_channel_type" "test" {
  name = "%s"
}
`, thingName, answer, typeName)
}
