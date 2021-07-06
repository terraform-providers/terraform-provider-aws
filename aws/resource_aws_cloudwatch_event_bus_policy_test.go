package aws

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	events "github.com/aws/aws-sdk-go/service/cloudwatchevents"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudwatchEventBusPolicy_basic(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatchEventBusPolicyConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudwatchEventBusPolicyExists(resourceName),
					testAccAWSCloudwatchEventBusPolicyDocument(resourceName),
				),
			},
			{
				Config: testAccAWSCloudwatchEventBusPolicyConfigUpdate(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudwatchEventBusPolicyExists(resourceName),
					testAccAWSCloudwatchEventBusPolicyDocument(resourceName),
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

func TestAccAWSCloudwatchEventBusPolicy_disappears(t *testing.T) {
	resourceName := "aws_cloudwatch_event_bus_policy.test"
	rstring := acctest.RandString(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, events.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchEventBusDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudwatchEventBusPolicyConfig(rstring),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudwatchEventBusPolicyExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchEventBusPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSCloudwatchEventBusPolicyExists(pr string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		eventBusResource, ok := state.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if eventBusResource.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		eventBusName := eventBusResource.Primary.ID

		input := &events.DescribeEventBusInput{
			Name: aws.String(eventBusName),
		}

		cloudWatchEventsConnection := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		describedEventBus, err := cloudWatchEventsConnection.DescribeEventBus(input)

		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events bus policy for '%s' failed: %w", pr, err)
		}
		if describedEventBus.Policy == nil || len(*describedEventBus.Policy) == 0 {
			return fmt.Errorf("Not found: %s", pr)
		}

		return nil
	}
}

func testAccAWSCloudwatchEventBusPolicyDocument(pr string) resource.TestCheckFunc {
	return func(state *terraform.State) error {
		eventBusPolicyResource, ok := state.RootModule().Resources[pr]
		if !ok {
			return fmt.Errorf("Not found: %s", pr)
		}

		if eventBusPolicyResource.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		var eventBusPolicyResourcePolicyDocument map[string]interface{}
		err := json.Unmarshal([]byte(eventBusPolicyResource.Primary.Attributes["policy"]), &eventBusPolicyResourcePolicyDocument)
		if err != nil {
			return fmt.Errorf("Parsing CloudWatch Events bus policy for '%s' failed: %w", pr, err)
		}

		eventBusName := eventBusPolicyResource.Primary.ID

		input := &events.DescribeEventBusInput{
			Name: aws.String(eventBusName),
		}

		cloudWatchEventsConnection := testAccProvider.Meta().(*AWSClient).cloudwatcheventsconn
		describedEventBus, err := cloudWatchEventsConnection.DescribeEventBus(input)
		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events bus policy for '%s' failed: %w", pr, err)
		}

		var describedEventBusPolicy map[string]interface{}
		err = json.Unmarshal([]byte(*describedEventBus.Policy), &describedEventBusPolicy)

		if err != nil {
			return fmt.Errorf("Reading CloudWatch Events bus policy for '%s' failed: %w", pr, err)
		}
		if describedEventBus.Policy == nil || len(*describedEventBus.Policy) == 0 {
			return fmt.Errorf("Not found: %s", pr)
		}

		if !reflect.DeepEqual(describedEventBusPolicy, eventBusPolicyResourcePolicyDocument) {
			return fmt.Errorf("CloudWatch Events bus policy mismatch for '%s'", pr)
		}

		return nil
	}
}

func testAccAWSCloudwatchEventBusPolicyConfig(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "access" {
  statement {
    sid    = "test-resource-policy"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutEvents",
      "events:PutRule"
    ]
    resources = [
      aws_cloudwatch_event_bus.test.arn,
    ]
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.access.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
`, name)
}

func testAccAWSCloudwatchEventBusPolicyConfigUpdate(name string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_event_bus" "test" {
  name = %[1]q
}

data "aws_iam_policy_document" "access" {
  statement {
    sid    = "test-resource-policy-1"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutEvents",
    ]
    resources = [
      aws_cloudwatch_event_bus.test.arn,
    ]
  }
  statement {
    sid    = "test-resource-policy-2"
    effect = "Allow"
    principals {
      identifiers = ["ecs.amazonaws.com"]
      type        = "Service"
    }
    actions = [
      "events:PutRule"
    ]
    resources = [
      aws_cloudwatch_event_bus.test.arn,
    ]
  }
}

resource "aws_cloudwatch_event_bus_policy" "test" {
  policy         = data.aws_iam_policy_document.access.json
  event_bus_name = aws_cloudwatch_event_bus.test.name
}
`, name)
}
