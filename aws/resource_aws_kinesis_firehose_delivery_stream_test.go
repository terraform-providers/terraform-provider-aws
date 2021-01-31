package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/firehose"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_kinesis_firehose_delivery_stream", &resource.Sweeper{
		Name: "aws_kinesis_firehose_delivery_stream",
		F:    testSweepKinesisFirehoseDeliveryStreams,
	})
}

func testSweepKinesisFirehoseDeliveryStreams(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}

	conn := client.(*AWSClient).firehoseconn
	input := &firehose.ListDeliveryStreamsInput{}

	for {
		output, err := conn.ListDeliveryStreams(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Kinesis Firehose Delivery Streams sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error listing Kinesis Firehose Delivery Streams: %s", err)
		}

		if len(output.DeliveryStreamNames) == 0 {
			log.Print("[DEBUG] No Kinesis Firehose Delivery Streams to sweep")
			return nil
		}

		for _, deliveryStreamNamePtr := range output.DeliveryStreamNames {
			input := &firehose.DeleteDeliveryStreamInput{
				DeliveryStreamName: deliveryStreamNamePtr,
			}
			name := aws.StringValue(deliveryStreamNamePtr)

			log.Printf("[INFO] Deleting Kinesis Firehose Delivery Stream: %s", name)
			_, err := conn.DeleteDeliveryStream(input)

			if isAWSErr(err, firehose.ErrCodeResourceNotFoundException, "") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting Kinesis Firehose Delivery Stream (%s): %s", name, err)
			}

			if err := waitForKinesisFirehoseDeliveryStreamDeletion(conn, name); err != nil {
				return fmt.Errorf("error waiting for Kinesis Firehose Delivery Stream (%s) deletion: %s", name, err)
			}
		}

		if !aws.BoolValue(output.HasMoreDeliveryStreams) {
			break
		}
	}

	return nil
}

