package aws

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_vpc_endpoint", &resource.Sweeper{
		Name: "aws_vpc_endpoint",
		F:    testSweepEc2VpcEndpoints,
	})
}

func testSweepEc2VpcEndpoints(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).ec2conn
	input := &ec2.DescribeVpcEndpointsInput{}

	for {
		output, err := conn.DescribeVpcEndpoints(input)

		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping EC2 VPC Endpoint sweep for %s: %s", region, err)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error retrieving EC2 VPC Endpoints: %s", err)
		}

		for _, vpcEndpoint := range output.VpcEndpoints {
			if aws.StringValue(vpcEndpoint.State) != "available" {
				continue
			}

			id := aws.StringValue(vpcEndpoint.VpcEndpointId)

			input := &ec2.DeleteVpcEndpointsInput{
				VpcEndpointIds: []*string{aws.String(id)},
			}

			log.Printf("[INFO] Deleting EC2 VPC Endpoint: %s", id)
			_, err := conn.DeleteVpcEndpoints(input)

			if isAWSErr(err, "InvalidVpcEndpointId.NotFound", "") {
				continue
			}

			if err != nil {
				return fmt.Errorf("error deleting EC2 VPC Endpoint (%s): %s", id, err)
			}

			if err := vpcEndpointWaitUntilDeleted(conn, id, 10*time.Minute); err != nil {
				return fmt.Errorf("error waiting for VPC Endpoint (%s) to delete: %s", id, err)
			}
		}

		if aws.StringValue(output.NextToken) == "" {
			break
		}

		input.NextToken = output.NextToken
	}

	return nil
}

func TestAccAWSVpcEndpoint_importBasic(t *testing.T) {
	resourceName := "aws_vpc_endpoint.s3"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicy,
			},

			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSVpcEndpoint_gatewayBasic(t *testing.T) {
	var endpoint ec2.VpcEndpoint

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpc_endpoint.s3",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_gatewayWithoutRouteTableOrPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.s3", &endpoint),
					testAccCheckVpcEndpointPrefixListAvailable("aws_vpc_endpoint.s3"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "private_dns_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpoint_gatewayWithRouteTableAndPolicy(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	var routeTable ec2.RouteTable

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpc_endpoint.s3",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.s3", &endpoint),
					testAccCheckRouteTableExists("aws_route_table.default", &routeTable),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "route_table_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "private_dns_enabled", "false"),
				),
			},
			{
				Config: testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicyModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.s3", &endpoint),
					testAccCheckRouteTableExists("aws_route_table.default", &routeTable),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "vpc_endpoint_type", "Gateway"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "security_group_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.s3", "private_dns_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpoint_interfaceBasic(t *testing.T) {
	var endpoint ec2.VpcEndpoint

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpc_endpoint.ec2",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_interfaceWithoutSubnet,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.ec2", &endpoint),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "private_dns_enabled", "false"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpoint_interfaceWithSubnetAndSecurityGroup(t *testing.T) {
	var endpoint ec2.VpcEndpoint

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpc_endpoint.ec2",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_interfaceWithSubnet,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.ec2", &endpoint),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "subnet_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "security_group_ids.#", "2"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "private_dns_enabled", "false"),
				),
			},
			{
				Config: testAccVpcEndpointConfig_interfaceWithSubnetModified,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.ec2", &endpoint),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "cidr_blocks.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "route_table_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "subnet_ids.#", "3"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.ec2", "private_dns_enabled", "true"),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpoint_interfaceNonAWSService(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsnlb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	var endpoint ec2.VpcEndpoint

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:      func() { testAccPreCheck(t) },
		IDRefreshName: "aws_vpc_endpoint.foo",
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_interfaceNonAWSService(lbName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.foo", &endpoint),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.foo", "vpc_endpoint_type", "Interface"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.foo", "subnet_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.foo", "network_interface_ids.#", "0"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.foo", "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.foo", "private_dns_enabled", "false"),
					resource.TestCheckResourceAttr("aws_vpc_endpoint.foo", "state", "available"),
				),
			},
		},
	})
}
func TestAccAWSVpcEndpoint_removed(t *testing.T) {
	var endpoint ec2.VpcEndpoint

	// reach out and DELETE the VPC Endpoint outside of Terraform
	testDestroy := func(*terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		input := &ec2.DeleteVpcEndpointsInput{
			VpcEndpointIds: []*string{endpoint.VpcEndpointId},
		}

		_, err := conn.DeleteVpcEndpoints(input)

		return err
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointConfig_gatewayWithoutRouteTableOrPolicy,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists("aws_vpc_endpoint.s3", &endpoint),
					testDestroy,
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckVpcEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint" {
			continue
		}

		// Try to find the VPC
		input := &ec2.DescribeVpcEndpointsInput{
			VpcEndpointIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcEndpoints(input)
		if err != nil {
			// Verify the error is what we want
			if ae, ok := err.(awserr.Error); ok && ae.Code() == "InvalidVpcEndpointId.NotFound" {
				continue
			}
			return err
		}
		if len(resp.VpcEndpoints) > 0 && aws.StringValue(resp.VpcEndpoints[0].State) != "deleted" {
			return fmt.Errorf("VPC Endpoints still exist.")
		}

		return err
	}

	return nil
}

