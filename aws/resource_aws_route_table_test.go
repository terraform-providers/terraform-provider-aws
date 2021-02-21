package aws

import (
	"fmt"
	"log"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/go-multierror"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func init() {
	resource.AddTestSweepers("aws_route_table", &resource.Sweeper{
		Name: "aws_route_table",
		F:    testSweepRouteTables,
	})
}

func testSweepRouteTables(region string) error {
	client, err := sharedClientForRegion(region)

	if err != nil {
		return fmt.Errorf("error getting client: %w", err)
	}

	conn := client.(*AWSClient).ec2conn

	var sweeperErrs *multierror.Error

	input := &ec2.DescribeRouteTablesInput{}

	err = conn.DescribeRouteTablesPages(input, func(page *ec2.DescribeRouteTablesOutput, lastPage bool) bool {
		if page == nil {
			return !lastPage
		}

		for _, routeTable := range page.RouteTables {
			if routeTable == nil {
				continue
			}

			id := aws.StringValue(routeTable.RouteTableId)
			isMainRouteTableAssociation := false

			for _, routeTableAssociation := range routeTable.Associations {
				if routeTableAssociation == nil {
					continue
				}

				if aws.BoolValue(routeTableAssociation.Main) {
					isMainRouteTableAssociation = true
					break
				}

				associationID := aws.StringValue(routeTableAssociation.RouteTableAssociationId)

				input := &ec2.DisassociateRouteTableInput{
					AssociationId: routeTableAssociation.RouteTableAssociationId,
				}

				log.Printf("[DEBUG] Deleting EC2 Route Table Association: %s", associationID)
				_, err := conn.DisassociateRouteTable(input)

				if err != nil {
					sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s) Association (%s): %w", id, associationID, err)
					log.Printf("[ERROR] %s", sweeperErr)
					sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
					continue
				}
			}

			if isMainRouteTableAssociation {
				for _, route := range routeTable.Routes {
					if route == nil {
						continue
					}

					if aws.StringValue(route.GatewayId) == "local" {
						continue
					}

					// Prevent deleting default VPC route for Internet Gateway
					// which some testing is still reliant on operating correctly
					if strings.HasPrefix(aws.StringValue(route.GatewayId), "igw-") && aws.StringValue(route.DestinationCidrBlock) == "0.0.0.0/0" {
						continue
					}

					input := &ec2.DeleteRouteInput{
						DestinationCidrBlock:     route.DestinationCidrBlock,
						DestinationIpv6CidrBlock: route.DestinationIpv6CidrBlock,
						RouteTableId:             routeTable.RouteTableId,
					}

					log.Printf("[DEBUG] Deleting EC2 Route Table (%s) Route", id)
					_, err := conn.DeleteRoute(input)

					if err != nil {
						sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s) Route: %w", id, err)
						log.Printf("[ERROR] %s", sweeperErr)
						sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
						continue
					}
				}

				continue
			}

			input := &ec2.DeleteRouteTableInput{
				RouteTableId: routeTable.RouteTableId,
			}

			log.Printf("[DEBUG] Deleting EC2 Route Table: %s", id)
			_, err := conn.DeleteRouteTable(input)

			if err != nil {
				sweeperErr := fmt.Errorf("error deleting EC2 Route Table (%s): %w", id, err)
				log.Printf("[ERROR] %s", sweeperErr)
				sweeperErrs = multierror.Append(sweeperErrs, sweeperErr)
				continue
			}
		}

		return !lastPage
	})

	if testSweepSkipSweepError(err) {
		log.Printf("[WARN] Skipping EC2 Route Table sweep for %s: %s", region, err)
		return sweeperErrs.ErrorOrNil()
	}

	if err != nil {
		sweeperErrs = multierror.Append(sweeperErrs, fmt.Errorf("error listing EC2 Route Tables: %w", err))
	}

	return sweeperErrs.ErrorOrNil()
}