func TestAccAWSKinesisFirehoseDeliveryStream_basic(t *testing.T) {
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	rInt := acctest.RandInt()

	funcName := fmt.Sprintf("aws_kinesis_firehose_ds_import_%d", rInt)
	policyName := fmt.Sprintf("tf_acc_policy_%d", rInt)
	roleName := fmt.Sprintf("tf_acc_role_%d", rInt)
	var stream firehose.DeliveryStreamDescription

	config := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic,
			rInt, rInt, rInt, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure we properly error on malformed import IDs
			{
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: "just-a-name",
				ExpectError:   regexp.MustCompile(`Expected ID in format`),
			},
			{
				ResourceName:  resourceName,
				ImportState:   true,
				ImportStateId: "arn:aws:firehose:us-east-1:123456789012:missing-slash", //lintignore:AWSAT003,AWSAT005
				ExpectError:   regexp.MustCompile(`Expected ID in format`),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_disappears(t *testing.T) {
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	rInt := acctest.RandInt()

	funcName := fmt.Sprintf("aws_kinesis_firehose_ds_import_%d", rInt)
	policyName := fmt.Sprintf("tf_acc_policy_%d", rInt)
	roleName := fmt.Sprintf("tf_acc_role_%d", rInt)
	var stream firehose.DeliveryStreamDescription

	config := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic,
			rInt, rInt, rInt, rInt)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsKinesisFirehoseDeliveryStream(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3basic(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	config := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3basic,
		ri, ri, ri, ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3basicWithSSE(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("terraform-kinesis-firehose-basictest-%d", rInt)
	config := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3basic,
		rInt, rInt, rInt, rInt)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName, rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName, rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "false"),
				),
			},
			{
				Config:   config,
				PlanOnly: true,
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName, rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", "AWS_OWNED_CMK"),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3basicWithSSEAndKeyArn(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("terraform-kinesis-firehose-basictest-%d", rInt)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSEAndKeyArn(rName, rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", firehose.KeyTypeCustomerManagedCmk),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.key_arn", "aws_kms_key.test", "arn"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName, rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "false"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSEAndKeyArn(rName, rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", firehose.KeyTypeCustomerManagedCmk),
					resource.TestCheckResourceAttrPair(resourceName, "server_side_encryption.0.key_arn", "aws_kms_key.test", "arn"),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3basicWithSSEAndKeyType(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("terraform-kinesis-firehose-basictest-%d", rInt)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSEAndKeyType(rName, rInt, true, firehose.KeyTypeAwsOwnedCmk),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", firehose.KeyTypeAwsOwnedCmk),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName, rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "false"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSEAndKeyType(rName, rInt, true, firehose.KeyTypeAwsOwnedCmk),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.key_type", firehose.KeyTypeAwsOwnedCmk),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3basicWithTags(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := fmt.Sprintf("terraform-kinesis-firehose-basictest-%d", rInt)
	config := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3basic,
		rInt, rInt, rInt, rInt)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithTags(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithTagsChanged(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3KinesisStreamSource(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	config := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3KinesisStreamSource,
		ri, ri, ri, ri, ri, ri, ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3WithCloudwatchLogging(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3WithCloudwatchLogging(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_s3ConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	ri := acctest.RandInt()
	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3basic,
		ri, ri, ri, ri)
	postConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_s3Updates,
		ri, ri, ri, ri)

	updatedS3DestinationConfig := &firehose.S3DestinationDescription{
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(400),
			SizeInMBs:         aws.Int64(10),
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, updatedS3DestinationConfig, nil, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3basic(t *testing.T) {
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	config := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic,
			ri, ri, ri, ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", ""),
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

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_DataFormatConversionConfiguration_Enabled(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_Enabled(rName, rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", "false"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_Enabled(rName, rInt, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", "true"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_Enabled(rName, rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_ExternalUpdate(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ExternalUpdate(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				PreConfig: func() {
					conn := testAccProvider.Meta().(*AWSClient).firehoseconn
					udi := firehose.UpdateDestinationInput{
						DeliveryStreamName:             aws.String(rName),
						DestinationId:                  aws.String("destinationId-000000000001"),
						CurrentDeliveryStreamVersionId: aws.String("1"),
						ExtendedS3DestinationUpdate: &firehose.ExtendedS3DestinationUpdate{
							DataFormatConversionConfiguration: &firehose.DataFormatConversionConfiguration{
								Enabled: aws.Bool(false),
							},
							ProcessingConfiguration: &firehose.ProcessingConfiguration{
								Enabled:    aws.Bool(false),
								Processors: []*firehose.Processor{},
							},
						},
					}
					_, err := conn.UpdateDestination(&udi)
					if err != nil {
						t.Fatalf("Unable to update firehose destination: %s", err)
					}
				},
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ExternalUpdate(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_DataFormatConversionConfiguration_Deserializer_Update(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_HiveJsonSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.hive_json_ser_de.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OpenXJsonSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.open_x_json_ser_de.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_DataFormatConversionConfiguration_HiveJsonSerDe_Empty(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_HiveJsonSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.hive_json_ser_de.#", "1"),
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

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_DataFormatConversionConfiguration_OpenXJsonSerDe_Empty(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OpenXJsonSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.input_format_configuration.0.deserializer.0.open_x_json_ser_de.#", "1"),
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

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_DataFormatConversionConfiguration_OrcSerDe_Empty(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OrcSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.orc_ser_de.#", "1"),
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

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_DataFormatConversionConfiguration_ParquetSerDe_Empty(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_ParquetSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.parquet_ser_de.#", "1"),
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

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_DataFormatConversionConfiguration_Serializer_Update(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OrcSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.orc_ser_de.#", "1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_ParquetSerDe_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.data_format_conversion_configuration.0.output_format_configuration.0.serializer.0.parquet_ser_de.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_ErrorOutputPrefix(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ErrorOutputPrefix(rName, rInt, "prefix1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", "prefix1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ErrorOutputPrefix(rName, rInt, "prefix2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.error_output_prefix", "prefix2"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/12600
func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_ProcessingConfiguration_Empty(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ProcessingConfiguration_Empty(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "extended_s3_configuration.0.processing_configuration.#", "1"),
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

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3KmsKeyArn(t *testing.T) {
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	config := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3KmsKeyArn,
			ri, ri, ri, ri, ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "extended_s3_configuration.0.kms_key_arn", "aws_kms_key.test", "arn"),
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

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3Updates(t *testing.T) {
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()

	preConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic,
			ri, ri, ri, ri)
	firstUpdateConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates_Initial,
			ri, ri, ri, ri)
	removeProcessorsConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates_RemoveProcessors,
			ri, ri, ri, ri)

	firstUpdateExtendedS3DestinationConfig := &firehose.ExtendedS3DestinationDescription{
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(400),
			SizeInMBs:         aws.Int64(10),
		},
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
		S3BackupMode: aws.String("Enabled"),
	}

	removeProcessorsExtendedS3DestinationConfig := &firehose.ExtendedS3DestinationDescription{
		BufferingHints: &firehose.BufferingHints{
			IntervalInSeconds: aws.Int64(400),
			SizeInMBs:         aws.Int64(10),
		},
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled:    aws.Bool(false),
			Processors: []*firehose.Processor{},
		},
		S3BackupMode: aws.String("Enabled"),
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: firstUpdateConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, firstUpdateExtendedS3DestinationConfig, nil, nil, nil, nil),
				),
			},
			{
				Config: removeProcessorsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, removeProcessorsExtendedS3DestinationConfig, nil, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ExtendedS3_KinesisStreamSource(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	config := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_extendedS3_KinesisStreamSource,
		ri, ri, ri, ri, ri, ri, ri)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
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

func TestAccAWSKinesisFirehoseDeliveryStream_RedshiftConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	updatedRedshiftConfig := &firehose.RedshiftDestinationDescription{
		CopyCommand: &firehose.CopyCommand{
			CopyOptions: aws.String("GZIP"),
		},
		S3BackupMode: aws.String("Enabled"),
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamRedshiftConfig(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"redshift_configuration.0.password"},
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamRedshiftConfigUpdates(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, updatedRedshiftConfig, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_SplunkConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	ri := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)

	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_SplunkBasic,
		ri, ri, ri, ri)
	postConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_SplunkUpdates,
			ri, ri, ri, ri)

	updatedSplunkConfig := &firehose.SplunkDestinationDescription{
		HECEndpointType:                   aws.String("Event"),
		HECAcknowledgmentTimeoutInSeconds: aws.Int64(600),
		S3BackupMode:                      aws.String("FailedEventsOnly"),
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, updatedSplunkConfig, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_HttpEndpointConfiguration(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	ri := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)

	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_HTTPEndpointBasic,
		ri, ri, ri, ri)
	postConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_HTTPEndpointUpdates,
			ri, ri, ri, ri)

	updatedHTTPEndpointConfig := &firehose.HttpEndpointDestinationDescription{
		EndpointConfiguration: &firehose.HttpEndpointDescription{
			Url:  aws.String("https://input-test.com:443"),
			Name: aws.String("HTTP_test"),
		},
		S3BackupMode: aws.String("FailedDataOnly"),
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, updatedHTTPEndpointConfig),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_HttpEndpointConfiguration_RetryDuration(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	rInt := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_HTTPEndpoint_RetryDuration(rInt, 301),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_HTTPEndpoint_RetryDuration(rInt, 302),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ElasticsearchConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	ri := acctest.RandInt()
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)
	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchBasic,
		ri, ri, ri, ri, ri, ri)
	postConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchUpdate,
			ri, ri, ri, ri, ri, ri)

	updatedElasticSearchConfig := &firehose.ElasticsearchDestinationDescription{
		BufferingHints: &firehose.ElasticsearchBufferingHints{
			IntervalInSeconds: aws.Int64(500),
		},
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticSearchConfig, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_ElasticsearchConfigEndpointUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	ri := acctest.RandInt()
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)
	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchEndpoint,
		ri, ri, ri, ri, ri, ri)
	postConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchEndpointUpdate,
			ri, ri, ri, ri, ri, ri)

	updatedElasticSearchConfig := &firehose.ElasticsearchDestinationDescription{
		BufferingHints: &firehose.ElasticsearchBufferingHints{
			IntervalInSeconds: aws.Int64(500),
		},
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticSearchConfig, nil, nil),
				),
			},
		},
	})
}

// This doesn't actually test updating VPC Configuration. It tests changing Elasticsearch configuration
// when the Kinesis Firehose delivery stream has a VPC Configuration.
func TestAccAWSKinesisFirehoseDeliveryStream_ElasticsearchWithVpcConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	ri := acctest.RandInt()
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)

	updatedElasticSearchConfig := &firehose.ElasticsearchDestinationDescription{
		BufferingHints: &firehose.ElasticsearchBufferingHints{
			IntervalInSeconds: aws.Int64(500),
		},
		ProcessingConfiguration: &firehose.ProcessingConfiguration{
			Enabled: aws.Bool(true),
			Processors: []*firehose.Processor{
				{
					Type: aws.String("Lambda"),
					Parameters: []*firehose.ProcessorParameter{
						{
							ParameterName:  aws.String("LambdaArn"),
							ParameterValue: aws.String("valueNotTested"),
						},
					},
				},
			},
		},
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckIamServiceLinkedRoleEs(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchVpcBasic(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.vpc_id", "aws_vpc.elasticsearch_in_vpc", "id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.role_arn", "aws_iam_role.firehose", "arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchVpcUpdate(funcName, policyName, roleName, ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticSearchConfig, nil, nil),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.vpc_id", "aws_vpc.elasticsearch_in_vpc", "id"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.subnet_ids.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "elasticsearch_configuration.0.vpc_config.0.security_group_ids.#", "2"),
					resource.TestCheckResourceAttrPair(resourceName, "elasticsearch_configuration.0.vpc_config.0.role_arn", "aws_iam_role.firehose", "arn"),
				),
			},
		},
	})
}

