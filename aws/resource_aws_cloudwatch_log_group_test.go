package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_cloudwatch_log_group", &resource.Sweeper{
		Name: "aws_cloudwatch_log_group",
		F:    testSweepCloudwatchLogGroups,
		Dependencies: []string{
			"aws_api_gateway_rest_api",
			"aws_cloudhsm_v2_cluster",
			"aws_cloudtrail",
			"aws_datasync_task",
			"aws_db_instance",
			"aws_directory_service_directory",
			"aws_ec2_client_vpn_endpoint",
			"aws_eks_cluster",
			"aws_elasticsearch_domain",
			"aws_flow_log",
			"aws_glue_job",
			"aws_kinesis_analytics_application",
			"aws_kinesis_firehose_delivery_stream",
			"aws_lambda_function",
			"aws_mq_broker",
			"aws_msk_cluster",
			"aws_rds_cluster",
			"aws_route53_query_log",
			"aws_sagemaker_endpoint",
			"aws_storagegateway_gateway",
		},
	})
}

func testSweepCloudwatchLogGroups(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).cloudwatchlogsconn
	var sweeperErrs *multierror.Error

	input := &cloudwatchlogs.DescribeLogGroupsInput{}

	err = conn.DescribeLogGroupsPages(input, func(page *cloudwatchlogs.DescribeLogGroupsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, logGroup := range page.LogGroups {
			if logGroup == nil {
				continue
			}

			input := &cloudwatchlogs.DeleteLogGroupInput{
				LogGroupName: logGroup.LogGroupName,
			}
			name := aws.StringValue(logGroup.LogGroupName)

			log.Printf("[INFO] Deleting CloudWatch Log Group: %s", name)
			_, err := conn.DeleteLogGroup(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting CloudWatch Log Group (%s): %w", name, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping CloudWatch Log Groups sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error retrieving CloudWatch Log Groups: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSCloudWatchLogGroup_basic(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "logs", fmt.Sprintf("log-group:foo-bar-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "name", fmt.Sprintf("foo-bar-%d", rInt)),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"}, //this has a default value
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_namePrefix(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroup_namePrefix,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "name_prefix"},
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_namePrefix_retention(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rName := acctest.RandString(5)
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroup_namePrefix_retention(rName, 365),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days", "name_prefix"},
			},
			{
				Config: testAccAWSCloudWatchLogGroup_namePrefix_retention(rName, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestMatchResourceAttr(resourceName, "name", regexp.MustCompile("^tf-test-")),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "7"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_generatedName(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroup_generatedName,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_retentionPolicy(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig_withRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "365"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigModified_withRetention(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "retention_in_days", "0"),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_multiple(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.alpha"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig_multiple(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists("aws_cloudwatch_log_group.alpha", &lg),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_group.alpha", "retention_in_days", "14"),
					testAccCheckCloudWatchLogGroupExists("aws_cloudwatch_log_group.beta", &lg),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_group.beta", "retention_in_days", "0"),
					testAccCheckCloudWatchLogGroupExists("aws_cloudwatch_log_group.charlie", &lg),
					resource.TestCheckResourceAttr("aws_cloudwatch_log_group.charlie", "retention_in_days", "3653"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_disappears(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					testAccCheckCloudWatchLogGroupDisappears(&lg),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_tagging(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTagsAdded(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", ""),
					resource.TestCheckResourceAttr(resourceName, "tags.Bar", "baz"),
				),
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTagsUpdated(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "4"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Development"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", "NotEmpty"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "UpdatedBar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Bar", "baz"),
				),
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithTags(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "3"),
					resource.TestCheckResourceAttr(resourceName, "tags.Environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.Foo", "Bar"),
					resource.TestCheckResourceAttr(resourceName, "tags.Empty", ""),
				),
			},
		},
	})
}

func TestAccAWSCloudWatchLogGroup_kmsKey(t *testing.T) {
	var lg cloudwatchlogs.LogGroup
	rInt := acctest.RandInt()
	resourceName := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSCloudWatchLogGroupDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSCloudWatchLogGroupConfig(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"retention_in_days"},
			},
			{
				Config: testAccAWSCloudWatchLogGroupConfigWithKmsKeyId(rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudWatchLogGroupExists(resourceName, &lg),
					resource.TestCheckResourceAttrSet(resourceName, "kms_key_id"),
				),
			},
		},
	})
}

