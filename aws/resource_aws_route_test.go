package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

// IPv4 to Internet Gateway.
func TestAccAWSRoute_basic(t *testing.T) {
	var route ec2.Route
	var routeTable ec2.RouteTable
	resourceName := "aws_route.test"
	igwResourceName := "aws_internet_gateway.test"
	rtResourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(rtResourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_disappears(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRoute(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute_disappears_RouteTable(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	rtResourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsRouteTable(), rtResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_EgressOnlyInternetGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6EgressOnlyInternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "egress_only_gateway_id", eoigwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				// Verify that expanded form of the destination CIDR causes no diff.
				Config:   testAccAWSRouteConfigIpv6EgressOnlyInternetGateway(rName, "::0/0"),
				PlanOnly: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_InternetGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_Instance(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	instanceResourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6Instance(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_NetworkInterface_Unattached(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6NetworkInterfaceUnattached(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateBlackhole),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_VpcPeeringConnection(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6VpcPeeringConnection(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_VpnGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6VpnGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_VpnGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4VpnGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_Instance(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	instanceResourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4Instance(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_NetworkInterface_Unattached(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4NetworkInterfaceUnattached(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateBlackhole),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_NetworkInterface_Attached(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	instanceResourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4NetworkInterfaceAttached(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_NetworkInterface_TwoAttachments(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eni1ResourceName := "aws_network_interface.test1"
	eni2ResourceName := "aws_network_interface.test2"
	instanceResourceName := "aws_instance.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4NetworkInterfaceTwoAttachments(rName, destinationCidr, eni1ResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eni1ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4NetworkInterfaceTwoAttachments(rName, destinationCidr, eni2ResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eni2ResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_VpcPeeringConnection(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4VpcPeeringConnection(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_NatGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	ngwResourceName := "aws_nat_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4NatGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", ngwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_DoesNotCrashWithVpcEndpoint(t *testing.T) {
	var route ec2.Route
	var routeTable ec2.RouteTable
	resourceName := "aws_route.test"
	rtResourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigWithVpcEndpoint(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(rtResourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					testAccCheckAWSRouteExists(resourceName, &route),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_TransitGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", tgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_TransitGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", tgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_CarrierGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	cgwResourceName := "aws_ec2_carrier_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "172.16.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSWavelengthZoneAvailable(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4CarrierGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "carrier_gateway_id", cgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_LocalGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	localGatewayDataSourceName := "data.aws_ec2_local_gateway.first"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "172.16.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteResourceConfigIpv4LocalGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayDataSourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_To_LocalGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	localGatewayDataSourceName := "data.aws_ec2_local_gateway.first"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "2002:bc9:1234:1a00::/56"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteResourceConfigIpv6LocalGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayDataSourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_ConditionalCidrBlock(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"
	destinationIpv6Cidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationIpv6Cidr),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_Update_Target(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	instanceResourceName := "aws_instance.test"
	igwResourceName := "aws_internet_gateway.test"
	eniResourceName := "aws_network_interface.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	ngwResourceName := "aws_nat_gateway.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID, "elasticloadbalancing"),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "instance_id", instanceResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "gateway_id", vgwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "gateway_id", igwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "nat_gateway_id", ngwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", ngwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "network_interface_id", eniResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateBlackhole),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "transit_gateway_id", tgwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", tgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "vpc_endpoint_id", vpcEndpointResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_id", vpcEndpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, "vpc_peering_connection_id", pcxResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv6_Update_Target(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	instanceResourceName := "aws_instance.test"
	igwResourceName := "aws_internet_gateway.test"
	eniResourceName := "aws_network_interface.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "::/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv6FlexiTarget(rName, destinationCidr, "instance_id", instanceResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv6FlexiTarget(rName, destinationCidr, "gateway_id", vgwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv6FlexiTarget(rName, destinationCidr, "gateway_id", igwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv6FlexiTarget(rName, destinationCidr, "egress_only_gateway_id", eoigwResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "egress_only_gateway_id", eoigwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv6FlexiTarget(rName, destinationCidr, "network_interface_id", eniResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateBlackhole),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				Config: testAccAWSRouteConfigIpv6FlexiTarget(rName, destinationCidr, "vpc_peering_connection_id", pcxResourceName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_IPv4_To_VpcEndpoint(t *testing.T) {
	var route ec2.Route
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_route.test"
	vpcEndpointResourceName := "aws_vpc_endpoint.test"
	destinationCidr := "172.16.1.0/24"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckElbv2GatewayLoadBalancer(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID, "elasticloadbalancing"),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteResourceConfigIpv4VpcEndpoint(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", destinationCidr),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_prefix_list_id", ""),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_endpoint_id", vpcEndpointResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

// https://github.com/terraform-providers/terraform-provider-aws/issues/11455.
func TestAccAWSRoute_LocalRoute(t *testing.T) {
	var routeTable ec2.RouteTable
	var vpc ec2.Vpc
	resourceName := "aws_route.test"
	rtResourceName := "aws_route_table.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigIpv4NoRoute(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcExists(vpcResourceName, &vpc),
					testAccCheckRouteTableExists(rtResourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
				),
			},
			{
				Config:       testAccAWSRouteConfigIpv4LocalRoute(rName),
				ResourceName: resourceName,
				ImportState:  true,
				ImportStateIdFunc: func(rt *ec2.RouteTable, v *ec2.Vpc) resource.ImportStateIdFunc {
					return func(s *terraform.State) (string, error) {
						return fmt.Sprintf("%s_%s", aws.StringValue(rt.RouteTableId), aws.StringValue(v.CidrBlock)), nil
					}
				}(&routeTable, &vpc),
				// Don't verify the state as the local route isn't actually in the pre-import state.
				// Just running ImportState verifies that we can import a local route.
				ImportStateVerify: false,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_InternetGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	igwResourceName := "aws_internet_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListInternetGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_VpnGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	vgwResourceName := "aws_vpn_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListVpnGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "gateway_id", vgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_Instance(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	instanceResourceName := "aws_instance.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListInstance(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", instanceResourceName, "primary_network_interface_id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_NetworkInterface_Unattached(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListNetworkInterfaceUnattached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateBlackhole),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_NetworkInterface_Attached(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eniResourceName := "aws_network_interface.test"
	instanceResourceName := "aws_instance.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListNetworkInterfaceAttached(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "instance_id", instanceResourceName, "id"),
					testAccCheckResourceAttrAccountID(resourceName, "instance_owner_id"),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "network_interface_id", eniResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_VpcPeeringConnection(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	pcxResourceName := "aws_vpc_peering_connection.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListVpcPeeringConnection(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_peering_connection_id", pcxResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_NatGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	ngwResourceName := "aws_nat_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListNatGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "nat_gateway_id", ngwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_TransitGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListTransitGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", tgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_CarrierGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	cgwResourceName := "aws_ec2_carrier_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckEc2ManagedPrefixList(t)
			testAccPreCheckAWSWavelengthZoneAvailable(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListCarrierGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttrPair(resourceName, "carrier_gateway_id", cgwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_LocalGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	localGatewayDataSourceName := "data.aws_ec2_local_gateway.first"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccPreCheckEc2ManagedPrefixList(t)
			testAccPreCheckAWSOutpostsOutposts(t)
		},
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListLocalGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "egress_only_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayDataSourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSRoute_PrefixList_To_EgressOnlyInternetGateway(t *testing.T) {
	var route ec2.Route
	resourceName := "aws_route.test"
	eoigwResourceName := "aws_egress_only_internet_gateway.test"
	plResourceName := "aws_ec2_managed_prefix_list.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckEc2ManagedPrefixList(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSRouteDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSRouteConfigPrefixListEgressOnlyInternetGateway(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSRouteExists(resourceName, &route),
					resource.TestCheckResourceAttr(resourceName, "carrier_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_cidr_block", ""),
					resource.TestCheckResourceAttr(resourceName, "destination_ipv6_cidr_block", ""),
					resource.TestCheckResourceAttrPair(resourceName, "destination_prefix_list_id", plResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "egress_only_gateway_id", eoigwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_id", ""),
					resource.TestCheckResourceAttr(resourceName, "instance_owner_id", ""),
					resource.TestCheckResourceAttr(resourceName, "local_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "nat_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "network_interface_id", ""),
					resource.TestCheckResourceAttr(resourceName, "origin", ec2.RouteOriginCreateRoute),
					resource.TestCheckResourceAttr(resourceName, "state", ec2.RouteStateActive),
					resource.TestCheckResourceAttr(resourceName, "transit_gateway_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_endpoint_id", ""),
					resource.TestCheckResourceAttr(resourceName, "vpc_peering_connection_id", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSRouteImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAWSRouteExists(n string, v *ec2.Route) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		var route *ec2.Route
		var err error
		if v := rs.Primary.Attributes["destination_cidr_block"]; v != "" {
			route, err = finder.RouteByIPv4Destination(conn, rs.Primary.Attributes["route_table_id"], v)
		} else if v := rs.Primary.Attributes["destination_ipv6_cidr_block"]; v != "" {
			route, err = finder.RouteByIPv6Destination(conn, rs.Primary.Attributes["route_table_id"], v)
		} else if v := rs.Primary.Attributes["destination_prefix_list_id"]; v != "" {
			route, err = finder.RouteByPrefixListIDDestination(conn, rs.Primary.Attributes["route_table_id"], v)
		}

		if err != nil {
			return err
		}

		*v = *route

		return nil
	}
}

func testAccCheckAWSRouteDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_route" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		var err error
		if v := rs.Primary.Attributes["destination_cidr_block"]; v != "" {
			_, err = finder.RouteByIPv4Destination(conn, rs.Primary.Attributes["route_table_id"], v)
		} else if v := rs.Primary.Attributes["destination_ipv6_cidr_block"]; v != "" {
			_, err = finder.RouteByIPv6Destination(conn, rs.Primary.Attributes["route_table_id"], v)
		} else if v := rs.Primary.Attributes["destination_prefix_list_id"]; v != "" {
			_, err = finder.RouteByPrefixListIDDestination(conn, rs.Primary.Attributes["route_table_id"], v)
		}

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Route still exists")
	}

	return nil
}

func testAccAWSRouteImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("not found: %s", resourceName)
		}

		destination := rs.Primary.Attributes["destination_cidr_block"]
		if v, ok := rs.Primary.Attributes["destination_ipv6_cidr_block"]; ok && v != "" {
			destination = v
		}
		if v, ok := rs.Primary.Attributes["destination_prefix_list_id"]; ok && v != "" {
			destination = v
		}

		return fmt.Sprintf("%s_%s", rs.Primary.Attributes["route_table_id"], destination), nil
	}
}

func testAccAWSRouteConfigIpv4InternetGateway(rName, destinationCidr string) string {
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

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  gateway_id             = aws_internet_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigIpv6InternetGateway(rName, destinationCidr string) string {
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

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  gateway_id                  = aws_internet_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigIpv6NetworkInterfaceUnattached(rName, destinationCidr string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  network_interface_id        = aws_network_interface.test.id
}
`, rName, destinationCidr))
}

func testAccAWSRouteConfigIpv6Instance(rName, destinationCidr string) string {
	return composeConfig(
		testAccLatestAmazonNatInstanceAmiConfig(),
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-nat-instance.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  ipv6_address_count = 1

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

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  instance_id                 = aws_instance.test.id
}
`, rName, destinationCidr))
}

func testAccAWSRouteConfigIpv6VpcPeeringConnection(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  vpc_peering_connection_id   = aws_vpc_peering_connection.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigIpv6EgressOnlyInternetGateway(rName, destinationCidr string) string {
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  egress_only_gateway_id      = aws_egress_only_internet_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigWithVpcEndpoint(rName string) string {
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

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = "10.3.0.0/16"
  gateway_id             = aws_internet_gateway.test.id

  # Forcing endpoint to create before route - without this the crash is a race.
  depends_on = [aws_vpc_endpoint.test]
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_route_table.test.id]
}
`, rName)
}

func testAccAWSRouteConfigIpv4TransitGateway(rName, destinationCidr string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		fmt.Sprintf(`
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_cidr_block = %[2]q
  route_table_id         = aws_route_table.test.id
  transit_gateway_id     = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
}
`, rName, destinationCidr))
}

func testAccAWSRouteConfigIpv6TransitGateway(rName, destinationCidr string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  destination_ipv6_cidr_block = %[2]q
  route_table_id              = aws_route_table.test.id
  transit_gateway_id          = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
}
`, rName, destinationCidr))
}

func testAccAWSRouteConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr string, ipv6Route bool) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
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

resource "aws_route_table" "test" {
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

resource "aws_route" "test" {
  route_table_id = aws_route_table.test.id
  gateway_id     = aws_internet_gateway.test.id

  destination_cidr_block      = local.ipv6 ? "" : local.destination
  destination_ipv6_cidr_block = local.ipv6 ? local.destination_ipv6 : ""
}
`, rName, destinationCidr, destinationIpv6Cidr, ipv6Route)
}

func testAccAWSRouteConfigIpv4Instance(rName, destinationCidr string) string {
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  instance_id            = aws_instance.test.id
}
`, rName, destinationCidr))
}

func testAccAWSRouteConfigIpv4NetworkInterfaceUnattached(rName, destinationCidr string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
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

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

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

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  network_interface_id   = aws_network_interface.test.id
}
`, rName, destinationCidr))
}

func testAccAWSRouteResourceConfigIpv4LocalGateway(rName, destinationCidr string) string {
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

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  local_gateway_id       = data.aws_ec2_local_gateway.first.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteResourceConfigIpv6LocalGateway(rName, destinationCidr string) string {
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
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

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

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  local_gateway_id            = data.aws_ec2_local_gateway.first.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigIpv4NetworkInterfaceAttached(rName, destinationCidr string) string {
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

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-nat-instance.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  network_interface {
    device_index         = 0
    network_interface_id = aws_network_interface.test.id
  }

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

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  network_interface_id   = aws_network_interface.test.id

  # Wait for the ENI attachment.
  depends_on = [aws_instance.test]
}
`, rName, destinationCidr))
}

func testAccAWSRouteConfigIpv4NetworkInterfaceTwoAttachments(rName, destinationCidr, targetResourceName string) string {
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

resource "aws_network_interface" "test1" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test2" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-nat-instance.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  network_interface {
    device_index         = 0
    network_interface_id = aws_network_interface.test1.id
  }

  network_interface {
    device_index         = 1
    network_interface_id = aws_network_interface.test2.id
  }

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

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  network_interface_id   = %[3]s.id

  # Wait for the ENI attachment.
  depends_on = [aws_instance.test]
}
`, rName, destinationCidr, targetResourceName))
}

func testAccAWSRouteConfigIpv4VpcPeeringConnection(rName, destinationCidr string) string {
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id            = aws_route_table.test.id
  destination_cidr_block    = %[2]q
  vpc_peering_connection_id = aws_vpc_peering_connection.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigIpv4NatGateway(rName, destinationCidr string) string {
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  nat_gateway_id         = aws_nat_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigIpv4VpnGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
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

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  gateway_id             = aws_vpn_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteConfigIpv6VpnGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
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

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q
  gateway_id                  = aws_vpn_gateway.test.id
}
`, rName, destinationCidr)
}

func testAccAWSRouteResourceConfigIpv4VpcEndpoint(rName, destinationCidr string) string {
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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q
  vpc_endpoint_id        = aws_vpc_endpoint.test.id
}
`, rName, destinationCidr))
}

func testAccAWSRouteConfigIpv4FlexiTarget(rName, destinationCidr, targetAttribute, targetValue string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
locals {
  target_attr  = %[3]q
  target_value = %[4]s.id
}

resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
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

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]

  map_public_ip_on_launch = true

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

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

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

data "aws_caller_identity" "current" {}

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = %[2]q

  egress_only_gateway_id    = (local.target_attr == "egress_only_gateway_id") ? local.target_value : null
  gateway_id                = (local.target_attr == "gateway_id") ? local.target_value : null
  instance_id               = (local.target_attr == "instance_id") ? local.target_value : null
  local_gateway_id          = (local.target_attr == "local_gateway_id") ? local.target_value : null
  nat_gateway_id            = (local.target_attr == "nat_gateway_id") ? local.target_value : null
  network_interface_id      = (local.target_attr == "network_interface_id") ? local.target_value : null
  transit_gateway_id        = (local.target_attr == "transit_gateway_id") ? local.target_value : null
  vpc_endpoint_id           = (local.target_attr == "vpc_endpoint_id") ? local.target_value : null
  vpc_peering_connection_id = (local.target_attr == "vpc_peering_connection_id") ? local.target_value : null
}
`, rName, destinationCidr, targetAttribute, targetValue))
}

func testAccAWSRouteConfigIpv6FlexiTarget(rName, destinationCidr, targetAttribute, targetValue string) string {
	return composeConfig(
		testAccLatestAmazonLinuxHvmEbsAmiConfig(),
		testAccAvailableAZsNoOptInConfig(),
		testAccAvailableEc2InstanceTypeForAvailabilityZone("data.aws_availability_zones.available.names[0]", "t3.micro", "t2.micro"),
		fmt.Sprintf(`
locals {
  target_attr  = %[3]q
  target_value = %[4]s.id
}

resource "aws_vpc" "test" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
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

resource "aws_subnet" "test" {
  cidr_block        = "10.1.1.0/24"
  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[0]
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 1)

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type
  subnet_id     = aws_subnet.test.id

  ipv6_address_count = 1

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

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "target" {
  cidr_block                       = "10.0.0.0/16"
  assign_generated_ipv6_cidr_block = true

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

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id              = aws_route_table.test.id
  destination_ipv6_cidr_block = %[2]q

  egress_only_gateway_id    = (local.target_attr == "egress_only_gateway_id") ? local.target_value : null
  gateway_id                = (local.target_attr == "gateway_id") ? local.target_value : null
  instance_id               = (local.target_attr == "instance_id") ? local.target_value : null
  local_gateway_id          = (local.target_attr == "local_gateway_id") ? local.target_value : null
  nat_gateway_id            = (local.target_attr == "nat_gateway_id") ? local.target_value : null
  network_interface_id      = (local.target_attr == "network_interface_id") ? local.target_value : null
  transit_gateway_id        = (local.target_attr == "transit_gateway_id") ? local.target_value : null
  vpc_endpoint_id           = (local.target_attr == "vpc_endpoint_id") ? local.target_value : null
  vpc_peering_connection_id = (local.target_attr == "vpc_peering_connection_id") ? local.target_value : null
}
`, rName, destinationCidr, targetAttribute, targetValue))
}

func testAccAWSRouteConfigIpv4NoRoute(rName string) string {
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
`, rName)
}

func testAccAWSRouteConfigIpv4LocalRoute(rName string) string {
	return composeConfig(
		testAccAWSRouteConfigIpv4NoRoute(rName),
		`
resource "aws_route" "test" {
  route_table_id         = aws_route_table.test.id
  destination_cidr_block = aws_vpc.test.cidr_block
  gateway_id             = "local"
}
`)
}

func testAccAWSRouteConfigPrefixListInternetGateway(rName string) string {
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

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  gateway_id                 = aws_internet_gateway.test.id
}
`, rName)
}

func testAccAWSRouteConfigPrefixListVpnGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpn_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  gateway_id                 = aws_vpn_gateway.test.id
}
`, rName)
}

func testAccAWSRouteConfigPrefixListInstance(rName string) string {
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

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  instance_id                = aws_instance.test.id
}
`, rName))
}

func testAccAWSRouteConfigPrefixListNetworkInterfaceUnattached(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
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

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  network_interface_id       = aws_network_interface.test.id
}
`, rName))
}

func testAccAWSRouteConfigPrefixListNetworkInterfaceAttached(rName string) string {
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

resource "aws_network_interface" "test" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_instance" "test" {
  ami           = data.aws_ami.amzn-ami-nat-instance.id
  instance_type = data.aws_ec2_instance_type_offering.available.instance_type

  network_interface {
    device_index         = 0
    network_interface_id = aws_network_interface.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  network_interface_id       = aws_network_interface.test.id

  # Wait for the ENI attachment.
  depends_on = [aws_instance.test]
}
`, rName))
}