func testAccCheckVpcEndpointExists(n string, endpoint *ec2.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		input := &ec2.DescribeVpcEndpointsInput{
			VpcEndpointIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeVpcEndpoints(input)
		if err != nil {
			return err
		}
		if len(resp.VpcEndpoints) == 0 {
			return fmt.Errorf("VPC Endpoint not found")
		}

		*endpoint = *resp.VpcEndpoints[0]

		return nil
	}
}

func testAccCheckVpcEndpointPrefixListAvailable(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		prefixListID := rs.Primary.Attributes["prefix_list_id"]
		if prefixListID == "" {
			return fmt.Errorf("Prefix list ID not available")
		}
		if !strings.HasPrefix(prefixListID, "pl") {
			return fmt.Errorf("Prefix list ID does not appear to be a valid value: '%s'", prefixListID)
		}

		var (
			cidrBlockSize int
			err           error
		)

		if cidrBlockSize, err = strconv.Atoi(rs.Primary.Attributes["cidr_blocks.#"]); err != nil {
			return err
		}
		if cidrBlockSize < 1 {
			return fmt.Errorf("cidr_blocks seem suspiciously low: %d", cidrBlockSize)
		}

		return nil
	}
}

const testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicy = `
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-endpoint-gw-w-route-table-and-policy"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-vpc-endpoint-gw-w-route-table-and-policy"
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = ["${aws_route_table.default.id}"]
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Sid": "AllowAll",
    "Effect": "Allow",
    "Principal": {"AWS": "*" },
    "Action": "*",
    "Resource": "*"
  }]
}
POLICY
}

resource "aws_route_table" "default" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route_table_association" "main" {
  subnet_id = "${aws_subnet.foo.id}"
  route_table_id = "${aws_route_table.default.id}"
}
`

const testAccVpcEndpointConfig_gatewayWithRouteTableAndPolicyModified = `
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-endpoint-gw-w-route-table-and-policy"
  }
}

resource "aws_subnet" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "10.0.1.0/24"
  tags = {
    Name = "tf-acc-vpc-endpoint-gw-w-route-table-and-policy"
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
  route_table_ids = []
  policy = ""
}

resource "aws_internet_gateway" "gw" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_route_table" "default" {
  vpc_id = "${aws_vpc.foo.id}"

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = "${aws_internet_gateway.gw.id}"
  }
}

resource "aws_route_table_association" "main" {
  subnet_id = "${aws_subnet.foo.id}"
  route_table_id = "${aws_route_table.default.id}"
}
`

const testAccVpcEndpointConfig_gatewayWithoutRouteTableOrPolicy = `
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-endpoint-gw-wout-route-table-or-policy"
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "s3" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}
`

const testAccVpcEndpointConfig_interfaceWithoutSubnet = `
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
  tags = {
    Name = "terraform-testacc-vpc-endpoint-iface-wout-subnet"
  }
}

data "aws_security_group" "default" {
  vpc_id = "${aws_vpc.foo.id}"
  name = "default"
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "ec2" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"
  security_group_ids = ["${data.aws_security_group.default.id}"]
}
`

