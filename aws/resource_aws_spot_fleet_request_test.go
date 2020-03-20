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
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
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
			err := deleteSpotFleetRequest(id, true, 5*time.Minute, conn)
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
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "excess_capacity_termination_policy", "Default"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "valid_until", validUntil),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_associatePublicIpAddress(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigAssociatePublicIpAddress(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.24370212.associate_public_ip_address", "true"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_instanceInterruptionBehavior(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "instance_interruption_behaviour", "stop"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_fleetType(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigFleetType(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "fleet_type", "request"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_iamInstanceProfileArn(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigIamInstanceProfileArn(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					testAccCheckAWSSpotFleetRequest_IamInstanceProfileArn(&sfr),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_changePriceForcesNewRequest(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &before),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_price", "0.005"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "1"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestConfigChangeSpotBidPrice(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &after),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_price", "0.01"),
					testAccCheckAWSSpotFleetRequestConfigRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_updateTargetCapacity(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &before),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "target_capacity", "2"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestConfigTargetCapacity(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &after),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "target_capacity", "3"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &before),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "target_capacity", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_updateExcessCapacityTerminationPolicy(t *testing.T) {
	var before, after ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &before),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "excess_capacity_termination_policy", "Default"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestConfigExcessCapacityTermination(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &after),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "excess_capacity_termination_policy", "NoTermination"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceAzOrSubnetInRegion(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceAzInGivenList(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithAzs(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.1991689378.availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.19404370.availability_zone", "us-west-2b"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceSubnetInGivenList(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithSubnet(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstanceTypesInSameAz(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameAz(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.1991689378.instance_type", "m1.small"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.1991689378.availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.590403189.instance_type", "m3.large"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.590403189.availability_zone", "us-west-2a"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstanceTypesInSameSubnet(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameSubnet(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_overriddingSpotPrice(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigOverridingSpotPrice(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_price", "0.035"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.4143232216.spot_price", "0.01"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.4143232216.instance_type", "m3.large"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.1991689378.spot_price", ""), //there will not be a value here since it's not overriding
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.1991689378.instance_type", "m1.small"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withoutSpotPrice(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithoutSpotPrice(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_diversifiedAllocation(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigDiversifiedAllocation(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "3"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "allocation_strategy", "diversified"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstancePools(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstancePools(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "3"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "instance_pools_to_use_count", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withWeightedCapacity(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

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
				Config: testAccAWSSpotFleetRequestConfigWithWeightedCapacity(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					fulfillSleep(),
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.4120185872.weighted_capacity", "3"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.4120185872.instance_type", "r3.large"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.590403189.weighted_capacity", "6"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.590403189.instance_type", "m3.large"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withEBSDisk(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestEBSConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &config),
					testAccCheckAWSSpotFleetRequest_EBSAttributes(
						&config),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_LaunchSpecification_EbsBlockDevice_KmsKeyId(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchSpecificationEbsBlockDeviceKmsKeyId(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &config),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_LaunchSpecification_RootBlockDevice_KmsKeyId(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
	resourceName := "aws_spot_fleet_request.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestLaunchSpecificationRootBlockDeviceKmsKeyId(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(resourceName, &config),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withTags(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestTagsConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &config),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.24370212.tags.%", "2"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.24370212.tags.First", "TfAccTest"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.24370212.tags.Second", "Terraform"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_placementTenancyAndGroup(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestTenancyGroupConfig(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists("aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					testAccCheckAWSSpotFleetRequest_PlacementAttributes(&sfr, rName),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_WithELBs(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithELBs(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "load_balancers.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_WithTargetGroups(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithTargetGroups(rName, rInt, validUntil),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr("aws_spot_fleet_request.foo", "target_group_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_WithInstanceStoreAmi(t *testing.T) {
	t.Skip("Test fails due to test harness constraints")
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()
	validUntil := time.Now().UTC().Add(24 * time.Hour).Format(time.RFC3339)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2SpotFleetRequest(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccAWSSpotFleetRequestLaunchSpecificationWithInstanceStoreAmi(rName, rInt, validUntil),
				ExpectError: regexp.MustCompile("Instance store backed AMIs do not provide a root device name"),
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
		if *placement.Tenancy != "dedicated" {
			return fmt.Errorf("Expected placement tenancy to be %q, got %q", "dedicated", *placement.Tenancy)
		}
		if aws.StringValue(placement.GroupName) != fmt.Sprintf("test-pg-%s", rName) {
			return fmt.Errorf("Expected placement group to be %q, got %q", fmt.Sprintf("test-pg-%s", rName), aws.StringValue(placement.GroupName))
		}

		return nil
	}
}

func testAccCheckAWSSpotFleetRequest_IamInstanceProfileArn(
	sfr *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
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
		re := regexp.MustCompile(`arn:aws:iam::\d{12}:instance-profile/?[a-zA-Z0-9+=,.@-_].*`)
		if !re.MatchString(*profile.Arn) {
			return fmt.Errorf("Expected IamInstanceProfile input as ARN, got %s", *profile.Arn)
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

func testAccAWSSpotFleetRequestConfigBase(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
  key_name   = "tmp-key-%[1]s"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_role" "test-role" {
  name = "test-role-%[1]s"

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": [
          "spotfleet.amazonaws.com",
          "ec2.amazonaws.com"
        ]
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOF
}

resource "aws_iam_policy" "test-policy" {
  name        = "test-policy-%[2]d"
  path        = "/"
  description = "Spot Fleet Request ACCTest Policy"

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [{
    "Effect": "Allow",
    "Action": [
       "ec2:DescribeImages",
       "ec2:DescribeSubnets",
       "ec2:RequestSpotInstances",
       "ec2:TerminateInstances",
       "ec2:DescribeInstanceStatus",
       "iam:PassRole"
        ],
    "Resource": ["*"]
  }]
}
EOF
}

resource "aws_iam_policy_attachment" "test-attach" {
  name       = "test-attachment-%[2]d"
  roles      = ["${aws_iam_role.test-role.name}"]
  policy_arn = "${aws_iam_policy.test-policy.arn}"
}
`, rName, rInt)
}

func testAccAWSSpotFleetRequestConfig(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    instance_interruption_behaviour = "stop"
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigAssociatePublicIpAddress(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.027"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        associate_public_ip_address = true
    }
	depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigTargetCapacity(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 3
    valid_until = %[1]q
    fleet_type = "request"
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigExcessCapacityTermination(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    excess_capacity_termination_policy = "NoTermination"
    valid_until = %[1]q
    fleet_type = "request"
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigFleetType(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = %[1]q
    fleet_type = "request"
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigIamInstanceProfileArn(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
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
          "spotfleet.amazonaws.com",
          "ec2.amazonaws.com"
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
	role = "${aws_iam_role.test-role1.name}"
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
	role = "${aws_iam_role.test-role1.name}"
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = %[2]q
    terminate_instances_with_expiration = true
    instance_interruption_behaviour = "stop"
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        iam_instance_profile_arn = "${aws_iam_instance_profile.test-iam-instance-profile1.arn}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, rName, validUntil)
}

func testAccAWSSpotFleetRequestConfigChangeSpotBidPrice(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.01"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithAzs(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2b"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithSubnet(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "10.1.0.0/16"
  tags = {
        Name = "terraform-testacc-spot-fleet-request-w-subnet"
    }
}

resource "aws_subnet" "foo" {
    cidr_block = "10.1.1.0/24"
    vpc_id = "${aws_vpc.foo.id}"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-spot-fleet-request-w-subnet-foo"
    }
}

resource "aws_subnet" "bar" {
    cidr_block = "10.1.20.0/24"
    vpc_id = "${aws_vpc.foo.id}"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-spot-fleet-request-w-subnet-bar"
    }
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.05"
    target_capacity = 4
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d0f506b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        subnet_id = "${aws_subnet.foo.id}"
    }
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d0f506b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        subnet_id = "${aws_subnet.bar.id}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithELBs(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "foo" {
    cidr_block = "10.1.1.0/24"
    vpc_id = "${aws_vpc.foo.id}"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-spot-fleet-request-with-elb-foo"
    }
}

resource "aws_subnet" "bar" {
    cidr_block = "10.1.20.0/24"
    vpc_id = "${aws_vpc.foo.id}"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-spot-fleet-request-with-elb-bar"
    }
}

resource "aws_elb" "elb" {
  name = "test-elb-%[1]s"
  subnets = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
  internal = true

  listener {
    instance_port      = 80
    instance_protocol  = "HTTP"
    lb_port            = 80
    lb_protocol        = "HTTP"
  }
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.5"
    target_capacity = 2
    valid_until = %[2]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    load_balancers = ["${aws_elb.elb.name}"]
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d0f506b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        subnet_id = "${aws_subnet.foo.id}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, rName, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithTargetGroups(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "10.1.0.0/16"
}

resource "aws_subnet" "foo" {
    cidr_block = "10.1.1.0/24"
    vpc_id = "${aws_vpc.foo.id}"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-spot-fleet-request-with-target-groups-foo"
    }
}

resource "aws_subnet" "bar" {
    cidr_block = "10.1.20.0/24"
    vpc_id = "${aws_vpc.foo.id}"
    availability_zone = "us-west-2b"
  tags = {
        Name = "tf-acc-spot-fleet-request-with-target-groups-bar"
    }
}

resource "aws_alb" "alb" {
  name            = "test-elb-%[1]s"
  internal        = true
  subnets         = ["${aws_subnet.foo.id}", "${aws_subnet.bar.id}"]
}

resource "aws_alb_listener" "listener" {
 load_balancer_arn = "${aws_alb.alb.arn}"
 port = 80
 protocol = "HTTP"

 default_action {
   target_group_arn = "${aws_alb_target_group.target_group.arn}"
   type             = "forward"
 }
}

resource "aws_alb_target_group" "target_group" {
 name     = "${aws_alb.alb.name}"
 port     = 80
 protocol = "HTTP"
 vpc_id   = "${aws_vpc.foo.id}"
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.5"
    target_capacity = 2
    valid_until = %[2]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    target_group_arns = ["${aws_alb_target_group.target_group.arn}"]
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d0f506b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        subnet_id = "${aws_subnet.foo.id}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, rName, validUntil)
}

func testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameAz(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.025"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameSubnet(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_vpc" "foo" {
    cidr_block = "10.1.0.0/16"
  tags = {
        Name = "terraform-testacc-spot-fleet-request-multi-instance-types"
    }
}

resource "aws_subnet" "foo" {
    cidr_block = "10.1.1.0/24"
    vpc_id = "${aws_vpc.foo.id}"
    availability_zone = "us-west-2a"
  tags = {
        Name = "tf-acc-spot-fleet-request-multi-instance-types"
    }
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.035"
    target_capacity = 4
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d0f506b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        subnet_id = "${aws_subnet.foo.id}"
    }
    launch_specification {
        instance_type = "r3.large"
        ami = "ami-d0f506b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        subnet_id = "${aws_subnet.foo.id}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigOverridingSpotPrice(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.035"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
        spot_price = "0.01"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithoutSpotPrice(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigMultipleInstancePools(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.7"
    target_capacity = 30
    valid_until = %[1]q
    instance_pools_to_use_count = 2
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "r3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigDiversifiedAllocation(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.7"
    target_capacity = 30
    valid_until = %[1]q
    allocation_strategy = "diversified"
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    launch_specification {
        instance_type = "r3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestConfigWithWeightedCapacity(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.7"
    target_capacity = 10
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
        weighted_capacity = "6"
    }
    launch_specification {
        instance_type = "r3.large"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        availability_zone = "us-west-2a"
        weighted_capacity = "3"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestEBSConfig(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 1
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"

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
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestLaunchSpecificationEbsBlockDeviceKmsKeyId(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = "${aws_iam_role.test-role.arn}"
  spot_price                          = "0.005"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  valid_until                         = %[1]q
  wait_for_fulfillment                = true

  launch_specification {
    ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
    instance_type = "t2.micro"

    ebs_block_device {
      device_name = "/dev/xvda"
      volume_type = "gp2"
      volume_size = 8
    }

    ebs_block_device {
      device_name = "/dev/xvdcz"
      encrypted   = true
      kms_key_id  = "${aws_kms_key.test.arn}"
      volume_type = "gp2"
      volume_size = 10
    }
  }

  depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestLaunchSpecificationRootBlockDeviceKmsKeyId(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
data "aws_ami" "amzn-ami-minimal-hvm-ebs" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-minimal-hvm-*"]
  }

  filter {
    name   = "root-device-type"
    values = ["ebs"]
  }
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
}

resource "aws_spot_fleet_request" "test" {
  iam_fleet_role                      = "${aws_iam_role.test-role.arn}"
  spot_price                          = "0.005"
  target_capacity                     = 1
  terminate_instances_with_expiration = true
  valid_until                         = %[1]q
  wait_for_fulfillment                = true

  launch_specification {
    ami           = "${data.aws_ami.amzn-ami-minimal-hvm-ebs.id}"
    instance_type = "t2.micro"

    root_block_device {
      encrypted   = true
      kms_key_id  = "${aws_kms_key.test.arn}"
      volume_type = "gp2"
      volume_size = 10
    }
  }

  depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestLaunchSpecificationWithInstanceStoreAmi(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
data "aws_ami" "ubuntu_instance_store" {
  most_recent = true
  owners      = ["099720109477"] # Canonical

  # Latest Ubuntu 18.04 LTS amd64 instance-store HVM AMI
  filter {
    name   = "name"
    values = ["ubuntu/images/hvm-instance/ubuntu-bionic-18.04-amd64-server*"]
  }
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.03"
    target_capacity = 2
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true

    launch_specification {
      ami           = "${data.aws_ami.ubuntu_instance_store.id}"
	  instance_type = "c3.large"
    }

	depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestTagsConfig(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 1
    valid_until = %[1]q
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
  tags = {
            First = "TfAccTest"
            Second = "Terraform"
        }
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, validUntil)
}

func testAccAWSSpotFleetRequestTenancyGroupConfig(rName string, rInt int, validUntil string) string {
	return testAccAWSSpotFleetRequestConfigBase(rName, rInt) + fmt.Sprintf(`
resource "aws_placement_group" "test" {
	name     = "test-pg-%[1]s"
	strategy = "cluster"
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = %[2]q
    terminate_instances_with_expiration = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
		placement_tenancy = "dedicated"
		placement_group = "${aws_placement_group.test.name}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, rName, validUntil)
}