func TestAccAWSRouteTable_basic(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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

func TestAccAWSRouteTable_disappears(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRouteTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRouteTable_disappears_SubnetAssociation(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigSubnetAssociation(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRouteTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRouteTable_IPv4_To_InternetGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr1 := "10.2.0.0/16"
	destinationCidr2 := "10.3.0.0/16"
	destinationCidr3 := "10.4.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "gateway_id", igwResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr2, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr2, destinationCidr3),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr2, "gateway_id", igwResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr3, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_Instance(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	instanceResourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4Instance(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr, "instance_id", instanceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv6_To_EgressOnlyInternetGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv6EgressOnlyInternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "ipv6_cidr_block", destinationCidr, "egress_only_gateway_id", eoigwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				// Verify that expanded form of the destination CIDR causes no diff.
				Config:   testAccAWSRouteTableConfigIpv6EgressOnlyInternetGateway(rName, "::0/0"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSRouteTable_tags(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAWSRouteTableConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSRouteTable_RequireRouteDestination(t *testing.T) {
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRouteTableConfigNoDestination(rName),
				ExpectError: regexp.MustCompile("error creating route: one of `cidr_block"),
			},
		},
	})
}

func TestAccAWSRouteTable_RequireRouteTarget(t *testing.T) {
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccRouteTableConfigNoTarget(rName),
				ExpectError: regexp.MustCompile("error creating route: one of `egress_only_gateway_id"),
			},
		},
	})
}

