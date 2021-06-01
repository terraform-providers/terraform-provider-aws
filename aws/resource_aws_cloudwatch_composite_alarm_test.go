package aws

import (
	"context"
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudwatch/finder"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_composite_alarm", &resource.Sweeper{
		Name: "aws_cloudwatch_composite_alarm",
		F:    testSweepCloudWatchCompositeAlarms,
	})
}

func testSweepCloudWatchCompositeAlarms(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).cloudwatchconn
	ctx := context.Background()

	input := &cloudwatch.DescribeAlarmsInput{
		AlarmTypes: aws.StringSlice([]string{cloudwatch.AlarmTypeCompositeAlarm}),
	}

	var sweeperErrs *multierror.Error

	err = conn.DescribeAlarmsPagesWithContext(ctx, input, func(page *cloudwatch.DescribeAlarmsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, compositeAlarm := range page.CompositeAlarms {
			if compositeAlarm == nil {
				continue
			}

			name := aws.StringValue(compositeAlarm.AlarmName)

			log.Printf("[INFO] Deleting CloudWatch Composite Alarm: %s", name)

			r := resourceAwsCloudWatchCompositeAlarm()
			d := r.Data(nil)
			d.SetId(name)

			diags := r.DeleteContext(ctx, d, client)

			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Composite Alarms sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudWatch Composite Alarms: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsCloudWatchCompositeAlarm_basic(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", ""),
					resource.TestCheckResourceAttr(resourceName, "alarm_name", "tf-test-composite-"+suffix),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(tf-test-metric-0-%[1]s) OR ALARM(tf-test-metric-1-%[1]s)", suffix)),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexp.MustCompile(`alarm:.+`)),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAwsCloudWatchCompositeAlarm_disappears(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchCompositeAlarm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsCloudWatchCompositeAlarm_actionsEnabled(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_actionsEnabled(false, suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_actionsEnabled(true, suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "actions_enabled", "true"),
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

func TestAccAwsCloudWatchCompositeAlarm_alarmActions(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_alarmActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_updateAlarmActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
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

func TestAccAwsCloudWatchCompositeAlarm_description(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_description("Test 1", suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test 1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_description("Test Updated", suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_description", "Test Updated"),
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

func TestAccAwsCloudWatchCompositeAlarm_insufficientDataActions(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_insufficientDataActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_updateInsufficientDataActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
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

func TestAccAwsCloudWatchCompositeAlarm_okActions(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_okActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_updateOkActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "0"),
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

func TestAccAwsCloudWatchCompositeAlarm_allActions(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_allActions(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "insufficient_data_actions.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "ok_actions.#", "0"),
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

func TestAccAwsCloudWatchCompositeAlarm_updateAlarmRule(t *testing.T) {
	suffix := acctest.RandString(8)
	resourceName := "aws_cloudwatch_composite_alarm.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsCloudWatchCompositeAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsCloudWatchCompositeAlarmConfig_updateAlarmRule(suffix),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "alarm_rule", fmt.Sprintf("ALARM(tf-test-metric-0-%[1]s)", suffix)),
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

func testAccCheckAwsCloudWatchCompositeAlarmDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_composite_alarm" {
			continue
		}

		alarm, err := finder.CompositeAlarmByName(context.Background(), conn, rs.Primary.ID)

		if tfawserr.ErrCodeEquals(err, cloudwatch.ErrCodeResourceNotFound) {
			continue
		}
		if err != nil {
			return fmt.Errorf("error reading CloudWatch composite alarm (%s): %w", rs.Primary.ID, err)
		}

		if alarm != nil {
			return fmt.Errorf("CloudWatch composite alarm (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsCloudWatchCompositeAlarmExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("resource %s has not set its id", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

		alarm, err := finder.CompositeAlarmByName(context.Background(), conn, rs.Primary.ID)

		if err != nil {
			return fmt.Errorf("error reading CloudWatch composite alarm (%s): %w", rs.Primary.ID, err)
		}

		if alarm == nil {
			return fmt.Errorf("CloudWatch composite alarm (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  count = 2

  alarm_name          = "tf-test-metric-${count.index}-%s"
  comparison_operator = "GreaterThanOrEqualToThreshold"
  evaluation_periods  = 2
  metric_name         = "CPUUtilization"
  namespace           = "AWS/EC2"
  period              = 120
  statistic           = "Average"
  threshold           = 80

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, suffix)
}

func testAccAwsCloudWatchCompositeAlarmConfig_actionsEnabled(enabled bool, suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  actions_enabled = %t
  alarm_name      = "tf-test-composite-%s"
  alarm_rule      = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
}
`, enabled, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_basic(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_description(description, suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_description = %q
  alarm_name        = "tf-test-composite-%s"
  alarm_rule        = join(" OR ", formatlist("ALARM(%%s)", aws_cloudwatch_metric_alarm.test.*.alarm_name))
}
`, description, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_alarmActions(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions = aws_sns_topic.test.*.arn
  alarm_name    = "tf-test-composite-%[1]s"
  alarm_rule    = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_updateAlarmActions(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions = [aws_sns_topic.test[0].arn]
  alarm_name    = "tf-test-composite-%[1]s"
  alarm_rule    = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_updateAlarmRule(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_insufficientDataActions(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = aws_sns_topic.test.*.arn
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_updateInsufficientDataActions(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = [aws_sns_topic.test[0].arn]
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_okActions(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  ok_actions = aws_sns_topic.test.*.arn
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_updateOkActions(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 2
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_name = "tf-test-composite-%[1]s"
  alarm_rule = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  ok_actions = [aws_sns_topic.test[0].arn]
}
`, suffix))
}

func testAccAwsCloudWatchCompositeAlarmConfig_allActions(suffix string) string {
	return composeConfig(
		testAccAwsCloudWatchCompositeAlarmBaseConfig(suffix),
		fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  count = 3
  name  = "tf-test-alarms-${count.index}-%[1]s"
}

resource "aws_cloudwatch_composite_alarm" "test" {
  alarm_actions             = [aws_sns_topic.test[0].arn]
  alarm_name                = "tf-test-composite-%[1]s"
  alarm_rule                = "ALARM(${aws_cloudwatch_metric_alarm.test[0].alarm_name})"
  insufficient_data_actions = [aws_sns_topic.test[1].arn]
  ok_actions                = [aws_sns_topic.test[2].arn]
}
`, suffix))
}
