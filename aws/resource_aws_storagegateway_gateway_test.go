package aws

import (
	"fmt"
	"log"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/storagegateway"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_storagegateway_gateway", &resource.Sweeper{
		Name: "aws_storagegateway_gateway",
		F:    testSweepStorageGatewayGateways,
	})
}

func testSweepStorageGatewayGateways(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}
	conn := client.(*AWSClient).storagegatewayconn

	err = conn.ListGatewaysPages(&storagegateway.ListGatewaysInput{}, func(page *storagegateway.ListGatewaysOutput, isLast bool) bool {
		if len(page.Gateways) == 0 {
			log.Print("[DEBUG] No Storage Gateway Gateways to sweep")
			return true
		}

		for _, gateway := range page.Gateways {
			name := aws.StringValue(gateway.GatewayName)

			log.Printf("[INFO] Deleting Storage Gateway Gateway: %s", name)
			input := &storagegateway.DeleteGatewayInput{
				GatewayARN: gateway.GatewayARN,
			}

			_, err := conn.DeleteGateway(input)
			if err != nil {
				if isAWSErr(err, storagegateway.ErrorCodeGatewayNotFound, "") {
					continue
				}
				log.Printf("[ERROR] Failed to delete Storage Gateway Gateway (%s): %s", name, err)
			}
		}

		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping Storage Gateway Gateway sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving Storage Gateway Gateways: %w", err)
	}
	return nil
}

func TestAccAWSStorageGatewayGateway_GatewayType_Cached(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "CACHED"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_security_strategy", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "gateway_network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_network_interface.0.ipv4_address", "aws_instance.test", "private_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayType_FileS3(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "FILE_S3"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "gateway_network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_network_interface.0.ipv4_address", "aws_instance.test", "private_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayType_Stored(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Stored(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "STORED"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
					resource.TestCheckResourceAttr(resourceName, "gateway_network_interface.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_network_interface.0.ipv4_address", "aws_instance.test", "private_ip"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayType_Vtl(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttrSet(resourceName, "gateway_id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT"),
					resource.TestCheckResourceAttr(resourceName, "gateway_type", "VTL"),
					resource.TestCheckResourceAttr(resourceName, "medium_changer_type", ""),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", ""),
					resource.TestCheckResourceAttr(resourceName, "tape_drive_type", ""),
					resource.TestCheckResourceAttrPair(resourceName, "ec2_instance_id", "aws_instance.test", "id"),
					resource.TestCheckResourceAttr(resourceName, "endpoint_type", "STANDARD"),
					resource.TestCheckResourceAttr(resourceName, "host_environment", "EC2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_tags(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					testAccMatchResourceAttrRegionalARN(resourceName, "arn", "storagegateway", regexp.MustCompile(`gateway/sgw-.+`)),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayName(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	rName2 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName1),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_name", rName2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_CloudWatchLogs(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName1 := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"
	resourceName2 := "aws_cloudwatch_log_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_Log_Group(rName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "cloudwatch_log_group_arn", resourceName2, "arn"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayTimezone(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayTimezone(rName, "GMT-1:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT-1:00"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayTimezone(rName, "GMT-2:00"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "gateway_timezone", "GMT-2:00"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_GatewayVpcEndpoint(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayVpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_vpc_endpoint", vpcEndpointResourceName, "dns_entry.0.dns_name"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_SmbActiveDirectorySettings(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.domain_name", "terraformtesting.com"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.username", "Administrator"),
					resource.TestCheckResourceAttrSet(resourceName, "smb_active_directory_settings.0.active_directory_status"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_active_directory_settings"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_SmbActiveDirectorySettings_timeout(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettingsTimeout(rName, 50),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.domain_name", "terraformtesting.com"),
					resource.TestCheckResourceAttr(resourceName, "smb_active_directory_settings.0.timeout_in_seconds", "50"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_active_directory_settings"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_SmbGuestPassword(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, "myguestpassword1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", "myguestpassword1"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, "myguestpassword2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_guest_password", "myguestpassword2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address", "smb_guest_password"},
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_SMBSecurityStrategy(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfigSMBSecurityStrategy(rName, "ClientSpecified"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_security_strategy", "ClientSpecified"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfigSMBSecurityStrategy(rName, "MandatorySigning"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "smb_security_strategy", "MandatorySigning"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_disappears(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsStorageGatewayGateway(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_bandwidthUpload(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayBandwidthConfigUpload(rName, 102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "102400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccAWSStorageGatewayGatewayBandwidthConfigUpload(rName, 2*102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "204800"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "0"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_bandwidthDownload(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayBandwidthConfigDownload(rName, 102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "102400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccAWSStorageGatewayGatewayBandwidthConfigDownload(rName, 2*102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "204800"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "0"),
				),
			},
		},
	})
}

func TestAccAWSStorageGatewayGateway_bandwidthAll(t *testing.T) {
	var gateway storagegateway.DescribeGatewayInformationOutput
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_storagegateway_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSStorageGatewayGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSStorageGatewayGatewayBandwidthConfigAll(rName, 102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "102400"),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "102400"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"activation_key", "gateway_ip_address"},
			},
			{
				Config: testAccAWSStorageGatewayGatewayBandwidthConfigAll(rName, 2*102400),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "204800"),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "204800"),
				),
			},
			{
				Config: testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSStorageGatewayGatewayExists(resourceName, &gateway),
					resource.TestCheckResourceAttr(resourceName, "average_upload_rate_limit_in_bits_per_sec", "0"),
					resource.TestCheckResourceAttr(resourceName, "average_download_rate_limit_in_bits_per_sec", "0"),
				),
			},
		},
	})
}

func testAccCheckAWSStorageGatewayGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_storagegateway_gateway" {
			continue
		}

		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(rs.Primary.ID),
		}

		_, err := conn.DescribeGatewayInformation(input)

		if err != nil {
			if isAWSErrStorageGatewayGatewayNotFound(err) {
				return nil
			}
			return err
		}
	}

	return nil

}

func testAccCheckAWSStorageGatewayGatewayExists(resourceName string, gateway *storagegateway.DescribeGatewayInformationOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).storagegatewayconn
		input := &storagegateway.DescribeGatewayInformationInput{
			GatewayARN: aws.String(rs.Primary.ID),
		}

		output, err := conn.DescribeGatewayInformation(input)

		if err != nil {
			return err
		}

		if output == nil {
			return fmt.Errorf("Gateway %q does not exist", rs.Primary.ID)
		}

		*gateway = *output

		return nil
	}
}

// testAccAWSStorageGateway_VPCBase provides a publicly accessible subnet
// and security group, suitable for Storage Gateway EC2 instances of any type
func testAccAWSStorageGateway_VPCBase(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSStorageGateway_FileGatewayBase(rName string) string {
	return composeConfig(
		testAccAWSStorageGateway_VPCBase(rName),
		// Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "m5.xlarge", "m4.xlarge"),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-file.html
data "aws_ssm_parameter" "aws_service_storagegateway_ami_FILE_S3_latest" {
  name = "/aws/service/storagegateway/ami/FILE_S3/latest"
}

resource "aws_instance" "test" {
  depends_on = [aws_route.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_FILE_S3_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName string) string {
	return composeConfig(
		testAccAWSStorageGateway_VPCBase(rName),
		// Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test.availability_zone", "m5.xlarge", "m4.xlarge"),
		fmt.Sprintf(`
# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-common.html
# NOTE: CACHED, STORED, and VTL Gateway Types share the same AMI
data "aws_ssm_parameter" "aws_service_storagegateway_ami_CACHED_latest" {
  name = "/aws/service/storagegateway/ami/CACHED/latest"
}

resource "aws_instance" "test" {
  depends_on = [aws_route.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_CACHED_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_Cached(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_FileS3(rName string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_Log_Group(rName string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address       = aws_instance.test.public_ip
  gateway_name             = %[1]q
  gateway_timezone         = "GMT"
  gateway_type             = "FILE_S3"
  cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_Stored(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "STORED"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayType_Vtl(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "VTL"
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayTimezone(rName, gatewayTimezone string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = %q
  gateway_type       = "FILE_S3"
}
`, rName, gatewayTimezone)
}

func testAccAWSStorageGatewayGatewayConfig_GatewayVpcEndpoint(rName string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
data "aws_vpc_endpoint_service" "storagegateway" {
  service = "storagegateway"
}

resource "aws_vpc_endpoint" "test" {
  security_group_ids = [aws_security_group.test.id]
  service_name       = data.aws_vpc_endpoint_service.storagegateway.service_name
  subnet_ids         = [aws_subnet.test.id]
  vpc_endpoint_type  = data.aws_vpc_endpoint_service.storagegateway.service_type
  vpc_id             = aws_vpc.test.id
}

resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address   = aws_instance.test.public_ip
  gateway_name         = %[1]q
  gateway_timezone     = "GMT"
  gateway_type         = "CACHED"
  gateway_vpc_endpoint = aws_vpc_endpoint.test.dns_entry[0].dns_name
}
`, rName)
}

func testAccAWSStorageGatewayGatewayConfigSmbActiveDirectorySettingsBase(rName string) string {
	return composeConfig(
		// Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/Requirements.html
		testAccAvailableEc2InstanceTypeForAvailabilityZone("aws_subnet.test[0].availability_zone", "m5.xlarge", "m4.xlarge"),
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
# Directory Service Directories must be deployed across multiple EC2 Availability Zones
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = "10.0.${count.index}.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = "0.0.0.0/0"
  gateway_id             = aws_internet_gateway.test.id
  route_table_id         = aws_vpc.test.main_route_table_id
}

resource "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_directory_service_directory" "test" {
  name     = "terraformtesting.com"
  password = "SuperSecretPassw0rd"
  size     = "Small"

  vpc_settings {
    subnet_ids = aws_subnet.test[*].id
    vpc_id     = aws_vpc.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options" "test" {
  domain_name         = aws_directory_service_directory.test.name
  domain_name_servers = aws_directory_service_directory.test.dns_ip_addresses

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_dhcp_options_association" "test" {
  dhcp_options_id = aws_vpc_dhcp_options.test.id
  vpc_id          = aws_vpc.test.id
}

# Reference: https://docs.aws.amazon.com/storagegateway/latest/userguide/ec2-gateway-file.html
data "aws_ssm_parameter" "aws_service_storagegateway_ami_FILE_S3_latest" {
  name = "/aws/service/storagegateway/ami/FILE_S3/latest"
}

resource "aws_instance" "test" {
  depends_on = [aws_route.test, aws_vpc_dhcp_options_association.test]

  ami                         = data.aws_ssm_parameter.aws_service_storagegateway_ami_FILE_S3_latest.value
  associate_public_ip_address = true
  instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
  vpc_security_group_ids      = [aws_security_group.test.id]
  subnet_id                   = aws_subnet.test[0].id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettings(rName string) string {
	return composeConfig(
		testAccAWSStorageGatewayGatewayConfigSmbActiveDirectorySettingsBase(rName),
		fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name = aws_directory_service_directory.test.name
    password    = aws_directory_service_directory.test.password
    username    = "Administrator"
  }
}
`, rName))
}

func testAccAWSStorageGatewayGatewayConfig_SmbActiveDirectorySettingsTimeout(rName string, timeout int) string {
	return composeConfig(
		testAccAWSStorageGatewayGatewayConfigSmbActiveDirectorySettingsBase(rName),
		fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %[1]q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"

  smb_active_directory_settings {
    domain_name        = aws_directory_service_directory.test.name
    password           = aws_directory_service_directory.test.password
    username           = "Administrator"
    timeout_in_seconds = %[2]d
  }
}
`, rName, timeout))
}

func testAccAWSStorageGatewayGatewayConfig_SmbGuestPassword(rName, smbGuestPassword string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "FILE_S3"
  smb_guest_password = %q
}
`, rName, smbGuestPassword)
}

func testAccAWSStorageGatewayGatewayConfigSMBSecurityStrategy(rName, strategy string) string {
	return testAccAWSStorageGateway_FileGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address    = aws_instance.test.public_ip
  gateway_name          = %q
  gateway_timezone      = "GMT"
  gateway_type          = "FILE_S3"
  smb_security_strategy = %q
}
`, rName, strategy)
}

func testAccAWSStorageGatewayGatewayConfigTags1(rName, tagKey1, tagValue1 string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"

  tags = {
    %q = %q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSStorageGatewayGatewayConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address = aws_instance.test.public_ip
  gateway_name       = %q
  gateway_timezone   = "GMT"
  gateway_type       = "CACHED"

  tags = {
    %q = %q
    %q = %q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSStorageGatewayGatewayBandwidthConfigUpload(rName string, rate int) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address                        = aws_instance.test.public_ip
  gateway_name                              = %[1]q
  gateway_timezone                          = "GMT"
  gateway_type                              = "CACHED"
  average_upload_rate_limit_in_bits_per_sec = %[2]d
}
`, rName, rate)
}

func testAccAWSStorageGatewayGatewayBandwidthConfigDownload(rName string, rate int) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address                          = aws_instance.test.public_ip
  gateway_name                                = %[1]q
  gateway_timezone                            = "GMT"
  gateway_type                                = "CACHED"
  average_download_rate_limit_in_bits_per_sec = %[2]d
}
`, rName, rate)
}

func testAccAWSStorageGatewayGatewayBandwidthConfigAll(rName string, rate int) string {
	return testAccAWSStorageGateway_TapeAndVolumeGatewayBase(rName) + fmt.Sprintf(`
resource "aws_storagegateway_gateway" "test" {
  gateway_ip_address                          = aws_instance.test.public_ip
  gateway_name                                = %[1]q
  gateway_timezone                            = "GMT"
  gateway_type                                = "CACHED"
  average_upload_rate_limit_in_bits_per_sec   = %[2]d
  average_download_rate_limit_in_bits_per_sec = %[2]d
}
`, rName, rate)
}