// Regression test for https://github.com/hashicorp/terraform-provider-aws/issues/1657
func TestAccAWSKinesisFirehoseDeliveryStream_missingProcessingConfiguration(t *testing.T) {
	var stream firehose.DeliveryStreamDescription
	ri := acctest.RandInt()
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckKinesisFirehoseDeliveryStreamDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_missingProcessingConfiguration(ri),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil, nil),
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

func testAccCheckKinesisFirehoseDeliveryStreamExists(n string, stream *firehose.DeliveryStreamDescription) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		log.Printf("State: %#v", s.RootModule().Resources)
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Kinesis Firehose ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).firehoseconn
		describeOpts := &firehose.DescribeDeliveryStreamInput{
			DeliveryStreamName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeDeliveryStream(describeOpts)
		if err != nil {
			return err
		}

		*stream = *resp.DeliveryStreamDescription

		return nil
	}
}

func testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(stream *firehose.DeliveryStreamDescription, s3config interface{}, extendedS3config interface{}, redshiftConfig interface{}, elasticsearchConfig interface{}, splunkConfig interface{}, httpEndpointConfig interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*stream.DeliveryStreamName, "terraform-kinesis-firehose") && !strings.HasPrefix(*stream.DeliveryStreamName, "tf-acc-test") {
			return fmt.Errorf("Bad Stream name: %s", *stream.DeliveryStreamName)
		}
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_kinesis_firehose_delivery_stream" {
				continue
			}
			if *stream.DeliveryStreamARN != rs.Primary.Attributes["arn"] {
				return fmt.Errorf("Bad Delivery Stream ARN\n\t expected: %s\n\tgot: %s\n", rs.Primary.Attributes["arn"], *stream.DeliveryStreamARN)
			}

			if s3config != nil {
				s := s3config.(*firehose.S3DestinationDescription)
				// Range over the Stream Destinations, looking for the matching S3
				// destination. For simplicity, our test only have a single S3 or
				// Redshift destination, so at this time it's safe to match on the first
				// one
				var match bool
				for _, d := range stream.Destinations {
					if d.S3DestinationDescription != nil {
						if *d.S3DestinationDescription.BufferingHints.SizeInMBs == *s.BufferingHints.SizeInMBs {
							match = true
						}
					}
				}
				if !match {
					return fmt.Errorf("Mismatch s3 buffer size, expected: %s, got: %s", s, stream.Destinations)
				}
			}

			if extendedS3config != nil {
				es := extendedS3config.(*firehose.ExtendedS3DestinationDescription)

				// Range over the Stream Destinations, looking for the matching S3
				// destination. For simplicity, our test only have a single S3 or
				// Redshift destination, so at this time it's safe to match on the first
				// one
				var match, processingConfigMatch, matchS3BackupMode bool
				for _, d := range stream.Destinations {
					if d.ExtendedS3DestinationDescription != nil {
						if *d.ExtendedS3DestinationDescription.BufferingHints.SizeInMBs == *es.BufferingHints.SizeInMBs {
							match = true
						}
						if *d.ExtendedS3DestinationDescription.S3BackupMode == *es.S3BackupMode {
							matchS3BackupMode = true
						}

						processingConfigMatch = len(es.ProcessingConfiguration.Processors) == len(d.ExtendedS3DestinationDescription.ProcessingConfiguration.Processors)
					}
				}
				if !match {
					return fmt.Errorf("Mismatch extended s3 buffer size, expected: %s, got: %s", es, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch extended s3 ProcessingConfiguration.Processors count, expected: %s, got: %s", es, stream.Destinations)
				}
				if !matchS3BackupMode {
					return fmt.Errorf("Mismatch extended s3 S3BackupMode, expected: %s, got: %s", es, stream.Destinations)
				}
			}

			if redshiftConfig != nil {
				r := redshiftConfig.(*firehose.RedshiftDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Redshift
				// destination
				var matchCopyOptions, matchS3BackupMode, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.RedshiftDestinationDescription != nil {
						if *d.RedshiftDestinationDescription.CopyCommand.CopyOptions == *r.CopyCommand.CopyOptions {
							matchCopyOptions = true
						}
						if *d.RedshiftDestinationDescription.S3BackupMode == *r.S3BackupMode {
							matchS3BackupMode = true
						}
						if r.ProcessingConfiguration != nil && d.RedshiftDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(r.ProcessingConfiguration.Processors) == len(d.RedshiftDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !matchCopyOptions || !matchS3BackupMode {
					return fmt.Errorf("Mismatch Redshift CopyOptions or S3BackupMode, expected: %s, got: %s", r, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch Redshift ProcessingConfiguration.Processors count, expected: %s, got: %s", r, stream.Destinations)
				}
			}

			if elasticsearchConfig != nil {
				es := elasticsearchConfig.(*firehose.ElasticsearchDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Elasticsearch destination
				var match, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.ElasticsearchDestinationDescription != nil {
						match = true
						if es.ProcessingConfiguration != nil && d.ElasticsearchDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(es.ProcessingConfiguration.Processors) == len(d.ElasticsearchDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !match {
					return fmt.Errorf("Mismatch Elasticsearch Buffering Interval, expected: %s, got: %s", es, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch Elasticsearch ProcessingConfiguration.Processors count, expected: %s, got: %s", es, stream.Destinations)
				}
			}

			if splunkConfig != nil {
				s := splunkConfig.(*firehose.SplunkDestinationDescription)
				// Range over the Stream Destinations, looking for the matching Splunk destination
				var matchHECEndpointType, matchHECAcknowledgmentTimeoutInSeconds, matchS3BackupMode, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.SplunkDestinationDescription != nil {
						if *d.SplunkDestinationDescription.HECEndpointType == *s.HECEndpointType {
							matchHECEndpointType = true
						}
						if *d.SplunkDestinationDescription.HECAcknowledgmentTimeoutInSeconds == *s.HECAcknowledgmentTimeoutInSeconds {
							matchHECAcknowledgmentTimeoutInSeconds = true
						}
						if *d.SplunkDestinationDescription.S3BackupMode == *s.S3BackupMode {
							matchS3BackupMode = true
						}
						if s.ProcessingConfiguration != nil && d.SplunkDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(s.ProcessingConfiguration.Processors) == len(d.SplunkDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !matchHECEndpointType || !matchHECAcknowledgmentTimeoutInSeconds || !matchS3BackupMode {
					return fmt.Errorf("Mismatch Splunk HECEndpointType or HECAcknowledgmentTimeoutInSeconds or S3BackupMode, expected: %s, got: %s", s, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch extended splunk ProcessingConfiguration.Processors count, expected: %s, got: %s", s, stream.Destinations)
				}
			}

			if httpEndpointConfig != nil {
				s := httpEndpointConfig.(*firehose.HttpEndpointDestinationDescription)
				// Range over the Stream Destinations, looking for the matching HttpEndpoint destination
				var matchS3BackupMode, matchUrl, matchName, processingConfigMatch bool
				for _, d := range stream.Destinations {
					if d.HttpEndpointDestinationDescription != nil {
						if *d.HttpEndpointDestinationDescription.S3BackupMode == *s.S3BackupMode {
							matchS3BackupMode = true
						}
						if *d.HttpEndpointDestinationDescription.EndpointConfiguration.Url == *s.EndpointConfiguration.Url {
							matchUrl = true
						}
						if *d.HttpEndpointDestinationDescription.EndpointConfiguration.Name == *s.EndpointConfiguration.Name {
							matchName = true
						}
						if s.ProcessingConfiguration != nil && d.HttpEndpointDestinationDescription.ProcessingConfiguration != nil {
							processingConfigMatch = len(s.ProcessingConfiguration.Processors) == len(d.HttpEndpointDestinationDescription.ProcessingConfiguration.Processors)
						}
					}
				}
				if !matchS3BackupMode {
					return fmt.Errorf("Mismatch HTTP Endpoint S3BackupMode, expected: %s, got: %s", s, stream.Destinations)
				}
				if !matchUrl || !matchName {
					return fmt.Errorf("Mismatch HTTP Endpoint EndpointConfiguration, expected: %s, got: %s", s, stream.Destinations)
				}
				if !processingConfigMatch {
					return fmt.Errorf("Mismatch HTTP Endpoint ProcessingConfiguration.Processors count, expected: %s, got: %s", s, stream.Destinations)
				}
			}

		}
		return nil
	}
}

func testAccCheckKinesisFirehoseDeliveryStreamDestroy_ExtendedS3(s *terraform.State) error {
	err := testAccCheckKinesisFirehoseDeliveryStreamDestroy(s)

	if err == nil {
		err = testAccCheckFirehoseLambdaFunctionDestroy(s)
	}

	return err
}

func testAccCheckKinesisFirehoseDeliveryStreamDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_kinesis_firehose_delivery_stream" {
			continue
		}
		conn := testAccProvider.Meta().(*AWSClient).firehoseconn
		describeOpts := &firehose.DescribeDeliveryStreamInput{
			DeliveryStreamName: aws.String(rs.Primary.Attributes["name"]),
		}
		resp, err := conn.DescribeDeliveryStream(describeOpts)
		if err == nil {
			if resp.DeliveryStreamDescription != nil && *resp.DeliveryStreamDescription.DeliveryStreamStatus != "DELETING" {
				return fmt.Errorf("Error: Delivery Stream still exists")
			}
		}

		return nil

	}

	return nil
}

func testAccCheckFirehoseLambdaFunctionDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).lambdaconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lambda_function" {
			continue
		}

		_, err := conn.GetFunction(&lambda.GetFunctionInput{
			FunctionName: aws.String(rs.Primary.ID),
		})

		if err == nil {
			return fmt.Errorf("Lambda Function still exists")
		}
	}

	return nil
}

func baseAccFirehoseAWSLambdaConfig(policyName, roleName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role_policy" "iam_policy_for_lambda" {
  name = "%s"
  role = aws_iam_role.iam_for_lambda.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "logs:CreateLogGroup",
        "logs:CreateLogStream",
        "logs:PutLogEvents"
      ],
      "Resource": "arn:${data.aws_partition.current.partition}:logs:*:*:*"
    },
    {
      "Effect": "Allow",
      "Action": [
        "xray:PutTraceSegments"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

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
`, policyName, roleName)
}

func testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName string) string {
	return fmt.Sprintf(baseAccFirehoseAWSLambdaConfig(policyName, roleName)+`
resource "aws_lambda_function" "lambda_function_test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%s"
  role          = aws_iam_role.iam_for_lambda.arn
  handler       = "exports.example"
  runtime       = "nodejs12.x"
}
`, funcName)
}

const testAccKinesisFirehoseDeliveryStreamBaseConfig = `
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = "tf_acctest_firehose_delivery_role_%d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%d"
  acl    = "private"
}

resource "aws_iam_role_policy" "firehose" {
  name = "tf_acctest_firehose_delivery_policy_%d"
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.bucket.arn}",
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Sid": "GlueAccess",
      "Effect": "Allow",
      "Action": [
        "glue:GetTable",
        "glue:GetTableVersion",
        "glue:GetTableVersions"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "LakeFormationDataAccess",
      "Effect": "Allow",
      "Action": [
        "lakeformation:GetDataAccess"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}
`

const testAccFirehoseKinesisStreamSource = `
resource "aws_kinesis_stream" "source" {
  name        = "terraform-kinesis-source-stream-basictest-%d"
  shard_count = 1
}

resource "aws_iam_role" "kinesis_source" {
  name = "tf_acctest_kinesis_source_role_%d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "kinesis_source" {
  name = "tf_acctest_kinesis_source_policy_%d"
  role = aws_iam_role.kinesis_source.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "kinesis:DescribeStream",
        "kinesis:GetShardIterator",
        "kinesis:GetRecords"
      ],
      "Resource": [
        "${aws_kinesis_stream.source.arn}"
      ]
    }
  ]
}
EOF
}
`

func testAccKinesisFirehoseDeliveryStreamConfig_s3WithCloudwatchLogging(rInt int) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = "tf-acc-test-%[1]d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  name = "tf-acc-test-%[1]d"
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.bucket.arn}",
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:putLogEvents"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs::log-group:/aws/kinesisfirehose/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"
}

resource "aws_cloudwatch_log_group" "test" {
  name = "tf-acc-test-%[1]d"
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = "tf-acc-test-%[1]d"
  log_group_name = aws_cloudwatch_log_group.test.name
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-%[1]d"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = aws_cloudwatch_log_group.test.name
      log_stream_name = aws_cloudwatch_log_stream.test.name
    }
  }
}
`, rInt)
}

