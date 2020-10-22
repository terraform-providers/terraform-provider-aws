package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_flow_log", &resource.Sweeper{
		Name: "aws_flow_log",
		F:    testSweepFlowLogs,
	})
}

func testSweepFlowLogs(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).ec2conn
	var sweeperErrs *multierror.Error

	err = conn.DescribeFlowLogsPages(&ec2.DescribeFlowLogsInput{}, func(page *ec2.DescribeFlowLogsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, flowLog := range page.FlowLogs {
			id := aws.StringValue(flowLog.FlowLogId)

			log.Printf("[INFO] Deleting Flow Log: %s", id)
			_, err := conn.DeleteFlowLogs(&ec2.DeleteFlowLogsInput{
				FlowLogIds: aws.StringSlice([]string{id}),
			})
			if isAWSErr(err, "InvalidFlowLogId.NotFound", "") {
				continue
			}
			if err != nil {
				sweeperErr := fmt.Errorf("error deleting Flow Log (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})
	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Flow Logs sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}
	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving Flow Logs: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSFlowLog_VPCID(t *testing.T) {
	var flowLog ec2.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_VPCID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "ec2", regexp.MustCompile(`vpc-flow-log/fl-.+`)),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_name", cloudwatchLogGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "600"),
					resource.TestCheckResourceAttr(resourceName, "traffic_type", "ALL"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSFlowLog_LogFormat(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	logFormat := "${version} ${vpc-id} ${subnet-id}"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogFormat(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					resource.TestCheckResourceAttr(resourceName, "log_format", logFormat),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config:             testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName),
				ExpectNonEmptyPlan: false,
			},
		},
	})
}

func TestAccAWSFlowLog_SubnetID(t *testing.T) {
	var flowLog ec2.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	iamRoleResourceName := "aws_iam_role.test"
	resourceName := "aws_flow_log.test"
	subnetResourceName := "aws_subnet.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_SubnetID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "iam_role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination", ""),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_name", cloudwatchLogGroupResourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "600"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_id", subnetResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "traffic_type", "ALL"),
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

func TestAccAWSFlowLog_LogDestinationType_CloudWatchLogs(t *testing.T) {
	var flowLog ec2.FlowLog
	cloudwatchLogGroupResourceName := "aws_cloudwatch_log_group.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					// We automatically trim :* from ARNs if present
					testAccCheckResourceAttrRegionalARN(resourceName, "log_destination", "logs", fmt.Sprintf("log-group:%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "cloud-watch-logs"),
					resource.TestCheckResourceAttrPair(resourceName, "log_group_name", cloudwatchLogGroupResourceName, "name"),
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

func TestAccAWSFlowLog_LogDestinationType_S3(t *testing.T) {
	var flowLog ec2.FlowLog
	s3ResourceName := "aws_s3_bucket.test"
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_LogDestinationType_S3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					resource.TestCheckResourceAttrPair(resourceName, "log_destination", s3ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "log_destination_type", "s3"),
					resource.TestCheckResourceAttr(resourceName, "log_group_name", ""),
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

func TestAccAWSFlowLog_LogDestinationType_S3_Invalid(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test-flow-log-s3-invalid")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccFlowLogConfig_LogDestinationType_S3_Invalid(rName),
				ExpectError: regexp.MustCompile(`(Access Denied for LogDestination|does not exist)`),
			},
		},
	})
}

func TestAccAWSFlowLog_LogDestinationType_MaxAggregationInterval(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_MaxAggregationInterval(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					resource.TestCheckResourceAttr(resourceName, "max_aggregation_interval", "60"),
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

func TestAccAWSFlowLog_tags(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
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
				Config: testAccFlowLogConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccFlowLogConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckAWSFlowLogAttributes(&flowLog),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSFlowLog_disappears(t *testing.T) {
	var flowLog ec2.FlowLog
	resourceName := "aws_flow_log.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckFlowLogDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccFlowLogConfig_VPCID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFlowLogExists(resourceName, &flowLog),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsFlowLog(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFlowLogExists(n string, flowLog *ec2.FlowLog) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Flow Log ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		describeOpts := &ec2.DescribeFlowLogsInput{
			FlowLogIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeFlowLogs(describeOpts)
		if err != nil {
			return err
		}

		if len(resp.FlowLogs) > 0 {
			*flowLog = *resp.FlowLogs[0]
			return nil
		}
		return fmt.Errorf("No Flow Logs found for id (%s)", rs.Primary.ID)
	}
}

func testAccCheckAWSFlowLogAttributes(flowLog *ec2.FlowLog) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if flowLog.FlowLogStatus != nil && *flowLog.FlowLogStatus == "ACTIVE" {
			return nil
		}
		if flowLog.FlowLogStatus == nil {
			return fmt.Errorf("Flow Log status is not ACTIVE, is nil")
		} else {
			return fmt.Errorf("Flow Log status is not ACTIVE, got: %s", *flowLog.FlowLogStatus)
		}
	}
}

func testAccCheckFlowLogDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_flow_log" {
			continue
		}

		return nil
	}

	return nil
}

func testAccFlowLogConfigBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_CloudWatchLogs(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn         = aws_iam_role.test.arn
  log_destination      = aws_cloudwatch_log_group.test.arn
  log_destination_type = "cloud-watch-logs"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
}
`, rName)
}

func testAccFlowLogConfig_LogDestinationType_S3_Invalid(rName string) string {
	return testAccFlowLogConfigBase(rName) + `
data "aws_partition" "current" {}

resource "aws_flow_log" "test" {
  log_destination      = "arn:${data.aws_partition.current.partition}:s3:::does-not-exist"
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
}
`
}

func testAccFlowLogConfig_SubnetID(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
resource "aws_subnet" "test" {
  cidr_block = "10.0.1.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  subnet_id      = aws_subnet.test.id
  traffic_type   = "ALL"
}
`, rName)
}

func testAccFlowLogConfig_VPCID(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id
}
`, rName)
}

func testAccFlowLogConfig_LogFormat(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_flow_log" "test" {
  log_destination      = aws_s3_bucket.test.arn
  log_destination_type = "s3"
  traffic_type         = "ALL"
  vpc_id               = aws_vpc.test.id
  log_format           = "$${version} $${vpc-id} $${subnet-id}"
}
`, rName)
}

func testAccFlowLogConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccFlowLogConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccFlowLogConfig_MaxAggregationInterval(rName string) string {
	return testAccFlowLogConfigBase(rName) + fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": [
        "sts:AssumeRole"
      ]
    }
  ]
}
EOF
}

resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_flow_log" "test" {
  iam_role_arn   = aws_iam_role.test.arn
  log_group_name = aws_cloudwatch_log_group.test.name
  traffic_type   = "ALL"
  vpc_id         = aws_vpc.test.id

  max_aggregation_interval = 60
}
`, rName)
}
