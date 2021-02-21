package aws

import (
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/kinesisanalytics"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalytics/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/kinesisanalytics/lister"
)

func init() {
	resource.AddTestSweepers("aws_kinesis_analytics_application", &resource.Sweeper{
		Name: "aws_kinesis_analytics_application",
		F:    testSweepKinesisAnalyticsApplications,
	})
}

func testSweepKinesisAnalyticsApplications(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).kinesisanalyticsconn
	input := &kinesisanalytics.ListApplicationsInput{}
	var sweeperErrs *multierror.Error

	err = lister.ListApplicationsPages(conn, input, func(page *kinesisanalytics.ListApplicationsOutput, isLast bool) bool {
		if page == nil {
			return !isLast
		}

		for _, applicationSummary := range page.ApplicationSummaries {
			arn := aws.StringValue(applicationSummary.ApplicationARN)
			name := aws.StringValue(applicationSummary.ApplicationName)

			application, err := finder.ApplicationByName(conn, name)

			if tfawserr.ErrMessageContains(err, kinesisanalytics.ErrCodeUnsupportedOperationException, "was created/updated by kinesisanalyticsv2 SDK") {
				continue
			}

			if err != nil {
				sweeperErr := fmt.Errorf("error reading Kinesis Analytics Application (%s): %w", arn, err)
				log.Printf("[ERROR] %s", err)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}

			r := resourceAwsKinesisAnalyticsApplication()
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
		log.Printf("[WARN] Skipping Kinesis Analytics Application sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil() // In case we have completed some pages, but had errors
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing Kinesis Analytics Applications: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSKinesisAnalyticsApplication_basic(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
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

func TestAccAWSKinesisAnalyticsApplication_disappears(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisAnalyticsApplication(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_Tags(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_Code_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigCode(rName, "SELECT 1;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", "SELECT 1;\n"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigCode(rName, "SELECT 2;\n"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", "SELECT 2;\n"),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", "test"),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_CloudWatchLoggingOptions_Add(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_CloudWatchLoggingOptions_Delete(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test.0"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_CloudWatchLoggingOptions_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	cloudWatchLogStream1ResourceName := "aws_cloudwatch_log_stream.test.0"
	cloudWatchLogStream2ResourceName := "aws_cloudwatch_log_stream.test.1"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigCloudWatchLoggingOptions(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigCloudWatchLoggingOptions(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStream2ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_Input_Add(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_Input_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigInputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_InputProcessingConfiguration_Add(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "3"), // Add input processing configuration + update input.
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

func TestAccAWSKinesisAnalyticsApplication_InputProcessingConfiguration_Delete(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigInput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "3"), // Delete input processing configuration + update input.
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

func TestAccAWSKinesisAnalyticsApplication_InputProcessingConfiguration_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	lambda1ResourceName := "aws_lambda_function.test.0"
	lambda2ResourceName := "aws_lambda_function.test.1"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigInputProcessingConfiguration(rName, 0),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambda1ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigInputProcessingConfiguration(rName, 1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambda2ResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_Multiple_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	cloudWatchLogStreamResourceName := "aws_cloudwatch_log_stream.test"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.log_stream_arn", cloudWatchLogStreamResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_logging_options.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.0.lambda.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.processing_configuration.0.lambda.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_firehose.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"name":                        "OUTPUT_1",
						"schema.#":                    "1",
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          "1",
						"kinesis_stream.#":            "0",
						"lambda.#":                    "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigMultipleUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.stream_names.#", "42"),
					resource.TestCheckResourceAttrSet(resourceName, "inputs.0.id"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_columns.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.mapping_parameters.0.json.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.name_prefix", "NAME_PREFIX_2"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.parallelism.0.count", "42"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.processing_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_firehose.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "inputs.0.kinesis_stream.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "inputs.0.kinesis_stream.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"name":                        "OUTPUT_2",
						"schema.#":                    "1",
						"schema.0.record_format_type": "JSON",
						"kinesis_firehose.#":          "0",
						"kinesis_stream.#":            "1",
						"lambda.#":                    "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"name":                        "OUTPUT_3",
						"schema.#":                    "1",
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          "0",
						"kinesis_stream.#":            "0",
						"lambda.#":                    "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "version", "8"), // Delete CloudWatch logging options + add reference data source + delete input processing configuration+ update application + delete output + 2 * add output.
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

func TestAccAWSKinesisAnalyticsApplication_Output_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	lambdaResourceName := "aws_lambda_function.test.0"
	firehoseResourceName := "aws_kinesis_firehose_delivery_stream.test"
	streamsResourceName := "aws_kinesis_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationOutput(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"name":                        "OUTPUT_1",
						"schema.#":                    "1",
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          "1",
						"kinesis_stream.#":            "0",
						"lambda.#":                    "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.resource_arn", firehoseResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_firehose.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationOutputUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"name":                        "OUTPUT_2",
						"schema.#":                    "1",
						"schema.0.record_format_type": "JSON",
						"kinesis_firehose.#":          "0",
						"kinesis_stream.#":            "1",
						"lambda.#":                    "0",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.resource_arn", streamsResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.kinesis_stream.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "outputs.*", map[string]string{
						"name":                        "OUTPUT_3",
						"schema.#":                    "1",
						"schema.0.record_format_type": "CSV",
						"kinesis_firehose.#":          "0",
						"kinesis_stream.#":            "0",
						"lambda.#":                    "1",
					}),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.resource_arn", lambdaResourceName, "arn"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "outputs.*.lambda.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "4"), // 1 * output deletion + 2 * output addition.
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "6"), // 2 * output deletion.
				),
			},
		},
	})
}

func TestAccAWSKinesisAnalyticsApplication_ReferenceDataSource_Add(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "1"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
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

func TestAccAWSKinesisAnalyticsApplication_ReferenceDataSource_Delete(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRoleResourceName := "aws_iam_role.test.0"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRoleResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "3"),
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

func TestAccAWSKinesisAnalyticsApplication_ReferenceDataSource_Update(t *testing.T) {
	var v kinesisanalytics.ApplicationDetail
	resourceName := "aws_kinesis_analytics_application.test"
	iamRole1ResourceName := "aws_iam_role.test.0"
	iamRole2ResourceName := "aws_iam_role.test.1"
	s3BucketResourceName := "aws_s3_bucket.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSKinesisAnalytics(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisAnalyticsApplicationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisAnalyticsApplicationReferenceDataSource(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "INTEGER"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", ""),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_column_delimiter", ","),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.0.record_row_delimiter", "|"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "CSV"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRole1ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-1"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "2"),
				),
			},
			{
				Config: testAccKinesisAnalyticsApplicationReferenceDataSourceUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisAnalyticsApplicationExists(resourceName, &v),
					testAccCheckResourceAttrRegionalARN(resourceName, "arn", "kinesisanalytics", fmt.Sprintf("application/%s", rName)),
					resource.TestCheckResourceAttr(resourceName, "cloudwatch_logging_options.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "code", ""),
					resource.TestCheckResourceAttrSet(resourceName, "create_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "description", ""),
					resource.TestCheckResourceAttrSet(resourceName, "last_update_timestamp"),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "inputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "outputs.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.mapping", "MAPPING-2"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.name", "COLUMN_2"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.0.sql_type", "VARCHAR(8)"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.1.mapping", "MAPPING-3"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.1.name", "COLUMN_3"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_columns.1.sql_type", "DOUBLE"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_encoding", "UTF-8"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.csv.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.mapping_parameters.0.json.0.record_row_path", "$path.to.record"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.schema.0.record_format.0.record_format_type", "JSON"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.bucket_arn", s3BucketResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.s3.0.file_key", "KEY-2"),
					resource.TestCheckResourceAttrPair(resourceName, "reference_data_sources.0.s3.0.role_arn", iamRole2ResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "reference_data_sources.0.table_name", "TABLE-2"),
					resource.TestCheckResourceAttrSet(resourceName, "reference_data_sources.0.id"),
					resource.TestCheckResourceAttr(resourceName, "status", "READY"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(resourceName, "version", "3"),
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

func testAccCheckKinesisAnalyticsApplicationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_analytics_application" {
			continue
		}

		_, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if isAWSErr(err, kinesisanalytics.ErrCodeResourceNotFoundException, "") {
			continue
		}
		if err != nil {
			return err
		}

		return fmt.Errorf("Kinesis Analytics Application %s still exists", rs.Primary.ID)
	}
	return nil
}

func testAccCheckKinesisAnalyticsApplicationExists(n string, v *kinesisanalytics.ApplicationDetail) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Analytics Application ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn

		application, err := finder.ApplicationByName(conn, rs.Primary.Attributes["name"])
		if err != nil {
			return err
		}

		*v = *application

		return nil
	}
}

func testAccPreCheckAWSKinesisAnalytics(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).kinesisanalyticsconn

	input := &kinesisanalytics.ListApplicationsInput{}

	_, err := conn.ListApplications(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName string) string {
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
      "Action": ["logs:*"],
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

func testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName string) string {
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

func testAccKinesisAnalyticsApplicationConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q
}
`, rName)
}

func testAccKinesisAnalyticsApplicationConfigCode(rName, code string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name        = %[1]q
  description = "test"
  code        = %[2]q
}
`, rName, code)
}

func testAccKinesisAnalyticsApplicationConfigCloudWatchLoggingOptions(rName string, streamIndex int) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  count = 2

  name           = "%[1]s.${count.index}"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.%[2]d.arn
    role_arn       = aws_iam_role.test.%[2]d.arn
  }
}
`, rName, streamIndex))
}

func testAccKinesisAnalyticsApplicationConfigInput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationConfigInputUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_2"

    parallelism {
      count = 42
    }

    schema {
      record_columns {
        name     = "COLUMN_2"
        sql_type = "VARCHAR(8)"
        mapping  = "MAPPING-2"
      }

      record_columns {
        name     = "COLUMN_3"
        sql_type = "DOUBLE"
        mapping  = "MAPPING-3"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$path.to.record"
          }
        }
      }
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationConfigInputProcessingConfiguration(rName string, lambdaIndex int) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    processing_configuration {
      lambda {
        resource_arn = aws_lambda_function.test.%[2]d.arn
        role_arn     = aws_iam_role.test.%[2]d.arn
      }
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName, lambdaIndex))
}

func testAccKinesisAnalyticsApplicationConfigMultiple(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = %[1]q
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  cloudwatch_logging_options {
    log_stream_arn = aws_cloudwatch_log_stream.test.arn
    role_arn       = aws_iam_role.test[1].arn
  }

  inputs {
    name_prefix = "NAME_PREFIX_1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    processing_configuration {
      lambda {
        resource_arn = aws_lambda_function.test[0].arn
        role_arn     = aws_iam_role.test[0].arn
      }
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }

  outputs {
    name = "OUTPUT_1"

    schema {
      record_format_type = "CSV"
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }

  tags = {
    Key1 = "Value1"
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationConfigMultipleUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  inputs {
    name_prefix = "NAME_PREFIX_2"

    parallelism {
      count = 42
    }

    schema {
      record_columns {
        name     = "COLUMN_2"
        sql_type = "VARCHAR(8)"
        mapping  = "MAPPING-2"
      }

      record_columns {
        name     = "COLUMN_3"
        sql_type = "DOUBLE"
        mapping  = "MAPPING-3"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$path.to.record"
          }
        }
      }
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }

  outputs {
    name = "OUTPUT_2"

    schema {
      record_format_type = "JSON"
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }

  outputs {
    name = "OUTPUT_3"

    schema {
      record_format_type = "CSV"
    }

    lambda {
      resource_arn = aws_lambda_function.test[0].arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }

  reference_data_sources {
    table_name = "TABLE-1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "KEY-1"
      role_arn   = aws_iam_role.test[1].arn
    }
  }

  tags = {
    Key2 = "Value2"
    Key3 = "Value3"
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationOutput(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  outputs {
    name = "OUTPUT_1"

    schema {
      record_format_type = "CSV"
    }

    kinesis_firehose {
      resource_arn = aws_kinesis_firehose_delivery_stream.test.arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationOutputUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		testAccKinesisAnalyticsApplicationConfigBaseInputOutput(rName),
		fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  outputs {
    name = "OUTPUT_2"

    schema {
      record_format_type = "JSON"
    }

    kinesis_stream {
      resource_arn = aws_kinesis_stream.test.arn
      role_arn     = aws_iam_role.test[1].arn
    }
  }

  outputs {
    name = "OUTPUT_3"

    schema {
      record_format_type = "CSV"
    }

    lambda {
      resource_arn = aws_lambda_function.test[0].arn
      role_arn     = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationReferenceDataSource(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  reference_data_sources {
    table_name = "TABLE-1"

    schema {
      record_columns {
        name     = "COLUMN_1"
        sql_type = "INTEGER"
      }

      record_format {
        mapping_parameters {
          csv {
            record_column_delimiter = ","
            record_row_delimiter    = "|"
          }
        }
      }
    }

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "KEY-1"
      role_arn   = aws_iam_role.test[0].arn
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationReferenceDataSourceUpdated(rName string) string {
	return composeConfig(
		testAccKinesisAnalyticsApplicationConfigBaseIamRole(rName),
		fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  reference_data_sources {
    table_name = "TABLE-2"

    schema {
      record_columns {
        name     = "COLUMN_2"
        sql_type = "VARCHAR(8)"
        mapping  = "MAPPING-2"
      }

      record_columns {
        name     = "COLUMN_3"
        sql_type = "DOUBLE"
        mapping  = "MAPPING-3"
      }

      record_encoding = "UTF-8"

      record_format {
        mapping_parameters {
          json {
            record_row_path = "$path.to.record"
          }
        }
      }
    }

    s3 {
      bucket_arn = aws_s3_bucket.test.arn
      file_key   = "KEY-2"
      role_arn   = aws_iam_role.test[1].arn
    }
  }
}
`, rName))
}

func testAccKinesisAnalyticsApplicationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccKinesisAnalyticsApplicationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_kinesis_analytics_application" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