func TestAccAWSRouteTable_Route_ConfigMode(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr1 := "10.2.0.0/16"
	destinationCidr2 := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "gateway_id", igwResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr2, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigRouteConfigModeNoBlocks(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "2"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "gateway_id", igwResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr2, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSRouteTableConfigRouteConfigModeZeroed(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_TransitGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr, "transit_gateway_id", tgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_VpcEndpoint(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	vpceResourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "0.0.0.0/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigRouteIpv4VpcEndpointId(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr, "vpc_endpoint_id", vpceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_LocalGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	lgwDataSourceName := "data.aws_ec2_local_gateway.first"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "0.0.0.0/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigRouteIpv4LocalGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr, "local_gateway_id", lgwDataSourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv4_To_VpcPeeringConnection(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4VpcPeeringConnection(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr, "vpc_peering_connection_id", pcxResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_vgwRoutePropagation(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	vgwResourceName1 := "aws_vpn_gateway.test1"
	vgwResourceName2 := "aws_vpn_gateway.test2"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigVgwRoutePropagation(rName, vgwResourceName1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "propagating_vgws.*", vgwResourceName1, "id"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSRouteTableConfigVgwRoutePropagation(rName, vgwResourceName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(resourceName, "propagating_vgws.*", vgwResourceName2, "id"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_ConditionalCidrBlock(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"
	destinationIpv6Cidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr, "gateway_id", igwResourceName, "id"),
				),
			},
			{
				Config: testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableRoute(resourceName, "ipv6_cidr_block", destinationIpv6Cidr, "gateway_id", igwResourceName, "id"),
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

func TestAccAWSRouteTable_IPv4_To_NatGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	ngwResourceName := "aws_nat_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv4NatGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr, "nat_gateway_id", ngwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_IPv6_To_NetworkInterface_Unattached(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	eniResourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigIpv6NetworkInterfaceUnattached(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckAWSRouteTableRoute(resourceName, "ipv6_cidr_block", destinationCidr, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_VpcMultipleCidrs(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigVpcMultipleCidrs(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_VpcClassicLink(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigVpcClassicLink(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_GatewayVpcEndpoint(t *testing.T) {
	var routeTable ec2.RouteTable
	var vpce ec2.VpcEndpoint
	resourceName := "aws_route_table.test"
	vpceResourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigGatewayVpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckVpcEndpointExists(vpceResourceName, &vpce),
					testAccCheckAWSRouteTableWaitForVpcEndpointRoute(&routeTable, &vpce),
					// Refresh the route table once the VPC endpoint route is present.
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func TestAccAWSRouteTable_MultipleRoutes(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_route_table.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	igwResourceName := "aws_internet_gateway.test"
	instanceResourceName := "aws_instance.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr1 := "10.2.0.0/16"
	destinationCidr2 := "10.3.0.0/16"
	destinationCidr3 := "10.4.0.0/16"
	destinationCidr4 := "2001:db8::/122"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteTableConfigMultipleRoutes(rName,
					"cidr_block", destinationCidr1, "gateway_id", igwResourceName,
					"cidr_block", destinationCidr2, "instance_id", instanceResourceName,
					"ipv6_cidr_block", destinationCidr4, "egress_only_gateway_id", eoigwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 5),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "3"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "gateway_id", igwResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr2, "instance_id", instanceResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "ipv6_cidr_block", destinationCidr4, "egress_only_gateway_id", eoigwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSRouteTableConfigMultipleRoutes(rName,
					"cidr_block", destinationCidr1, "vpc_peering_connection_id", pcxResourceName,
					"cidr_block", destinationCidr3, "instance_id", instanceResourceName,
					"ipv6_cidr_block", destinationCidr4, "egress_only_gateway_id", eoigwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 5),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "3"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "vpc_peering_connection_id", pcxResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr3, "instance_id", instanceResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "ipv6_cidr_block", destinationCidr4, "egress_only_gateway_id", eoigwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccAWSRouteTableConfigMultipleRoutes(rName,
					"ipv6_cidr_block", destinationCidr4, "vpc_peering_connection_id", pcxResourceName,
					"cidr_block", destinationCidr3, "gateway_id", igwResourceName,
					"cidr_block", destinationCidr2, "instance_id", instanceResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 5),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "3"),
					testAccCheckAWSRouteTableRoute(resourceName, "ipv6_cidr_block", destinationCidr4, "vpc_peering_connection_id", pcxResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr3, "gateway_id", igwResourceName, "id"),
					testAccCheckAWSRouteTableRoute(resourceName, "cidr_block", destinationCidr2, "instance_id", instanceResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
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

func testAccCheckRouteTableExists(n string, v *ec2.RouteTable) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.RouteTables) == 0 {
			return fmt.Errorf("RouteTable not found")
		}

		*v = *resp.RouteTables[0]

		return nil
	}
}

func testAccCheckRouteTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route_table" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
			RouteTableIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err == nil {
			if len(resp.RouteTables) > 0 {
				return fmt.Errorf("still exist.")
			}

			return nil
		}

		// Verify the error is what we want
		if !isAWSErr(err, "InvalidRouteTableID.NotFound", "") {
			return err
		}
	}

	return nil
}

func testAccCheckAWSRouteTableNumberOfRoutes(routeTable *ec2.RouteTable, n int) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if len := len(routeTable.Routes); len != n {
			return fmt.Errorf("Route Table has incorrect number of routes (Expected=%d, Actual=%d)\n", n, len)
		}

		return nil
	}
}

func testAccCheckAWSRouteTableRoute(resourceName, destinationAttr, destination, targetAttr, targetResourceName, targetResourceAttr string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[targetResourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", targetResourceName)
		}

		target := rs.Primary.Attributes[targetResourceAttr]
		if target == "" {
			return fmt.Errorf("Not found: %s.%s", targetResourceName, targetResourceAttr)
		}

		return resource.TestCheckTypeSetElemNestedAttrs(resourceName, "route.*", map[string]string{
			destinationAttr: destination,
			targetAttr:      target,
		})(s)
	}
}

// testAccCheckAWSRouteTableWaitForVpcEndpointRoute returns a TestCheckFunc which waits for
// a route to the specified VPC endpoint's prefix list to appear in the specified route table.
func testAccCheckAWSRouteTableWaitForVpcEndpointRoute(routeTable *ec2.RouteTable, vpce *ec2.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribePrefixLists(&ec2.DescribePrefixListsInput{
			Filters: buildEC2AttributeFilterList(map[string]string{
				"prefix-list-name": aws.StringValue(vpce.ServiceName),
			}),
		})
		if err != nil {
			return err
		}

		if resp == nil || len(resp.PrefixLists) == 0 {
			return fmt.Errorf("Prefix List not found")
		}

		plId := aws.StringValue(resp.PrefixLists[0].PrefixListId)

		err = resource.Retry(3*time.Minute, func() *resource.RetryError {
			resp, err := conn.DescribeRouteTables(&ec2.DescribeRouteTablesInput{
				RouteTableIds: []*string{routeTable.RouteTableId},
			})
			if err != nil {
				return resource.NonRetryableError(err)
			}
			if resp == nil || len(resp.RouteTables) == 0 {
				return resource.NonRetryableError(fmt.Errorf("Route Table not found"))
			}

			for _, route := range resp.RouteTables[0].Routes {
				if aws.StringValue(route.DestinationPrefixListId) == plId {
					return nil
				}
			}

			return resource.RetryableError(fmt.Errorf("Route not found"))
		})

		return err
	}
}

func testAccAWSRouteTableConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id
}
`, rName)
}

func testAccAWSRouteTableConfigSubnetAssociation(rName string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
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

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table_association" "test" {
  route_table_id = aws_route_table.test.id
  subnet_id      = aws_subnet.test.id
}
`, rName))
}

