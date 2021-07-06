package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/ec2/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAWSVpcEndpointSubnetAssociation_basic(t *testing.T) {
	var vpce ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_subnet_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointSubnetAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointSubnetAssociationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointSubnetAssociationExists(resourceName, &vpce),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointSubnetAssociation_disappears(t *testing.T) {
	var vpce ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_subnet_association.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointSubnetAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointSubnetAssociationConfigBasic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointSubnetAssociationExists(resourceName, &vpce),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpcEndpointSubnetAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpcEndpointSubnetAssociation_multiple(t *testing.T) {
	var vpce ec2.VpcEndpoint
	resourceName0 := "aws_vpc_endpoint_subnet_association.test.0"
	resourceName1 := "aws_vpc_endpoint_subnet_association.test.1"
	resourceName2 := "aws_vpc_endpoint_subnet_association.test.2"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointSubnetAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointSubnetAssociationConfigMultiple(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointSubnetAssociationExists(resourceName0, &vpce),
					testAccCheckVpcEndpointSubnetAssociationExists(resourceName1, &vpce),
					testAccCheckVpcEndpointSubnetAssociationExists(resourceName2, &vpce),
				),
			},
		},
	})
}

func testAccCheckVpcEndpointSubnetAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_subnet_association" {
			continue
		}

		err := finder.VpcEndpointSubnetAssociationExists(conn, rs.Primary.Attributes["vpc_endpoint_id"], rs.Primary.Attributes["subnet_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("VPC Endpoint Subnet Association %s still exists", rs.Primary.ID)
	}

	return nil
}

func testAccCheckVpcEndpointSubnetAssociationExists(n string, vpce *ec2.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		out, err := finder.VpcEndpointByID(conn, rs.Primary.Attributes["vpc_endpoint_id"])

		if err != nil {
			return err
		}

		err = finder.VpcEndpointSubnetAssociationExists(conn, rs.Primary.Attributes["vpc_endpoint_id"], rs.Primary.Attributes["subnet_id"])

		if err != nil {
			return err
		}

		*vpce = *out

		return nil
	}
}

func testAccVpcEndpointSubnetAssociationConfigBase(rName string) string {
	return composeConfig(
		testAccAvailableAZsNoOptInConfig(),
		fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_security_group" "test" {
  vpc_id = aws_vpc.test.id
  name   = "default"
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id              = aws_vpc.test.id
  vpc_endpoint_type   = "Interface"
  service_name        = "com.amazonaws.${data.aws_region.current.name}.ec2"
  security_group_ids  = [data.aws_security_group.test.id]
  private_dns_enabled = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 2, count.index)

  tags = {
    Name = %[1]q
  }
}
`, rName))
}

func testAccVpcEndpointSubnetAssociationConfigBasic(rName string) string {
	return composeConfig(
		testAccVpcEndpointSubnetAssociationConfigBase(rName),
		`
resource "aws_vpc_endpoint_subnet_association" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  subnet_id       = aws_subnet.test[0].id
}
`)
}

func testAccVpcEndpointSubnetAssociationConfigMultiple(rName string) string {
	return composeConfig(
		testAccVpcEndpointSubnetAssociationConfigBase(rName),
		`
resource "aws_vpc_endpoint_subnet_association" "test" {
  count = 3

  vpc_endpoint_id = aws_vpc_endpoint.test.id
  subnet_id       = aws_subnet.test[count.index].id
}
`)
}
