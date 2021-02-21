package aws

import (
	"context"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/networkfirewall"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/networkfirewall/finder"
)

func init() {
	resource.AddTestSweepers("aws_networkfirewall_logging_configuration", &resource.Sweeper{
		Name: "aws_networkfirewall_logging_configuration",
		F:    testSweepNetworkFirewallLoggingConfigurations,
	})
}

func testSweepNetworkFirewallLoggingConfigurations(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).networkfirewallconn
	ctx := context.TODO()
	input := &networkfirewall.ListFirewallsInput{MaxResults: aws.Int64(100)}
	var sweeperErrs *multierror.Error

	for {
		resp, err := conn.ListFirewallsWithContext(ctx, input)
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping NetworkFirewall Logging Configuration sweep for %s: %s", region, err)
			return nil
		}
		if err != nil {
			return fmt.Errorf("error retrieving NetworkFirewall firewalls: %s", err)
		}

		for _, f := range resp.Firewalls {
			if f == nil {
				continue
			}

			arn := aws.StringValue(f.FirewallArn)

			log.Printf("[INFO] Deleting NetworkFirewall Logging Configuration for firewall: %s", arn)

			r := resourceAwsNetworkFirewallLoggingConfiguration()
			d := r.Data(nil)
			d.SetId(arn)
			diags := r.DeleteContext(ctx, d, client)
			for i := range diags {
				if diags[i].Severity == diag.Error {
					log.Printf("[ERROR] %s", diags[i].Summary)
					sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf(diags[i].Summary))
					continue
				}
			}
		}

		if aws.StringValue(resp.NextToken) == "" {
			break
		}
		input.NextToken = resp.NextToken
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAwsNetworkFirewallLoggingConfiguration_cloudwatchLogDestination_logGroup(t *testing.T) {
	logGroupName := acctest.RandomWithPrefix("tf-acc-test")
	updatedLogGroupName := fmt.Sprintf("%s-updated", logGroupName)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_cloudwatch(logGroupName, rName, networkfirewall.LogDestinationTypeCloudWatchLogs, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        "1",
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     networkfirewall.LogDestinationTypeCloudWatchLogs,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_cloudwatch(updatedLogGroupName, rName, networkfirewall.LogDestinationTypeCloudWatchLogs, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        "1",
						"log_destination.logGroup": updatedLogGroupName,
						"log_destination_type":     networkfirewall.LogDestinationTypeCloudWatchLogs,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_cloudwatchLogDestination_logType(t *testing.T) {
	logGroupName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_cloudwatch(logGroupName, rName, networkfirewall.LogDestinationTypeCloudWatchLogs, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": networkfirewall.LogTypeFlow,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_cloudwatch(logGroupName, rName, networkfirewall.LogDestinationTypeCloudWatchLogs, networkfirewall.LogTypeAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": networkfirewall.LogTypeAlert,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_kinesisLogDestination_deliveryStream(t *testing.T) {
	streamName := acctest.RandomWithPrefix("tf-acc-test")
	updatedStreamName := fmt.Sprintf("%s-updated", streamName)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_kinesis(streamName, rName, networkfirewall.LogDestinationTypeKinesisDataFirehose, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              "1",
						"log_destination.deliveryStream": streamName,
						"log_destination_type":           networkfirewall.LogDestinationTypeKinesisDataFirehose,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_kinesis(updatedStreamName, rName, networkfirewall.LogDestinationTypeKinesisDataFirehose, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              "1",
						"log_destination.deliveryStream": updatedStreamName,
						"log_destination_type":           networkfirewall.LogDestinationTypeKinesisDataFirehose,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_kinesisLogDestination_logType(t *testing.T) {
	streamName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_kinesis(streamName, rName, networkfirewall.LogDestinationTypeKinesisDataFirehose, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": networkfirewall.LogTypeFlow,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_kinesis(streamName, rName, networkfirewall.LogDestinationTypeKinesisDataFirehose, networkfirewall.LogTypeAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": networkfirewall.LogTypeAlert,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_s3LogDestination_bucketName(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	updatedBucketName := fmt.Sprintf("%s-updated", bucketName)
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{

			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "1",
						"log_destination.bucketName": bucketName,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(updatedBucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "1",
						"log_destination.bucketName": updatedBucketName,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_s3LogDestination_logType(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{

			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": networkfirewall.LogTypeFlow,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_type": networkfirewall.LogTypeAlert,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_s3LogDestination_prefix(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{

			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "1",
						"log_destination.bucketName": bucketName,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3_updatePrefix(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "2",
						"log_destination.bucketName": bucketName,
						"log_destination.prefix":     "update-prefix",
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_updateFirewallArn(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"
	firewallResourceName := "aws_networkfirewall_firewall.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_arn", firewallResourceName, "arn"),
				),
			},
			{
				// ForceNew Firewall i.e. LoggingConfiguration Resource
				Config: testAccNetworkFirewallLoggingConfiguration_s3_updateFirewallArn(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "firewall_arn", firewallResourceName, "arn"),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_updateLogDestinationType(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	logGroupName := acctest.RandomWithPrefix("tf-acc-test")
	streamName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_cloudwatch(logGroupName, rName, networkfirewall.LogDestinationTypeCloudWatchLogs, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        "1",
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     networkfirewall.LogDestinationTypeCloudWatchLogs,
						"log_type":                 networkfirewall.LogTypeFlow,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_kinesis(streamName, rName, networkfirewall.LogDestinationTypeKinesisDataFirehose, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              "1",
						"log_destination.deliveryStream": streamName,
						"log_destination_type":           networkfirewall.LogDestinationTypeKinesisDataFirehose,
						"log_type":                       networkfirewall.LogTypeFlow,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination_type": networkfirewall.LogDestinationTypeS3,
						"log_type":             networkfirewall.LogTypeFlow,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_updateToMultipleLogDestinationConfigs(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	streamName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3AndKinesis(bucketName, streamName, rName, networkfirewall.LogTypeAlert, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":              "1",
						"log_destination.deliveryStream": streamName,
						"log_destination_type":           networkfirewall.LogDestinationTypeKinesisDataFirehose,
						"log_type":                       networkfirewall.LogTypeFlow,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "1",
						"log_destination.bucketName": bucketName,
						"log_destination_type":       networkfirewall.LogDestinationTypeS3,
						"log_type":                   networkfirewall.LogTypeAlert,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", "1"),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_updateToSingleAlertTypeLogDestinationConfig(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	logGroupName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3AndCloudWatch(bucketName, logGroupName, rName, networkfirewall.LogTypeAlert, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        "1",
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     networkfirewall.LogDestinationTypeCloudWatchLogs,
						"log_type":                 networkfirewall.LogTypeFlow,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "1",
						"log_destination.bucketName": bucketName,
						"log_destination_type":       networkfirewall.LogDestinationTypeS3,
						"log_type":                   networkfirewall.LogTypeAlert,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeAlert),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "1",
						"log_destination.bucketName": bucketName,
						"log_destination_type":       networkfirewall.LogDestinationTypeS3,
						"log_type":                   networkfirewall.LogTypeAlert,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_updateToSingleFlowTypeLogDestinationConfig(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	logGroupName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3AndCloudWatch(bucketName, logGroupName, rName, networkfirewall.LogTypeAlert, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        "1",
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     networkfirewall.LogDestinationTypeCloudWatchLogs,
						"log_type":                 networkfirewall.LogTypeFlow,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":          "1",
						"log_destination.bucketName": bucketName,
						"log_destination_type":       networkfirewall.LogDestinationTypeS3,
						"log_type":                   networkfirewall.LogTypeAlert,
					}),
				),
			},
			{
				Config: testAccNetworkFirewallLoggingConfiguration_cloudwatch(logGroupName, rName, networkfirewall.LogDestinationTypeCloudWatchLogs, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging_configuration.0.log_destination_config.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "logging_configuration.0.log_destination_config.*", map[string]string{
						"log_destination.%":        "1",
						"log_destination.logGroup": logGroupName,
						"log_destination_type":     networkfirewall.LogDestinationTypeCloudWatchLogs,
						"log_type":                 networkfirewall.LogTypeFlow,
					}),
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

func TestAccAwsNetworkFirewallLoggingConfiguration_disappears(t *testing.T) {
	bucketName := acctest.RandomWithPrefix("tf-acc-test")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_networkfirewall_logging_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAwsNetworkFirewall(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, networkfirewall.LogDestinationTypeS3, networkfirewall.LogTypeFlow),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsNetworkFirewallLoggingConfigurationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsNetworkFirewallLoggingConfiguration(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsNetworkFirewallLoggingConfigurationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_networkfirewall_logging_configuration" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.LoggingConfiguration(context.Background(), conn, rs.Primary.ID)
		if tfawserr.ErrCodeEquals(err, networkfirewall.ErrCodeResourceNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}
		if output != nil {
			return fmt.Errorf("NetworkFirewall Logging Configuration for firewall (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsNetworkFirewallLoggingConfigurationExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No NetworkFirewall Logging Configuration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).networkfirewallconn
		output, err := finder.LoggingConfiguration(context.Background(), conn, rs.Primary.ID)
		if err != nil {
			return err
		}
		if output == nil && output.LoggingConfiguration == nil {
			return fmt.Errorf("NetworkFirewall Logging Configuration for firewall (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccNetworkFirewallLoggingConfigurationBaseConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = %[1]q
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName)
}

func testAccNetworkFirewallLoggingConfigurationBaseConfig_updateFirewall(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "test" {
  cidr_block = "192.168.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_networkfirewall_firewall_policy" "test" {
  name = %[1]q
  firewall_policy {
    stateless_fragment_default_actions = ["aws:drop"]
    stateless_default_actions          = ["aws:pass"]
  }
}

resource "aws_networkfirewall_firewall" "test" {
  name                = "%[1]s-updated"
  firewall_policy_arn = aws_networkfirewall_firewall_policy.test.arn
  vpc_id              = aws_vpc.test.id

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}
`, rName)
}

func testAccNetworkFirewallLoggingConfigurationS3BucketDependencyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %q
  acl           = "private"
  force_destroy = true

  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccNetworkFirewallLoggingConfigurationCloudWatchDependencyConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %q

  lifecycle {
    create_before_destroy = true
  }
}
`, rName)
}

func testAccNetworkFirewallLoggingConfiguration_kinesisDependenciesConfig(rName, streamName string) string {
	return fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.${data.aws_partition.current.dns_suffix}"
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

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = aws_iam_role.test.id

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
        "${aws_s3_bucket.logs.arn}",
        "${aws_s3_bucket.logs.arn}/*"
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

resource "aws_s3_bucket" "logs" {
  bucket        = %[1]q
  acl           = "private"
  force_destroy = true
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.test]
  name        = %[2]q
  destination = "s3"

  s3_configuration {
    role_arn   = aws_iam_role.test.arn
    bucket_arn = aws_s3_bucket.logs.arn
  }

  tags = {
    LogDeliveryEnabled = "placeholder"
  }

  lifecycle {
    create_before_destroy = true
    ignore_changes = [
      # Ignore changes to LogDeliveryEnabled tag as API adds this tag when broker log delivery is enabled
      tags["LogDeliveryEnabled"],
    ]
  }
}
`, rName, streamName)
}

func testAccNetworkFirewallLoggingConfiguration_s3(bucketName, rName, destinationType, logType string) string {
	return composeConfig(
		testAccNetworkFirewallLoggingConfigurationBaseConfig(rName),
		testAccNetworkFirewallLoggingConfigurationS3BucketDependencyConfig(bucketName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccNetworkFirewallLoggingConfiguration_s3_updateFirewallArn(bucketName, rName, destinationType, logType string) string {
	return composeConfig(
		testAccNetworkFirewallLoggingConfigurationBaseConfig_updateFirewall(rName),
		testAccNetworkFirewallLoggingConfigurationS3BucketDependencyConfig(bucketName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccNetworkFirewallLoggingConfiguration_s3_updatePrefix(bucketName, rName, destinationType, logType string) string {
	return composeConfig(
		testAccNetworkFirewallLoggingConfigurationBaseConfig(rName),
		testAccNetworkFirewallLoggingConfigurationS3BucketDependencyConfig(bucketName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
        prefix     = "update-prefix"
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccNetworkFirewallLoggingConfiguration_kinesis(streamName, rName, destinationType, logType string) string {
	return composeConfig(
		testAccNetworkFirewallLoggingConfigurationBaseConfig(rName),
		testAccNetworkFirewallLoggingConfiguration_kinesisDependenciesConfig(rName, streamName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        deliveryStream = aws_kinesis_firehose_delivery_stream.test.name
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccNetworkFirewallLoggingConfiguration_cloudwatch(logGroupName, rName, destinationType, logType string) string {
	return composeConfig(
		testAccNetworkFirewallLoggingConfigurationBaseConfig(rName),
		testAccNetworkFirewallLoggingConfigurationCloudWatchDependencyConfig(logGroupName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        logGroup = aws_cloudwatch_log_group.test.name
      }
      log_destination_type = %[2]q
      log_type             = %[3]q
    }
  }
}
`, rName, destinationType, logType))
}

func testAccNetworkFirewallLoggingConfiguration_s3AndKinesis(bucketName, streamName, rName, logTypeS3, logTypeKinesis string) string {
	return composeConfig(
		testAccNetworkFirewallLoggingConfigurationS3BucketDependencyConfig(bucketName),
		testAccNetworkFirewallLoggingConfiguration_kinesisDependenciesConfig(rName, streamName),
		testAccNetworkFirewallLoggingConfigurationBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = "S3"
      log_type             = %[2]q
    }

    log_destination_config {
      log_destination = {
        deliveryStream = aws_kinesis_firehose_delivery_stream.test.name
      }
      log_destination_type = "KinesisDataFirehose"
      log_type             = %[3]q
    }
  }
}
`, rName, logTypeS3, logTypeKinesis))
}

func testAccNetworkFirewallLoggingConfiguration_s3AndCloudWatch(bucketName, logGroupName, rName, logTypeS3, logTypeCloudWatch string) string {
	return composeConfig(
		testAccNetworkFirewallLoggingConfigurationBaseConfig(rName),
		testAccNetworkFirewallLoggingConfigurationS3BucketDependencyConfig(bucketName),
		testAccNetworkFirewallLoggingConfigurationCloudWatchDependencyConfig(logGroupName),
		fmt.Sprintf(`
resource "aws_networkfirewall_logging_configuration" "test" {
  firewall_arn = aws_networkfirewall_firewall.test.arn

  logging_configuration {
    log_destination_config {
      log_destination = {
        bucketName = aws_s3_bucket.test.bucket
      }
      log_destination_type = "S3"
      log_type             = %[2]q
    }

    log_destination_config {
      log_destination = {
        logGroup = aws_cloudwatch_log_group.test.name
      }
      log_destination_type = "CloudWatchLogs"
      log_type             = %[3]q
    }
  }
}
`, rName, logTypeS3, logTypeCloudWatch))
}