func testAccAWSRouteTableConfigIpv4InternetGateway(rName, destinationCidr1, destinationCidr2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

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

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }

  route {
    cidr_block = %[3]q
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr1, destinationCidr2)
}

func testAccAWSRouteTableConfigIpv6EgressOnlyInternetGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    ipv6_cidr_block        = %[2]q
    egress_only_gateway_id = aws_egress_only_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigIpv4Instance(rName, destinationCidr string) string {
	return composeConfig(
		testAccLatestAmazonNatInstanceAmiConfig(),
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
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

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-nat-instance.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block  = %[2]q
    instance_id = aws_instance.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr))
}

func testAccAWSRouteTableConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccAWSRouteTableConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccAWSRouteTableConfigIpv4VpcPeeringConnection(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block                = %[2]q
    vpc_peering_connection_id = aws_vpc_peering_connection.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigVgwRoutePropagation(rName, vgwResourceName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test1" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test2" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway_attachment" "test" {
  vpc_id         = aws_vpc.test.id
  vpn_gateway_id = %[2]s.id
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  propagating_vgws = [aws_vpn_gateway_attachment.test.vpn_gateway_id]

  tags = {
    Name = %[1]q
  }
}
`, rName, vgwResourceName)
}

func testAccRouteTableConfigNoDestination(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		fmt.Sprintf(`
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    instance_id = aws_instance.test.id
  }

  tags = {
    Name = %[1]q
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

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccRouteTableConfigNoTarget(rName string) string {
	return fmt.Sprintf(`
resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block = "10.1.0.0/16"
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigRouteConfigModeNoBlocks(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

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

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigRouteConfigModeZeroed(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

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

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route = []

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigIpv4TransitGateway(rName, destinationCidr string) string {
	return composeConfig(testAccAvailableAZsNoOptInDefaultExcludeConfig(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block         = %[2]q
    transit_gateway_id = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr))
}

func testAccAWSRouteTableConfigRouteIpv4VpcEndpointId(rName, destinationCidr string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.10.10.0/25"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, 0)
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_lb" "test" {
  load_balancer_type = "gateway"
  name               = %[1]q

  subnet_mapping {
    subnet_id = aws_subnet.test.id
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required        = false
  allowed_principals         = [data.aws_caller_identity.current.arn]
  gateway_load_balancer_arns = [aws_lb.test.arn]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name      = aws_vpc_endpoint_service.test.service_name
  subnet_ids        = [aws_subnet.test.id]
  vpc_endpoint_type = aws_vpc_endpoint_service.test.service_type
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block      = %[2]q
    vpc_endpoint_id = aws_vpc_endpoint.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr))
}

func testAccAWSRouteTableConfigRouteIpv4LocalGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
data "aws_ec2_local_gateways" "all" {}

data "aws_ec2_local_gateway" "first" {
  id = tolist(data.aws_ec2_local_gateways.all.ids)[0]
}

data "aws_ec2_local_gateway_route_tables" "all" {}

data "aws_ec2_local_gateway_route_table" "first" {
  local_gateway_route_table_id = tolist(data.aws_ec2_local_gateway_route_tables.all.ids)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block       = %[2]q
    local_gateway_id = data.aws_ec2_local_gateway.first.id
  }

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr string, ipv6Route bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  assign_generated_ipv6_cidr_block = true

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

locals {
  ipv6             = %[4]t
  destination      = %[2]q
  destination_ipv6 = %[3]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block      = local.ipv6 ? "" : local.destination
    ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : ""
    gateway_id      = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr, destinationIpv6Cidr, ipv6Route)
}

func testAccAWSRouteTableConfigIpv4NatGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.1.1.0/24"
  vpc_id     = aws_vpc.test.id

  map_public_ip_on_launch = true

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

resource "aws_eip" "test" {
  vpc = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_nat_gateway" "test" {
  allocation_id = aws_eip.test.id
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    cidr_block     = %[2]q
    nat_gateway_id = aws_nat_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigIpv6NetworkInterfaceUnattached(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block      = "10.1.1.0/24"
  vpc_id          = aws_vpc.test.id
  ipv6_cidr_block = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  route {
    ipv6_cidr_block      = %[2]q
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationCidr)
}

func testAccAWSRouteTableConfigVpcMultipleCidrs(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_ipv4_cidr_block_association" "test" {
  vpc_id     = aws_vpc.test.id
  cidr_block = "172.2.0.0/16"
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc_ipv4_cidr_block_association.test.vpc_id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigVpcClassicLink(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block         = "10.1.0.0/16"
  enable_classiclink = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAWSRouteTableConfigGatewayVpcEndpoint(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]
}
`, rName)
}

func testAccAWSRouteTableConfigMultipleRoutes(rName,
	destinationAttr1, destinationValue1, targetAttribute1, targetValue1,
	destinationAttr2, destinationValue2, targetAttribute2, targetValue2,
	destinationAttr3, destinationValue3, targetAttribute3, targetValue3 string) string {
	return composeConfig(
		testAccLatestAmazonNatInstanceAmiConfig(),
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block = "10.0.0.0/16"

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

resource "aws_egress_only_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-nat-instance.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.test.id
  peer_vpc_id = aws_vpc.target.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

locals {
  routes = [
    {
      destination_attr  = %[2]q
      destination_value = %[3]q
      target_attr       = %[4]q
      target_value      = %[5]s.id
    },
    {
      destination_attr  = %[6]q
      destination_value = %[7]q
      target_attr       = %[8]q
      target_value      = %[9]s.id
    },
    {
      destination_attr  = %[10]q
      destination_value = %[11]q
      target_attr       = %[12]q
      target_value      = %[13]s.id
    }
  ]
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  dynamic "route" {
    for_each = local.routes
    content {
      # Destination.
      cidr_block      = (route.value["destination_attr"] == "cidr_block") ? route.value["destination_value"] : null
      ipv6_cidr_block = (route.value["destination_attr"] == "ipv6_cidr_block") ? route.value["destination_value"] : null

      # Target.
      egress_only_gateway_id    = (route.value["target_attr"] == "egress_only_gateway_id") ? route.value["target_value"] : null
      gateway_id                = (route.value["target_attr"] == "gateway_id") ? route.value["target_value"] : null
      instance_id               = (route.value["target_attr"] == "instance_id") ? route.value["target_value"] : null
      local_gateway_id          = (route.value["target_attr"] == "local_gateway_id") ? route.value["target_value"] : null
      nat_gateway_id            = (route.value["target_attr"] == "nat_gateway_id") ? route.value["target_value"] : null
      network_interface_id      = (route.value["target_attr"] == "network_interface_id") ? route.value["target_value"] : null
      transit_gateway_id        = (route.value["target_attr"] == "transit_gateway_id") ? route.value["target_value"] : null
      vpc_endpoint_id           = (route.value["target_attr"] == "vpc_endpoint_id") ? route.value["target_value"] : null
      vpc_peering_connection_id = (route.value["target_attr"] == "vpc_peering_connection_id") ? route.value["target_value"] : null
    }
  }

  tags = {
    Name = %[1]q
  }
}
`, rName, destinationAttr1, destinationValue1, targetAttribute1, targetValue1, destinationAttr2, destinationValue2, targetAttribute2, targetValue2, destinationAttr3, destinationValue3, targetAttribute3, targetValue3))
}

// testAccLatestAmazonNatInstanceAmiConfig returns the configuration for a data source that
// describes the latest Amazon NAT instance AMI.
// See https://docs.aws.amazon.com/vpc/latest/userguide/VPC_NAT_Instance.html#nat-instance-ami.
// The data source is named 'amzn-ami-nat-instance'.
func testAccLatestAmazonNatInstanceAmiConfig() string {
	return `
data "aws_ami" "amzn-ami-nat-instance" {
  most_recent = true
  owners      = ["amazon"]

  filter {
    name   = "name"
    values = ["amzn-ami-vpc-nat-*"]
  }
}
`
}
