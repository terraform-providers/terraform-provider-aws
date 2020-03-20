package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSRoute53HealthCheck_basic(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "measure_latency", "true"),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "true"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckConfigUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "failure_threshold", "5"),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "false"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_tags(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigTags1("key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
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
				Config: testAccRoute53HealthCheckConfigTags2("key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccRoute53HealthCheckConfigTags1("key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_withSearchString(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigWithSearchString,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "false"),
					resource.TestCheckResourceAttr(resourceName, "search_string", "OK"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckConfigWithSearchStringUpdate,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "invert_healthcheck", "true"),
					resource.TestCheckResourceAttr(resourceName, "search_string", "FAILED"),
				),
			},
		},
	})
}

func TestAccAWSRoute53HealthCheck_withChildHealthChecks(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfig_withChildHealthChecks,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
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

func TestAccAWSRoute53HealthCheck_withHealthCheckRegions(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfig_withHealthCheckRegions,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "regions.#", "3"),
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

func TestAccAWSRoute53HealthCheck_IpConfig(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckIpConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
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

func TestAccAWSRoute53HealthCheck_CloudWatchAlarmCheck(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckCloudWatchAlarm,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_alarm_name", "cloudwatch-healthcheck-alarm"),
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

func TestAccAWSRoute53HealthCheck_withSNI(t *testing.T) {
	resourceName := "aws_route53_health_check.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRoute53HealthCheckDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccRoute53HealthCheckConfigWithoutSNI,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccRoute53HealthCheckConfigWithSNIDisabled,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", "false"),
				),
			},
			{
				Config: testAccRoute53HealthCheckConfigWithSNI,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRoute53HealthCheckExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "enable_sni", "true"),
				),
			},
		},
	})
}

func testAccCheckRoute53HealthCheckDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).r53conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route53_health_check" {
			continue
		}

		lopts := &route53.ListHealthChecksInput{}
		resp, err := conn.ListHealthChecks(lopts)
		if err != nil {
			return err
		}
		if len(resp.HealthChecks) == 0 {
			return nil
		}

		for _, check := range resp.HealthChecks {
			if *check.Id == rs.Primary.ID {
				return fmt.Errorf("Record still exists: %#v", check)
			}

		}

	}
	return nil
}

func testAccCheckRoute53HealthCheckExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).r53conn

		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No health check ID is set")
		}

		lopts := &route53.ListHealthChecksInput{}
		resp, err := conn.ListHealthChecks(lopts)
		if err != nil {
			return err
		}
		if len(resp.HealthChecks) == 0 {
			return fmt.Errorf("Health Check does not exist")
		}

		for _, check := range resp.HealthChecks {
			if *check.Id == rs.Primary.ID {
				return nil
			}

		}
		return fmt.Errorf("Health Check does not exist")
	}
}

const testAccRoute53HealthCheckConfig = `
resource "aws_route53_health_check" "test" {
  fqdn = "dev.notexample.com"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = true
}
`

func testAccRoute53HealthCheckConfigTags1(tag1Key, tag1Value string) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  fqdn = "dev.notexample.com"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = true

  tags = {
    %[1]q = %[2]q
  }
}
`, tag1Key, tag1Value)
}

func testAccRoute53HealthCheckConfigTags2(tag1Key, tag1Value, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_route53_health_check" "test" {
  fqdn = "dev.notexample.com"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = true

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tag1Key, tag1Value, tagKey2, tagValue2)
}

const testAccRoute53HealthCheckConfigUpdate = `
resource "aws_route53_health_check" "test" {
  fqdn = "dev.notexample.com"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "5"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = false
}
`

const testAccRoute53HealthCheckIpConfig = `
resource "aws_route53_health_check" "test" {
  ip_address = "1.2.3.4"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"

  tags = {
    Name = "tf-test-health-check"
   }
}
`

const testAccRoute53HealthCheckConfig_withChildHealthChecks = `
resource "aws_route53_health_check" "child1" {
  fqdn = "child1.notexample.com"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
}

resource "aws_route53_health_check" "test" {
  type = "CALCULATED"
  child_health_threshold = 1
  child_healthchecks = ["${aws_route53_health_check.child1.id}"]

  tags = {
    Name = "tf-test-calculated-health-check"
   }
}
`

const testAccRoute53HealthCheckConfig_withHealthCheckRegions = `
resource "aws_route53_health_check" "test" {
  ip_address = "1.2.3.4"
  port = 80
  type = "HTTP"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"

  regions = ["us-west-1","us-east-1","eu-west-1"]

  tags = {
    Name = "tf-test-check-with-regions"
   }
}
`

const testAccRoute53HealthCheckCloudWatchAlarm = `
resource "aws_cloudwatch_metric_alarm" "test" {
    alarm_name = "cloudwatch-healthcheck-alarm"
    comparison_operator = "GreaterThanOrEqualToThreshold"
    evaluation_periods = "2"
    metric_name = "CPUUtilization"
    namespace = "AWS/EC2"
    period = "120"
    statistic = "Average"
    threshold = "80"
    alarm_description = "This metric monitors ec2 cpu utilization"
}

resource "aws_route53_health_check" "test" {
  type = "CLOUDWATCH_METRIC"
  cloudwatch_alarm_name = "${aws_cloudwatch_metric_alarm.test.alarm_name}"
  cloudwatch_alarm_region = "us-west-2"
  insufficient_data_health_status = "Healthy"
}
`

const testAccRoute53HealthCheckConfigWithSearchString = `
resource "aws_route53_health_check" "test" {
  fqdn = "dev.notexample.com"
  port = 80
  type = "HTTP_STR_MATCH"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = false
  search_string = "OK"

  tags = {
    Name = "tf-test-health-check"
   }
}
`

const testAccRoute53HealthCheckConfigWithSearchStringUpdate = `
resource "aws_route53_health_check" "test" {
  fqdn = "dev.notexample.com"
  port = 80
  type = "HTTP_STR_MATCH"
  resource_path = "/"
  failure_threshold = "5"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = true
  search_string = "FAILED"

  tags = {
    Name = "tf-test-health-check"
   }
}
`

const testAccRoute53HealthCheckConfigWithoutSNI = `
resource "aws_route53_health_check" "test" {
  fqdn = "dev.notexample.com"
  port = 443
  type = "HTTPS"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = true

  tags = {
    Name = "tf-test-health-check"
   }
}
`

const testAccRoute53HealthCheckConfigWithSNI = `
resource "aws_route53_health_check" "test" {
	fqdn = "dev.notexample.com"
  port = 443
  type = "HTTPS"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = true
  enable_sni = true

  tags = {
    Name = "tf-test-health-check"
   }
}
`

const testAccRoute53HealthCheckConfigWithSNIDisabled = `
resource "aws_route53_health_check" "test" {
	fqdn = "dev.notexample.com"
  port = 443
  type = "HTTPS"
  resource_path = "/"
  failure_threshold = "2"
  request_interval = "30"
  measure_latency = true
  invert_healthcheck = true
  enable_sni = false

  tags = {
    Name = "tf-test-health-check"
   }
}
`
