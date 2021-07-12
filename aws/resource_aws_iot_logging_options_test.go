package aws

import (
	"fmt"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"log"
	"testing"
)

type iotlog struct {
	level   string
	disable bool
	rolearn string
}

func TestAccAWSIotLoggingOptions_set(t *testing.T) {
	var logging iot.GetV2LoggingOptionsOutput
	resourceName := "aws_iot_logging_options.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		CheckDestroy: testdestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIotLoggingOptionsConfig(),
				Check:  check(resourceName, &logging),
			},
		},
	})
}

func testAccAWSIotLoggingOptionsConfig() string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test_role" {
	name = "test1"

	assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
	{
	  "Action": "sts:AssumeRole",
	  "Principal": { "Service": "logs.us-west-2.amazonaws.com" },
	  "Effect": "Allow"
	}
  ]
}
	EOF
}

resource "aws_iot_logging_options" "test" {
	default_log_level 	= "DEBUG"
	disable_all_logs 	= "true"
	role_arn = aws_iam_role.test_role.arn
}
`)
}

func check(n string, loggin *iot.GetV2LoggingOptionsOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[n]
		if !ok {
			log.Printf("[INFO] %s the loggind resource wasn't create", n)
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		out, err := conn.GetV2LoggingOptions(nil)
		if err != nil {
			return err
		}
		if out == nil {
			log.Printf("[INFO] %s the loggind resource wasn't get from AWS side", n)
			return fmt.Errorf("Not found: %s", n)
		}

		*loggin = *out

		return nil
	}
}

func testdestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_logging_options" {
			continue
		}

		_, err := conn.GetV2LoggingOptions(nil)

		if isAWSErr(err, iot.ErrCodeResourceNotFoundException, "") {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("IoT Logging (%s) still exists", rs.Primary.ID)
	}

	return nil
}
