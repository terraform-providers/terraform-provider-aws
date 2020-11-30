package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalyticsv2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalyticsv2/lister"
)

func init() {
	resource.AddTestSweepers("aws_kinesisanalyticsv2_application", &resource.Sweeper{
		Name: "aws_kinesisanalyticsv2_application",
		F:    testSweepKinesisAnalyticsV2Application,
	})
}

func testSweepKinesisAnalyticsV2Application(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).kinesisanalyticsv2conn
	input := &kinesisanalyticsv2.ListApplicationsInput{}
	var sweeperErrs *multierror.Error

	err = lister.ListApplicationsPages(conn, input, func(page *kinesisanalyticsv2.ListApplicationsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, applicationSummary := range page.ApplicationSummaries {
			arn := aws.StringValue(applicationSummary.ApplicationARN)
			name := aws.StringValue(applicationSummary.ApplicationName)

			application, err := finder.ApplicationByName(conn, name)

			if err != nil {
				sweeperErr := fmt.Errorf("error reading Kinesis Analytics v2 Application (%s): %w", arn, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			r := resourceAwsKinesisAnalyticsV2Application()
			d := r.Data(nil)
			d.SetId(arn)
			d.Set("create_timestamp", aws.TimeValue(application.CreateTimestamp).Format(time.RFC3339))
			d.Set("name", name)
			err = r.Delete(d, client)

			if err != nil {
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, err)
				continue
			}
		}

		return !isLast
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping Kinesis Analytics v2 Application sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Kinesis Analytics v2 Applications: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSKinesisAnalyticsV2Application_basicFlinkApplication(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicFlinkApplication(rName, "FLINK-1_6"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_6"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicFlinkApplication(rName, "FLINK-1_8"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicFlinkApplication(rName, "FLINK-1_11"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_11"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_basicSQLApplication(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
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

func TestAccAWSKinesisAnalyticsV2Application_disappears(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisAnalyticsV2Application(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsV2Application_Tags(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsV2Application_ApplicationCodeConfiguration_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigApplicationCodeConfiguration(rName, "SELECT 1;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigApplicationCodeConfiguration(rName, "SELECT 2;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 2;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_CloudWatchLoggingOptions_Add(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_CloudWatchLoggingOptions_Delete(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplication(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_CloudWatchLoggingOptions_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStream1ResourceName := "aws_cloudwatch_log_stream.test.0"
	cloudWatchLogStream2ResourceName := "aws_cloudwatch_log_stream.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream1ResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream2ResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_EnvironmentProperties_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketObjectResourceName := "aws_s3_bucket_object.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigEnvironmentProperties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID1",
						"property_map.%":    "2",
						"property_map.Key9": "Value1",
						"property_map.Key8": "Value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    "3",
						"property_map.KeyA": "ValueZ",
						"property_map.KeyB": "ValueY",
						"property_map.KeyC": "ValueX",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigEnvironmentPropertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    "2",
						"property_map.KeyA": "ValueZ",
						"property_map.KeyC": "ValueW",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID3",
						"property_map.%":    "1",
						"property_map.Key":  "Value",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":     "PROPERTY-GROUP-ID4",
						"property_map.%":        "1",
						"property_map.KeyAlpha": "ValueOmega",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigEnvironmentPropertiesNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "3"),
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsV2Application_FlinkApplicationConfiguration_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketObject1ResourceName := "aws_s3_bucket_object.test.0"
	s3BucketObject2ResourceName := "aws_s3_bucket_object.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObject1ResourceName, "key"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3BucketObject1ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "10"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "4"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObject2ResourceName, "key"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3BucketObject2ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "55000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5500"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "OPERATOR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_FlinkApplicationConfiguration_EnvironmentProperties_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketObject1ResourceName := "aws_s3_bucket_object.test.0"
	s3BucketObject2ResourceName := "aws_s3_bucket_object.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfigurationEnvironmentProperties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObject1ResourceName, "key"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3BucketObject1ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID1",
						"property_map.%":    "2",
						"property_map.Key9": "Value1",
						"property_map.Key8": "Value2",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    "3",
						"property_map.KeyA": "ValueZ",
						"property_map.KeyB": "ValueY",
						"property_map.KeyC": "ValueX",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "DEBUG"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "TASK"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "10"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "4"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfigurationEnvironmentPropertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObject2ResourceName, "key"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", s3BucketObject2ResourceName, "version_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID2",
						"property_map.%":    "2",
						"property_map.KeyA": "ValueZ",
						"property_map.KeyC": "ValueW",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id": "PROPERTY-GROUP-ID3",
						"property_map.%":    "1",
						"property_map.Key":  "Value",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.environment_properties.0.property_group.*", map[string]string{
						"property_group_id":     "PROPERTY-GROUP-ID4",
						"property_map.%":        "1",
						"property_map.KeyAlpha": "ValueOmega",
					}),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "55000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5500"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "CUSTOM"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "ERROR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "OPERATOR"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_ServiceExecutionRole_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplicationPlusDescription(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "Testing"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplicationServiceExecutionRoleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "Testing"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_Input_Add(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_Input_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_InputProcessingConfiguration_Add(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "3"), // Add input processing configuration + update input.
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_InputProcessingConfiguration_Delete(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "3"), // Delete input processing configuration + update input.
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_InputProcessingConfiguration_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambda1ResourceName := "aws_lambda_function.test.0"
	lambda2ResourceName := "aws_lambda_function.test.1"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambda1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInputProcessingConfiguration(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambda2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_Multiple_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.0.input_lambda_processor.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                 "OUTPUT_1",
						"destination_schema.#": "1",
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               "1",
						"kinesis_streams_output.#":                "0",
						"lambda_output.#":                         "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_firehose_output.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, "arn"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationMultipleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 2;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.in_app_stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_column.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.input_processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_firehose_input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.input.0.kinesis_streams_input.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                 "OUTPUT_2",
						"destination_schema.#": "1",
						"destination_schema.0.record_format_type": "JSON",
						"kinesis_firehose_output.#":               "0",
						"kinesis_streams_output.#":                "1",
						"lambda_output.#":                         "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_streams_output.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                 "OUTPUT_3",
						"destination_schema.#": "1",
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               "0",
						"kinesis_streams_output.#":                "0",
						"lambda_output.#":                         "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.lambda_output.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "8"), // Delete CloudWatch logging options + add reference data source + delete input processing configuration+ update application + delete output + 2 * add output.
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_Output_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                 "OUTPUT_1",
						"destination_schema.#": "1",
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               "1",
						"kinesis_streams_output.#":                "0",
						"lambda_output.#":                         "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_firehose_output.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationOutputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                 "OUTPUT_2",
						"destination_schema.#": "1",
						"destination_schema.0.record_format_type": "JSON",
						"kinesis_firehose_output.#":               "0",
						"kinesis_streams_output.#":                "1",
						"lambda_output.#":                         "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.kinesis_streams_output.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "application_configuration.0.sql_application_configuration.0.output.*", map[string]string{
						"name":                 "OUTPUT_3",
						"destination_schema.#": "1",
						"destination_schema.0.record_format_type": "CSV",
						"kinesis_firehose_output.#":               "0",
						"kinesis_streams_output.#":                "0",
						"lambda_output.#":                         "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.output.*.lambda_output.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "4"), // 1 * output deletion + 2 * output addition.
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "6"), // 2 * output deletion.
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_ReferenceDataSource_Add(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_ReferenceDataSource_Delete(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_ReferenceDataSource_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationReferenceDataSourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.text_content", "SELECT 1;\n"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "PLAINTEXT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.input.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.output.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_column.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.csv_mapping_parameters.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.mapping_parameters.0.json_mapping_parameters.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.s3_reference_data_source.0.file_key", "KEY-2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.table_name", "TABLE-2"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.sql_application_configuration.0.reference_data_source.0.reference_id"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "SQL-1_0"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_VPCConfiguration_Add(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketObjectResourceName := "aws_s3_bucket_object.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigVPCConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigVPCConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, "id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_VPCConfiguration_Delete(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketObjectResourceName := "aws_s3_bucket_object.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigVPCConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, "id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigVPCConfigurationNotSpecified(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "0"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func TestAccAWSKinesisAnalyticsV2Application_SQLApplicationConfiguration_VPCConfiguration_Update(t *testing.T) {
	var v kinesisanalyticsv2.ApplicationDetail
	resourceName := "aws_kinesisanalyticsv2_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	s3BucketObjectResourceName := "aws_s3_bucket_object.test.0"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalyticsV2(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsV2ApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigVPCConfiguration(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, "id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsV2ApplicationConfigVPCConfigurationUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsV2ApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.file_key", s3BucketObjectResourceName, "key"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content.0.s3_content_location.0.object_version", ""),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_code_configuration.0.code_content_type", "ZIPFILE"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.application_snapshot_configuration.0.snapshots_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.environment_properties.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpointing_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.checkpoint_interval", "60000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.checkpoint_configuration.0.min_pause_between_checkpoints", "5000"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.log_level", "INFO"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.monitoring_configuration.0.metrics_level", "APPLICATION"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.auto_scaling_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.configuration_type", "DEFAULT"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.flink_application_configuration.0.parallelism_configuration.0.parallelism_per_kpu", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.sql_application_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "application_configuration.0.vpc_configuration.0.subnet_ids.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "application_configuration.0.vpc_configuration.0.vpc_configuration_id"),
					resource.TestCheckResourceAttrPair(resourceName, "application_configuration.0.vpc_configuration.0.vpc_id", vpcResourceName, "id"),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "runtime_environment", "FLINK-1_8"),
					resource.TestCheckResourceAttrPair(resourceName, "service_execution_role", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version_id", "2"),
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

func testAccCheckKinesisAnalyticsV2ApplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesisanalyticsv2_application" {
			continue
		}

		_, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if isAWSErr(err, kinesisanalyticsv2.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Kinesis Analytics v2 Application %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckKinesisAnalyticsV2ApplicationExists(n string, v *kinesisanalyticsv2.ApplicationDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics v2 Application ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

		application, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		*v = *application

		return nil
	}
}

func testAccPreCheckAWSKinesisAnalyticsV2(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsv2conn

	input := &kinesisanalyticsv2.ListApplicationsInput{}

	_, err := conn.ListApplications(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  count = 2

  name               = "%[1]s.${count.index}"
  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "firehose.amazonaws.com"}
    },
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "kinesisanalytics.amazonaws.com"}
    },
    {
      "Effect": "Allow",
      "Action": ["sts:AssumeRole"],
      "Principal": {"Service": "lambda.amazonaws.com"}
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test" {
  name   = %[1]q
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["ec2:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["firehose:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["lambda:*"],
      "Resource": ["*"]
    },
    {
      "Effect": "Allow",
      "Action": ["s3:*"],
      "Resource": ["*"]
    }
  ]
}
EOF
}

resource "aws_iam_role_policy_attachment" "test" {
  count = 2

  role       = aws_iam_role.test[count.index].name
  policy_arn = aws_iam_policy.test.arn
}
`, rName)
}

func testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_object" "test" {
  count = 2

  bucket = aws_s3_bucket.test.bucket
  key    = "%[1]s.${count.index}"
  source = "test-fixtures/flink-app.jar"
}
`, rName)
}

func testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_lambda_function" "test" {
  count = 2

  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s_${count.index}"
  handler       = "exports.example"
  role          = aws_iam_role.test[0].arn
  runtime       = "nodejs12.x"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.test.arn
    role_arn   = aws_iam_role.test[0].arn
  }
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}
`, rName)
}

func testAccKinesisAnalyticsV2ApplicationConfigBaseVPC(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  vpc_id            = aws_vpc.test.id
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  count = 2

  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigBasicFlinkApplication(rName, runtimeEnvironment string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = %[2]q
  service_execution_role = aws_iam_role.test[0].arn
}
`, rName, runtimeEnvironment))
}

func testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplication(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplicationPlusDescription(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn
  description            = "Testing"
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigBasicSQLApplicationServiceExecutionRoleUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[1].arn
  description            = "Testing"
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigApplicationCodeConfiguration(rName, textContent string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = %[2]q
      }

      code_content_type = "PLAINTEXT"
    }
  }
}
`, rName, textContent))
}

func testAccKinesisAnalyticsV2ApplicationConfigCloudWatchLoggingOptions(rName string, streamIndex int) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  count = 2

  name           = "%[1]s.${count.index}"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.%[2]d.arn
  }
}
`, rName, streamIndex))
}

