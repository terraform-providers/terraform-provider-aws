package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudWatchLogMetricFilter_basic(t *testing.T) {
	var mf cloudwatchlogs.MetricFilter
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_metric_filter.foobar"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, cloudwatchlogs.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogMetricFilterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogMetricFilterConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogMetricFilterExists(resourceName, &mf),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					testAccCheckCloudWatchLogMetricFilterName(&mf, fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "pattern", ""),
					testAccCheckCloudWatchLogMetricFilterPattern(&mf, ""),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", fmt.Sprintf("MyApp/access-%d.log", rInt)),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.name", "EventCount"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.namespace", "YourNamespace"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.value", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.%", "0"),
					testAccCheckCloudWatchLogMetricFilterTransformation(&mf, &cloudwatchlogs.MetricTransformation{
						MetricName:      aws.String("EventCount"),
						MetricNamespace: aws.String("YourNamespace"),
						MetricValue:     aws.String("1"),
					}),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSCloudwatchLogMetricFilterImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSCloudWatchLogMetricFilterConfigModified(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogMetricFilterExists(resourceName, &mf),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					testAccCheckCloudWatchLogMetricFilterName(&mf, fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "pattern", "{ $.errorCode = \"AccessDenied\" }"),
					testAccCheckCloudWatchLogMetricFilterPattern(&mf, "{ $.errorCode = \"AccessDenied\" }"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", fmt.Sprintf("MyApp/access-%d.log", rInt)),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.name", "AccessDeniedCount"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.namespace", "MyNamespace"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.value", "2"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.default_value", "1"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.%", "0"),
					testAccCheckCloudWatchLogMetricFilterTransformation(&mf, &cloudwatchlogs.MetricTransformation{
						MetricName:      aws.String("AccessDeniedCount"),
						MetricNamespace: aws.String("MyNamespace"),
						MetricValue:     aws.String("2"),
						DefaultValue:    aws.Float64(1),
					}),
				),
			},
			{
				Config: testAccAWSCloudWatchLogMetricFilterConfigModifiedWithDimensions(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogMetricFilterExists(resourceName, &mf),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					testAccCheckCloudWatchLogMetricFilterName(&mf, fmt.Sprintf("MyAppAccessCount-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "pattern", "{ $.errorCode = \"AccessDenied\" }"),
					testAccCheckCloudWatchLogMetricFilterPattern(&mf, "{ $.errorCode = \"AccessDenied\" }"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", fmt.Sprintf("MyApp/access-%d.log", rInt)),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.name", "AccessDeniedCount"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.namespace", "MyNamespace"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.value", "2"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.ErrorCode", "$.errorCode"),
					resource.TestCheckResourceAttr(resourceName, "metric_transformation.0.dimensions.Dummy", "$.dummy"),
					testAccCheckCloudWatchLogMetricFilterTransformation(&mf, &cloudwatchlogs.MetricTransformation{
						MetricName:      aws.String("AccessDeniedCount"),
						MetricNamespace: aws.String("MyNamespace"),
						MetricValue:     aws.String("2"),
						Dimensions: aws.StringMap(map[string]string{
							"ErrorCode": "$.errorCode",
							"Dummy":     "$.dummy",
						}),
					}),
				),
			},
			{
				Config: testAccAWSCloudwatchLogMetricFilterConfigMany(rInt),
				Check:  testAccCheckCloudwatchLogMetricFilterManyExist("aws_cloudwatch_log_metric_filter.count_dracula", &mf),
			},
		},
	})
}

func testAccCheckCloudWatchLogMetricFilterName(mf *cloudwatchlogs.MetricFilter, name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if name != *mf.FilterName {
			return fmt.Errorf("Expected filter name: %q, given: %q", name, *mf.FilterName)
		}
		return nil
	}
}

func testAccCheckCloudWatchLogMetricFilterPattern(mf *cloudwatchlogs.MetricFilter, pattern string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if mf.FilterPattern == nil {
			if pattern != "" {
				return fmt.Errorf("Received empty filter pattern, expected: %q", pattern)
			}
			return nil
		}

		if pattern != *mf.FilterPattern {
			return fmt.Errorf("Expected filter pattern: %q, given: %q", pattern, *mf.FilterPattern)
		}
		return nil
	}
}