func testAccCheckCloudWatchLogGroupDisappears(lg *cloudwatchlogs.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		opts := &cloudwatchlogs.DeleteLogGroupInput{
			LogGroupName: lg.LogGroupName,
		}
		_, err := conn.DeleteLogGroup(opts)
		return err
	}
}

func testAccCheckCloudWatchLogGroupExists(n string, lg *cloudwatchlogs.LogGroup) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn
		logGroup, err := lookupCloudWatchLogGroup(conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if logGroup == nil {
			return fmt.Errorf("Bad: LogGroup %q does not exist", rs.Primary.ID)
		}

		*lg = *logGroup

		return nil
	}
}

func testAccCheckAWSCloudWatchLogGroupDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).cloudwatchlogsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_cloudwatch_log_group" {
			continue
		}
		logGroup, err := lookupCloudWatchLogGroup(conn, rs.Primary.ID)
		if err != nil {
			return nil
		}

		if logGroup != nil {
			return fmt.Errorf("Bad: LogGroup still exists: %q", rs.Primary.ID)
		}

	}

	return nil
}

func testAccAWSCloudWatchLogGroupConfig(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigWithTags(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"

  tags = {
    Environment = "Production"
    Foo         = "Bar"
    Empty       = ""
  }
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigWithTagsAdded(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"

  tags = {
    Environment = "Development"
    Foo         = "Bar"
    Empty       = ""
    Bar         = "baz"
  }
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigWithTagsUpdated(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"

  tags = {
    Environment = "Development"
    Foo         = "UpdatedBar"
    Empty       = "NotEmpty"
    Bar         = "baz"
  }
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfig_withRetention(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name              = "foo-bar-%d"
  retention_in_days = 365
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfigModified_withRetention(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = "foo-bar-%d"
}
`, rInt)
}

func testAccAWSCloudWatchLogGroupConfig_multiple(rInt int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "alpha" {
  name              = "foo-bar-%d"
  retention_in_days = 14
}

resource "aws_cloudwatch_log_group" "beta" {
  name = "foo-bar-%d"
}

resource "aws_cloudwatch_log_group" "charlie" {
  name              = "foo-bar-%d"
  retention_in_days = 3653
}
`, rInt, rInt+1, rInt+2)
}

func testAccAWSCloudWatchLogGroupConfigWithKmsKeyId(rInt int) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "foo" {
  description             = "Terraform acc test %d"
  deletion_window_in_days = 7

  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Id": "kms-tf-1",
  "Statement": [
    {
      "Sid": "Enable IAM User Permissions",
      "Effect": "Allow",
      "Principal": {
        "AWS": "*"
      },
      "Action": "kms:*",
      "Resource": "*"
    }
  ]
}
POLICY
}

resource "aws_cloudwatch_log_group" "test" {
  name       = "foo-bar-%d"
  kms_key_id = aws_kms_key.foo.arn
}
`, rInt, rInt)
}

const testAccAWSCloudWatchLogGroup_namePrefix = `
resource "aws_cloudwatch_log_group" "test" {
  name_prefix = "tf-test-"
}
`

func testAccAWSCloudWatchLogGroup_namePrefix_retention(rName string, retention int) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name_prefix       = "tf-test-%s"
  retention_in_days = %d
}
`, rName, retention)
}

const testAccAWSCloudWatchLogGroup_generatedName = `
resource "aws_cloudwatch_log_group" "test" {}
`
