package aws

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/lambda/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSLambdaEventSourceMapping_Kinesis_basic(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	functionResourceName := "aws_lambda_function.test"
	functionResourceNameUpdated := "aws_lambda_function.test_update"
	eventSourceResourceName := "aws_kinesis_stream.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisBatchSize(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", "0"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", "0"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccAWSLambdaEventSourceMappingConfigKinesisBatchSize(rName, "null"),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisUpdateFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "200"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceNameUpdated, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceNameUpdated, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_SQS_basic(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	functionResourceName := "aws_lambda_function.test"
	functionResourceNameUpdated := "aws_lambda_function.test_update"
	eventSourceResourceName := "aws_sqs_queue.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsBatchSize(rName, "10"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "10"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceName, "arn"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccAWSLambdaEventSourceMappingConfigSqsBatchSize(rName, "null"),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsUpdateFunctionName(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "5"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceNameUpdated, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceNameUpdated, "arn"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_DynamoDB_basic(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	functionResourceName := "aws_lambda_function.test"
	eventSourceResourceName := "aws_dynamodb_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigDynamoDbBatchSize(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "stream_arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", "0"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", "0"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccAWSLambdaEventSourceMappingConfigDynamoDbBatchSize(rName, "null"),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_DynamoDB_FunctionResponseTypes(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigDynamoDbFunctionResponseTypes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "function_response_types.*", "ReportBatchItemFailures"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigDynamoDbNoFunctionResponseTypes(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "function_response_types.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_SQS_BatchWindow(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	batchWindow := int64(0)
	batchWindowUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsBatchWindow(rName, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindow))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsBatchWindow(rName, batchWindowUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindowUpdate))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_disappears(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsBatchSize(rName, "7"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLambdaEventSourceMapping(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_SQS_changesInEnabledAreDetected(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsBatchSize(rName, "9"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					testAccCheckAWSLambdaEventSourceMappingIsBeingDisabled(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_StartingPositionTimestamp(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	startingPositionTimestamp := time.Now().UTC().Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisStartingPositionTimestamp(rName, startingPositionTimestamp),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "starting_position", "AT_TIMESTAMP"),
					resource.TestCheckResourceAttr(resourceName, "starting_position_timestamp", startingPositionTimestamp),
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

func TestAccAWSLambdaEventSourceMapping_Kinesis_BatchWindow(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	batchWindow := int64(5)
	batchWindowUpdate := int64(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisBatchWindow(rName, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindow))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisBatchWindow(rName, batchWindowUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindowUpdate))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_ParallelizationFactor(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	parallelizationFactor := int64(1)
	parallelizationFactorUpdate := int64(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisParallelizationFactor(rName, parallelizationFactor),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "parallelization_factor", strconv.Itoa(int(parallelizationFactor))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisParallelizationFactor(rName, parallelizationFactorUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "parallelization_factor", strconv.Itoa(int(parallelizationFactorUpdate))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_TumblingWindowInSeconds(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	tumblingWindowInSeconds := int64(30)
	tumblingWindowInSecondsUpdate := int64(300)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisTumblingWindowInSeconds(rName, tumblingWindowInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", strconv.Itoa(int(tumblingWindowInSeconds))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisTumblingWindowInSeconds(rName, tumblingWindowInSecondsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "tumbling_window_in_seconds", strconv.Itoa(int(tumblingWindowInSecondsUpdate))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_MaximumRetryAttempts(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(10000)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttemptsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttemptsUpdate))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_MaximumRetryAttemptsZero(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(0)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttemptsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttemptsUpdate))),
				),
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_MaximumRetryAttemptsNegativeOne(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(-1)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttemptsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttemptsUpdate))),
				),
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName, maximumRetryAttempts),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_retry_attempts", strconv.Itoa(int(maximumRetryAttempts))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_MaximumRecordAgeInSeconds(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRecordAgeInSeconds := int64(604800)
	maximumRecordAgeInSecondsUpdate := int64(3600)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSeconds))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSecondsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSecondsUpdate))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_MaximumRecordAgeInSecondsNegativeOne(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRecordAgeInSeconds := int64(-1)
	maximumRecordAgeInSecondsUpdate := int64(3600)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSeconds),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSeconds))),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRecordAgeInSeconds(rName, maximumRecordAgeInSecondsUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_record_age_in_seconds", strconv.Itoa(int(maximumRecordAgeInSecondsUpdate))),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_BisectBatch(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	bisectBatch := false
	bisectBatchUpdate := true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisBisectBatch(rName, bisectBatch),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "bisect_batch_on_function_error", strconv.FormatBool(bisectBatch)),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisBisectBatch(rName, bisectBatchUpdate),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "bisect_batch_on_function_error", strconv.FormatBool(bisectBatchUpdate)),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_Kinesis_DestinationConfig(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	snsResourceName := "aws_sns_topic.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisDestinationConfig(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination_arn", snsResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisDestinationConfig(rName, rName+"-update"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination_arn", snsResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_MSK(t *testing.T) {
	var v lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	eventSourceResourceName := "aws_msk_cluster.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSMsk(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID, "kafka"),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigMsk(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccAWSLambdaEventSourceMappingConfigMsk(rName, "null"),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigMsk(rName, "9999"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "9999"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_SelfManagedKafka(t *testing.T) {
	var v lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSelfManagedKafka(rName, "100", "test1:9092,test2:9092"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "false"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_event_source.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "self_managed_event_source.0.endpoints.KAFKA_BOOTSTRAP_SERVERS", "test1:9092,test2:9092"),
					resource.TestCheckResourceAttr(resourceName, "source_access_configuration.#", "3"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			// Verify also that bootstrap broker order does not matter.
			{
				PlanOnly: true,
				Config:   testAccAWSLambdaEventSourceMappingConfigSelfManagedKafka(rName, "null", "test2:9092,test1:9092"),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_ActiveMQ(t *testing.T) {
	var v lambda.EventSourceMappingConfiguration
	resourceName := "aws_lambda_event_source_mapping.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckAWSSecretsManager(t)
			testAccPartitionHasServicePreCheck("mq", t)
			testAccPreCheckAWSMq(t)
		},
		ErrorCheck:   testAccErrorCheck(t, lambda.EndpointsID, "mq", "secretsmanager"),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigActiveMQ(rName, "100"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "batch_size", "100"),
					resource.TestCheckResourceAttr(resourceName, "enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "queues.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "queues.*", "test"),
					resource.TestCheckResourceAttr(resourceName, "source_access_configuration.#", "1"),
				),
			},
			// batch_size became optional.  Ensure that if the user supplies the default
			// value, but then moves to not providing the value, that we don't consider this
			// a diff.
			{
				PlanOnly: true,
				Config:   testAccAWSLambdaEventSourceMappingConfigActiveMQ(rName, "null"),
			},
		},
	})
}

