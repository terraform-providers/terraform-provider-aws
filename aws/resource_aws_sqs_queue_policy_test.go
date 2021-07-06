package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSSQSQueuePolicy_basic(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:   testAccAWSSQSPolicyConfig(rName),
				PlanOnly: true,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "policy", queueResourceName, "policy"),
				),
			},
		},
	})
}

func TestAccAWSSQSQueuePolicy_disappears(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSqsQueuePolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSQSQueuePolicy_disappears_queue(t *testing.T) {
	var queueAttributes map[string]string
	queueResourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSqsQueue(), queueResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSSQSQueuePolicy_Update(t *testing.T) {
	var queueAttributes map[string]string
	resourceName := "aws_sqs_queue_policy.test"
	queueResourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, sqs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSQSQueueDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSQSPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSSQSQueueExists(queueResourceName, &queueAttributes),
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSSQSPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "policy"),
				),
			},
		},
	})
}

func testAccAWSSQSPolicyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": "sqs:*",
    "Resource": "${aws_sqs_queue.test.arn}",
    "Condition": {
      "ArnEquals": {
        "aws:SourceArn": "${aws_sqs_queue.test.arn}"
      }
    }
  }]
}
POLICY
}
`, rName)
}

func testAccAWSSQSPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_sqs_queue_policy" "test" {
  queue_url = aws_sqs_queue.test.id

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Principal": "*",
    "Action": [
      "sqs:SendMessage",
      "sqs:ReceiveMessage"
    ],
    "Resource": "${aws_sqs_queue.test.arn}"
  }]
}
POLICY
}
`, rName)
}
