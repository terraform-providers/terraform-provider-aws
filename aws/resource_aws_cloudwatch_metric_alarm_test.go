package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudWatchMetricAlarm_basic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "statistic", "Average"),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "cloudwatch", regexp.MustCompile(`alarm:.+`)),
					testAccCheckCloudWatchMetricAlarmDimension(resourceName, "InstanceId", "i-abc123"),
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

func TestAccAWSCloudWatchMetricAlarm_AlarmActions_EC2Automate(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigAlarmActionsEC2Automate(rName, "reboot"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigAlarmActionsEC2Automate(rName, "recover"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigAlarmActionsEC2Automate(rName, "stop"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigAlarmActionsEC2Automate(rName, "terminate"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_AlarmActions_SNSTopic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigAlarmActionsSNSTopic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
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

func TestAccAWSCloudWatchMetricAlarm_AlarmActions_SWFAction(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigAlarmActionsSWFAction(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "alarm_actions.#", "1"),
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

func TestAccAWSCloudWatchMetricAlarm_datapointsToAlarm(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigDatapointsToAlarm(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "datapoints_to_alarm", "2"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_treatMissingData(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatMissingData(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "treat_missing_data", "missing"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatMissingDataUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "treat_missing_data", "breaching"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_evaluateLowSampleCountPercentiles(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentiles(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "evaluate_low_sample_count_percentiles", "evaluate"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentilesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "evaluate_low_sample_count_percentiles", "ignore"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_extendedStatistic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigExtendedStatistic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "extended_statistic", "p88.0"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_expression(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudWatchMetricAlarmConfigWithBadExpression(rName),
				ExpectError: regexp.MustCompile("No metric_query may have both `expression` and a `metric` specified"),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigWithExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigWithExpressionUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "3"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigWithExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigWithAnomalyDetectionExpression(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigWithExpressionWithQueryUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_query.#", "2"),
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

func TestAccAWSCloudWatchMetricAlarm_missingStatistic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudWatchMetricAlarmConfigMissingStatistic(rName),
				ExpectError: regexp.MustCompile("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm"),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_tags(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_disappears(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudWatchMetricAlarm(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckCloudWatchMetricAlarmDimension(n, k, v string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}
		key := fmt.Sprintf("dimensions.%s", k)
		val, ok := rs.Primary.Attributes[key]
		if !ok {
			return fmt.Errorf("Could not find dimension: %s", k)
		}
		if val != v {
			return fmt.Errorf("Expected dimension %s => %s; got: %s", k, v, val)
		}
		return nil
	}
}

func testAccCheckCloudWatchMetricAlarmExists(n string, alarm *cloudwatch.MetricAlarm) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn
		params := cloudwatch.DescribeAlarmsInput{
			AlarmNames: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeAlarms(&params)
		if err != nil {
			return err
		}
		if len(resp.MetricAlarms) == 0 {
			return fmt.Errorf("Alarm not found")
		}
		*alarm = *resp.MetricAlarms[0]

		return nil
	}
}

func testAccCheckAWSCloudWatchMetricAlarmDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_metric_alarm" {
			continue
		}

		params := cloudwatch.DescribeAlarmsInput{
			AlarmNames: []*string{aws.String(rs.Primary.ID)},
		}

		resp, err := conn.DescribeAlarms(&params)

		if err == nil {
			if len(resp.MetricAlarms) != 0 &&
				*resp.MetricAlarms[0].AlarmName == rs.Primary.ID {
				return fmt.Errorf("Alarm Still Exists: %s", rs.Primary.ID)
			}
		}
	}

	return nil
}

func testAccAWSCloudWatchMetricAlarmConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigDatapointsToAlarm(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  datapoints_to_alarm       = "2"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatMissingData(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  treat_missing_data        = "missing"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatMissingDataUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  treat_missing_data        = "breaching"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentiles(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                            = "%s"
  comparison_operator                   = "GreaterThanOrEqualToThreshold"
  evaluation_periods                    = "2"
  metric_name                           = "CPUUtilization"
  namespace                             = "AWS/EC2"
  period                                = "120"
  extended_statistic                    = "p88.0"
  threshold                             = "80"
  alarm_description                     = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "evaluate"
  insufficient_data_actions             = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentilesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                            = "%s"
  comparison_operator                   = "GreaterThanOrEqualToThreshold"
  evaluation_periods                    = "2"
  metric_name                           = "CPUUtilization"
  namespace                             = "AWS/EC2"
  period                                = "120"
  extended_statistic                    = "p88.0"
  threshold                             = "80"
  alarm_description                     = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "ignore"
  insufficient_data_actions             = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigExtendedStatistic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  extended_statistic        = "p88.0"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigMissingStatistic(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigWithExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "m1"
    label       = "cat"
    return_data = "true"
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigWithAnomalyDetectionExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanUpperThreshold"
  evaluation_periods        = "2"
  threshold_metric_id       = "e1"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "ANOMALY_DETECTION_BAND(m1)"
    label       = "CPUUtilization (Expected)"
    return_data = "true"
  }

  metric_query {
    id          = "m1"
    return_data = "true"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigWithExpressionUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id         = "e1"
    expression = "m1"
    label      = "cat"
  }

  metric_query {
    id          = "e2"
    expression  = "e1"
    label       = "bug"
    return_data = "true"
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigWithExpressionWithQueryUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id          = "e1"
    expression  = "m1"
    label       = "cat"
    return_data = "true"
  }

  metric_query {
    id = "m1"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Maximum"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigWithBadExpression(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = "%s"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  metric_query {
    id         = "e1"
    expression = "m1"
    label      = "cat"

    metric {
      metric_name = "CPUUtilization"
      namespace   = "AWS/EC2"
      period      = "120"
      stat        = "Average"
      unit        = "Count"

      dimensions = {
        InstanceId = "i-abc123"
      }
    }
  }
}
`, rName)
}

// EC2 Automate requires a valid EC2 instance
// ValidationError: Invalid use of EC2 'Recover' action. i-abc123 is not a valid EC2 instance.
func testAccAWSCloudWatchMetricAlarmConfigAlarmActionsEC2Automate(rName, action string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "172.16.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["arn:${data.aws_partition.current.partition}:automate:${data.aws_region.current.name}:ec2:%[2]s"]
  alarm_description   = "Status checks have failed for system"
  alarm_name          = %[1]q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "StatusCheckFailed_System"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "0"
  unit                = "Count"

  dimensions = {
    InstanceId = aws_instance.test.id
  }
}
`, rName, action))
}

func testAccAWSCloudWatchMetricAlarmConfigAlarmActionsSNSTopic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %q
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = [aws_sns_topic.test.arn]
  alarm_description   = "Status checks have failed for system"
  alarm_name          = %q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "2"
  metric_name         = "StatusCheckFailed_System"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "0"
  unit                = "Count"

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigAlarmActionsSWFAction(rName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {
}

data "aws_partition" "current" {
}

data "aws_region" "current" {
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["arn:${data.aws_partition.current.partition}:swf:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:action/actions/AWS_EC2.InstanceId.Reboot/1.0"]
  alarm_description   = "Status checks have failed, rebooting system."
  alarm_name          = %q
  comparison_operator = "GreaterThanThreshold"
  evaluation_periods  = "5"
  metric_name         = "StatusCheckFailed_Instance"
  namespace           = "AWS/EC2"
  period              = "60"
  statistic           = "Minimum"
  threshold           = "0"
  unit                = "Count"

  dimensions = {
    InstanceId = "i-abc123"
  }
}
`, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSCloudWatchMetricAlarmConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_name                = %[1]q
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  statistic                 = "Average"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []

  dimensions = {
    InstanceId = "i-abc123"
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