func testAccKinesisAnalyticsV2ApplicationConfigEnvironmentProperties(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID1"

        property_map = {
          Key9 = "Value1"
          Key8 = "Value2"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyB = "ValueY"
          KeyC = "ValueX"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigEnvironmentPropertiesUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID3"

        property_map = {
          Key = "Value"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyC = "ValueW"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID4"

        property_map = {
          KeyAlpha = "ValueOmega"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigEnvironmentPropertiesNotSpecified(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfiguration(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_bucket_object.test[0].key
          object_version = aws_s3_bucket_object.test[0].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = false
    }

    flink_application_configuration {
      checkpoint_configuration {
        configuration_type = "DEFAULT"
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "DEBUG"
        metrics_level      = "TASK"
      }

      parallelism_configuration {
        auto_scaling_enabled = true
        configuration_type   = "CUSTOM"
        parallelism          = 10
        parallelism_per_kpu  = 4
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfigurationUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_bucket_object.test[1].key
          object_version = aws_s3_bucket_object.test[1].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = true
    }

    flink_application_configuration {
      checkpoint_configuration {
        checkpointing_enabled         = true
        configuration_type            = "CUSTOM"
        checkpoint_interval           = 55000
        min_pause_between_checkpoints = 5500
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "ERROR"
        metrics_level      = "OPERATOR"
      }

      parallelism_configuration {
        configuration_type = "DEFAULT"
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfigurationEnvironmentProperties(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_bucket_object.test[0].key
          object_version = aws_s3_bucket_object.test[0].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = false
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID1"

        property_map = {
          Key9 = "Value1"
          Key8 = "Value2"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyB = "ValueY"
          KeyC = "ValueX"
        }
      }
    }

    flink_application_configuration {
      checkpoint_configuration {
        configuration_type = "DEFAULT"
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "DEBUG"
        metrics_level      = "TASK"
      }

      parallelism_configuration {
        auto_scaling_enabled = true
        configuration_type   = "CUSTOM"
        parallelism          = 10
        parallelism_per_kpu  = 4
      }
    }
  }

  tags = {
    Key1 = "Value1"
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigFlinkApplicationConfigurationEnvironmentPropertiesUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[1].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn     = aws_s3_bucket.test.arn
          file_key       = aws_s3_bucket_object.test[1].key
          object_version = aws_s3_bucket_object.test[1].version_id
        }
      }

      code_content_type = "ZIPFILE"
    }

    application_snapshot_configuration {
      snapshots_enabled = true
    }

    environment_properties {
      property_group {
        property_group_id = "PROPERTY-GROUP-ID3"

        property_map = {
          Key = "Value"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID2"

        property_map = {
          KeyA = "ValueZ"
          KeyC = "ValueW"
        }
      }

      property_group {
        property_group_id = "PROPERTY-GROUP-ID4"

        property_map = {
          KeyAlpha = "ValueOmega"
        }
      }
    }

    flink_application_configuration {
      checkpoint_configuration {
        checkpointing_enabled         = true
        configuration_type            = "CUSTOM"
        checkpoint_interval           = 55000
        min_pause_between_checkpoints = 5500
      }

      monitoring_configuration {
        configuration_type = "CUSTOM"
        log_level          = "ERROR"
        metrics_level      = "OPERATOR"
      }

      parallelism_configuration {
        configuration_type = "DEFAULT"
      }
    }
  }

  tags = {
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationNotSpecified(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_1"

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        kinesis_firehose_input {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInputUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_2"

        input_parallelism {
          count = 42
        }

        input_schema {
          record_column {
            name     = "COLUMN_2"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-2"
          }

          record_column {
            name     = "COLUMN_3"
            sql_type = "DOUBLE"
            mapping  = "MAPPING-3"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$path.to.record"
              }
            }
          }
        }

        kinesis_streams_input {
          resource_arn = aws_kinesis_stream.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationInputProcessingConfiguration(rName string, lambdaIndex int) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_1"

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        input_processing_configuration {
          input_lambda_processor {
            resource_arn = aws_lambda_function.test.%[2]d.arn
          }
        }

        kinesis_firehose_input {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }
}
`, rName, lambdaIndex))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationMultiple(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_1"

        input_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        input_processing_configuration {
          input_lambda_processor {
            resource_arn = aws_lambda_function.test[0].arn
          }
        }

        kinesis_firehose_input {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }

      output {
        name = "OUTPUT_1"

        destination_schema {
          record_format_type = "CSV"
        }

        kinesis_firehose_output {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.arn
  }

  tags = {
    Key1 = "Value1"
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationMultipleUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[1].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 2;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      input {
        name_prefix = "NAME_PREFIX_2"

        input_parallelism {
          count = 42
        }

        input_schema {
          record_column {
            name     = "COLUMN_2"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-2"
          }

          record_column {
            name     = "COLUMN_3"
            sql_type = "DOUBLE"
            mapping  = "MAPPING-3"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$path.to.record"
              }
            }
          }
        }

        kinesis_streams_input {
          resource_arn = aws_kinesis_stream.test.arn
        }
      }

      output {
        name = "OUTPUT_2"

        destination_schema {
          record_format_type = "JSON"
        }

        kinesis_streams_output {
          resource_arn = aws_kinesis_stream.test.arn
        }
      }

      output {
        name = "OUTPUT_3"

        destination_schema {
          record_format_type = "CSV"
        }

        lambda_output {
          resource_arn = aws_lambda_function.test[0].arn
        }
      }

      reference_data_source {
        table_name = "TABLE-1"

        reference_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = "KEY-1"
        }
      }
    }
  }

  tags = {
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationOutput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "OUTPUT_1"

        destination_schema {
          record_format_type = "CSV"
        }

        kinesis_firehose_output {
          resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationOutputUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      output {
        name = "OUTPUT_2"

        destination_schema {
          record_format_type = "JSON"
        }

        kinesis_streams_output {
          resource_arn = aws_kinesis_stream.test.arn
        }
      }

      output {
        name = "OUTPUT_3"

        destination_schema {
          record_format_type = "CSV"
        }

        lambda_output {
          resource_arn = aws_lambda_function.test[0].arn
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationReferenceDataSource(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      reference_data_source {
        table_name = "TABLE-1"

        reference_schema {
          record_column {
            name     = "COLUMN_1"
            sql_type = "INTEGER"
          }

          record_format {
            record_format_type = "CSV"

            mapping_parameters {
              csv_mapping_parameters {
                record_column_delimiter = ","
                record_row_delimiter    = "|"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = "KEY-1"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigSQLApplicationConfigurationReferenceDataSourceUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseSQLApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        text_content = "SELECT 1;\n"
      }

      code_content_type = "PLAINTEXT"
    }

    sql_application_configuration {
      reference_data_source {
        table_name = "TABLE-2"

        reference_schema {
          record_column {
            name     = "COLUMN_2"
            sql_type = "VARCHAR(8)"
            mapping  = "MAPPING-2"
          }

          record_column {
            name     = "COLUMN_3"
            sql_type = "DOUBLE"
            mapping  = "MAPPING-3"
          }

          record_encoding = "UTF-8"

          record_format {
            record_format_type = "JSON"

            mapping_parameters {
              json_mapping_parameters {
                record_row_path = "$path.to.record"
              }
            }
          }
        }

        s3_reference_data_source {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = "KEY-2"
        }
      }
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccKinesisAnalyticsV2ApplicationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "SQL-1_0"
  service_execution_role = aws_iam_role.test[0].arn

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccKinesisAnalyticsV2ApplicationConfigVPCConfiguration(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseVPC(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    vpc_configuration {
      security_group_ids = aws_security_group.test.*.id
      subnet_ids         = aws_subnet.test.*.id
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigVPCConfigurationUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseVPC(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }

    vpc_configuration {
      security_group_ids = [aws_security_group.test[0].id]
      subnet_ids         = [aws_subnet.test[0].id]
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsV2ApplicationConfigVPCConfigurationNotSpecified(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsV2ApplicationConfigBaseServiceExecutionIamRole(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseVPC(rName),
		testAccKinesisAnalyticsV2ApplicationConfigBaseFlinkApplication(rName),
		fmt.Sprintf(`
resource "aws_kinesisanalyticsv2_application" "test" {
  name                   = %[1]q
  runtime_environment    = "FLINK-1_8"
  service_execution_role = aws_iam_role.test[0].arn

  application_configuration {
    application_code_configuration {
      code_content {
        s3_content_location {
          bucket_arn = aws_s3_bucket.test.arn
          file_key   = aws_s3_bucket_object.test[0].key
        }
      }

      code_content_type = "ZIPFILE"
    }
  }
}
`, rName))
}
