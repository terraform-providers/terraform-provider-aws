package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/rds"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/rds/waiter"
)

func init() {
	resource.AddTestSweepers("aws_db_event_subscription", &resource.Sweeper{
		Name: "aws_db_event_subscription",
		F:    testSweepDbEventSubscriptions,
	})
}

func testSweepDbEventSubscriptions(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).rdsconn
	var sweeperErrs *multierror.Error

	err = conn.DescribeEventSubscriptionsPages(&rds.DescribeEventSubscriptionsInput{}, func(page *rds.DescribeEventSubscriptionsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, eventSubscription := range page.EventSubscriptionsList {
			name := aws.StringValue(eventSubscription.CustSubscriptionId)

			log.Printf("[INFO] Deleting RDS Event Subscription: %s", name)
			_, err = conn.DeleteEventSubscription(&rds.DeleteEventSubscriptionInput{
				SubscriptionName: aws.String(name),
			})
			if isAWSErr(err, rds.ErrCodeSubscriptionNotFoundFault, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting RDS Event Subscription (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			_, err = waiter.EventSubscriptionDeleted(conn, name)
			if isAWSErr(err, rds.ErrCodeSubscriptionNotFoundFault, "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error waiting for RDS Event Subscription (%s) deletion: %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping RDS Event Subscriptions sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving RDS Event Subscriptions: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSDBEventSubscription_basicUpdate(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.test"
	subscriptionName := fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "rds", regexp.MustCompile(fmt.Sprintf("es:%s$", subscriptionName))),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "name", subscriptionName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     subscriptionName,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "new-name"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_disappears(t *testing.T) {
	var eventSubscription rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &eventSubscription),
					testAccCheckAWSDBEventSubscriptionDisappears(&eventSubscription),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDBEventSubscription_withPrefix(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	startsWithPrefix := regexp.MustCompile("^tf-acc-test-rds-event-subs-")
	resourceName := "aws_db_event_subscription.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigWithPrefix(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestMatchResourceAttr(resourceName, "name", startsWithPrefix),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", "name"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_withSourceIds(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.test"
	subscriptionName := fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfigWithSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     subscriptionName,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdateSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-parameter-group"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-acc-test-rds-event-subs-with-ids-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "source_ids.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSDBEventSubscription_categoryUpdate(t *testing.T) {
	var v rds.EventSubscription
	rInt := acctest.RandInt()
	resourceName := "aws_db_event_subscription.test"
	subscriptionName := fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSDBEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("tf-acc-test-rds-event-subs-%d", rInt)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateId:     subscriptionName,
			},
			{
				Config: testAccAWSDBEventSubscriptionConfigUpdateCategories(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSDBEventSubscriptionExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "source_type", "db-instance"),
				),
			},
		},
	})
}

func testAccCheckAWSDBEventSubscriptionExists(n string, v *rds.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No RDS Event Subscription is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		eventSubscription, err := resourceAwsDbEventSubscriptionRetrieve(rs.Primary.ID, conn)

		if err != nil {
			return err
		}

		if eventSubscription == nil {
			return fmt.Errorf("RDS Event Subscription not found")
		}

		*v = *eventSubscription

		return nil
	}
}

func testAccCheckAWSDBEventSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).rdsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_db_event_subscription" {
			continue
		}

		eventSubscription, err := resourceAwsDbEventSubscriptionRetrieve(rs.Primary.ID, conn)

		if isAWSErr(err, rds.ErrCodeSubscriptionNotFoundFault, "") {
			continue
		}

		if err != nil {
			return err
		}

		if eventSubscription != nil {
			return fmt.Errorf("RDS Event Subscription (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAWSDBEventSubscriptionDisappears(eventSubscription *rds.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).rdsconn

		input := &rds.DeleteEventSubscriptionInput{
			SubscriptionName: eventSubscription.CustSubscriptionId,
		}

		_, err := conn.DeleteEventSubscription(input)

		if err != nil {
			return err
		}

		return waitForRdsEventSubscriptionDeletion(conn, aws.StringValue(eventSubscription.CustSubscriptionId), 10*time.Minute)
	}
}

func testAccAWSDBEventSubscriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigWithPrefix(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%d"
}

resource "aws_db_event_subscription" "test" {
  name_prefix = "tf-acc-test-rds-event-subs-"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-instance"

  event_categories = [
    "availability",
    "backup",
    "creation",
    "deletion",
    "maintenance",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  enabled     = false
  source_type = "db-parameter-group"

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "new-name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigWithSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_parameter_group" "test" {
  name        = "db-parameter-group-event-%[1]d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-with-ids-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-parameter-group"
  source_ids  = [aws_db_parameter_group.test.id]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdateSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_parameter_group" "test" {
  name        = "db-parameter-group-event-%[1]d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_parameter_group" "test2" {
  name        = "db-parameter-group-event-2-%[1]d"
  family      = "mysql5.6"
  description = "Test parameter group for terraform"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-with-ids-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-parameter-group"
  source_ids  = [aws_db_parameter_group.test.id, aws_db_parameter_group.test2.id]

  event_categories = [
    "configuration change",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}

func testAccAWSDBEventSubscriptionConfigUpdateCategories(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-rds-event-subs-sns-topic-%[1]d"
}

resource "aws_db_event_subscription" "test" {
  name        = "tf-acc-test-rds-event-subs-%[1]d"
  sns_topic   = aws_sns_topic.aws_sns_topic.arn
  source_type = "db-instance"

  event_categories = [
    "availability",
  ]

  tags = {
    Name = "name"
  }
}
`, rInt)
}