const testAccVpcEndpointConfig_interfaceWithSubnet = `
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
  enable_dns_support = true
  enable_dns_hostnames = true
  tags = {
    Name = "terraform-testacc-vpc-endpoint-iface-w-subnet"
  }
}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "sn1" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-vpc-endpoint-iface-w-subnet-1"
  }
}

resource "aws_subnet" "sn2" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  tags = {
    Name = "tf-acc-vpc-endpoint-iface-w-subnet-2"
  }
}

resource "aws_subnet" "sn3" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 2)}"
  availability_zone = "${data.aws_availability_zones.available.names[2]}"
  tags = {
    Name = "tf-acc-vpc-endpoint-iface-w-subnet-3"
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_security_group" "sg2" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "ec2" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"
  subnet_ids = ["${aws_subnet.sn1.id}"]
  security_group_ids = ["${aws_security_group.sg1.id}", "${aws_security_group.sg2.id}"]
  private_dns_enabled = false
}
`

const testAccVpcEndpointConfig_interfaceWithSubnetModified = `
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"
  enable_dns_support = true
  enable_dns_hostnames = true
  tags = {
    Name = "terraform-testacc-vpc-endpoint-iface-w-subnet"
  }
}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "sn1" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 0)}"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"
  tags = {
    Name = "tf-acc-vpc-endpoint-iface-w-subnet-1"
  }
}

resource "aws_subnet" "sn2" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 1)}"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"
  tags = {
    Name = "tf-acc-vpc-endpoint-iface-w-subnet-2"
  }
}

resource "aws_subnet" "sn3" {
  vpc_id = "${aws_vpc.foo.id}"
  cidr_block = "${cidrsubnet(aws_vpc.foo.cidr_block, 2, 2)}"
  availability_zone = "${data.aws_availability_zones.available.names[2]}"
  tags = {
    Name = "tf-acc-vpc-endpoint-iface-w-subnet-3"
  }
}

resource "aws_security_group" "sg1" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_security_group" "sg2" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "ec2" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "com.amazonaws.${data.aws_region.current.name}.ec2"
  vpc_endpoint_type = "Interface"
  subnet_ids = ["${aws_subnet.sn1.id}", "${aws_subnet.sn2.id}", "${aws_subnet.sn3.id}"]
  security_group_ids = ["${aws_security_group.sg1.id}"]
  private_dns_enabled = true
}
`

func testAccVpcEndpointConfig_interfaceNonAWSService(lbName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "foo" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-endpoint-iface-non-aws-svc"
  }
}

resource "aws_lb" "nlb_test_1" {
  name = "%s"

  subnets = [
    "${aws_subnet.nlb_test_1.id}",
    "${aws_subnet.nlb_test_2.id}",
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = "testAccVpcEndpointServiceBasicConfig_nlb1"
  }
}

data "aws_region" "current" {}

data "aws_availability_zones" "available" {}

resource "aws_subnet" "nlb_test_1" {
  vpc_id            = "${aws_vpc.foo.id}"
  cidr_block        = "10.0.1.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[0]}"

  tags = {
    Name = "tf-acc-vpc-endpoint-iface-non-aws-svc-1"
  }
}

resource "aws_subnet" "nlb_test_2" {
  vpc_id            = "${aws_vpc.foo.id}"
  cidr_block        = "10.0.2.0/24"
  availability_zone = "${data.aws_availability_zones.available.names[1]}"

  tags = {
    Name = "tf-acc-vpc-endpoint-iface-non-aws-svc-2"
  }
}

resource "aws_vpc_endpoint_service" "foo" {
  acceptance_required = true

  network_load_balancer_arns = [
    "${aws_lb.nlb_test_1.id}",
  ]
}

resource "aws_security_group" "sg1" {
  vpc_id = "${aws_vpc.foo.id}"
}

resource "aws_vpc_endpoint" "foo" {
  vpc_id = "${aws_vpc.foo.id}"
  service_name = "${aws_vpc_endpoint_service.foo.service_name}"
  vpc_endpoint_type = "Interface"
  security_group_ids = ["${aws_security_group.sg1.id}"]
  private_dns_enabled = false
  auto_accept = true
}
  `, lbName)
}
