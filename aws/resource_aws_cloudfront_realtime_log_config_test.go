package aws

import (
	"fmt"
	"log"
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudfront"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/cloudfront/finder"
)

func init() {
	resource.AddTestSweepers("aws_cloudfront_realtime_log_config", &resource.Sweeper{
		Name: "aws_cloudfront_realtime_log_config",
		F:    testSweepCloudFrontRealtimeLogConfigs,
	})
}

func testSweepCloudFrontRealtimeLogConfigs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudfrontconn
	input := &cloudfront.ListRealtimeLogConfigsInput{}
	var sweeperErrs *multierror.Error

	for {
		output, err := conn.ListRealtimeLogConfigs(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping CloudFront Real-time Log Configs sweep for %s: %s", region, err)
			return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
		}

		if err != nil {
			sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudFront Real-time Log Configs: %w", err))
			return sweeperErrs
		}

		for _, config := range output.RealtimeLogConfigs.Items {
			id := aws.StringValue(config.ARN)

			log.Printf("[INFO] Deleting CloudFront Real-time Log Config: %s", id)
			r := resourceAwsCloudFrontRealtimeLogConfig()
			d := r.Data(nil)
			d.SetId(id)
			err := r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		if aws.StringValue(output.RealtimeLogConfigs.NextMarker) == "" {
			break
		}
		input.Marker = output.RealtimeLogConfigs.NextMarker
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudFrontRealtimeLogConfig_basic(t *testing.T) {
	var v cloudfront.RealtimeLogConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	samplingRate := acctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	roleResourceName := "aws_iam_role.test.0"
	streamResourceName := "aws_kinesis_stream.test.0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontRealtimeLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontRealtimeLogConfigConfig(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontRealtimeLogConfigExists(resourceName, &v),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", roleResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", streamResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "timestamp"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sampling_rate", strconv.Itoa(samplingRate)),
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

func TestAccAWSCloudFrontRealtimeLogConfig_disappears(t *testing.T) {
	var v cloudfront.RealtimeLogConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	samplingRate := acctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontRealtimeLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontRealtimeLogConfigConfig(rName, samplingRate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontRealtimeLogConfigExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsCloudFrontRealtimeLogConfig(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudFrontRealtimeLogConfig_updates(t *testing.T) {
	var v cloudfront.RealtimeLogConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	samplingRate1 := acctest.RandIntRange(1, 100)
	samplingRate2 := acctest.RandIntRange(1, 100)
	resourceName := "aws_cloudfront_realtime_log_config.test"
	role1ResourceName := "aws_iam_role.test.0"
	stream1ResourceName := "aws_kinesis_stream.test.0"
	role2ResourceName := "aws_iam_role.test.1"
	stream2ResourceName := "aws_kinesis_stream.test.1"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(cloudfront.EndpointsID, t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudFrontRealtimeLogConfigDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudFrontRealtimeLogConfigConfig(rName, samplingRate1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontRealtimeLogConfigExists(resourceName, &v),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", role1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", stream1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "2"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "timestamp"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sampling_rate", strconv.Itoa(samplingRate1)),
				),
			},
			{
				Config: testAccAWSCloudFrontRealtimeLogConfigConfigUpdated(rName, samplingRate2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudFrontRealtimeLogConfigExists(resourceName, &v),
					testAccCheckResourceAttrGlobalARN(resourceName, "arn", "cloudfront", fmt.Sprintf("realtime-log-config/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "endpoint.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.stream_type", "Kinesis"),
					resource.TestCheckResourceAttr(resourceName, "endpoint.0.kinesis_stream_config.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.role_arn", role2ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "endpoint.0.kinesis_stream_config.0.stream_arn", stream2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "fields.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "c-ip"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "cs-host"),
					resource.TestCheckTypeSetElemAttr(resourceName, "fields.*", "sc-status"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "sampling_rate", strconv.Itoa(samplingRate2)),
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

func testAccCheckCloudFrontRealtimeLogConfigDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudfront_realtime_log_config" {
			continue
		}

		// Try to find the resource
		_, err := finder.RealtimeLogConfigByARN(conn, rs.Primary.ID)
		// Verify the error is what we want
		if isAWSErr(err, cloudfront.ErrCodeNoSuchRealtimeLogConfig, "") {
			continue
		}
		if err != nil {
			return err
		}
		return fmt.Errorf("CloudFront Real-time Log Config still exists: %s", rs.Primary.ID)
	}

	return nil
}

func testAccCheckCloudFrontRealtimeLogConfigExists(n string, v *cloudfront.RealtimeLogConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No CloudFront Real-time Log Config ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudfrontconn
		out, err := finder.RealtimeLogConfigByARN(conn, rs.Primary.ID)
		if err != nil {
			return err
		}

		*v = *out

		return nil
	}
}

func testAccAWSCloudFrontRealtimeLogConfigConfigBase(rName string, count int) string {
	return fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  count = %[2]d

  name        = format("%%s-%%d", %[1]q, count.index)
  shard_count = 2
}

resource "aws_iam_role" "test" {
  count = %[2]d

  name = format("%%s-%%d", %[1]q, count.index)

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "cloudfront.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  count = %[2]d

  name = format("%%s-%%d", %[1]q, count.index)
  role = aws_iam_role.test[count.index].id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
      "kinesis:DescribeStreamSummary",
      "kinesis:DescribeStream",
      "kinesis:PutRecord",
      "kinesis:PutRecords"
    ],
    "Resource": "${aws_kinesis_stream.test[count.index].arn}"
  }]
}
EOF
}
`, rName, count)
}

func testAccAWSCloudFrontRealtimeLogConfigConfig(rName string, samplingRate int) string {
	return composeConfig(
		testAccAWSCloudFrontRealtimeLogConfigConfigBase(rName, 1),
		fmt.Sprintf(`
resource "aws_cloudfront_realtime_log_config" "test" {
  name          = %[1]q
  sampling_rate = %[2]d
  fields        = ["timestamp", "c-ip"]

  endpoint {
    stream_type = "Kinesis"

    kinesis_stream_config {
      role_arn   = aws_iam_role.test[0].arn
      stream_arn = aws_kinesis_stream.test[0].arn
    }
  }

  depends_on = [aws_iam_role_policy.test[0]]
}
`, rName, samplingRate))
}

func testAccAWSCloudFrontRealtimeLogConfigConfigUpdated(rName string, samplingRate int) string {
	return composeConfig(
		testAccAWSCloudFrontRealtimeLogConfigConfigBase(rName, 2),
		fmt.Sprintf(`
resource "aws_cloudfront_realtime_log_config" "test" {
  name          = %[1]q
  sampling_rate = %[2]d
  fields        = ["c-ip", "cs-host", "sc-status"]

  endpoint {
    stream_type = "Kinesis"

    kinesis_stream_config {
      role_arn   = aws_iam_role.test[1].arn
      stream_arn = aws_kinesis_stream.test[1].arn
    }
  }

  depends_on = [aws_iam_role_policy.test[1]]
}
`, rName, samplingRate))
}