func testAccAWSRouteConfigPrefixListVpcPeeringConnection(rName string) string {
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

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  vpc_peering_connection_id  = aws_vpc_peering_connection.test.id
}
`, rName)
}

func testAccAWSRouteConfigPrefixListNatGateway(rName string) string {
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

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  nat_gateway_id             = aws_nat_gateway.test.id
}
`, rName)
}

func testAccAWSRouteConfigPrefixListTransitGateway(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInDefaultExcludeConfig(),
		fmt.Sprintf(`
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

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  transit_gateway_id         = aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id
}
`, rName))
}

func testAccAWSRouteConfigPrefixListCarrierGateway(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  carrier_gateway_id         = aws_ec2_carrier_gateway.test.id
}
`, rName)
}

func testAccAWSRouteConfigPrefixListLocalGateway(rName string) string {
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

resource "aws_ec2_local_gateway_route_table_vpc_association" "example" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.first.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv4"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_ec2_local_gateway_route_table_vpc_association.example]
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  local_gateway_id           = data.aws_ec2_local_gateway.first.id
}
`, rName)
}

func testAccAWSRouteConfigPrefixListEgressOnlyInternetGateway(rName string) string {
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

resource "aws_ec2_managed_prefix_list" "test" {
  address_family = "IPv6"
  max_entries    = 1
  name           = %[1]q
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_route" "test" {
  route_table_id             = aws_route_table.test.id
  destination_prefix_list_id = aws_ec2_managed_prefix_list.test.id
  egress_only_gateway_id     = aws_egress_only_internet_gateway.test.id
}
`, rName)
}

func testAccAWSRouteConfigIpv4CarrierGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_carrier_gateway" "test" {
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

resource "aws_route" "test" {
  destination_cidr_block = %[2]q
  route_table_id         = aws_route_table.test.id
  carrier_gateway_id     = aws_ec2_carrier_gateway.test.id
}
`, rName, destinationCidr)
}
