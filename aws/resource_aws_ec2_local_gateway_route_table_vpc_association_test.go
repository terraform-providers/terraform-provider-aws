package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAwsEc2LocalGatewayRouteTableVpcAssociation_basic(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	localGatewayRouteTableDataSourceName := "data.aws_ec2_local_gateway_route_table.test"
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"
	vpcResourceName := "aws_vpc.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_id", localGatewayRouteTableDataSourceName, "local_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "local_gateway_route_table_id", localGatewayRouteTableDataSourceName, "id"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", vpcResourceName, "id"),
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

func TestAccAwsEc2LocalGatewayRouteTableVpcAssociation_disappears(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsEc2LocalGatewayRouteTableVpcAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAwsEc2LocalGatewayRouteTableVpcAssociation_Tags(t *testing.T) {
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_ec2_local_gateway_route_table_vpc_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags1(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
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
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags2(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
			{
				Config: testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags1(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
				),
			},
		},
	})
}

func testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationExists(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("%s: missing resource ID", resourceName)
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		association, err := getEc2LocalGatewayRouteTableVpcAssociation(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if association == nil {
			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association (%s) not found", rs.Primary.ID)
		}

		if aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeAssociated {
			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association (%s) not in associated state: %s", rs.Primary.ID, aws.StringValue(association.State))
		}

		return nil
	}
}

func testAccCheckAwsEc2LocalGatewayRouteTableVpcAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ec2_local_gateway_route_table_vpc_association" {
			continue
		}

		association, err := getEc2LocalGatewayRouteTableVpcAssociation(conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		if association != nil && aws.StringValue(association.State) != ec2.RouteTableAssociationStateCodeDisassociated {
			return fmt.Errorf("EC2 Local Gateway Route Table VPC Association (%s) still exists in state: %s", rs.Primary.ID, aws.StringValue(association.State))
		}
	}

	return nil
}

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_ec2_local_gateway_route_table" "test" {
  outpost_arn = tolist(data.aws_outposts_outposts.test.arns)[0]
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfig(rName string) string {
	return composeConfig(
		testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName),
		`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id
}
`)
}

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags1(rName, tagKey1, tagValue1 string) string {
	return composeConfig(
		testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
  }
}
`, tagKey1, tagValue1))
}

func testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigTags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return composeConfig(
		testAccAwsEc2LocalGatewayRouteTableVpcAssociationConfigBase(rName),
		fmt.Sprintf(`
resource "aws_ec2_local_gateway_route_table_vpc_association" "test" {
  local_gateway_route_table_id = data.aws_ec2_local_gateway_route_table.test.id
  vpc_id                       = aws_vpc.test.id

  tags = {
    %[1]q = %[2]q
    %[3]q = %[4]q
  }
}
`, tagKey1, tagValue1, tagKey2, tagValue2))
}
