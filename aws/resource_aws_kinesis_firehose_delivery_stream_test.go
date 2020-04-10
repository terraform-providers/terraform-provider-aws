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
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
				ImportStateId: "arn:aws:firehose:us-east-1:123456789012:missing-slash",
				ExpectError:   regexp.MustCompile(`Expected ID in format`),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName, rInt, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "server_side_encryption.0.enabled", "true"),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "original"),
				),
			},
			{
				Config: testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithTagsChanged(rName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Usage", "changed"),
				),
			},
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
				),
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, updatedS3DestinationConfig, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
					resource.TestMatchResourceAttr(resourceName, "extended_s3_configuration.0.kms_key_arn", regexp.MustCompile(`^arn:[^:]+:kms:[^:]+:[^:]+:key/.+$`)),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, firstUpdateExtendedS3DestinationConfig, nil, nil, nil),
				),
			},
			{
				Config: removeProcessorsConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, removeProcessorsExtendedS3DestinationConfig, nil, nil, nil),
				),
			},
		},
	})
}

func TestAccAWSKinesisFirehoseDeliveryStream_RedshiftConfigUpdates(t *testing.T) {
	var stream firehose.DeliveryStreamDescription

	resourceName := "aws_kinesis_firehose_delivery_stream.test"
	ri := acctest.RandInt()
	rString := acctest.RandString(8)
	funcName := fmt.Sprintf("aws_kinesis_firehose_delivery_stream_test_%s", rString)
	policyName := fmt.Sprintf("tf_acc_policy_%s", rString)
	roleName := fmt.Sprintf("tf_acc_role_%s", rString)
	preConfig := fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_RedshiftBasic,
		ri, ri, ri, ri, ri)
	postConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_RedshiftUpdates,
			ri, ri, ri, ri, ri)

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
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"redshift_configuration.0.password"},
			},
			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckKinesisFirehoseDeliveryStreamExists(resourceName, &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, updatedRedshiftConfig, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, updatedSplunkConfig),
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
		ri, ri, ri, ri, ri)
	postConfig := testAccFirehoseAWSLambdaConfigBasic(funcName, policyName, roleName) +
		fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchUpdate,
			ri, ri, ri, ri, ri)

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
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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
					testAccCheckKinesisFirehoseDeliveryStreamExists("aws_kinesis_firehose_delivery_stream.test", &stream),
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, updatedElasticSearchConfig, nil),
				),
			},
		},
	})
}

// Regression test for https://github.com/terraform-providers/terraform-provider-aws/issues/1657
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
					testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(&stream, nil, nil, nil, nil, nil),
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

func testAccCheckAWSKinesisFirehoseDeliveryStreamAttributes(stream *firehose.DeliveryStreamDescription, s3config interface{}, extendedS3config interface{}, redshiftConfig interface{}, elasticsearchConfig interface{}, splunkConfig interface{}) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !strings.HasPrefix(*stream.DeliveryStreamName, "terraform-kinesis-firehose") {
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
  role = "${aws_iam_role.iam_for_lambda.id}"

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
    filename = "test-fixtures/lambdatest.zip"
    function_name = "%s"
    role = "${aws_iam_role.iam_for_lambda.arn}"
    handler = "exports.example"
    runtime = "nodejs12.x"
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
  acl = "private"
}

resource "aws_iam_role_policy" "firehose" {
  name = "tf_acctest_firehose_delivery_policy_%d"
  role = "${aws_iam_role.firehose.id}"
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
        "glue:GetTableVersions"
      ],
      "Resource": [
        "*"
      ]
    }
  ]
}
EOF
}

