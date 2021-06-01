package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/neptune"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/neptune/waiter"
)

func init() {
	resource.AddTestSweepers("aws_neptune_event_subscription", &resource.Sweeper{
		Name: "aws_neptune_event_subscription",
		F:    testSweepNeptuneEventSubscriptions,
	})
}

func testSweepNeptuneEventSubscriptions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).neptuneconn
	var sweeperErrs *multierror.Error

	err = conn.DescribeEventSubscriptionsPages(&neptune.DescribeEventSubscriptionsInput{}, func(page *neptune.DescribeEventSubscriptionsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			name := aws.StringValue(eventSubscription.CustSubscriptionId)

			log.Printf("[INFO] Deleting Neptune Event Subscription: %s", name)
			_, err = conn.DeleteEventSubscription(&neptune.DeleteEventSubscriptionInput{
				SubscriptionName: aws.String(name),
			})
			if isAWSErr(err, neptune.ErrCodeSubscriptionNotFoundFault, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Neptune Event Subscription (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = waiter.EventSubscriptionDeleted(conn, name)
			if isAWSErr(err, neptune.ErrCodeSubscriptionNotFoundFault, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error waiting for Neptune Event Subscription (%s) deletion: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Neptune Event Subscriptions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Neptune Event Subscriptions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSNeptuneEventSubscription_basic(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-neptune-event-subs-%d", rInt)

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "rds", fmt.Sprintf("es:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-test"),
				),
			},
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigUpdate(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "tf-acc-test1"),
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

func TestAccAWSNeptuneEventSubscription_withPrefix(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()
	startsWithPrefix := regexp.MustCompile("^tf-acc-test-neptune-event-subs-")

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigWithPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestMatchResourceAttr(resourceName, "name", startsWithPrefix),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"name_prefix"},
			},
		},
	})
}

func TestAccAWSNeptuneEventSubscription_withSourceIds(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigWithSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigUpdateSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "2"),
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

func TestAccAWSNeptuneEventSubscription_withCategories(t *testing.T) {
	var v neptune.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-neptune-event-subs-%d", rInt)

	resourceName := "aws_neptune_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSNeptuneEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSNeptuneEventSubscriptionConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "5"),
				),
			},
			{
				Config: testAccAWSNeptuneEventSubscriptionConfigUpdateCategories(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSNeptuneEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "event_categories.#", "1"),
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

func testAccCheckAWSNeptuneEventSubscriptionExists(n string, v *neptune.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Neptune Event Subscription is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).neptuneconn

		opts := neptune.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventSubscriptions(&opts)

		if err != nil {
			return err
		}

		if len(resp.EventSubscriptionsList) != 1 ||
			aws.StringValue(resp.EventSubscriptionsList[0].CustSubscriptionId) != rs.Primary.ID {
			return fmt.Errorf("Neptune Event Subscription not found")
		}

		*v = *resp.EventSubscriptionsList[0]
		return nil
	}
}

func testAccCheckAWSNeptuneEventSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).neptuneconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_neptune_event_subscription" {
			continue
		}

		var err error
		resp, err := conn.DescribeEventSubscriptions(
			&neptune.DescribeEventSubscriptionsInput{
				SubscriptionName: aws.String(rs.Primary.ID),
			})

		if isAWSErr(err, neptune.ErrCodeSubscriptionNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if len(resp.EventSubscriptionsList) != 0 &&
			aws.StringValue(resp.EventSubscriptionsList[0].CustSubscriptionId) == rs.Primary.ID {
			return fmt.Errorf("Event Subscription still exists")
		}
	}

	return nil
}

func testAccAWSNeptuneEventSubscriptionConfig(subscriptionName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[2]q
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt, subscriptionName)
}

func testAccAWSNeptuneEventSubscriptionConfigUpdate(subscriptionName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[2]q
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  enabled       = false
  source_type   = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "tf-acc-test1"
  }
}
`, rInt, subscriptionName)
}

func testAccAWSNeptuneEventSubscriptionConfigWithPrefix(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%d"
}

resource "aws_neptune_event_subscription" "test" {
  name_prefix   = "tf-acc-test-neptune-event-subs-"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt)
}

func testAccAWSNeptuneEventSubscriptionConfigWithSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_parameter_group" "test" {
  name        = "neptune-parameter-group-event-%[1]d"
  family      = "neptune1"
  description = "Test parameter group for terraform"
}

resource "aws_neptune_event_subscription" "test" {
  name          = "tf-acc-test-neptune-event-subs-with-ids-%[1]d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "db-parameter-group"
  source_ids    = [aws_neptune_parameter_group.test.id]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt)
}

func testAccAWSNeptuneEventSubscriptionConfigUpdateSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_parameter_group" "test" {
  name        = "neptune-parameter-group-event-%[1]d"
  family      = "neptune1"
  description = "Test parameter group for terraform"
}

resource "aws_neptune_parameter_group" "test_2" {
  name        = "neptune-parameter-group-event-2-%[1]d"
  family      = "neptune1"
  description = "Test parameter group for terraform"
}

resource "aws_neptune_event_subscription" "test" {
  name          = "tf-acc-test-neptune-event-subs-with-ids-%[1]d"
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "db-parameter-group"
  source_ids    = [aws_neptune_parameter_group.test.id, aws_neptune_parameter_group.test_2.id]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt)
}

func testAccAWSNeptuneEventSubscriptionConfigUpdateCategories(subscriptionName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-neptune-event-subs-sns-topic-%[1]d"
}

resource "aws_neptune_event_subscription" "test" {
  name          = %[2]q
  sns_topic_arn = aws_sns_topic.aws_sns_topic.arn
  source_type   = "db-instance"

  event_categories = [
    "availability",
  ]

  tags = {
    Name = "tf-acc-test"
  }
}
`, rInt, subscriptionName)
}
