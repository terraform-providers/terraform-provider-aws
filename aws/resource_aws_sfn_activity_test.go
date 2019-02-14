package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/sfn"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSSfnActivity_importBasic(t *testing.T) {
	resourceName := "aws_sfn_activity.foo"
	rName := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnActivityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnActivityBasicConfig(rName),
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSSfnActivity_basic(t *testing.T) {
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnActivityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnActivityBasicConfig(name),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnActivityExists("aws_sfn_activity.foo"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "name", name),
					resource.TestCheckResourceAttrSet("aws_sfn_activity.foo", "creation_date"),
				),
			},
		},
	})
}

func TestAccAWSSfnActivity_Tags(t *testing.T) {
	name := acctest.RandString(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSfnActivityDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSfnActivityBasicConfigTags1(name, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnActivityExists("aws_sfn_activity.foo"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "tags.key1", "value1"),
				),
			},
			{
				Config: testAccAWSSfnActivityBasicConfigTags2(name, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnActivityExists("aws_sfn_activity.foo"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "tags.%", "2"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSfnActivityBasicConfigTags1(name, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSfnActivityExists("aws_sfn_activity.foo"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_sfn_activity.foo", "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAWSSfnActivityExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Step Function ID set")
		}

		conn := testAccProvider.Meta().(*AWSClient).sfnconn

		_, err := conn.DescribeActivity(&sfn.DescribeActivityInput{
			ActivityArn: aws.String(rs.Primary.ID),
		})

		return err
	}
}

func testAccCheckAWSSfnActivityDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).sfnconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_sfn_activity" {
			continue
		}

		// Retrying as Read after Delete is not always consistent
		retryErr := resource.Retry(1*time.Minute, func() *resource.RetryError {
			var err error

			_, err = conn.DescribeActivity(&sfn.DescribeActivityInput{
				ActivityArn: aws.String(rs.Primary.ID),
			})

			if err != nil {
				if awsErr, ok := err.(awserr.Error); ok && awsErr.Code() == "ActivityDoesNotExist" {
					return nil
				}

				return resource.NonRetryableError(err)
			}

			// If there are no errors, the removal failed
			// and the object is not yet removed.
			return resource.RetryableError(fmt.Errorf("Expected AWS Step Function Activity to be destroyed, but was still found, retrying"))
		})

		return retryErr
	}

	return fmt.Errorf("Default error in Step Function Test")
}

func testAccAWSSfnActivityBasicConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "foo" {
  name = "%s"
}
`, rName)
}

func testAccAWSSfnActivityBasicConfigTags1(rName, tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "foo" {
  name = "%s"
  tags = {
	%q = %q
}
}
`, rName, tag1Key, tag1Value)
}

func testAccAWSSfnActivityBasicConfigTags2(rName, tag1Key, tag1Value, tag2Key, tag2Value string) string {
	return fmt.Sprintf(`
resource "aws_sfn_activity" "foo" {
  name = "%s"
  tags = {
	%q = %q
	%q = %q
}
}
`, rName, tag1Key, tag1Value, tag2Key, tag2Value)
}