`

const testAccFirehoseKinesisStreamSource = `
resource "aws_kinesis_stream" "source" {
  name = "terraform-kinesis-source-stream-basictest-%d"
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
  role = "${aws_iam_role.kinesis_source.id}"
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

resource "aws_iam_role_policy" "firehose" {
  name = "tf_acctest_firehose_delivery_policy_%d"
  role = "${aws_iam_role.firehose.id}"

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
  bucket = "tf-test-bucket-%d"
  acl    = "private"
}

resource "aws_cloudwatch_log_group" "test" {
  name = "example-%d"
}

resource "aws_cloudwatch_log_stream" "test" {
  name           = "sample-log-stream-test-%d"
  log_group_name = "${aws_cloudwatch_log_group.test.name}"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = ["aws_iam_role_policy.firehose"]
  name        = "terraform-kinesis-firehose-cloudwatch-%d"
  destination = "s3"

  s3_configuration {
    role_arn   = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"

    cloudwatch_logging_options {
      enabled         = true
      log_group_name  = "${aws_cloudwatch_log_group.test.name}"
      log_stream_name = "${aws_cloudwatch_log_stream.test.name}"
    }
  }
}
`, rInt, rInt, rInt, rInt, rInt, rInt)
}

var testAccKinesisFirehoseDeliveryStreamConfig_s3basic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
}`

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithSSE(rName string, rInt int, sseEnabled bool) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "%s"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }

  server_side_encryption {
    enabled = %t
  }
}
`, rName, sseEnabled)
}

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithTags(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
	resource "aws_kinesis_firehose_delivery_stream" "test" {
		depends_on = ["aws_iam_role_policy.firehose"]
		name = "%s"
		destination = "s3"
		s3_configuration {
			role_arn = "${aws_iam_role.firehose.arn}"
			bucket_arn = "${aws_s3_bucket.bucket.arn}"
		}
	tags = {
			Environment = "production"
			Usage = "original"
		}
	}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_s3basicWithTagsChanged(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) +
		fmt.Sprintf(`
	resource "aws_kinesis_firehose_delivery_stream" "test" {
		depends_on = ["aws_iam_role_policy.firehose"]
		name = "%s"
		destination = "s3"
		s3_configuration {
			role_arn = "${aws_iam_role.firehose.arn}"
			bucket_arn = "${aws_s3_bucket.bucket.arn}"
		}
	tags = {
			Usage = "changed"
		}
	}
`, rName)
}

var testAccKinesisFirehoseDeliveryStreamConfig_s3KinesisStreamSource = testAccKinesisFirehoseDeliveryStreamBaseConfig + testAccFirehoseKinesisStreamSource + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose", "aws_iam_role_policy.kinesis_source"]
  name = "terraform-kinesis-firehose-basictest-%d"
  kinesis_source_configuration {
    kinesis_stream_arn = "${aws_kinesis_stream.source.arn}"
    role_arn = "${aws_iam_role.kinesis_source.arn}"
  }
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_s3Updates = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-s3test-%d"
  destination = "s3"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3basic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    processing_configuration {
      enabled = true
      processors {
        type = "Lambda"
        parameters {
          parameter_name = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
    s3_backup_mode = "Disabled"
  }
}
`

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_Enabled(rName string, rInt int, enabled bool) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "%s"
}

resource "aws_glue_catalog_table" "test" {
  database_name = "${aws_glue_catalog_database.test.name}"
  name          = "%s"

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "%s"

  extended_s3_configuration {
    bucket_arn  = "${aws_s3_bucket.bucket.arn}"
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = "${aws_iam_role.firehose.arn}"

    data_format_conversion_configuration {
      enabled = %t

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
        database_name = "${aws_glue_catalog_table.test.database_name}"
        role_arn      = "${aws_iam_role.firehose.arn}"
        table_name    = "${aws_glue_catalog_table.test.name}"
      }
    }
  }

  depends_on = ["aws_iam_role_policy.firehose"]
}
`, rName, rName, rName, enabled)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_HiveJsonSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "%s"
}

resource "aws_glue_catalog_table" "test" {
  database_name = "${aws_glue_catalog_database.test.name}"
  name          = "%s"

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "%s"

  extended_s3_configuration {
    bucket_arn  = "${aws_s3_bucket.bucket.arn}"
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = "${aws_iam_role.firehose.arn}"

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
        database_name = "${aws_glue_catalog_table.test.database_name}"
        role_arn      = "${aws_iam_role.firehose.arn}"
        table_name    = "${aws_glue_catalog_table.test.name}"
      }
    }
  }

  depends_on = ["aws_iam_role_policy.firehose"]
}
`, rName, rName, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ExternalUpdate(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "%s"

  extended_s3_configuration {
    bucket_arn  = "${aws_s3_bucket.bucket.arn}"
    role_arn    = "${aws_iam_role.firehose.arn}"
  }
}
`, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OpenXJsonSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "%s"
}

resource "aws_glue_catalog_table" "test" {
  database_name = "${aws_glue_catalog_database.test.name}"
  name          = "%s"

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "%s"

  extended_s3_configuration {
    bucket_arn  = "${aws_s3_bucket.bucket.arn}"
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = "${aws_iam_role.firehose.arn}"

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
        database_name = "${aws_glue_catalog_table.test.database_name}"
        role_arn      = "${aws_iam_role.firehose.arn}"
        table_name    = "${aws_glue_catalog_table.test.name}"
      }
    }
  }

  depends_on = ["aws_iam_role_policy.firehose"]
}
`, rName, rName, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_OrcSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "%s"
}

