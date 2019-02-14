package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/redshift"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSRedshiftEventSubscription_basicUpdate(t *testing.T) {
	var v redshift.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "source_type", "cluster"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.Name", "name"),
				),
			},
			{
				Config: testAccAWSRedshiftEventSubscriptionConfigUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "enabled", "false"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "source_type", "cluster-snapshot"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.%", "1"),
					resource.TestCheckResourceAttr("aws_redshift_event_subscription.bar", "tags.Name", "new-name"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRedshiftEventSubscription_withPrefix(t *testing.T) {
	var v redshift.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "tags.Name", "name"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRedshiftEventSubscription_withSourceIds(t *testing.T) {
	var v redshift.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-with-ids-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftEventSubscriptionConfigWithSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster-parameter-group"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_ids.#", "1"),
				),
			},
			{
				Config: testAccAWSRedshiftEventSubscriptionConfigUpdateSourceIds(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster-parameter-group"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_ids.#", "2"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRedshiftEventSubscription_categoryUpdate(t *testing.T) {
	var v redshift.EventSubscription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("tf-acc-test-redshift-event-subs-%d", rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRedshiftEventSubscriptionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRedshiftEventSubscriptionConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "name", rName),
				),
			},
			{
				Config: testAccAWSRedshiftEventSubscriptionConfigUpdateCategories(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRedshiftEventSubscriptionExists("aws_redshift_event_subscription.bar", &v),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "enabled", "true"),
					resource.TestCheckResourceAttr(
						"aws_redshift_event_subscription.bar", "source_type", "cluster"),
				),
			},
			{
				ResourceName:      "aws_redshift_event_subscription.bar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSRedshiftEventSubscriptionExists(n string, v *redshift.EventSubscription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Redshift Event Subscription is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).redshiftconn

		opts := redshift.DescribeEventSubscriptionsInput{
			SubscriptionName: aws.String(rs.Primary.ID),
		}

		resp, err := conn.DescribeEventSubscriptions(&opts)

		if err != nil {
			return err
		}

		if len(resp.EventSubscriptionsList) != 1 ||
			*resp.EventSubscriptionsList[0].CustSubscriptionId != rs.Primary.ID {
			return fmt.Errorf("Redshift Event Subscription not found")
		}

		*v = *resp.EventSubscriptionsList[0]
		return nil
	}
}

func testAccCheckAWSRedshiftEventSubscriptionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).redshiftconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_redshift_event_subscription" {
			continue
		}

		var err error
		resp, err := conn.DescribeEventSubscriptions(
			&redshift.DescribeEventSubscriptionsInput{
				SubscriptionName: aws.String(rs.Primary.ID),
			})

		if ae, ok := err.(awserr.Error); ok && ae.Code() == "SubscriptionNotFound" {
			continue
		}

		if err == nil {
			if len(resp.EventSubscriptionsList) != 0 &&
				*resp.EventSubscriptionsList[0].CustSubscriptionId == rs.Primary.ID {
				return fmt.Errorf("Event Subscription still exists")
			}
		}

		// Verify the error
		newerr, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if newerr.Code() != "SubscriptionNotFound" {
			return err
		}
	}

	return nil
}

func testAccAWSRedshiftEventSubscriptionConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_event_subscription" "bar" {
  name = "tf-acc-test-redshift-event-subs-%d"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "cluster"
  severity = "INFO"
  event_categories = [
    "configuration",
    "management",
    "monitoring",
    "security",
  ]
  tags = {
    Name = "name"
  }
}`, rInt, rInt)
}

func testAccAWSRedshiftEventSubscriptionConfigUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_event_subscription" "bar" {
  name = "tf-acc-test-redshift-event-subs-%d"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  enabled = false
  source_type = "cluster-snapshot"
  severity = "INFO"
  event_categories = [
    "monitoring",
  ]
  tags = {
    Name = "new-name"
  }
}`, rInt, rInt)
}

func testAccAWSRedshiftEventSubscriptionConfigWithSourceIds(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_parameter_group" "bar" {
  name = "redshift-parameter-group-event-%d"
  family = "redshift-1.0"
  description = "Test parameter group for terraform"
}

resource "aws_redshift_event_subscription" "bar" {
  name = "tf-acc-test-redshift-event-subs-with-ids-%d"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "cluster-parameter-group"
  severity = "INFO"
  source_ids = ["${aws_redshift_parameter_group.bar.id}"]
  event_categories = [
    "configuration",
  ]
  tags = {
    Name = "name"
  }
}`, rInt, rInt, rInt)
}

func testAccAWSRedshiftEventSubscriptionConfigUpdateSourceIds(rInt int) string {
	return fmt.Sprintf(`
    resource "aws_sns_topic" "aws_sns_topic" {
        name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
    }

    resource "aws_redshift_parameter_group" "bar" {
        name = "tf-acc-redshift-parameter-group-event-%d"
        family = "redshift-1.0"
        description = "Test parameter group for terraform"
    }

    resource "aws_redshift_parameter_group" "foo" {
        name = "tf-acc-redshift-parameter-group-event-2-%d"
        family = "redshift-1.0"
        description = "Test parameter group for terraform"
    }

    resource "aws_redshift_event_subscription" "bar" {
        name = "tf-acc-test-redshift-event-subs-with-ids-%d"
        sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
        source_type = "cluster-parameter-group"
        severity = "INFO"
        source_ids = ["${aws_redshift_parameter_group.bar.id}","${aws_redshift_parameter_group.foo.id}"]
        event_categories = [
            "configuration",
        ]
  tags = {
            Name = "name"
        }
    }`, rInt, rInt, rInt, rInt)
}

func testAccAWSRedshiftEventSubscriptionConfigUpdateCategories(rInt int) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "aws_sns_topic" {
  name = "tf-acc-test-redshift-event-subs-sns-topic-%d"
}

resource "aws_redshift_event_subscription" "bar" {
  name = "tf-acc-test-redshift-event-subs-%d"
  sns_topic_arn = "${aws_sns_topic.aws_sns_topic.arn}"
  source_type = "cluster"
  severity = "INFO"
  event_categories = [
    "monitoring",
  ]
  tags = {
    Name = "name"
  }
}`, rInt, rInt)
}