var testAccKinesisFirehoseDeliveryStreamConfig_s3basic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-basictest-%d"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName string, rInt int, sseEnabled bool) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  server_side_encryption {
    enabled = %[2]t
  }
}
`, rName, sseEnabled)
}

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSEAndKeyArn(rName string, rInt int, sseEnabled bool) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = %[1]q
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  server_side_encryption {
    enabled  = %[2]t
    key_arn  = aws_kms_key.test.arn
    key_type = "CUSTOMER_MANAGED_CMK"
  }
}
`, rName, sseEnabled)
}

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSEAndKeyType(rName string, rInt int, sseEnabled bool, keyType string) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  server_side_encryption {
    enabled  = %[2]t
    key_type = %[3]q
  }
}
`, rName, sseEnabled, keyType)
}

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithTags(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "%s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    Environment = "production"
    Usage       = "original"
  }
}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithTagsChanged(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "%s"
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  tags = {
    Usage = "changed"
  }
}
`, rName)
}

var testAccKinesisFirehoseDeliveryStreamConfig_s3KinesisStreamSource = testAccKinesisFirehoseDeliveryStreamBaseConfig + testAccFirehoseKinesisStreamSource + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose, aws_iam_role_policy.kinesis_source]
  name       = "terraform-kinesis-firehose-basictest-%d"

  kinesis_source_configuration {
    kinesis_stream_arn = aws_kinesis_stream.source.arn
    role_arn           = aws_iam_role.kinesis_source.arn
  }

  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_s3Updates = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-s3test-%d"
  destination = "s3"

  s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    bucket_arn         = aws_s3_bucket.bucket.arn
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }

    s3_backup_mode = "Disabled"
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3_KinesisStreamSource = testAccKinesisFirehoseDeliveryStreamBaseConfig + testAccFirehoseKinesisStreamSource + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose, aws_iam_role_policy.kinesis_source]
  name       = "terraform-kinesis-firehose-basictest-%d"

  kinesis_source_configuration {
    kinesis_stream_arn = aws_kinesis_stream.source.arn
    role_arn           = aws_iam_role.kinesis_source.arn
  }

  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}