resource "aws_glue_catalog_table" "test" {
  database_name = "${aws_glue_catalog_database.test.name}"
  name          = "%s"

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "%s"

  extended_s3_configuration {
    bucket_arn  = "${aws_s3_bucket.bucket.arn}"
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = "${aws_iam_role.firehose.arn}"

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
        database_name = "${aws_glue_catalog_table.test.database_name}"
        role_arn      = "${aws_iam_role.firehose.arn}"
        table_name    = "${aws_glue_catalog_table.test.name}"
      }
    }
  }

  depends_on = ["aws_iam_role_policy.firehose"]
}
`, rName, rName, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_DataFormatConversionConfiguration_ParquetSerDe_Empty(rName string, rInt int) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = "%s"
}

resource "aws_glue_catalog_table" "test" {
  database_name = "${aws_glue_catalog_database.test.name}"
  name          = "%s"

  storage_descriptor {
    columns {
      name = "test"
      type = "string"
    }
  }
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = "%s"

  extended_s3_configuration {
    bucket_arn  = "${aws_s3_bucket.bucket.arn}"
    # InvalidArgumentException: BufferingHints.SizeInMBs must be at least 64 when data format conversion is enabled.
    buffer_size = 128
    role_arn    = "${aws_iam_role.firehose.arn}"

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
        database_name = "${aws_glue_catalog_table.test.database_name}"
        role_arn      = "${aws_iam_role.firehose.arn}"
        table_name    = "${aws_glue_catalog_table.test.name}"
      }
    }
  }

  depends_on = ["aws_iam_role_policy.firehose"]
}
`, rName, rName, rName)
}

func testAccKinesisFirehoseDeliveryStreamConfig_ExtendedS3_ErrorOutputPrefix(rName string, rInt int, errorOutputPrefix string) string {
	return fmt.Sprintf(testAccKinesisFirehoseDeliveryStreamBaseConfig, rInt, rInt, rInt) + fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  destination = "extended_s3"
  name        = %q

  extended_s3_configuration {
    bucket_arn          = "${aws_s3_bucket.bucket.arn}"
    error_output_prefix = %q
    role_arn            = "${aws_iam_role.firehose.arn}"
  }

  depends_on = ["aws_iam_role_policy.firehose"]
}
`, rName, errorOutputPrefix)
}

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3KmsKeyArn = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kms_key" "test" {
  description = "Terraform acc test %s"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    kms_key_arn = "${aws_kms_key.test.arn}"
    processing_configuration {
      enabled = false
      processors {
        type = "Lambda"
        parameters {
          parameter_name = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates_Initial = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    processing_configuration {
      enabled = true
      processors {
        type = "Lambda"
        parameters {
          parameter_name = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
    s3_backup_mode = "Enabled"
    s3_backup_configuration {
      role_arn = "${aws_iam_role.firehose.arn}"
      bucket_arn = "${aws_s3_bucket.bucket.arn}"
    }
  }
}
`

var testAccKinesisFirehoseDeliveryStreamConfig_extendedS3Updates_RemoveProcessors = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basictest-%d"
  destination = "extended_s3"
  extended_s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
    s3_backup_mode = "Enabled"
    s3_backup_configuration {
      role_arn = "${aws_iam_role.firehose.arn}"
      bucket_arn = "${aws_s3_bucket.bucket.arn}"
    }
  }
}
`

var testAccKinesisFirehoseDeliveryStreamBaseRedshiftConfig = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_redshift_cluster" "test_cluster" {
  cluster_identifier = "tf-redshift-cluster-%d"
  database_name = "test"
  master_username = "testuser"
  master_password = "T3stPass"
  node_type = "dc1.large"
  cluster_type = "single-node"
	skip_final_snapshot = true
}`