func testAccCheckAWSLambdaEventSourceMappingIsBeingDisabled(conf *lambda.EventSourceMappingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lambdaconn
		// Disable enabled state
		err := resource.Retry(10*time.Minute, func() *resource.RetryError {
			params := &lambda.UpdateEventSourceMappingInput{
				UUID:    conf.UUID,
				Enabled: aws.Bool(false),
			}

			_, err := conn.UpdateEventSourceMapping(params)

			if err != nil {
				if isAWSErr(err, lambda.ErrCodeResourceInUseException, "") {
					return resource.RetryableError(fmt.Errorf(
						"Waiting for Lambda Event Source Mapping to be ready to be updated: %v", conf.UUID))
				}

				return resource.NonRetryableError(
					fmt.Errorf("Error updating Lambda Event Source Mapping: %w", err))
			}

			return nil
		})

		if err != nil {
			return err
		}

		// wait for state to be propagated
		return resource.Retry(10*time.Minute, func() *resource.RetryError {
			params := &lambda.GetEventSourceMappingInput{
				UUID: conf.UUID,
			}
			newConf, err := conn.GetEventSourceMapping(params)
			if err != nil {
				return resource.NonRetryableError(
					fmt.Errorf("Error getting Lambda Event Source Mapping: %s", err))
			}

			if *newConf.State != "Disabled" {
				return resource.RetryableError(fmt.Errorf(
					"Waiting to get Lambda Event Source Mapping to be fully enabled, it's currently %s: %v", *newConf.State, conf.UUID))

			}

			return nil
		})

	}
}

func testAccCheckLambdaEventSourceMappingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_event_source_mapping" {
			continue
		}

		_, err := finder.EventSourceMappingConfigurationByID(conn, rs.Primary.ID)

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return fmt.Errorf("error reading Lambda Event Source Mapping (%s): %w", rs.Primary.ID, err)
		}

		return fmt.Errorf("Lambda Event Source Mapping %s still exists", rs.Primary.ID)
	}

	return nil

}

func testAccCheckAwsLambdaEventSourceMappingExists(n string, v *lambda.EventSourceMappingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf(" Lambda Event Source Mapping resource not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no Lambda Event Source Mapping ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		eventSourceMappingConfiguration, err := finder.EventSourceMappingConfigurationByID(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *eventSourceMappingConfiguration

		return nil
	}
}

func testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:*"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sns:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 1
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}
`, rName)
}

func testAccAWSLambdaEventSourceMappingConfigSQSBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "sqs:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_sqs_queue" "test" {
  name = %[1]q
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}
`, rName)
}