`

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_Enabled(rName string, rInt int, enabled bool) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      enabled = %[2]t

      input_format_configuration {
        deserializer {
          hive_json_ser_de {} # we have to pick one
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {} # we have to pick one
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName, enabled)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_HiveJsonSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {} # we have to pick one
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ExternalUpdate(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "%s"

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    role_arn   = aws_iam_role.firehose.arn
  }
}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OpenXJsonSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          open_x_json_ser_de {}
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {} # we have to pick one
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OrcSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {} # we have to pick one
        }
      }

      output_format_configuration {
        serializer {
          orc_ser_de {}
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_ParquetSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[1]q

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_lakeformation_permissions" "test" {
  permissions = ["ALL"]
  principal   = aws_iam_role.firehose.arn

  table {
    database_name = aws_glue_catalog_database.test.name
    name          = aws_glue_catalog_table.test.name
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = aws_iam_role.firehose.arn

    data_format_conversion_configuration {
      input_format_configuration {
        deserializer {
          hive_json_ser_de {} # we have to pick one
        }
      }

      output_format_configuration {
        serializer {
          parquet_ser_de {}
        }
      }

      schema_configuration {
        database_name = aws_glue_catalog_table.test.database_name
        role_arn      = aws_iam_role.firehose.arn
        table_name    = aws_glue_catalog_table.test.name
      }
    }
  }

  depends_on = [aws_iam_role_policy.firehose, aws_lakeformation_permissions.test]
}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ErrorOutputPrefix(rName string, rInt int, errorOutputPrefix string) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %q

  extended_s3_configuration {
    bucket_arn          = aws_s3_bucket.bucket.arn
    error_output_prefix = %q
    role_arn            = aws_iam_role.firehose.arn
  }

  depends_on = [aws_iam_role_policy.firehose]
}
`, rName, errorOutputPrefix)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ProcessingConfiguration_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %[1]q

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.bucket.arn
    role_arn   = aws_iam_role.firehose.arn

    processing_configuration {}
  }

  depends_on = [aws_iam_role_policy.firehose]
}
`, rName)
}

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3KmsKeyArn = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kms_key" "test" {
  description = "Terraform acc test %s"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn    = aws_iam_role.firehose.arn
    bucket_arn  = aws_s3_bucket.bucket.arn
    kms_key_arn = aws_kms_key.test.arn

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates_Initial = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }

    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
    s3_backup_mode     = "Enabled"

    s3_backup_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates_RemoveProcessors = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    bucket_arn         = aws_s3_bucket.bucket.arn
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
    s3_backup_mode     = "Enabled"

    s3_backup_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}
`