var testAccKinesisFirehoseDeliveryStreamConfig_RedshiftBasic = testAccKinesisFirehoseDeliveryStreamBaseRedshiftConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose", "aws_redshift_cluster.test_cluster"]
  name = "terraform-kinesis-firehose-basicredshifttest-%d"
  destination = "redshift"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
  redshift_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test_cluster.endpoint}/${aws_redshift_cluster.test_cluster.database_name}"
    username = "testuser"
    password = "T3stPass"
    data_table_name = "test-table"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_RedshiftUpdates = testAccKinesisFirehoseDeliveryStreamBaseRedshiftConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose", "aws_redshift_cluster.test_cluster"]
  name = "terraform-kinesis-firehose-basicredshifttest-%d"
  destination = "redshift"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
  }
  redshift_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    cluster_jdbcurl = "jdbc:redshift://${aws_redshift_cluster.test_cluster.endpoint}/${aws_redshift_cluster.test_cluster.database_name}"
    username = "testuser"
    password = "T3stPass"
    s3_backup_mode = "Enabled"
    s3_backup_configuration {
      role_arn = "${aws_iam_role.firehose.arn}"
      bucket_arn = "${aws_s3_bucket.bucket.arn}"
    }
    data_table_name = "test-table"
    copy_options = "GZIP"
    data_table_columns = "test-col"
    processing_configuration {
      enabled = false
      processors {
        type = "Lambda"
        parameters {
          parameter_name = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
      }
    }
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_SplunkBasic = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basicsplunktest-%d"
  destination = "splunk"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
  splunk_configuration {
    hec_endpoint = "https://input-test.com:443"
    hec_token = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_SplunkUpdates = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose"]
  name = "terraform-kinesis-firehose-basicsplunktest-%d"
  destination = "splunk"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
    buffer_size = 10
    buffer_interval = 400
    compression_format = "GZIP"
  }
  splunk_configuration {
    hec_endpoint = "https://input-test.com:443"
    hec_token = "51D4DA16-C61B-4F5F-8EC7-ED4301342A4A"
    hec_acknowledgment_timeout = 600
    hec_endpoint_type = "Event"
    s3_backup_mode = "FailedEventsOnly"
    processing_configuration {
      enabled = true
      processors {
        type = "Lambda"

        parameters {
          parameter_name = "LambdaArn"
          parameter_value = "${aws_lambda_function.lambda_function_test.arn}:$LATEST"
        }
        parameters {
          parameter_name = "RoleArn"
          parameter_value = "${aws_iam_role.firehose.arn}"
        }
        parameters {
          parameter_name = "BufferSizeInMBs"
          parameter_value = 1
        }
        parameters {
          parameter_name = "BufferIntervalInSeconds"
          parameter_value = 120
        }
      }
    }
  }
}`

var testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig = testAccKinesisFirehoseDeliveryStreamBaseConfig + `
resource "aws_elasticsearch_domain" "test_cluster" {
  domain_name = "es-test-%d"

  cluster_config {
    instance_type = "m4.large.elasticsearch"
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }
}

resource "aws_iam_role_policy" "firehose-elasticsearch" {
  name   = "elasticsearch"
  role   = "${aws_iam_role.firehose.id}"
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

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchBasic = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose-elasticsearch"]

  name = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
  elasticsearch_configuration {
    domain_arn = "${aws_elasticsearch_domain.test_cluster.arn}"
    role_arn = "${aws_iam_role.firehose.arn}"
    index_name = "test"
    type_name = "test"
  }
}`

var testAccKinesisFirehoseDeliveryStreamConfig_ElasticsearchUpdate = testAccKinesisFirehoseDeliveryStreamBaseElasticsearchConfig + `
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on = ["aws_iam_role_policy.firehose-elasticsearch"]

  name = "terraform-kinesis-firehose-es-%d"
  destination = "elasticsearch"
  s3_configuration {
    role_arn = "${aws_iam_role.firehose.arn}"
    bucket_arn = "${aws_s3_bucket.bucket.arn}"
  }
  elasticsearch_configuration {
    domain_arn = "${aws_elasticsearch_domain.test_cluster.arn}"
    role_arn = "${aws_iam_role.firehose.arn}"
    index_name = "test"
    type_name = "test"
    buffering_interval = 500
    processing_configuration {
      enabled = false
      processors {
        type = "Lambda"
        parameters {
          parameter_name = "LambdaArn"
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
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "firehose" {
  name = "tf_acctest_firehose_delivery_policy_%d"
  role = "${aws_iam_role.firehose.id}"

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
  bucket = "tf-test-bucket-%d"
  acl    = "private"
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  name        = "terraform-kinesis-firehose-mpc-%d"
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn           = "${aws_iam_role.firehose.arn}"
    prefix             = "tracking/autocomplete_stream/"
    buffer_interval    = 300
    buffer_size        = 5
    compression_format = "GZIP"
    bucket_arn         = "${aws_s3_bucket.bucket.arn}"
  }
}
`, rInt, rInt, rInt, rInt)
}