func testAccCheckCloudWatchLogMetricFilterTransformation(mf *cloudwatchlogs.MetricFilter,
	t *cloudwatchlogs.MetricTransformation) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		given := mf.MetricTransformations[0]
		expected := t

		if *given.MetricName != *expected.MetricName {
			return fmt.Errorf("Expected metric name: %q, received: %q",
				*expected.MetricName, *given.MetricName)
		}

		if *given.MetricNamespace != *expected.MetricNamespace {
			return fmt.Errorf("Expected metric namespace: %q, received: %q",
				*expected.MetricNamespace, *given.MetricNamespace)
		}

		if *given.MetricValue != *expected.MetricValue {
			return fmt.Errorf("Expected metric value: %q, received: %q",
				*expected.MetricValue, *given.MetricValue)
		}

		if (given.DefaultValue != nil) != (expected.DefaultValue != nil) {
			return fmt.Errorf("Expected default value to be present: %t, received: %t",
				expected.DefaultValue != nil, given.DefaultValue != nil)
		} else if (given.DefaultValue != nil) && *given.DefaultValue != *expected.DefaultValue {
			return fmt.Errorf("Expected metric value: %g, received: %g",
				*expected.DefaultValue, *given.DefaultValue)
		}

		if len(expected.Dimensions) > 0 || len(given.Dimensions) > 0 {
			e, g := aws.StringValueMap(expected.Dimensions), aws.StringValueMap(given.Dimensions)

			if len(e) != len(g) {
				return fmt.Errorf("Expected %d dimensions, received %d", len(e), len(g))
			}

			for ek, ev := range e {
				gv, ok := g[ek]
				if !ok {
					return fmt.Errorf("Expected dimension %s, received nothing", ek)
				}
				if gv != ev {
					return fmt.Errorf("Expected dimension %s to be %s, received %s", ek, ev, gv)
				}
			}
		}

		return nil
	}
}

func testAccCheckCloudWatchLogMetricFilterExists(n string, mf *cloudwatchlogs.MetricFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		metricFilter, err := lookupCloudWatchLogMetricFilter(conn, rs.Primary.Attributes["name"], rs.Primary.Attributes["log_group_name"], nil)
		if err != nil {
			return err
		}

		*mf = *metricFilter

		return nil
	}
}

func testAccCheckAWSCloudWatchLogMetricFilterDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_metric_filter" {
			continue
		}

		_, err := lookupCloudWatchLogMetricFilter(conn, rs.Primary.Attributes["name"], rs.Primary.Attributes["log_group_name"], nil)
		if err == nil {
			return fmt.Errorf("MetricFilter Still Exists: %s", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckCloudwatchLogMetricFilterManyExist(basename string, mf *cloudwatchlogs.MetricFilter) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for i := 0; i < 15; i++ {
			n := fmt.Sprintf("%s.%d", basename, i)
			testfunc := testAccCheckCloudWatchLogMetricFilterExists(n, mf)
			err := testfunc(s)
			if err != nil {
				return err
			}
		}

		return nil
	}
}

func testAccAWSCloudWatchLogMetricFilterConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_metric_filter" "foobar" {
  name           = "MyAppAccessCount-%d"
  pattern        = ""
  log_group_name = aws_cloudwatch_log_group.dada.name

  metric_transformation {
    name      = "EventCount"
    namespace = "YourNamespace"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "dada" {
  name = "MyApp/access-%d.log"
}
`, rInt, rInt)
}

func testAccAWSCloudWatchLogMetricFilterConfigModified(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_metric_filter" "foobar" {
  name = "MyAppAccessCount-%d"

  pattern = <<PATTERN
{ $.errorCode = "AccessDenied" }
PATTERN


  log_group_name = aws_cloudwatch_log_group.dada.name

  metric_transformation {
    name          = "AccessDeniedCount"
    namespace     = "MyNamespace"
    value         = "2"
    default_value = "1"
  }
}

resource "aws_cloudwatch_log_group" "dada" {
  name = "MyApp/access-%d.log"
}
`, rInt, rInt)
}

func testAccAWSCloudWatchLogMetricFilterConfigModifiedWithDimensions(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_metric_filter" "foobar" {
  name = "MyAppAccessCount-%d"

  pattern = <<PATTERN
{ $.errorCode = "AccessDenied" }
PATTERN


  log_group_name = aws_cloudwatch_log_group.dada.name

  metric_transformation {
    name      = "AccessDeniedCount"
    namespace = "MyNamespace"
    value     = "2"
    dimensions = {
      ErrorCode = "$.errorCode"
      Dummy     = "$.dummy"
    }
  }
}

resource "aws_cloudwatch_log_group" "dada" {
  name = "MyApp/access-%d.log"
}
`, rInt, rInt)
}

func testAccAWSCloudwatchLogMetricFilterConfigMany(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_metric_filter" "count_dracula" {
  count          = 15
  name           = "MyAppCountLog-${count.index}-%d"
  pattern        = "count ${count.index}"
  log_group_name = aws_cloudwatch_log_group.mama.name

  metric_transformation {
    name      = "CountDracula-${count.index}"
    namespace = "CountNamespace"
    value     = "1"
  }
}

resource "aws_cloudwatch_log_group" "mama" {
  name = "MyApp/count-log-%d.log"
}
`, rInt, rInt)
}

func testAccAWSCloudwatchLogMetricFilterImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["log_group_name"] + ":" + rs.Primary.Attributes["name"], nil
	}
}