func testAccAWSLambdaEventSourceMappingConfigDynamoDBBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "dynamodb:*"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_dynamodb_table" "test" {
  name           = %[1]q
  read_capacity  = 1
  write_capacity = 2
  hash_key       = "TestTableHashKey"

  attribute {
    name = "TestTableHashKey"
    type = "S"
  }

  stream_enabled   = true
  stream_view_type = "KEYS_ONLY"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}
`, rName)
}

func testAccAWSLambdaEventSourceMappingConfigKafkaBase(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_policy" "test" {
  name = %[1]q

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kafka:DescribeCluster",
        "kafka:GetBootstrapBrokers",
        "ec2:CreateNetworkInterface",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeSecurityGroups",
        "ec2:DescribeSubnets",
        "ec2:DescribeVpcs",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "test" {
  name       = %[1]q
  roles      = [aws_iam_role.test.name]
  policy_arn = aws_iam_policy.test.arn
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/22"

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
  name   = %[1]q
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}
`, rName))
}

func testAccAWSLambdaEventSourceMappingConfigMQBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": "sts:AssumeRole",
    "Principal": {
      "Service": "lambda.amazonaws.com"
    },
    "Effect": "Allow"
  }]
}
EOF
}

resource "aws_iam_role_policy" "test" {
  role = aws_iam_role.test.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "mq:DescribeBroker",
        "secretsmanager:GetSecretValue",
        "ec2:CreateNetworkInterface",
        "ec2:DescribeNetworkInterfaces",
        "ec2:DescribeVpcs",
        "ec2:DeleteNetworkInterface",
        "ec2:DescribeSubnets",
        "ec2:DescribeSecurityGroups",
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}

resource "aws_security_group" "test" {
  name = %[1]q

  ingress {
    from_port   = 61617
    to_port     = 61617
    protocol    = "tcp"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_mq_broker" "test" {
  broker_name             = %[1]q
  engine_type             = "ActiveMQ"
  engine_version          = "5.15.0"
  host_instance_type      = "mq.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }

  publicly_accessible = true
}

resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id     = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({ username = "Test", password = "TestTest1234" })
}
`, rName)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisStartingPositionTimestamp(rName, startingPositionTimestamp string) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                  = 100
  enabled                     = true
  event_source_arn            = aws_kinesis_stream.test.arn
  function_name               = aws_lambda_function.test.arn
  starting_position           = "AT_TIMESTAMP"
  starting_position_timestamp = %[1]q
}
`, startingPositionTimestamp))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisBatchWindow(rName string, batchWindow int64) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                         = 100
  maximum_batching_window_in_seconds = %[1]d
  enabled                            = true
  event_source_arn                   = aws_kinesis_stream.test.arn
  function_name                      = aws_lambda_function.test.arn
  starting_position                  = "TRIM_HORIZON"
}
`, batchWindow))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisParallelizationFactor(rName string, parallelizationFactor int64) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size             = 100
  parallelization_factor = %[1]d
  enabled                = true
  event_source_arn       = aws_kinesis_stream.test.arn
  function_name          = aws_lambda_function.test.arn
  starting_position      = "TRIM_HORIZON"
}
`, parallelizationFactor))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisTumblingWindowInSeconds(rName string, tumblingWindowInSeconds int64) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                 = 100
  tumbling_window_in_seconds = %[1]d
  enabled                    = true
  event_source_arn           = aws_kinesis_stream.test.arn
  function_name              = aws_lambda_function.test.arn
  starting_position          = "TRIM_HORIZON"
}
`, tumblingWindowInSeconds))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName string, maximumRetryAttempts int64) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size             = 100
  maximum_retry_attempts = %[1]d
  enabled                = true
  event_source_arn       = aws_kinesis_stream.test.arn
  function_name          = aws_lambda_function.test.arn
  starting_position      = "TRIM_HORIZON"
}
`, maximumRetryAttempts))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisBisectBatch(rName string, bisectBatch bool) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                     = 100
  bisect_batch_on_function_error = %[1]t
  enabled                        = true
  event_source_arn               = aws_kinesis_stream.test.arn
  function_name                  = aws_lambda_function.test.arn
  starting_position              = "TRIM_HORIZON"
}
`, bisectBatch))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRecordAgeInSeconds(rName string, maximumRecordAgeInSeconds int64) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                    = 100
  maximum_record_age_in_seconds = %[1]d
  enabled                       = true
  event_source_arn              = aws_kinesis_stream.test.arn
  function_name                 = aws_lambda_function.test.arn
  starting_position             = "TRIM_HORIZON"
}
`, maximumRecordAgeInSeconds))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisDestinationConfig(rName string, topicName string) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_sns_topic" "test" {
  name = %[1]q
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 100
  enabled           = true
  event_source_arn  = aws_kinesis_stream.test.arn
  function_name     = aws_lambda_function.test.arn
  starting_position = "TRIM_HORIZON"

  destination_config {
    on_failure {
      destination_arn = aws_sns_topic.test.arn
    }
  }
}
`, topicName))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisBatchSize(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[1]s
  enabled           = true
  event_source_arn  = aws_kinesis_stream.test.arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "TRIM_HORIZON"
}
`, batchSize))
}

