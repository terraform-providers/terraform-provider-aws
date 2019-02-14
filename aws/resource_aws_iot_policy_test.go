package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSIoTPolicy_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("PubSubToAnyTopic-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSIoTPolicyConfigInitialState(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_policy.pubsub", "name", rName),
					resource.TestCheckResourceAttrSet("aws_iot_policy.pubsub", "arn"),
					resource.TestCheckResourceAttrSet("aws_iot_policy.pubsub", "default_version_id"),
					resource.TestCheckResourceAttrSet("aws_iot_policy.pubsub", "policy"),
				),
			},
		},
	})
}

func TestAccAWSIoTPolicy_invalidJson(t *testing.T) {
	rName := acctest.RandomWithPrefix("PubSubToAnyTopic-")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSIoTPolicyDestroy_basic,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSIoTPolicyInvalidJsonConfig(rName),
				ExpectError: regexp.MustCompile("MalformedPolicyException.*"),
			},
		},
	})
}

func testAccCheckAWSIoTPolicyDestroy_basic(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).iotconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_iot_policy" {
			continue
		}

		// Try to find the Policy
		GetPolicyOpts := &iot.GetPolicyInput{
			PolicyName: aws.String(rs.Primary.Attributes["name"]),
		}

		resp, err := conn.GetPolicy(GetPolicyOpts)

		if err == nil {
			if resp.PolicyName != nil {
				return fmt.Errorf("IoT Policy still exists")
			}
		}

		// Verify the error is what we want
		if err != nil {
			iotErr, ok := err.(awserr.Error)
			if !ok || iotErr.Code() != "ResourceNotFoundException" {
				return err
			}
		}
	}

	return nil
}

func testAccAWSIoTPolicyConfigInitialState(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "pubsub" {
  name = "%s"
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["*"]
  }]
}
EOF
}
`, rName)
}

func testAccAWSIoTPolicyInvalidJsonConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "pubsub" {
  name = "%s"
  policy = <<EOF
	{
	  "Version": "2012-10-17",
	  "Statement": [{
		"Effect": "Allow",
		"Action": ["iot:*"],
		"Resource": ["*"]
	  }]
	}
EOF
}
`, rName)
}
