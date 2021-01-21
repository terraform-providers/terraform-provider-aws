package aws

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSSNSTopicPolicy_basic(t *testing.T) {
	attributes := make(map[string]string)
	resourceName := "aws_sns_topic_policy.custom"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSNSTopicDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSNSTopicConfig_withPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSNSTopicExists("aws_sns_topic.test", attributes),
					resource.TestMatchResourceAttr(resourceName, "policy",
						regexp.MustCompile("^{\"Version\":\"2012-10-17\".+")),
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

const testAccAWSSNSTopicConfig_withPolicy = `
resource "aws_sns_topic" "test" {
  name = "tf-acc-test-topic-with-policy"
}

resource "aws_sns_topic_policy" "custom" {
  arn = aws_sns_topic.test.arn

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "default",
  "Statement": [
    {
      "Sid": "default",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": [
        "SNS:GetTopicAttributes",
        "SNS:SetTopicAttributes",
        "SNS:AddPermission",
        "SNS:RemovePermission",
        "SNS:DeleteTopic"
      ],
      "Resource": "${aws_sns_topic.test.arn}"
    }
  ]
}
POLICY
}
`
