package aws

import (
	"errors"
	"fmt"
	"log"
	"regexp"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfawsresource"
)

func init() {
	resource.AddTestSweepers("aws_spot_fleet_request", &resource.Sweeper{
		Name: "aws_spot_fleet_request",
		F:    testSweepSpotFleetRequests,
	})
}

func testSweepSpotFleetRequests(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn

	err = conn.DescribeSpotFleetRequestsPages(&ec2.DescribeSpotFleetRequestsInput{}, func(page *ec2.DescribeSpotFleetRequestsOutput, isLast bool) bool {
		if len(page.SpotFleetRequestConfigs) == 0 {
			log.Print("[DEBUG] No Spot Fleet Requests to sweep")
			return false
		}

		for _, config := range page.SpotFleetRequestConfigs {
			id := aws.StringValue(config.SpotFleetRequestId)

			log.Printf("[INFO] Deleting Spot Fleet Request: %s", id)
			err := deleteSpotFleetRequest(id, true, 15*time.Minute, conn)
			if err != nil {
				log.Printf("[ERROR] Failed to delete Spot Fleet Request (%s): %s", id, err)
			}
		}
		return !isLast
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 Spot Fleet Requests sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving EC2 Spot Fleet Requests: %s", err)
	}
	return nil
}

func TestAccAWSSpotFleetRequest_basic(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "Default"),
					resource.TestCheckResourceAttr(resourceName, "valid_until", validUntil),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_tags(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigTags1(rName, validUntil, "key1", "value1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccAWSSpotFleetRequestConfigTags2(rName, validUntil, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestConfigTags1(rName, validUntil, "key2", "value2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_associatePublicIpAddress(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigAssociatePublicIpAddress(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"associate_public_ip_address": "true",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_launchTemplate(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchTemplateConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_launchTemplate_multiple(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchTemplateMultipleConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_launchTemplateWithOverrides(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchTemplateConfigWithOverrides(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_launchTemplateToLaunchSpec(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchTemplateConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					testAccCheckAWSSpotFleetRequestConfigRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_launchSpecToLaunchTemplate(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestLaunchTemplateConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "launch_template_config.#", "1"),
					testAccCheckAWSSpotFleetRequestConfigRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_instanceInterruptionBehavior(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "instance_interruption_behaviour", "stop"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_fleetType(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigFleetType(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "fleet_type", "request"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_iamInstanceProfileArn(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigIamInstanceProfileArn(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					testAccCheckAWSSpotFleetRequest_IamInstanceProfileArn(&sfr),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_changePriceForcesNewRequest(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccAWSSpotFleetRequestConfigChangeSpotBidPrice(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					testAccCheckAWSSpotFleetRequestConfigRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_updateTargetCapacity(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccAWSSpotFleetRequestConfigTargetCapacity(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "3"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_updateExcessCapacityTerminationPolicy(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &before),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "Default"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccAWSSpotFleetRequestConfigExcessCapacityTermination(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &after),
					resource.TestCheckResourceAttr(resourceName, "excess_capacity_termination_policy", "NoTermination"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceAzOrSubnetInRegion(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceAzInGivenList(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	availabilityZonesDataSource := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithAzs(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.availability_zone", availabilityZonesDataSource, "names.0"),
					tfawsresource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.availability_zone", availabilityZonesDataSource, "names.1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceSubnetInGivenList(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithSubnet(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstanceTypesInSameAz(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	instanceTypeDataSource := "data.aws_ec2_instance_type_offering.available"
	availabilityZonesDataSource := "data.aws_availability_zones.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameAz(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					tfawsresource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.availability_zone", availabilityZonesDataSource, "names.0"),
					tfawsresource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.instance_type", instanceTypeDataSource, "instance_type"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"instance_type": "m3.large",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstanceTypesInSameSubnet(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameSubnet(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_overriddingSpotPrice(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	instanceTypeDataSourceName := "data.aws_ec2_instance_type_offering.available"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigOverridingSpotPrice(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "spot_price", "0.05"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"spot_price":    "0.05",
						"instance_type": "m3.large",
					}),
					tfawsresource.TestCheckTypeSetElemAttrPair(resourceName, "launch_specification.*.instance_type", instanceTypeDataSourceName, "instance_type"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withoutSpotPrice(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithoutSpotPrice(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_diversifiedAllocation(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigDiversifiedAllocation(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "allocation_strategy", "diversified"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstancePools(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstancePools(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "3"),
					resource.TestCheckResourceAttr(resourceName, "allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr(resourceName, "instance_pools_to_use_count", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withWeightedCapacity(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	fulfillSleep := func() resource.TestCheckFunc {
		// sleep so that EC2 can fuflill the request. We do this to guard against a
		// regression and possible leak where we'll destroy the request and the
		// associated IAM role before anything is actually provisioned and running,
		// thus leaking when those newly started instances are attempted to be
		// destroyed
		// See https://github.com/hashicorp/terraform/pull/8938
		return func(s *terraform.State) error {
			log.Print("[DEBUG] Test: Sleep to allow EC2 to actually begin fulfilling TestAccAWSSpotFleetRequest_withWeightedCapacity request")
			time.Sleep(1 * time.Minute)
			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithWeightedCapacity(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					fulfillSleep(),
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "2"),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"weighted_capacity": "3",
						"instance_type":     "r3.large",
					}),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"weighted_capacity": "6",
						"instance_type":     "m3.large",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withEBSDisk(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestEBSConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &config),
					testAccCheckAWSSpotFleetRequest_EBSAttributes(&config),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_LaunchSpecification_EbsBlockDevice_KmsKeyId(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchSpecificationEbsBlockDeviceKmsKeyId(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &config),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_LaunchSpecification_RootBlockDevice_KmsKeyId(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchSpecificationRootBlockDeviceKmsKeyId(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &config),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withTags(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestTagsConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &config),
					tfawsresource.TestCheckTypeSetElemNestedAttrs(resourceName, "launch_specification.*", map[string]string{
						"tags.%":      "2",
						"tags.First":  "TfAccTest",
						"tags.Second": "Terraform",
					}),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_placementTenancyAndGroup(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestTenancyGroupConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					testAccCheckAWSSpotFleetRequest_PlacementAttributes(&sfr, rName),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_WithELBs(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithELBs(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "load_balancers.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_WithTargetGroups(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithTargetGroups(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "spot_request_state", "active"),
					resource.TestCheckResourceAttr(resourceName, "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "target_group_arns.#", "1"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_zero_capacity(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestZeroCapacityConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "0"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"wait_for_fulfillment"},
			},
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "2"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestZeroCapacityConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					resource.TestCheckResourceAttr(resourceName, "target_capacity", "0"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_WithInstanceStoreAmi(t *testing.T) {
	TestAccSkip(t, "Test fails due to test harness constraints")
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSpotFleetRequestLaunchSpecificationWithInstanceStoreAmi(rName, validUntil),
				ExpectError: regexp.MustCompile("Instance store backed AMIs do not provide a root device name"),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_disappears(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandomWithPrefix("tf-acc-test")
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &sfr),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsSpotFleetRequest(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAWSSpotFleetRequestConfigRecreated(t *testing.T,
	before, after *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before.SpotFleetRequestId == after.SpotFleetRequestId {
			t.Fatalf("Expected change of Spot Fleet Request IDs, but both were %v", before.SpotFleetRequestId)
		}
		return nil
	}
}

func testAccCheckAWSSpotFleetRequestExists(
	n string, sfr *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No Spot fleet request with that id exists")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		params := &ec2.DescribeSpotFleetRequestsInput{
			SpotFleetRequestIds: []*string{&rs.Primary.ID},
		}
		resp, err := conn.DescribeSpotFleetRequests(params)

		if err != nil {
			return err
		}

		if v := len(resp.SpotFleetRequestConfigs); v != 1 {
			return fmt.Errorf("Expected 1 request returned, got %d", v)
		}

		*sfr = *resp.SpotFleetRequestConfigs[0]

		return nil
	}
}

func testAccCheckAWSSpotFleetRequestDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_spot_fleet_request" {
			continue
		}

		_, err := conn.CancelSpotFleetRequests(&ec2.CancelSpotFleetRequestsInput{
			SpotFleetRequestIds: []*string{aws.String(rs.Primary.ID)},
			TerminateInstances:  aws.Bool(true),
		})

		if err != nil {
			return fmt.Errorf("Error cancelling spot request (%s): %s", rs.Primary.ID, err)
		}
	}

	return nil
}

func testAccCheckAWSSpotFleetRequest_EBSAttributes(
	sfr *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sfr.SpotFleetRequestConfig.LaunchSpecifications) == 0 {
			return errors.New("Missing launch specification")
		}

		spec := *sfr.SpotFleetRequestConfig.LaunchSpecifications[0]

		ebs := spec.BlockDeviceMappings
		if len(ebs) < 2 {
			return fmt.Errorf("Expected %d block device mappings, got %d", 2, len(ebs))
		}

		if *ebs[0].DeviceName != "/dev/xvda" {
			return fmt.Errorf("Expected device 0's name to be %s, got %s", "/dev/xvda", *ebs[0].DeviceName)
		}
		if *ebs[1].DeviceName != "/dev/xvdcz" {
			return fmt.Errorf("Expected device 1's name to be %s, got %s", "/dev/xvdcz", *ebs[1].DeviceName)
		}

		return nil
	}
}

func testAccCheckAWSSpotFleetRequest_PlacementAttributes(
	sfr *ec2.SpotFleetRequestConfig, rName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sfr.SpotFleetRequestConfig.LaunchSpecifications) == 0 {
			return errors.New("Missing launch specification")
		}

		spec := *sfr.SpotFleetRequestConfig.LaunchSpecifications[0]

		placement := spec.Placement
		if placement == nil {
			return fmt.Errorf("Expected placement to be set, got nil")
		}
		if *placement.Tenancy != ec2.TenancyDedicated {
			return fmt.Errorf("Expected placement tenancy to be %q, got %q", "dedicated", *placement.Tenancy)
		}

		if aws.StringValue(placement.GroupName) != fmt.Sprintf("test-pg-%s", rName) {
			return fmt.Errorf("Expected placement group to be %q, got %q", fmt.Sprintf("test-pg-%s", rName), aws.StringValue(placement.GroupName))
		}

		return nil
	}

}

func testAccPreCheckAWSEc2SpotFleetRequest(t *testing.T) {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	input := &ec2.DescribeSpotFleetRequestsInput{}

	_, err := conn.DescribeSpotFleetRequests(input)

	if testAccPreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}

	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccCheckAWSSpotFleetRequest_IamInstanceProfileArn(sfr *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len(sfr.SpotFleetRequestConfig.LaunchSpecifications) == 0 {
			return errors.New("Missing launch specification")
		}

		spec := *sfr.SpotFleetRequestConfig.LaunchSpecifications[0]

		profile := spec.IamInstanceProfile
		if profile == nil {
			return fmt.Errorf("Expected IamInstanceProfile to be set, got nil")
		}
		//Validate the string whether it is ARN
		re := regexp.MustCompile(fmt.Sprintf(`arn:%s:iam::\d{12}:instance-profile/?[a-zA-Z0-9+=,.@-_].*`, testAccGetPartition()))
		if !re.MatchString(*profile.Arn) {
			return fmt.Errorf("Expected IamInstanceProfile input as ARN, got %s", *profile.Arn)
		}

		return nil
	}
}

func testAccAWSSpotFleetRequestConfigBase(rName string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForRegion("t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = %[1]q
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"

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
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF

  tags = {
    Name = %[1]q
  }
}

resource "aws_iam_policy" "test" {
  name        = %[1]q
  path        = "/"
  description = "Spot Fleet Request ACCTest Policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "ec2:DescribeImages",
        "ec2:DescribeSubnets",
        "ec2:RequestSpotInstances",
        "ec2:TerminateInstances",
        "ec2:DescribeInstanceStatus",
        "iam:PassRole"
      ],
      "Resource": [
        "*"
      ]
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
`, rName))
}

func testAccAWSSpotFleetRequestConfig(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigTags1(rName, validUntil, tagKey1, tagValue1 string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
  }

  tags = {
    %[2]q = %[3]q
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, tagKey1, tagValue1)
}

func testAccAWSSpotFleetRequestConfigTags2(rName, validUntil, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
  }

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSSpotFleetRequestConfigAssociatePublicIpAddress(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type               = data.aws_ec2_instance_type_offering.available.instance_type
    ami                         = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name                    = aws_key_pair.test.key_name
    associate_public_ip_address = true
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigTargetCapacity(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 3
  valid_until                         = %[1]q
  fleet_type                          = "request"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestLaunchTemplateConfig(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[2]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, rName)
}

func testAccAWSSpotFleetRequestLaunchTemplateMultipleConfig(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) +
		fmt.Sprintf(`
data "aws_ec2_instance_type_offering" "test" {
  filter {
    name   = "instance-type"
    values = ["t1.micro"]
  }

  preferred_instance_types = ["t1.micro"]
}

resource "aws_launch_template" "test1" {
  name          = "%[2]s-1"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name
}

resource "aws_launch_template" "test2" {
  name          = "%[2]s-2"
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.test.instance_type
  key_name      = aws_key_pair.test.key_name
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test1.name
      version = aws_launch_template.test1.latest_version
    }
  }

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test2.name
      version = aws_launch_template.test2.latest_version
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, rName)
}

func testAccAWSSpotFleetRequestLaunchTemplateConfigWithOverrides(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_launch_template" "test" {
  name          = %[2]q
  image_id      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  key_name      = aws_key_pair.test.key_name
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_template_config {
    launch_template_specification {
      name    = aws_launch_template.test.name
      version = aws_launch_template.test.latest_version
    }

    overrides {
      instance_type     = "t1.micro"
      weighted_capacity = "2"
    }

    overrides {
      instance_type = "m3.medium"
      priority      = 1
      spot_price    = "0.26"
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, rName)
}

func testAccAWSSpotFleetRequestConfigExcessCapacityTermination(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  excess_capacity_termination_policy  = "NoTermination"
  valid_until                         = %[1]q
  fleet_type                          = "request"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigFleetType(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  fleet_type                          = "request"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigIamInstanceProfileArn(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_iam_role" "test-role1" {
  name = "tf-test-role1-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.${data.aws_partition.current.dns_suffix}",
          "ec2.${data.aws_partition.current.dns_suffix}"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_role_policy" "test-role-policy1" {
  name = "tf-test-role-policy1-%[1]s"
  role = aws_iam_role.test-role1.name

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": {
    "Effect": "Allow",
    "Action": "*",
    "Resource": "*"
  }
}
EOF
}

resource "aws_iam_instance_profile" "test-iam-instance-profile1" {
  name = "tf-test-profile1-%[1]s"
  role = aws_iam_role.test-role1.name
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.25"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type            = data.aws_ec2_instance_type_offering.available.instance_type
    ami                      = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name                 = aws_key_pair.test.key_name
    iam_instance_profile_arn = aws_iam_instance_profile.test-iam-instance-profile1.arn
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil)
}

func testAccAWSSpotFleetRequestConfigChangeSpotBidPrice(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithAzs(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[1]
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithSubnet(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[2]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[2]q
  }
}

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.20.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[2]q
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 4
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test.id
  }

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.bar.id
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, rName)
}

func testAccAWSSpotFleetRequestConfigWithELBs(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
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

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.20.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_elb" "elb" {
  name     = %[1]q
  subnets  = [aws_subnet.test.id, aws_subnet.bar.id]
  internal = true

  listener {
    instance_port     = 80
    instance_protocol = "HTTP"
    lb_port           = 80
    lb_protocol       = "HTTP"
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.5"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true
  load_balancers                      = [aws_elb.elb.name]

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test.id
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithTargetGroups(rName string, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
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

resource "aws_subnet" "bar" {
  cidr_block        = "10.1.20.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_alb" "alb" {
  name     = %[1]q
  internal = true
  subnets  = [aws_subnet.test.id, aws_subnet.bar.id]
}

resource "aws_alb_listener" "listener" {
  load_balancer_arn = aws_alb.alb.arn
  port              = 80
  protocol          = "HTTP"

  default_action {
    target_group_arn = aws_alb_target_group.target_group.arn
    type             = "forward"
  }
}

resource "aws_alb_target_group" "target_group" {
  name     = aws_alb.alb.name
  port     = 80
  protocol = "HTTP"
  vpc_id   = aws_vpc.test.id
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.5"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true
  target_group_arns                   = [aws_alb_target_group.target_group.arn]

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test.id
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil)
}

func testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameAz(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameSubnet(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[2]q
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

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 4
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = "m3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test.id
  }

  launch_specification {
    instance_type = "r3.large"
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
    subnet_id     = aws_subnet.test.id
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, rName)
}

func testAccAWSSpotFleetRequestConfigOverridingSpotPrice(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
    spot_price        = "0.05"
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithoutSpotPrice(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigMultipleInstancePools(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.7"
  target_capacity                     = 30
  valid_until                         = %[1]q
  instance_pools_to_use_count         = 2
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = "r3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigDiversifiedAllocation(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.7"
  target_capacity                     = 30
  valid_until                         = %[1]q
  allocation_strategy                 = "diversified"
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  launch_specification {
    instance_type     = "r3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithWeightedCapacity(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.7"
  target_capacity                     = 10
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type     = "m3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
    weighted_capacity = "6"
  }

  launch_specification {
    instance_type     = "r3.large"
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    availability_zone = data.aws_availability_zones.available.names[0]
    weighted_capacity = "3"
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestEBSConfig(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

    ebs_block_device {
      device_name = "/dev/xvda"
      volume_type = "gp2"
      volume_size = "8"
    }

    ebs_block_device {
      device_name = "/dev/xvdcz"
      volume_type = "gp2"
      volume_size = "100"
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestLaunchSpecificationEbsBlockDeviceKmsKeyId(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags = {
    Name = %[2]q
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  valid_until                         = %[1]q
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    instance_type = "t2.micro"

    ebs_block_device {
      device_name = "/dev/xvda"
      volume_type = "gp2"
      volume_size = 8
    }

    ebs_block_device {
      device_name = "/dev/xvdcz"
      encrypted   = true
      kms_key_id  = aws_kms_key.test.arn
      volume_type = "gp2"
      volume_size = 10
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, rName)
}

func testAccAWSSpotFleetRequestLaunchSpecificationRootBlockDeviceKmsKeyId(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_kms_key" "test" {
  deletion_window_in_days = 7

  tags = {
    Name = %[2]q
  }
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  valid_until                         = %[1]q
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    instance_type = "t2.micro"

    root_block_device {
      encrypted   = true
      kms_key_id  = aws_kms_key.test.arn
      volume_type = "gp2"
      volume_size = 10
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil, rName)
}

func testAccAWSSpotFleetRequestLaunchSpecificationWithInstanceStoreAmi(rName string, validUntil string) string {
	return testAccLatestAmazonLinuxHvmInstanceStoreAmiConfig() +
		testAccAWSSpotFleetRequestConfigBase(rName) +
		fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test-role.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    ami           = data.aws_ami.amzn-ami-minimal-hvm-instance-store.id
    instance_type = "c3.large"
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestTagsConfig(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 1
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

    tags = {
      First  = "TfAccTest"
      Second = "Terraform"
    }
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestTenancyGroupConfig(rName, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_placement_group" "test" {
  name     = "test-pg-%[1]s"
  strategy = "cluster"
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 2
  valid_until                         = %[2]q
  terminate_instances_with_expiration = true

  launch_specification {
    instance_type     = data.aws_ec2_instance_type_offering.available.instance_type
    ami               = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name          = aws_key_pair.test.key_name
    placement_tenancy = "dedicated"
    placement_group   = aws_placement_group.test.name
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, rName, validUntil)
}

func testAccAWSSpotFleetRequestZeroCapacityConfig(rName string, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = aws_iam_role.test.arn
  spot_price                          = "0.05"
  target_capacity                     = 0
  valid_until                         = %[1]q
  terminate_instances_with_expiration = true
  instance_interruption_behaviour     = "stop"
  wait_for_fulfillment                = true

  launch_specification {
    instance_type = data.aws_ec2_instance_type_offering.available.instance_type
    ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
    key_name      = aws_key_pair.test.key_name
  }

  depends_on = [aws_iam_policy_attachment.test]
}
`, validUntil)
}
