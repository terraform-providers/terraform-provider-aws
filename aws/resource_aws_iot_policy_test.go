package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/iot"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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

func TestAccAWSIoTPolicy_update(t *testing.T) {
	rName := acctest.RandomWithPrefix("PubSubToAnyTopic-")
	expectedVersions := []string{"2", "3", "4", "5", "6"}

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
					resource.TestCheckResourceAttr("aws_iot_policy.pubsub", "default_version_id", "1"),
					resource.TestCheckResourceAttrSet("aws_iot_policy.pubsub", "policy"),
				),
			},
			{
				Config: testAccAWSIoTPolicyConfig_updatePolicy(rName, "topic2"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_policy.pubsub", "default_version_id", "2"),
				),
			},
			{
				Config: testAccAWSIoTPolicyConfig_updatePolicy(rName, "topic3"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_policy.pubsub", "default_version_id", "3"),
				),
			},
			{
				Config: testAccAWSIoTPolicyConfig_updatePolicy(rName, "topic4"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_policy.pubsub", "default_version_id", "4"),
				),
			},
			{
				Config: testAccAWSIoTPolicyConfig_updatePolicy(rName, "topic5"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_policy.pubsub", "default_version_id", "5"),
				),
			},
			{
				Config: testAccAWSIoTPolicyConfig_updatePolicy(rName, "topic6"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("aws_iot_policy.pubsub", "default_version_id", "6"),
					testAccCheckAWSIoTPolicyVersions("aws_iot_policy.pubsub", expectedVersions),
				),
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

func testAccCheckAWSIoTPolicyVersions(rName string, expVersions []string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		conn := testAccProvider.Meta().(*AWSClient).iotconn
		params := &iot.ListPolicyVersionsInput{
			PolicyName: aws.String(rs.Primary.Attributes["name"]),
		}

		resp, err := conn.ListPolicyVersions(params)
		if err != nil {
			return err
		}

		if len(expVersions) != len(resp.PolicyVersions) {
			return fmt.Errorf("Expected %d versions, got %d", len(expVersions), len(resp.PolicyVersions))
		}

		var actVersions []string
		for _, actVer := range resp.PolicyVersions {
			actVersions = append(actVersions, *(actVer.VersionId))
		}

		matchedValue := false
		for _, actVer := range actVersions {
			matchedValue = false
			for _, expVer := range expVersions {
				if actVer == expVer {
					matchedValue = true
					break
				}
			}
			if !matchedValue {
				return fmt.Errorf("Expected: %v / Got: %v", expVersions, actVersions)
			}
		}

		return nil
	}
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

func testAccAWSIoTPolicyConfig_updatePolicy(rName string, topicName string) string {
	return fmt.Sprintf(`
resource "aws_iot_policy" "pubsub" {
  name = "%s"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": ["iot:*"],
    "Resource": ["arn:aws:iot:*:*:topic/%s"]
}]
}
EOF
}
`, rName, topicName)

}
