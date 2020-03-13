package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccAWSInspectorTemplateEventSubscriptions_basic(t *testing.T) {
	prefix := resource.PrefixedUniqueId("test")
	resourceName := "aws_inspector_assessment_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInspectorTemplateAssessmentConfigTwoEventSubscriptions(prefix),

				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscribe_to_event.#", "2"),
					testAccCheckIsSubscribedToEvent(resourceName, "ASSESSMENT_RUN_STARTED"),
					testAccCheckIsSubscribedToEvent(resourceName, "ASSESSMENT_RUN_COMPLETED"),
				),
			},
		},
	})
}

func TestAccAWSInspectorTemplateEventSubscriptions_update(t *testing.T) {
	prefix := resource.PrefixedUniqueId("test")
	resourceName := "aws_inspector_assessment_template.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSInspectorTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSInspectorTemplateAssessmentConfigBasic(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscribe_to_event.#", "0"),
				),
			},
			{
				Config: testAccAWSInspectorTemplateAssessmentConfigTwoEventSubscriptions(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscribe_to_event.#", "2"),
					testAccCheckIsSubscribedToEvent(resourceName, "ASSESSMENT_RUN_STARTED"),
					testAccCheckIsSubscribedToEvent(resourceName, "ASSESSMENT_RUN_COMPLETED"),
				),
			},
			{
				Config: testAccAWSInspectorTemplateAssessmentConfigReplaceOneEventSubscription(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscribe_to_event.#", "2"),
					testAccCheckIsSubscribedToEvent(resourceName, "ASSESSMENT_RUN_STARTED"),
					testAccCheckIsSubscribedToEvent(resourceName, "FINDING_REPORTED"),
				),
			},
			{
				Config: testAccAWSInspectorTemplateAssessmentConfigBasic(prefix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSInspectorTemplateExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "subscribe_to_event.#", "0"),
				),
			},
		},
	})
}

func testAccAWSInspectorTemplateAssessmentConfigBasic(prefix string) string {
	return testAccAWSInspectorTemplateAssessmentConfig(prefix, "")
}

func testAccAWSInspectorTemplateAssessmentConfigTwoEventSubscriptions(prefix string) string {
	return testAccAWSInspectorTemplateAssessmentConfig(prefix, awsInspectorTwoEventSubscriptions) +
		awsInspectorEventSubscriptionsSNSTopicAndIAMPolicy(prefix)

}

func testAccAWSInspectorTemplateAssessmentConfigReplaceOneEventSubscription(prefix string) string {
	return testAccAWSInspectorTemplateAssessmentConfig(prefix, awsInspectorReplacedEventSubscriptions) +
		awsInspectorEventSubscriptionsSNSTopicAndIAMPolicy(prefix)
}

func testAccAWSInspectorTemplateAssessmentConfig(prefix string, topicSubscriptions string) string {
	return fmt.Sprintf(
		`
data "aws_inspector_rules_packages" "rules" {}

resource "aws_inspector_resource_group" "test" {
  tags = {
    Name = "bar"
  }
}

resource "aws_inspector_assessment_target" "test" {
  name               = "%s"
  resource_group_arn = "${aws_inspector_resource_group.test.arn}"
}

resource "aws_inspector_assessment_template" "test" {
  name       = "template %s"
  target_arn = "${aws_inspector_assessment_target.test.arn}"
  duration   = 3600

  rules_package_arns = data.aws_inspector_rules_packages.rules.arns

%s
}`, prefix, prefix, topicSubscriptions)
}

var awsInspectorTwoEventSubscriptions = `
subscribe_to_event {
  event     = "ASSESSMENT_RUN_STARTED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}

subscribe_to_event {
  event     = "ASSESSMENT_RUN_COMPLETED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}
`

var awsInspectorReplacedEventSubscriptions = `
subscribe_to_event {
  event     = "ASSESSMENT_RUN_STARTED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}

subscribe_to_event {
  event     = "FINDING_REPORTED"
  topic_arn = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
}
`

func awsInspectorEventSubscriptionsSNSTopicAndIAMPolicy(prefix string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" { }

resource "aws_sns_topic" "test_sns_topic_for_inspector" {
  name = "inspector_%s"
}

resource "aws_sns_topic_policy" "test_sns_topic_for_inspector" {
  arn    = "${aws_sns_topic.test_sns_topic_for_inspector.arn}"
  policy = "${data.aws_iam_policy_document.inspector-allow-write-to-test-sns-topic.json}"
}

data "aws_iam_policy_document" "inspector-allow-write-to-test-sns-topic" {
  policy_id = "__default_policy_ID"

  statement {
    sid = "inspector"

    principals {
      type = "AWS"

      identifiers = [
		"arn:aws:iam::162588757376:root",
		"arn:aws:iam::526946625049:root",
		"arn:aws:iam::454640832652:root",
		"arn:aws:iam::406045910587:root",
		"arn:aws:iam::537503971621:root",
		"arn:aws:iam::357557129151:root",
		"arn:aws:iam::316112463485:root",
		"arn:aws:iam::646659390643:root",
		"arn:aws:iam::166987590008:root",
		"arn:aws:iam::758058086616:root"
      ]
    }

    actions = [
      "SNS:Subscribe",
      "SNS:Receive",
      "SNS:Publish",
    ]

    effect    = "Allow"
    resources = ["${aws_sns_topic.test_sns_topic_for_inspector.arn}"]
  }
}
`, prefix)
}
