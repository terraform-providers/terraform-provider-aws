package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccAWSDefaultRouteTable_basic(t *testing.T) {
	var routeTable1 ec2.RouteTable
	resourceName := "aws_default_route_table.test"
	vpcResourceName := "aws_vpc.test"

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
				Config: testAccDefaultRouteTableConfigRequired(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable1),
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
	var routeTable1 ec2.RouteTable
	var vpc1 ec2.Vpc
	resourceName := "aws_default_route_table.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfigRequired(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable1),
					testAccCheckVpcExists(vpcResourceName, &vpc1),
					testAccCheckVpcDisappears(&vpc1),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_Route(t *testing.T) {
	var v ec2.RouteTable
	resourceName := "aws_default_route_table.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDefaultRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTableConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &v),
					testAccCheckResourceAttrAccountID(resourceName, "owner_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSDefaultRouteTableImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccDefaultRouteTableConfig_noRouteBlock,
				Check: resource.ComposeTestCheckFunc(
					// The route block from the previous step should still be
					// present, because no blocks means "ignore existing blocks".
					resource.TestCheckResourceAttr(resourceName, "route.#", "1"),
				),
			},
			{
				Config: testAccDefaultRouteTableConfig_routeBlocksExplicitZero,
				Check: resource.ComposeTestCheckFunc(
					// This config uses attribute syntax to set zero routes
					// explicitly, so should remove the one we created before.
					resource.TestCheckResourceAttr(resourceName, "route.#", "0"),
				),
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_swap(t *testing.T) {
	var v ec2.RouteTable
	resourceName := "aws_default_route_table.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDefaultRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTable_change,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(
						resourceName, &v),
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
				Config: testAccDefaultRouteTable_change_mod,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(
						resourceName, &v),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSDefaultRouteTable_Route_TransitGatewayID(t *testing.T) {
	var routeTable1 ec2.RouteTable
	resourceName := "aws_default_route_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDefaultRouteTableConfigRouteTransitGatewayID(),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(resourceName, &routeTable1),
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

func TestAccAWSDefaultRouteTable_vpc_endpoint(t *testing.T) {
	var v ec2.RouteTable
	resourceName := "aws_default_route_table.foo"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckDefaultRouteTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccDefaultRouteTable_vpc_endpoint,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckRouteTableExists(
						resourceName, &v),
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

func testAccDefaultRouteTableConfigDefaultRouteTableId(defaultRouteTableId string) string {
	return fmt.Sprintf(`
resource "aws_default_route_table" "test" {
  default_route_table_id = %[1]q
}
`, defaultRouteTableId)
}

func testAccDefaultRouteTableConfigRequired() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.1.0.0/16"
}

resource "aws_default_route_table" "test" {
  default_route_table_id = aws_vpc.test.default_route_table_id
}
`)
}

const testAccDefaultRouteTableConfig = `
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-default-route-table"
  }
}

resource "aws_default_route_table" "foo" {
  default_route_table_id = "${aws_vpc.foo.default_route_table_id}"

  route {
    cidr_block = "10.0.1.0/32"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }

  tags = {
    Name = "tf-default-route-table-test"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-default-route-table-test"
  }
}`

const testAccDefaultRouteTableConfig_noRouteBlock = `
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-default-route-table"
  }
}

resource "aws_default_route_table" "foo" {
  default_route_table_id = "${aws_vpc.foo.default_route_table_id}"

  tags = {
    Name = "tf-default-route-table-test"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-default-route-table-test"
  }
}`

const testAccDefaultRouteTableConfig_routeBlocksExplicitZero = `
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-default-route-table"
  }
}

resource "aws_default_route_table" "foo" {
  default_route_table_id = "${aws_vpc.foo.default_route_table_id}"

  route = []

  tags = {
    Name = "tf-default-route-table-test"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "tf-default-route-table-test"
  }
}`

const testAccDefaultRouteTable_change = `
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-default-route-table-change"
  }
}

resource "aws_default_route_table" "foo" {
  default_route_table_id = "${aws_vpc.foo.default_route_table_id}"

  route {
    cidr_block = "10.0.1.0/32"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }

  tags = {
    Name = "this was the first main"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "main-igw"
  }
}

# Thing to help testing changes
resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.foo.id}"

  route {
    cidr_block = "10.0.1.0/24"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }

  tags = {
    Name = "other"
  }
}
`

const testAccDefaultRouteTable_change_mod = `
resource "aws_vpc" "foo" {
  cidr_block           = "10.1.0.0/16"
  enable_dns_hostnames = true

  tags = {
    Name = "terraform-testacc-default-route-table-change"
  }
}

resource "aws_default_route_table" "foo" {
  default_route_table_id = "${aws_vpc.foo.default_route_table_id}"

  route {
    cidr_block = "10.0.1.0/32"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }

  tags = {
    Name = "this was the first main"
  }
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"

  tags = {
    Name = "main-igw"
  }
}

# Thing to help testing changes
resource "aws_route_table" "r" {
  vpc_id = "${aws_vpc.foo.id}"

  route {
    cidr_block = "10.0.1.0/24"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }

  tags = {
    Name = "other"
  }
}

resource "aws_main_route_table_association" "a" {
  vpc_id         = "${aws_vpc.foo.id}"
  route_table_id = "${aws_route_table.r.id}"
}
`

func testAccAWSDefaultRouteTableConfigRouteTransitGatewayID() string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-default-route-table-transit-gateway-id"
  }
}

resource "aws_subnet" "test" {
  cidr_block = "10.0.0.0/24"
  vpc_id     = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-default-route-table-transit-gateway-id"
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = "tf-acc-test-ec2-default-route-table-transit-gateway-id"
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = ["${aws_subnet.test.id}"]
  transit_gateway_id = "${aws_ec2_transit_gateway.test.id}"
  vpc_id             = "${aws_vpc.test.id}"

  tags = {
    Name = "tf-acc-test-ec2-default-route-table-transit-gateway-id"
  }
}

resource "aws_default_route_table" "test" {
  default_route_table_id = "${aws_vpc.test.default_route_table_id}"

  route {
    cidr_block         = "0.0.0.0/0"
    transit_gateway_id = "${aws_ec2_transit_gateway_vpc_attachment.test.transit_gateway_id}"
  }
}
`)
}

const testAccDefaultRouteTable_vpc_endpoint = `
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-default-route-table-vpc-endpoint"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = "${aws_vpc.test.id}"

  tags = {
    Name = "terraform-testacc-default-route-table-vpc-endpoint"
  }
}

resource "aws_vpc_endpoint" "s3" {
    vpc_id          = "${aws_vpc.test.id}"
    service_name    = "com.amazonaws.${data.aws_region.current.name}.s3"
    route_table_ids = ["${aws_vpc.test.default_route_table_id}"]

  tags = {
    Name = "terraform-testacc-default-route-table-vpc-endpoint"
  }
}

resource "aws_default_route_table" "foo" {
  default_route_table_id = "${aws_vpc.test.default_route_table_id}"

  tags = {
        Name = "test"
  }

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.igw.id}"
  }
}
`

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

func testAccAWSDefaultRouteTableImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return rs.Primary.Attributes["vpc_id"], nil
	}
}