func testAccKinesisFirehoseDeliveryStreamRedshiftConfigBase(rName string, rInt int) string {
	return composeConfig(
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt),
		fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_redshift_subnet_group" "test" {
  name        = %[1]q
  description = "test"
  subnet_ids  = [aws_subnet.test.id]
}

resource "aws_redshift_cluster" "test" {
  cluster_identifier        = %[1]q
  availability_zone         = data.aws_availability_zones.available.names[0]
  database_name             = "test"
  master_username           = "testuser"
  master_password           = "T3stPass"
  node_type                 = "dc1.large"
  cluster_type              = "single-node"
  skip_final_snapshot       = true
  cluster_subnet_group_name = aws_redshift_subnet_group.test.id
  publicly_accessible       = false
}
`, rName))
}

func testAccKinesisFirehoseDeliveryStreamRedshiftConfig(rName string, rInt int) string {
	return composeConfig(
		testAccKinesisFirehoseDeliveryStreamRedshiftConfigBase(rName, rInt),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "redshift"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  redshift_configuration {
    role_arn        = aws_iam_role.firehose.arn
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/${aws_redshift_cluster.test.database_name}"
    username        = "testuser"
    password        = "T3stPass"
    data_table_name = "test-table"
  }
}
`, rName))
}

func testAccKinesisFirehoseDeliveryStreamRedshiftConfigUpdates(rName string, rInt int) string {
	return composeConfig(
		testAccFirehoseAWSLambdaConfigBasic(rName, rName, rName),
		testAccKinesisFirehoseDeliveryStreamRedshiftConfigBase(rName, rInt),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = %[1]q
  destination = "redshift"

  s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    bucket_arn         = aws_s3_bucket.bucket.arn
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }

  redshift_configuration {
    role_arn        = aws_iam_role.firehose.arn
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test.endpoint}/${aws_redshift_cluster.test.database_name}"
    username        = "testuser"
    password        = "T3stPass"
    s3_backup_mode  = "Enabled"

    s3_backup_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }

    data_table_name    = "test-table"
    copy_options       = "GZIP"
    data_table_columns = "test-col"

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
  }
}
`, rName))
}

var testAccKinesisFirehoseDeliveryStreamConfig_SplunkBasic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-basicsplunktest-%d"
  destination = "splunk"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  splunk_configuration {
    hec_endpoint = "https://input-test.com:443"
    hec_token    = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_SplunkUpdates = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-basicsplunktest-%d"
  destination = "splunk"

  s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    bucket_arn         = aws_s3_bucket.bucket.arn
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }

  splunk_configuration {
    hec_endpoint               = "https://input-test.com:443"
    hec_token                  = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"
    hec_acknowledgment_timeout = 600
    hec_endpoint_type          = "Event"
    s3_backup_mode             = "FailedEventsOnly"

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }

        parameters {
          parameter_name  = "RoleArn"
          parameter_value = aws_iam_role.firehose.arn
        }

        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = 1
        }

        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = 120
        }
      }
    }
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_HTTPEndpointBasic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-httpendpoint-%d"
  destination = "http_endpoint"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  http_endpoint_configuration {
    url      = "https://input-test.com:443"
    name     = "HTTP_test"
    role_arn = aws_iam_role.firehose.arn
  }
}
`