func testAccAWSLambdaEventSourceMappingConfigKinesisUpdateFunctionName(rName string) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-update"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 200
  enabled           = false
  event_source_arn  = aws_kinesis_stream.test.arn
  function_name     = aws_lambda_function.test_update.arn
  starting_position = "TRIM_HORIZON"
}
`, rName))
}

func testAccAWSLambdaEventSourceMappingConfigSqsBatchSize(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return composeConfig(testAccAWSLambdaEventSourceMappingConfigSQSBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = %[1]s
  enabled          = true
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test.function_name
}
`, batchSize))
}

func testAccAWSLambdaEventSourceMappingConfigSqsUpdateFunctionName(rName string) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigSQSBase(rName), fmt.Sprintf(`
resource "aws_lambda_function" "test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s-update"
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = 5
  enabled          = false
  event_source_arn = aws_sqs_queue.test.arn
  function_name    = aws_lambda_function.test_update.arn
}
`, rName))
}

func testAccAWSLambdaEventSourceMappingConfigSqsBatchWindow(rName string, batchWindow int64) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigSQSBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                         = 10
  maximum_batching_window_in_seconds = %[1]d
  event_source_arn                   = aws_sqs_queue.test.arn
  enabled                            = false
  function_name                      = aws_lambda_function.test.arn
}
`, batchWindow))
}

func testAccAWSLambdaEventSourceMappingConfigMsk(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKafkaBase(rName), fmt.Sprintf(`
resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.7.1"
  number_of_broker_nodes = 2

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    ebs_volume_size = 10
    instance_type   = "kafka.m5.large"
    security_groups = [aws_security_group.test.id]
  }
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[2]s
  event_source_arn  = aws_msk_cluster.test.arn
  enabled           = true
  function_name     = aws_lambda_function.test.arn
  topics            = ["test"]
  starting_position = "TRIM_HORIZON"

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, batchSize))
}

func testAccAWSLambdaEventSourceMappingConfigSelfManagedKafka(rName, batchSize, kafkaBootstrapServers string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return composeConfig(testAccAWSLambdaEventSourceMappingConfigKafkaBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[2]s
  enabled           = false
  function_name     = aws_lambda_function.test.arn
  topics            = ["test"]
  starting_position = "TRIM_HORIZON"

  self_managed_event_source {
    endpoints = {
      KAFKA_BOOTSTRAP_SERVERS = %[3]q
    }
  }

  dynamic "source_access_configuration" {
    for_each = aws_subnet.test.*.id
    content {
      type = "VPC_SUBNET"
      uri  = "subnet:${source_access_configuration.value}"
    }
  }

  source_access_configuration {
    type = "VPC_SECURITY_GROUP"
    uri  = aws_security_group.test.id
  }
}
`, rName, batchSize, kafkaBootstrapServers))
}

func testAccAWSLambdaEventSourceMappingConfigDynamoDbBatchSize(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return composeConfig(testAccAWSLambdaEventSourceMappingConfigDynamoDBBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = %[1]s
  enabled           = true
  event_source_arn  = aws_dynamodb_table.test.stream_arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "LATEST"
}
`, batchSize))
}

func testAccAWSLambdaEventSourceMappingConfigDynamoDbFunctionResponseTypes(rName string) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigDynamoDBBase(rName), `
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 150
  enabled           = true
  event_source_arn  = aws_dynamodb_table.test.stream_arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "LATEST"

  function_response_types = ["ReportBatchItemFailures"]
}
`)
}

func testAccAWSLambdaEventSourceMappingConfigDynamoDbNoFunctionResponseTypes(rName string) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigDynamoDBBase(rName), `
resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 150
  enabled           = true
  event_source_arn  = aws_dynamodb_table.test.stream_arn
  function_name     = aws_lambda_function.test.function_name
  starting_position = "LATEST"
}
`)
}

func testAccAWSLambdaEventSourceMappingConfigActiveMQ(rName, batchSize string) string {
	if batchSize == "" {
		batchSize = "null"
	}

	return composeConfig(testAccAWSLambdaEventSourceMappingConfigMQBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = %[1]s
  event_source_arn = aws_mq_broker.test.arn
  enabled          = true
  function_name    = aws_lambda_function.test.arn
  queues           = ["test"]

  source_access_configuration {
    type = "BASIC_AUTH"
    uri  = aws_secretsmanager_secret_version.test.arn
  }
}
`, batchSize))
}
