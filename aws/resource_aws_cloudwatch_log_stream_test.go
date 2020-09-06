package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudWatchLogStream_basic(t *testing.T) {
	var ls cloudwatchlogs.LogStream
	resourceName := "aws_cloudwatch_log_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogStreamExists(resourceName, &ls),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateIdFunc: testAccAWSCloudWatchLogStreamImportStateIdFunc(resourceName),
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCloudWatchLogStream_disappears(t *testing.T) {
	var ls cloudwatchlogs.LogStream
	resourceName := "aws_cloudwatch_log_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogStreamExists(resourceName, &ls),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchLogStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudWatchLogStream_disappears_LogGroup(t *testing.T) {
	var ls cloudwatchlogs.LogStream
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_stream.test"
	logGroupResourceName := "aws_cloudwatch_log_group.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogStreamConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogStreamExists(resourceName, &ls),
					testAccCheckCloudWatchLogGroupExists(logGroupResourceName, &lg),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchLogGroup(), logGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudWatchLogStreamExists(n string, ls *cloudwatchlogs.LogStream) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		logGroupName := rs.Primary.Attributes["log_group_name"]
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		logGroup, exists, err := lookupCloudWatchLogStream(conn, rs.Primary.ID, logGroupName, nil)
		if err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("Bad: LogStream %q does not exist", rs.Primary.ID)
		}

		*ls = *logGroup

		return nil
	}
}

func testAccCheckAWSCloudWatchLogStreamDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_stream" {
			continue
		}

		logGroupName := rs.Primary.Attributes["log_group_name"]
		_, exists, err := lookupCloudWatchLogStream(conn, rs.Primary.ID, logGroupName, nil)
		if err != nil {
			return nil
		}

		if exists {
			return fmt.Errorf("Bad: LogStream still exists: %q", rs.Primary.ID)
		}

	}

	return nil
}

func testAccAWSCloudWatchLogStreamImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not Found: %s", resourceName)
		}

		return fmt.Sprintf("%s:%s", rs.Primary.Attributes["log_group_name"], rs.Primary.ID), nil
	}
}

func TestValidateCloudWatchLogStreamName(t *testing.T) {
	validNames := []string{
		"test-log-stream",
		"my_sample_log_stream",
		"012345678",
		"logstream/1234",
	}
	for _, v := range validNames {
		_, errors := validateCloudWatchLogStreamName(v, "name")
		if len(errors) != 0 {
			t.Fatalf("%q should be a valid CloudWatch LogStream name: %q", v, errors)
		}
	}

	invalidNames := []string{
		acctest.RandString(513),
		"",
		"stringwith:colon",
	}
	for _, v := range invalidNames {
		_, errors := validateCloudWatchLogStreamName(v, "name")
		if len(errors) == 0 {
			t.Fatalf("%q should be an invalid CloudWatch LogStream name", v)
		}
	}
}

func testAccAWSCloudWatchLogStreamConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.id
}
`, rName)
}
