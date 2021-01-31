package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSDefaultRouteTable_basic(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			// Verify non-existent Route Table ID behavior
			{
				Config:      testAccDefaultRouteTableConfigDefaultRouteTableId("rtb-00000000"),
				ExpectError: regexp.MustCompile(`EC2 Default Route Table \(rtb-00000000\): not found`),
			},
			// Verify invalid Route Table ID behavior
			{
				Config:      testAccDefaultRouteTableConfigDefaultRouteTableId("vpc-00000000"),
				ExpectError: regexp.MustCompile(`EC2 Default Route Table \(vpc-00000000\): not found`),
			},
			{
				Config: testAccDefaultRouteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_disappears_Vpc(t *testing.T) {
	var routeTable ec2.RouteTable
	var vpc ec2.Vpc
	resourceName := "aws_default_route_table.test"
	vpcResourceName := "aws_vpc.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckVpcExists(vpcResourceName, &vpc),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpc(), vpcResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_Route_ConfigMode(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDefaultRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigIpv4InternetGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDefaultRouteTableConfigNoRouteBlock(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					// The route block from the previous step should still be
					// present, because no blocks means "ignore existing blocks".
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				Config: testAccDefaultRouteTableConfigRouteBlocksExplicitZero(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 1),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					// This config uses attribute syntax to set zero routes
					// explicitly, so should remove the one we created before.
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_swap(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rtResourceName := "aws_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr1 := "10.2.0.0/16"
	destinationCidr2 := "10.3.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDefaultRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigIpv4InternetGateway(rName, destinationCidr1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
					resource.TestCheckResourceAttr(resourceName, "propagating_vgws.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.Name", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},

			// This config will swap out the original Default Route Table and replace
			// it with the custom route table. While this is not advised, it's a
			// behavior that may happen, in which case a follow up plan will show (in
			// this case) a diff as the table now needs to be updated to match the
			// config
			{
				Config: testAccDefaultRouteTableConfigSwap(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "gateway_id", igwResourceName, "id"),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccDefaultRouteTableConfigSwap(rName, destinationCidr1, destinationCidr2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr1, "gateway_id", igwResourceName, "id"),
					resource.TestCheckResourceAttrPair(resourceName, "id", rtResourceName, "id"),
				),
				// Follow up plan will now show a diff as the destination CIDR on the aws_route_table
				// (now also the aws_default_route_table) will change from destinationCidr1 to destinationCidr2.
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_IPv4_To_TransitGateway(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	tgwResourceName := "aws_ec2_transit_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigIpv4TransitGateway(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr, "transit_gateway_id", tgwResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_IPv4_To_VpcEndpoint(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	vpceResourceName := "aws_vpc_endpoint.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "0.0.0.0/0"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigIpv4VpcEndpoint(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 2),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr, "vpc_endpoint_id", vpceResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			// Default route tables do not currently have a method to remove routes during deletion.
			// VPC Endpoints will not delete unless the route is removed prior, otherwise will error:
			// InvalidParameter: Endpoint must be removed from route table before deletion
			{
				Config: testAccDefaultRouteTableConfigIpv4VpcEndpointNoRoute(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
				),
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_VpcEndpointAssociation(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	igwResourceName := "aws_internet_gateway.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")
	destinationCidr := "10.2.0.0/16"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDefaultRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigVpcEndpointAssociation(rName, destinationCidr),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckAWSRouteTableNumberOfRoutes(&routeTable, 3),
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr, "gateway_id", igwResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_tags(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
				),
			},
			{
				Config: testAccDefaultRouteTableConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccDefaultRouteTableConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_ConditionalCidrBlock(t *testing.T) {
	var routeTable ec2.RouteTable
	resourceName := "aws_default_route_table.test"
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
				Config: testAccDefaultRouteTableConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckDefaultRouteTableRoute(resourceName, "cidr_block", destinationCidr, "gateway_id", igwResourceName, "id"),
				),
			},
			{
				Config: testAccDefaultRouteTableConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable),
					testAccCheckDefaultRouteTableRoute(resourceName, "ipv6_cidr_block", destinationIpv6Cidr, "gateway_id", igwResourceName, "id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckDefaultRouteTableDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_default_route_table" {
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

func testAccCheckDefaultRouteTableRoute(resourceName, destinationAttr, destination, targetAttr, targetResourceName, targetResourceAttr string) resource.TestCheckFunc {
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

func testAccDefaultRouteTableConfigDefaultRouteTableId(defaultRouteTableId string) string {
	return fmt.Sprintf(`
resource "aws_default_route_table" "test" {
  default_route_table_id = %[1]q
}
`, defaultRouteTableId)
}

func testAccDefaultRouteTableConfigBasic(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id
}
`, rName)
}

func testAccDefaultRouteTableConfigIpv4InternetGateway(rName, destinationCidr string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }

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
`, rName, destinationCidr)
}

func testAccDefaultRouteTableConfigNoRouteBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccDefaultRouteTableConfigRouteBlocksExplicitZero(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route = []

  tags = {
    Name = %[1]q
  }
}

resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}`, rName)
}

func testAccDefaultRouteTableConfigSwap(rName, destinationCidr1, destinationCidr2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }

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
    cidr_block = %[3]q
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}

resource "aws_main_route_table_association" "test" {
  vpc_id         = aws_vpc.test.id
  route_table_id = aws_route_table.test.id
}
`, rName, destinationCidr1, destinationCidr2)
}

func testAccDefaultRouteTableConfigIpv4TransitGateway(rName, destinationCidr string) string {
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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

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

func testAccDefaultRouteTableConfigIpv4VpcEndpoint(rName, destinationCidr string) string {
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

# Another route destination for update
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

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

func testAccDefaultRouteTableConfigIpv4VpcEndpointNoRoute(rName string) string {
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

# Another route destination for update
resource "aws_internet_gateway" "test" {
  vpc_id = aws_vpc.test.id

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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.test.id
  }

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccDefaultRouteTableConfigVpcEndpointAssociation(rName, destinationCidr string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

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

resource "aws_vpc_endpoint" "test" {
  vpc_id          = aws_vpc.test.id
  service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = [aws_vpc.test.default_route_table_id]

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    Name = %[1]q
  }

  route {
    cidr_block = %[2]q
    gateway_id = aws_internet_gateway.test.id
  }
}
`, rName, destinationCidr)
}

func testAccDefaultRouteTableConfigTags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccDefaultRouteTableConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccDefaultRouteTableConfigConditionalIpv4Ipv6(rName, destinationCidr, destinationIpv6Cidr string, ipv6Route bool) string {
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

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id

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

func testAccAWSDefaultRouteTableImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["vpc_id"], nil
	}
}
