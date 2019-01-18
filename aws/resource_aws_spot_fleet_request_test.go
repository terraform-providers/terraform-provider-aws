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
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
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

func TestAccAWSSpotFleetRequest_associatePublicIpAddress(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigAssociatePublicIpAddress(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.24370212.associate_public_ip_address", "true"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_instanceInterruptionBehavior(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "instance_interruption_behaviour", "stop"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_fleetType(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigFleetType(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "fleet_type", "request"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_iamInstanceProfileArn(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigIamInstanceProfileArn(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
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

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &before),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_price", "0.005"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "1"),
				),
			},
			{
				Config: testAccAWSSpotFleetRequestConfigChangeSpotBidPrice(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &after),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_price", "0.01"),
					testAccCheckAWSSpotFleetRequestConfigRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceAzOrSubnetInRegion(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfig(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceAzInGivenList(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithAzs(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.1991689378.availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.19404370.availability_zone", "us-west-2b"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_lowestPriceSubnetInGivenList(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithSubnet(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstanceTypesInSameAz(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameAz(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.1991689378.instance_type", "m1.small"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.1991689378.availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.590403189.instance_type", "m3.large"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.590403189.availability_zone", "us-west-2a"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstanceTypesInSameSubnet(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameSubnet(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_overriddingSpotPrice(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigOverridingSpotPrice(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_price", "0.035"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.4143232216.spot_price", "0.01"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.4143232216.instance_type", "m3.large"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.1991689378.spot_price", ""), //there will not be a value here since it's not overriding
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.1991689378.instance_type", "m1.small"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withoutSpotPrice(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithoutSpotPrice(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_diversifiedAllocation(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigDiversifiedAllocation(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "allocation_strategy", "diversified"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_multipleInstancePools(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigMultipleInstancePools(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "3"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "allocation_strategy", "lowestPrice"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "instance_pools_to_use_count", "2"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withWeightedCapacity(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

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
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithWeightedCapacity(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					fulfillSleep(),
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "2"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.4120185872.weighted_capacity", "3"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.4120185872.instance_type", "r3.large"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.590403189.weighted_capacity", "6"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.590403189.instance_type", "m3.large"),
				),
			},
		},
	})
}

func TestAccAWSSpotFleetRequest_withEBSDisk(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestEBSConfig(rName, rInt),
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

func TestAccAWSSpotFleetRequest_withTags(t *testing.T) {
	var config ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestTagsConfig(rName, rInt),
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

func TestAccAWSSpotFleetRequest_placementTenancy(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestTenancyConfig(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					testAccCheckAWSSpotFleetRequest_PlacementAttributes(&sfr),
				),
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
	sfr *ec2.SpotFleetRequestConfig) resource.TestCheckFunc {
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

func TestAccAWSSpotFleetRequest_WithELBs(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithELBs(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "load_balancers.#", "1"),
				),
			},
		},
	})
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

func TestAccAWSSpotFleetRequest_WithTargetGroups(t *testing.T) {
	var sfr ec2.SpotFleetRequestConfig
	rName := acctest.RandString(10)
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSSpotFleetRequestDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSSpotFleetRequestConfigWithTargetGroups(rName, rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSSpotFleetRequestExists(
						"aws_spot_fleet_request.foo", &sfr),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "spot_request_state", "active"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "launch_specification.#", "1"),
					resource.TestCheckResourceAttr(
						"aws_spot_fleet_request.foo", "target_group_arns.#", "1"),
				),
			},
		},
	})
}

func testAccAWSSpotFleetRequestConfigAssociatePublicIpAddress(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy_attachment" "test-attach" {
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonEC2SpotFleetRole"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.027"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
        associate_public_ip_address = true
    }
	depends_on = ["aws_iam_policy_attachment.test-attach"]
}`, rName, rInt, rName)
}

func testAccAWSSpotFleetRequestConfig(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigFleetType(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
    fleet_type = "request"
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigIamInstanceProfileArn(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_iam_role" "test-role1" {
    name = "tf-test-role1-%s"
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
	name = "tf-test-role-policy1-%s"
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
	name = "tf-test-profile1-%s"
	role = "${aws_iam_role.test-role1.name}"
}

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName, rName, rName, rName)
}

func testAccAWSSpotFleetRequestConfigChangeSpotBidPrice(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.01"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
    terminate_instances_with_expiration = true
    wait_for_fulfillment = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-516b9131"
        key_name = "${aws_key_pair.debugging.key_name}"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigWithAzs(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigWithSubnet(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigWithELBs(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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
  name = "test-elb-%s"
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
    valid_until = "2029-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName, rName)
}

func testAccAWSSpotFleetRequestConfigWithTargetGroups(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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
  name            = "test-elb-%s"
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
    valid_until = "2029-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName, rName)
}

func testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameAz(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.025"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigMultipleInstanceTypesinSameSubnet(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigOverridingSpotPrice(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.035"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigWithoutSpotPrice(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigMultipleInstancePools(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.7"
    target_capacity = 30
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigDiversifiedAllocation(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.7"
    target_capacity = 30
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestConfigWithWeightedCapacity(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.7"
    target_capacity = 10
    valid_until = "2019-11-04T20:44:20Z"
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
`, rName, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestEBSConfig(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 1
    valid_until = "2019-11-04T20:44:20Z"
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
`, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestTagsConfig(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 1
    valid_until = "2019-11-04T20:44:20Z"
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
`, rInt, rInt, rName)
}

func testAccAWSSpotFleetRequestTenancyConfig(rName string, rInt int) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "debugging" {
	key_name = "tmp-key-%s"
	public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 phodgson@thoughtworks.com"
}

resource "aws_iam_policy" "test-policy" {
  name = "test-policy-%d"
  path = "/"
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
    name = "test-attachment-%d"
    roles = ["${aws_iam_role.test-role.name}"]
    policy_arn = "${aws_iam_policy.test-policy.arn}"
}

resource "aws_iam_role" "test-role" {
    name = "test-role-%s"
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

resource "aws_spot_fleet_request" "foo" {
    iam_fleet_role = "${aws_iam_role.test-role.arn}"
    spot_price = "0.005"
    target_capacity = 2
    valid_until = "2019-11-04T20:44:20Z"
    terminate_instances_with_expiration = true
    launch_specification {
        instance_type = "m1.small"
        ami = "ami-d06a90b0"
        key_name = "${aws_key_pair.debugging.key_name}"
        placement_tenancy = "dedicated"
    }
    depends_on = ["aws_iam_policy_attachment.test-attach"]
}
`, rName, rInt, rInt, rName)
}
