package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestAccAWSVpcEndpointRouteTableAssociation_basic(t *testing.T) {
	var vpce ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_route_table_association.test"
	rName := fmt.Sprintf("tf-testacc-vpce-%s", acctest.RandStringFromCharSet(16, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointRouteTableAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointRouteTableAssociationConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointRouteTableAssociationExists(resourceName, &vpce),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSVpcEndpointRouteTableAssociationImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckVpcEndpointRouteTableAssociationDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_vpc_endpoint_route_table_association" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{
			VpcEndpointIds: aws.StringSlice([]string{rs.Primary.Attributes["vpc_endpoint_id"]}),
		})
		if err != nil {
			// Verify the error is what we want
			ec2err, ok := err.(awserr.Error)
			if !ok {
				return err
			}
			if ec2err.Code() != "InvalidVpcEndpointId.NotFound" {
				return err
			}
			return nil
		}

		vpce := resp.VpcEndpoints[0]
		if len(vpce.RouteTableIds) > 0 {
			return fmt.Errorf(
				"VPC endpoint %s has route tables", *vpce.VpcEndpointId)
		}
	}

	return nil
}

func testAccCheckVpcEndpointRouteTableAssociationExists(n string, vpce *ec2.VpcEndpoint) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPC Endpoint Route Table Association ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn
		resp, err := conn.DescribeVpcEndpoints(&ec2.DescribeVpcEndpointsInput{
			VpcEndpointIds: aws.StringSlice([]string{rs.Primary.Attributes["vpc_endpoint_id"]}),
		})
		if err != nil {
			return err
		}
		if len(resp.VpcEndpoints) == 0 {
			return fmt.Errorf("VPC Endpoint not found")
		}

		*vpce = *resp.VpcEndpoints[0]

		if len(vpce.RouteTableIds) == 0 {
			return fmt.Errorf("No VPC Endpoint Route Table Associations")
		}

		for _, rtId := range vpce.RouteTableIds {
			if aws.StringValue(rtId) == rs.Primary.Attributes["route_table_id"] {
				return nil
			}
		}

		return fmt.Errorf("VPC Endpoint Route Table Association not found")
	}
}

func testAccAWSVpcEndpointRouteTableAssociationImportStateIdFunc(n string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return "", fmt.Errorf("Not found: %s", n)
		}

		id := fmt.Sprintf("%s/%s", rs.Primary.Attributes["vpc_endpoint_id"], rs.Primary.Attributes["route_table_id"])
		return id, nil
	}
}

func testAccVpcEndpointRouteTableAssociationConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_region" "current" {}

resource "aws_vpc_endpoint" "test" {
  vpc_id       = aws_vpc.test.id
  service_name = "com.amazonaws.${data.aws_region.current.name}.s3"
}

resource "aws_route_table" "test" {
  vpc_id = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_route_table_association" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  route_table_id  = aws_route_table.test.id
}
`, rName)
}
