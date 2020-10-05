package aws

import (
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatch"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSCloudwatchMetricsDataSource_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "data.aws_cloudwatch_metrics.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccCheckAWSCloudwatchMetricsPutConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSCloudwatchMetricsDataPut(rName),
				),
			},
			{
				Config: testAccCheckAWSCloudwatchMetricsDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "metrics.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "metrics.0.namespace", rName),
					resource.TestCheckResourceAttr(resourceName, "metrics.0.metric_name", "test"),
					resource.TestCheckResourceAttr(resourceName, "metrics.0.dimensions.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "metrics.0.dimensions.0.name"),
					resource.TestCheckResourceAttr(resourceName, "metrics.0.dimensions.0.value", rName),
					resource.TestCheckResourceAttr(resourceName, "metrics.0.dimensions.1.value", rName),
				),
			},
		},
	})
}

func testAccCheckAWSCloudwatchMetricsDataPut(rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchconn

		input := &cloudwatch.PutMetricDataInput{
			Namespace: aws.String(rName),
			MetricData: []*cloudwatch.MetricDatum{
				{
					Dimensions: []*cloudwatch.Dimension{
						{
							Name:  aws.String("test"),
							Value: aws.String(rName),
						},
						{
							Name:  aws.String("test2"),
							Value: aws.String(rName),
						},
					},
					MetricName: aws.String("test"),
					Timestamp:  aws.Time(time.Now()),
					Unit:       aws.String(cloudwatch.StandardUnitBytes),
					Value:      aws.Float64(1),
				},
			},
		}

		_, err := conn.PutMetricData(input)
		if err != nil {
			return fmt.Errorf("Couldn't perform a PutMetricData: %s", err)
		}

		input2 := &cloudwatch.ListMetricsInput{
			Namespace:  aws.String(rName),
			MetricName: aws.String("test"),
		}

		// wait until metrics are visible
		var metrics []*cloudwatch.Metric
		for i := 0; i < 10; i++ {
			metrics, err = dataSourceAwsCloudwatchMetricsLookup(conn, input2)

			if err != nil {
				return err
			}
			if len(metrics) > 0 {
				break
			}
			time.Sleep(15 * time.Second)
		}

		if len(metrics) == 0 {
			return fmt.Errorf("Couldn't retrieve metric after PutMetricData")
		}

		return nil
	}
}

func testAccCheckAWSCloudwatchMetricsPutConfig(rName string) string {
	return `
	`
}

func testAccCheckAWSCloudwatchMetricsDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
data aws_cloudwatch_metrics "test" {
  namespace = "%s"
  metric_name = "test"
  dimensions {
	  name  = "test"
	  value = "%s"
  }
}
`, rName, rName)
}
