package aws

import (
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSLambdaEventSourceMapping_kinesis_basic(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration

	resourceName := "aws_lambda_event_source_mapping.lambda_event_source_mapping_test"
	functionResourceName := "aws_lambda_function.lambda_function_test_create"
	functionResourceNameUpdated := "aws_lambda_function.lambda_function_test_update"
	eventSourceResourceName := "aws_kinesis_stream.kinesis_stream_test"

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_esm_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_esm_basic_%s", rString)
	attName := fmt.Sprintf("tf_acc_att_lambda_esm_basic_%s", rString)
	streamName := fmt.Sprintf("tf_acc_stream_lambda_esm_basic_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_esm_basic_%s", rString)
	uFuncName := fmt.Sprintf("tf_acc_lambda_esm_basic_updated_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfig_kinesis(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					testAccCheckAWSLambdaEventSourceMappingAttributes(&conf),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "starting_position"},
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigUpdate_kinesis(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", strconv.Itoa(200)),
					resource.TestCheckResourceAttr(resourceName, "enabled", strconv.FormatBool(false)),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceNameUpdated, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceNameUpdated, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					resource.TestCheckResourceAttr(resourceName, "starting_position", "TRIM_HORIZON"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_kinesis_removeBatchSize(t *testing.T) {
	// batch_size became optional.  Ensure that if the user supplies the default
	// value, but then moves to not providing the value, that we don't consider this
	// a diff.

	var conf lambda.EventSourceMappingConfiguration

	resourceName := "aws_lambda_event_source_mapping.lambda_event_source_mapping_test"

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_esm_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_esm_basic_%s", rString)
	attName := fmt.Sprintf("tf_acc_att_lambda_esm_basic_%s", rString)
	streamName := fmt.Sprintf("tf_acc_stream_lambda_esm_basic_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_esm_basic_%s", rString)
	uFuncName := fmt.Sprintf("tf_acc_lambda_esm_basic_updated_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfig_kinesis(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					testAccCheckAWSLambdaEventSourceMappingAttributes(&conf),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "starting_position"},
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigUpdate_kinesis_removeBatchSize(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", strconv.Itoa(100)),
					resource.TestCheckResourceAttr(resourceName, "enabled", strconv.FormatBool(true)),
					resource.TestCheckResourceAttr(resourceName, "starting_position", "TRIM_HORIZON"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_sqs_basic(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration

	resourceName := "aws_lambda_event_source_mapping.lambda_event_source_mapping_test"
	functionResourceName := "aws_lambda_function.lambda_function_test_create"
	functionResourceNameUpdated := "aws_lambda_function.lambda_function_test_update"
	eventSourceResourceName := "aws_sqs_queue.sqs_queue_test"

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_sqs_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_sqs_basic_%s", rString)
	attName := fmt.Sprintf("tf_acc_att_lambda_sqs_basic_%s", rString)
	streamName := fmt.Sprintf("tf_acc_stream_lambda_sqs_basic_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_sqs_basic_%s", rString)
	uFuncName := fmt.Sprintf("tf_acc_lambda_sqs_basic_updated_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfig_sqs(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					testAccCheckAWSLambdaEventSourceMappingAttributes(&conf),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckNoResourceAttr(resourceName, "starting_position"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "starting_position"},
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigUpdate_sqs(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "batch_size", strconv.Itoa(5)),
					resource.TestCheckResourceAttr(resourceName, "enabled", strconv.FormatBool(false)),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceNameUpdated, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceNameUpdated, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "starting_position"),
				),
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_sqs_withFunctionName(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration

	resourceName := "aws_lambda_event_source_mapping.lambda_event_source_mapping_test"
	functionResourceName := "aws_lambda_function.lambda_function_test_create"

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_sqs_basic_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_sqs_basic_%s", rString)
	attName := fmt.Sprintf("tf_acc_att_lambda_sqs_basic_%s", rString)
	streamName := fmt.Sprintf("tf_acc_stream_lambda_sqs_basic_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_sqs_basic_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfig_sqs_testWithFunctionName(roleName, policyName, attName, streamName, funcName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					testAccCheckAWSLambdaEventSourceMappingAttributes(&conf),
					resource.TestCheckResourceAttrPair(resourceName, "function_name", functionResourceName, "arn"),
					resource.TestCheckResourceAttrPair(resourceName, "function_arn", functionResourceName, "arn"),
					resource.TestCheckNoResourceAttr(resourceName, "starting_position"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "starting_position"},
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_SQSBatchWindow(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration

	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	batchWindow := int64(0)
	batchWindowUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsWithBatchWindow(rName, batchWindow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "maximum_batching_window_in_seconds", strconv.Itoa(int(batchWindow))),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "starting_position"},
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigSqsWithBatchWindow(rName, batchWindowUpdate),
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

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_sqs_import_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_sqs_import_%s", rString)
	attName := fmt.Sprintf("tf_acc_att_lambda_sqs_import_%s", rString)
	streamName := fmt.Sprintf("tf_acc_stream_lambda_sqs_import_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_sqs_import_%s", rString)
	uFuncName := fmt.Sprintf("tf_acc_lambda_sqs_import_updated_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfig_sqs(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists("aws_lambda_event_source_mapping.lambda_event_source_mapping_test", &conf),
					testAccCheckAWSLambdaEventSourceMappingDisappears(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_changesInEnabledAreDetected(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration

	rString := acctest.RandString(8)
	roleName := fmt.Sprintf("tf_acc_role_lambda_sqs_import_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_lambda_sqs_import_%s", rString)
	attName := fmt.Sprintf("tf_acc_att_lambda_sqs_import_%s", rString)
	streamName := fmt.Sprintf("tf_acc_stream_lambda_sqs_import_%s", rString)
	funcName := fmt.Sprintf("tf_acc_lambda_sqs_import_%s", rString)
	uFuncName := fmt.Sprintf("tf_acc_lambda_sqs_import_updated_%s", rString)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfig_sqs(roleName, policyName, attName, streamName, funcName, uFuncName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists("aws_lambda_event_source_mapping.lambda_event_source_mapping_test", &conf),
					testAccCheckAWSLambdaEventSourceMappingIsBeingDisabled(&conf),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_StartingPositionTimestamp(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	startingPositionTimestamp := time.Now().UTC().Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
			},
		},
	})
}

func TestAccAWSLambdaEventSourceMapping_KinesisBatchWindow(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	batchWindow := int64(5)
	batchWindowUpdate := int64(10)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
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

func TestAccAWSLambdaEventSourceMapping_ParallelizationFactor(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	parallelizationFactor := int64(1)
	parallelizationFactorUpdate := int64(5)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
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

func TestAccAWSLambdaEventSourceMapping_MaximumRetryAttempts(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(10000)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
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

func TestAccAWSLambdaEventSourceMapping_MaximumRetryAttemptsZero(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRetryAttempts := int64(0)
	maximumRetryAttemptsUpdate := int64(100)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
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

func TestAccAWSLambdaEventSourceMapping_MaximumRecordAgeInSeconds(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	maximumRecordAgeInSeconds := int64(604800)
	maximumRecordAgeInSecondsUpdate := int64(3600)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
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

func TestAccAWSLambdaEventSourceMapping_BisectBatch(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_lambda_event_source_mapping.test"
	bisectBatch := false
	bisectBatchUpdate := true

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
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
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
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

func TestAccAWSLambdaEventSourceMapping_KinesisDestinationConfig(t *testing.T) {
	var conf lambda.EventSourceMappingConfiguration
	rName := acctest.RandomWithPrefix("tf-acc-test")
	rString := acctest.RandString(8)
	streamName := fmt.Sprintf("tf_acc_stream_dest_config_%s", rString)
	streamNameUpdated := fmt.Sprintf("tf_acc_stream_dest_config_updated_%s", rString)

	resourceName := "aws_lambda_event_source_mapping.test"
	sqsResourceName := "aws_sqs_queue.sqs_queue_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisDestinationConfig(rName, streamName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination_arn", sqsResourceName, "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"starting_position",
					"starting_position_timestamp",
				},
			},
			{
				Config: testAccAWSLambdaEventSourceMappingConfigKinesisDestinationConfig(rName, streamNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "destination_config.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "destination_config.0.on_failure.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "destination_config.0.on_failure.0.destination_arn", sqsResourceName, "arn"),
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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckLambdaEventSourceMappingDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLambdaEventSourceMappingConfigMsk(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLambdaEventSourceMappingExists(resourceName, &v),
					testAccCheckAWSLambdaEventSourceMappingAttributes(&v),
					resource.TestCheckResourceAttrPair(resourceName, "event_source_arn", eventSourceResourceName, "arn"),
					testAccCheckResourceAttrRfc3339(resourceName, "last_modified"),
					resource.TestCheckNoResourceAttr(resourceName, "starting_position"),
					resource.TestCheckResourceAttr(resourceName, "topics.#", "1"),
					resource.TestCheckTypeSetElemAttr(resourceName, "topics.*", "test"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"enabled", "starting_position"},
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

func testAccCheckAWSLambdaEventSourceMappingDisappears(conf *lambda.EventSourceMappingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		err := resource.Retry(10*time.Minute, func() *resource.RetryError {
			params := &lambda.DeleteEventSourceMappingInput{
				UUID: conf.UUID,
			}
			_, err := conn.DeleteEventSourceMapping(params)
			if err != nil {
				cgw, ok := err.(awserr.Error)
				if ok {
					if cgw.Code() == "ResourceNotFoundException" {
						return nil
					}

					if cgw.Code() == "ResourceInUseException" {
						return resource.RetryableError(fmt.Errorf(
							"Waiting for Lambda Event Source Mapping to delete: %v", conf.UUID))
					}
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error deleting Lambda Event Source Mapping: %s", err))
			}

			return nil
		})

		if err != nil {
			return err
		}

		return resource.Retry(10*time.Minute, func() *resource.RetryError {
			params := &lambda.GetEventSourceMappingInput{
				UUID: conf.UUID,
			}
			_, err = conn.GetEventSourceMapping(params)
			if err != nil {
				cgw, ok := err.(awserr.Error)
				if ok && cgw.Code() == "ResourceNotFoundException" {
					return nil
				}
				return resource.NonRetryableError(
					fmt.Errorf("Error getting Lambda Event Source Mapping: %s", err))
			}
			return resource.RetryableError(fmt.Errorf(
				"Waiting to get Lambda Event Source Mapping: %v", conf.UUID))
		})
	}
}

func testAccCheckLambdaEventSourceMappingDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_event_source_mapping" {
			continue
		}

		_, err := conn.GetEventSourceMapping(&lambda.GetEventSourceMappingInput{
			UUID: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Lambda event source mapping was not deleted")
		}

	}

	return nil

}

func testAccCheckAwsLambdaEventSourceMappingExists(n string, mapping *lambda.EventSourceMappingConfiguration) resource.TestCheckFunc {
	// Wait for IAM role
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Lambda event source mapping not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Lambda event source mapping ID not set")
		}

		conn := testAccProvider.Meta().(*AWSClient).lambdaconn

		params := &lambda.GetEventSourceMappingInput{
			UUID: aws.String(rs.Primary.ID),
		}

		getSourceMappingConfiguration, err := conn.GetEventSourceMapping(params)
		if err != nil {
			return err
		}

		*mapping = *getSourceMappingConfiguration

		return nil
	}
}

func testAccCheckAWSLambdaEventSourceMappingAttributes(mapping *lambda.EventSourceMappingConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		uuid := *mapping.UUID
		if uuid == "" {
			return fmt.Errorf("Could not read Lambda event source mapping's UUID")
		}

		return nil
	}
}

func testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
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
        "kinesis:GetRecords",
        "kinesis:GetShardIterator",
        "kinesis:DescribeStream"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:ListStreams"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "sqs:SendMessage"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_kinesis_stream" "test" {
  name        = %q
  shard_count = 1
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %q
  handler       = "exports.example"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs12.x"
}
`, rName, rName, rName)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisStartingPositionTimestamp(rName, startingPositionTimestamp string) string {
	return testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                  = 100
  enabled                     = true
  event_source_arn            = aws_kinesis_stream.test.arn
  function_name               = aws_lambda_function.test.arn
  starting_position           = "AT_TIMESTAMP"
  starting_position_timestamp = %q
}
`, startingPositionTimestamp)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisBatchWindow(rName string, batchWindow int64) string {
	return testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                         = 100
  maximum_batching_window_in_seconds = %d
  enabled                            = true
  event_source_arn                   = aws_kinesis_stream.test.arn
  function_name                      = aws_lambda_function.test.arn
  starting_position                  = "TRIM_HORIZON"
}
`, batchWindow)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisParallelizationFactor(rName string, parallelizationFactor int64) string {
	return testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size             = 100
  parallelization_factor = %d
  enabled                = true
  event_source_arn       = aws_kinesis_stream.test.arn
  function_name          = aws_lambda_function.test.arn
  starting_position      = "TRIM_HORIZON"
}
`, parallelizationFactor)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRetryAttempts(rName string, maximumRetryAttempts int64) string {
	return testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size             = 100
  maximum_retry_attempts = %d
  enabled                = true
  event_source_arn       = aws_kinesis_stream.test.arn
  function_name          = aws_lambda_function.test.arn
  starting_position      = "TRIM_HORIZON"
}
`, maximumRetryAttempts)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisBisectBatch(rName string, bisectBatch bool) string {
	return testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                     = 100
  bisect_batch_on_function_error = %t
  enabled                        = true
  event_source_arn               = aws_kinesis_stream.test.arn
  function_name                  = aws_lambda_function.test.arn
  starting_position              = "TRIM_HORIZON"
}
`, bisectBatch)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisMaximumRecordAgeInSeconds(rName string, maximumRecordAgeInSeconds int64) string {
	return testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                    = 100
  maximum_record_age_in_seconds = %d
  enabled                       = true
  event_source_arn              = aws_kinesis_stream.test.arn
  function_name                 = aws_lambda_function.test.arn
  starting_position             = "TRIM_HORIZON"
}
`, maximumRecordAgeInSeconds)
}

func testAccAWSLambdaEventSourceMappingConfigKinesisDestinationConfig(rName string, streamName string) string {
	return testAccAWSLambdaEventSourceMappingConfigKinesisBase(rName) + fmt.Sprintf(`
resource "aws_sqs_queue" "sqs_queue_test" {
  name = "%s"
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size        = 100
  enabled           = true
  event_source_arn  = aws_kinesis_stream.test.arn
  function_name     = aws_lambda_function.test.arn
  starting_position = "TRIM_HORIZON"

  destination_config {
    on_failure {
      destination_arn = aws_sqs_queue.sqs_queue_test.arn
    }
  }
}
`, streamName)
}

func testAccAWSLambdaEventSourceMappingConfig_kinesis(roleName, policyName, attName, streamName,
	funcName, uFuncName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy_for_role" {
  name        = "%s"
  path        = "/"
  description = "IAM policy for Lambda event mapping testing"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:GetRecords",
        "kinesis:GetShardIterator",
        "kinesis:DescribeStream"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:ListStreams"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%s"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}

resource "aws_kinesis_stream" "kinesis_stream_test" {
  name        = "%s"
  shard_count = 1
}

resource "aws_lambda_function" "lambda_function_test_create" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_function_test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "lambda_event_source_mapping_test" {
  batch_size        = 100
  event_source_arn  = aws_kinesis_stream.kinesis_stream_test.arn
  enabled           = true
  depends_on        = [aws_iam_policy_attachment.policy_attachment_for_role]
  function_name     = aws_lambda_function.lambda_function_test_create.arn
  starting_position = "TRIM_HORIZON"
}
`, roleName, policyName, attName, streamName, funcName, uFuncName)
}

func testAccAWSLambdaEventSourceMappingConfigUpdate_kinesis_removeBatchSize(roleName, policyName, attName, streamName,
	funcName, uFuncName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy_for_role" {
  name        = "%s"
  path        = "/"
  description = "IAM policy for Lambda event mapping testing"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:GetRecords",
        "kinesis:GetShardIterator",
        "kinesis:DescribeStream"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:ListStreams"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%s"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}

resource "aws_kinesis_stream" "kinesis_stream_test" {
  name        = "%s"
  shard_count = 1
}

resource "aws_lambda_function" "lambda_function_test_create" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_function_test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "lambda_event_source_mapping_test" {
  event_source_arn  = aws_kinesis_stream.kinesis_stream_test.arn
  enabled           = true
  depends_on        = [aws_iam_policy_attachment.policy_attachment_for_role]
  function_name     = aws_lambda_function.lambda_function_test_create.arn
  starting_position = "TRIM_HORIZON"
}
`, roleName, policyName, attName, streamName, funcName, uFuncName)
}

func testAccAWSLambdaEventSourceMappingConfigUpdate_kinesis(roleName, policyName, attName, streamName,
	funcName, uFuncName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy_for_role" {
  name        = "%s"
  path        = "/"
  description = "IAM policy for Lambda event mapping testing"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:GetRecords",
        "kinesis:GetShardIterator",
        "kinesis:DescribeStream"
      ],
      "Resource": "*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "kinesis:ListStreams"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%s"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}

resource "aws_kinesis_stream" "kinesis_stream_test" {
  name        = "%s"
  shard_count = 1
}

resource "aws_lambda_function" "lambda_function_test_create" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_function_test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "lambda_event_source_mapping_test" {
  batch_size        = 200
  event_source_arn  = aws_kinesis_stream.kinesis_stream_test.arn
  enabled           = false
  depends_on        = [aws_iam_policy_attachment.policy_attachment_for_role]
  function_name     = aws_lambda_function.lambda_function_test_update.arn
  starting_position = "TRIM_HORIZON"
}
`, roleName, policyName, attName, streamName, funcName, uFuncName)
}

func testAccAWSLambdaEventSourceMappingConfig_sqs(roleName, policyName, attName, streamName,
	funcName, uFuncName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy_for_role" {
  name        = "%s"
  path        = "/"
  description = "IAM policy for Lambda event mapping testing"

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

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%s"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}

resource "aws_sqs_queue" "sqs_queue_test" {
  name = "%s"
}

resource "aws_lambda_function" "lambda_function_test_create" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_function_test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "lambda_event_source_mapping_test" {
  batch_size       = 10
  event_source_arn = aws_sqs_queue.sqs_queue_test.arn
  enabled          = true
  depends_on       = [aws_iam_policy_attachment.policy_attachment_for_role]
  function_name    = aws_lambda_function.lambda_function_test_create.arn
}
`, roleName, policyName, attName, streamName, funcName, uFuncName)
}

func testAccAWSLambdaEventSourceMappingConfigUpdate_sqs(roleName, policyName, attName, streamName,
	funcName, uFuncName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy_for_role" {
  name        = "%s"
  path        = "/"
  description = "IAM policy for Lambda event mapping testing"

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

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%s"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}

resource "aws_sqs_queue" "sqs_queue_test" {
  name = "%s"
}

resource "aws_lambda_function" "lambda_function_test_create" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_function" "lambda_function_test_update" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "lambda_event_source_mapping_test" {
  batch_size       = 5
  event_source_arn = aws_sqs_queue.sqs_queue_test.arn
  enabled          = false
  depends_on       = [aws_iam_policy_attachment.policy_attachment_for_role]
  function_name    = aws_lambda_function.lambda_function_test_update.arn
}
`, roleName, policyName, attName, streamName, funcName, uFuncName)
}

func testAccAWSLambdaEventSourceMappingConfig_sqs_testWithFunctionName(roleName, policyName, attName, streamName, funcName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "iam_for_lambda" {
  name = "%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
EOF
}

resource "aws_iam_policy" "policy_for_role" {
  name        = "%[2]s"
  path        = "/"
  description = "IAM policy for Lambda event mapping testing"

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

resource "aws_iam_policy_attachment" "policy_attachment_for_role" {
  name       = "%[3]s"
  roles      = [aws_iam_role.iam_for_lambda.name]
  policy_arn = aws_iam_policy.policy_for_role.arn
}

resource "aws_sqs_queue" "sqs_queue_test" {
  name = "%[4]s"
}

resource "aws_lambda_function" "lambda_function_test_create" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[5]s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "lambda_event_source_mapping_test" {
  batch_size       = 5
  event_source_arn = aws_sqs_queue.sqs_queue_test.arn
  depends_on       = [aws_iam_policy_attachment.policy_attachment_for_role]
  enabled          = false
  function_name    = "%[5]s"
}
`, roleName, policyName, attName, streamName, funcName)
}

func testAccAWSLambdaEventSourceMappingConfigSQSBase(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
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

func testAccAWSLambdaEventSourceMappingConfigSqsWithBatchWindow(rName string, batchWindow int64) string {
	return composeConfig(testAccAWSLambdaEventSourceMappingConfigSQSBase(rName), fmt.Sprintf(`
resource "aws_lambda_event_source_mapping" "test" {
  batch_size                         = 10
  maximum_batching_window_in_seconds = %d
  event_source_arn                   = aws_sqs_queue.test.arn
  enabled                            = false
  function_name                      = aws_lambda_function.test.arn
}
`, batchWindow))
}

func testAccAWSLambdaEventSourceMappingConfigMsk(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "lambda.amazonaws.com"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
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

resource "aws_msk_cluster" "test" {
  cluster_name           = %[1]q
  kafka_version          = "2.2.1"
  number_of_broker_nodes = 2

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    ebs_volume_size = 10
    instance_type   = "kafka.t3.small"
    security_groups = [aws_security_group.test.id]
  }
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = %[1]q
  role          = aws_iam_role.test.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}

resource "aws_lambda_event_source_mapping" "test" {
  batch_size       = 100
  event_source_arn = aws_msk_cluster.test.arn
  enabled          = true
  function_name    = aws_lambda_function.test.arn
  topics           = ["test"]

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName))
}
