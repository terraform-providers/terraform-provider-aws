package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSCloudWatchMetricAlarm_basic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	resourceName := "aws_cloudwatch_metric_alarm.foobar"
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists(resourceName, &alarm),
					resource.TestCheckResourceAttr(resourceName, "metric_name", "CPUUtilization"),
					resource.TestCheckResourceAttr(resourceName, "statistic", "Average"),
					resource.TestMatchResourceAttr(resourceName, "arn",
						regexp.MustCompile(`^arn:[\w-]+:cloudwatch:[^:]+:\d{12}:alarm:.+$`)),
					testAccCheckCloudWatchMetricAlarmDimension(
						resourceName, "InstanceId", "i-abc123"),
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
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigDatapointsToAlarm(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "datapoints_to_alarm", "2"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_treatMissingData(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatMissingData(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "treat_missing_data", "missing"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatMissingDataUpdate(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "treat_missing_data", "breaching"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_evaluateLowSampleCountPercentiles(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentiles(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "evaluate_low_sample_count_percentiles", "evaluate"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentilesUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "evaluate_low_sample_count_percentiles", "ignore"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_extendedStatistic(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigExtendedStatistic(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "extended_statistic", "p88.0"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_expression(t *testing.T) {
	var alarm cloudwatch.MetricAlarm
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudWatchMetricAlarmConfigWithBadExpression(rInt),
				ExpectError: regexp.MustCompile("No metric_query may have both `expression` and a `metric` specified"),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigWithExpression(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "metric_query.#", "2"),
				),
			},
			{
				Config: testAccAWSCloudWatchMetricAlarmConfigWithExpressionUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchMetricAlarmExists("aws_cloudwatch_metric_alarm.foobar", &alarm),
					resource.TestCheckResourceAttr("aws_cloudwatch_metric_alarm.foobar", "metric_query.#", "3"),
				),
			},
			{
				ResourceName:      "aws_cloudwatch_metric_alarm.foobar",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSCloudWatchMetricAlarm_missingStatistic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchMetricAlarmDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSCloudWatchMetricAlarmConfigMissingStatistic(rInt),
				ExpectError: regexp.MustCompile("One of `statistic` or `extended_statistic` must be set for a cloudwatch metric alarm"),
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

func testAccAWSCloudWatchMetricAlarmConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
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
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigDatapointsToAlarm(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
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
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatMissingData(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
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
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatMissingDataUpdate(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
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
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentiles(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  extended_statistic        = "p88.0"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "evaluate"
  insufficient_data_actions = []
  dimensions = {
    InstanceId = "i-abc123"
  }
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigTreatEvaluateLowSampleCountPercentilesUpdated(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  extended_statistic        = "p88.0"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  evaluate_low_sample_count_percentiles = "ignore"
  insufficient_data_actions = []
  dimensions = {
    InstanceId = "i-abc123"
  }
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigExtendedStatistic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
  metric_name               = "CPUUtilization"
  namespace                 = "AWS/EC2"
  period                    = "120"
  extended_statistic	    = "p88.0"
  threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []
  dimensions = {
    InstanceId = "i-abc123"
  }
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigMissingStatistic(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
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
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigWithExpression(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
	threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []
	metric_query {
		id = "e1"
		expression = "m1"
		label = "cat"
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
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigWithExpressionUpdated(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
	threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []
	metric_query {
		id = "e1"
		expression = "m1"
		label = "cat"
	}
	metric_query {
		id = "e2"
		expression = "e1"
		label = "bug"
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
}`, rInt)
}

func testAccAWSCloudWatchMetricAlarmConfigWithBadExpression(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_metric_alarm" "foobar" {
  alarm_name                = "terraform-test-foobar%d"
  comparison_operator       = "GreaterThanOrEqualToThreshold"
  evaluation_periods        = "2"
	threshold                 = "80"
  alarm_description         = "This metric monitors ec2 cpu utilization"
  insufficient_data_actions = []
	metric_query {
		id = "e1"
		expression = "m1"
		label = "cat"
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
}`, rInt)
}

// EC2 Automate requires a valid EC2 instance
// ValidationError: Invalid use of EC2 'Recover' action. i-abc123 is not a valid EC2 instance.
func testAccAWSCloudWatchMetricAlarmConfigAlarmActionsEC2Automate(rName, action string) string {
	return fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }
  filter {
    name = "root-device-type"
    values = ["ebs"]
  }
}

data "aws_partition" "current" {}

data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "172.16.0.0/16"

  tags = {
    Name = %q
  }
}

resource "aws_subnet" "test" {
  vpc_id     = "${aws_vpc.test.id}"
  cidr_block = "172.16.0.0/24"

  tags = {
    Name = %q
  }
}

resource "aws_instance" "test" {
  ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
  instance_type = "t2.micro"
  subnet_id     = "${aws_subnet.test.id}"

  tags = {
    Name = %q
  }
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["arn:${data.aws_partition.current.partition}:automate:${data.aws_region.current.name}:ec2:%s"]
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
    InstanceId = "${aws_instance.test.id}"
  }
}
`, rName, rName, rName, action, rName)
}

func testAccAWSCloudWatchMetricAlarmConfigAlarmActionsSNSTopic(rName string) string {
	return fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %q
}

resource "aws_cloudwatch_metric_alarm" "test" {
  alarm_actions       = ["${aws_sns_topic.test.arn}"]
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
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

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