func testAccKinesisFirehoseDeliveryStreamConfig_HTTPEndpoint_RetryDuration(rInt, retryDuration int) string {
	return composeConfig(
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt),
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-httpendpoint-%[1]d"
  destination = "http_endpoint"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  http_endpoint_configuration {
    url            = "https://input-test.com:443"
    name           = "HTTP_test"
    retry_duration = %[2]d
    role_arn       = aws_iam_role.firehose.arn
  }
}
`, rInt, retryDuration))
}

var testAccKinesisFirehoseDeliveryStreamConfig_HTTPEndpointUpdates = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = "terraform-kinesis-firehose-httpendpoint-%d"
  destination = "http_endpoint"

  s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    bucket_arn         = aws_s3_bucket.bucket.arn
    buffer_size        = 10
    buffer_interval    = 400
    compression_format = "GZIP"
  }

  http_endpoint_configuration {
    url            = "https://input-test.com:443"
    name           = "HTTP_test"
    access_key     = "test_key"
    role_arn       = aws_iam_role.firehose.arn
    s3_backup_mode = "FailedDataOnly"

    processing_configuration {
      enabled = true

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }

        parameters {
          parameter_name  = "BufferSizeInMBs"
          parameter_value = 1
        }

        parameters {
          parameter_name  = "BufferIntervalInSeconds"
          parameter_value = 120
        }
      }
    }
  }
}
`

var testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "es-test-%d"

  cluster_config {
    instance_type = "m4.large.elasticsearch"
  }

  domain_endpoint_options {
    enforce_https       = true
    tls_security_policy = "Policy-Min-TLS-1-2-2019-07"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role_policy" "firehose-elasticsearch" {
  name   = "tf-acc-test-%d"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "es:*"
      ],
      "Resource": [
        "${aws_elasticsearch_domain.test_cluster.arn}",
        "${aws_elasticsearch_domain.test_cluster.arn}/*"
      ]
    }
  ]
}
EOF
}
`

// ElasticSearch associated with VPC
var testAccKinesisFirehoseDeliveryStreamBaseElasticsearchVpcConfig = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "elasticsearch_in_vpc" {
  cidr_block = "192.168.0.0/22"

  tags = {
    Name = "terraform-testacc-elasticsearch-domain-in-vpc"
  }
}

resource "aws_subnet" "first" {
  vpc_id            = aws_vpc.elasticsearch_in_vpc.id
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "192.168.0.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-first"
  }
}

resource "aws_subnet" "second" {
  vpc_id            = aws_vpc.elasticsearch_in_vpc.id
  availability_zone = data.aws_availability_zones.available.names[1]
  cidr_block        = "192.168.1.0/24"

  tags = {
    Name = "tf-acc-elasticsearch-domain-in-vpc-second"
  }
}

resource "aws_security_group" "first" {
  vpc_id = aws_vpc.elasticsearch_in_vpc.id
}

resource "aws_security_group" "second" {
  vpc_id = aws_vpc.elasticsearch_in_vpc.id
}

resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "es-test-%d"

  cluster_config {
    instance_count         = 2
    zone_awareness_enabled = true
    instance_type          = "t2.small.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  vpc_options {
    security_group_ids = [aws_security_group.first.id, aws_security_group.second.id]
    subnet_ids         = [aws_subnet.first.id, aws_subnet.second.id]
  }
}

resource "aws_iam_role_policy" "firehose-elasticsearch" {
  name   = "elasticsearch"
  role   = aws_iam_role.firehose.id
  policy = <<EOF
{
	"Version":"2012-10-17",
	"Statement":[
	   {
		  "Effect":"Allow",
		  "Action":[
			 "es:*"
		  ],
		  "Resource":[
			"${aws_elasticsearch_domain.test_cluster.arn}",
			"${aws_elasticsearch_domain.test_cluster.arn}/*"
		  ]
	   },
	   {
		  "Effect":"Allow",
		  "Action":[
			 "ec2:Describe*",
			 "ec2:CreateNetworkInterface",
			 "ec2:CreateNetworkInterfacePermission",
			 "ec2:DeleteNetworkInterface"
		  ],
		  "Resource":[
			 "*"
		  ]
	   }
	]
}
EOF
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchBasic = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"
  }
}
`

func testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchVpcBasic(ri int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseElasticsearchVpcConfig+`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  elasticsearch_configuration {
    domain_arn = aws_elasticsearch_domain.test_cluster.arn
    role_arn   = aws_iam_role.firehose.arn
    index_name = "test"
    type_name  = "test"

    vpc_config {
      subnet_ids         = [aws_subnet.first.id, aws_subnet.second.id]
      security_group_ids = [aws_security_group.first.id, aws_security_group.second.id]
      role_arn           = aws_iam_role.firehose.arn
    }
  }
}
`, ri, ri, ri, ri, ri)
}

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchUpdate = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  elasticsearch_configuration {
    domain_arn         = aws_elasticsearch_domain.test_cluster.arn
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    type_name          = "test"
    buffering_interval = 500

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
  }
}
`

func testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchVpcUpdate(funcName, policyName, roleName string, ri int) string {
	return composeConfig(
		testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName),
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseElasticsearchVpcConfig+`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  elasticsearch_configuration {
    domain_arn         = aws_elasticsearch_domain.test_cluster.arn
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    type_name          = "test"
    buffering_interval = 500

    vpc_config {
      subnet_ids         = [aws_subnet.first.id, aws_subnet.second.id]
      security_group_ids = [aws_security_group.first.id, aws_security_group.second.id]
      role_arn           = aws_iam_role.firehose.arn
    }

    processing_configuration {
      enabled = false
      processors {
        type = "Lambda"
        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
  }
}`, ri, ri, ri, ri, ri))
}

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchEndpoint = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  elasticsearch_configuration {
    cluster_endpoint = "https://${aws_elasticsearch_domain.test_cluster.endpoint}"
    role_arn         = aws_iam_role.firehose.arn
    index_name       = "test"
    type_name        = "test"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchEndpointUpdate = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = [aws_iam_role_policy.firehose-elasticsearch]

  name        = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"

  s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }

  elasticsearch_configuration {
    cluster_endpoint   = "https://${aws_elasticsearch_domain.test_cluster.endpoint}"
    role_arn           = aws_iam_role.firehose.arn
    index_name         = "test"
    type_name          = "test"
    buffering_interval = 500

    processing_configuration {
      enabled = false

      processors {
        type = "Lambda"

        parameters {
          parameter_name  = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
  }
}`

func testAccKinesisFirehoseDeliveryStreamConfig_missingProcessingConfiguration(rInt int) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = "tf_acctest_firehose_delivery_role_%[1]d"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  name = "tf_acctest_firehose_delivery_policy_%[1]d"
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.bucket.arn}",
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Effect": "Allow",
      "Action": [
        "logs:putLogEvents"
      ],
      "Resource": [
        "arn:${data.aws_partition.current.partition}:logs::log-group:/aws/kinesisfirehose/*"
      ]
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = "tf-test-bucket-%[1]d"
  acl    = "private"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = "terraform-kinesis-firehose-mpc-%[1]d"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn           = aws_iam_role.firehose.arn
    prefix             = "tracking/autocomplete_stream/"
    buffer_interval    = 300
    buffer_size        = 5
    compression_format = "GZIP"
    bucket_arn         = aws_s3_bucket.bucket.arn
  }
}
`, rInt)
}
