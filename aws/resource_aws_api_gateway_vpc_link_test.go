package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSAPIGatewayVpcLink_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAPIGatewayVpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists("aws_api_gateway_vpc_link.test"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "name", fmt.Sprintf("tf-apigateway-%s", rName)),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "description", "test"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "target_arns.#", "1"),
				),
			},
			{
				Config: testAccAPIGatewayVpcLinkConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists("aws_api_gateway_vpc_link.test"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "name", fmt.Sprintf("tf-apigateway-update-%s", rName)),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "description", "test update"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "target_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayVpcLink_importBasic(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccCheckAwsAPIGatewayVpcLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_vpc_link" {
			continue
		}

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetVpcLink(input)
		if err != nil {
			if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected VPC Link to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsAPIGatewayVpcLinkExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetVpcLink(input)
		return err
	}
}

func testAccAPIGatewayVpcLinkConfig_basis(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test_a" {
  name = "tf-lb-%s"
  internal = true
  load_balancer_type = "network"
  subnets = ["${aws_subnet.test.id}"]
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
}

data "aws_availability_zones" "test" {}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.10.0.0/21"
  availability_zone = "${data.aws_availability_zones.test.names[0]}"
  tags = {
    Name = "tf-acc-api-gateway-vpc-link"
  }
}
`, rName)
}

func testAccAPIGatewayVpcLinkConfig(rName string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name = "tf-apigateway-%s"
  description = "test"
  target_arns = ["${aws_lb.test_a.arn}"]
}
`, rName)
}

func testAccAPIGatewayVpcLinkConfig_Update(rName string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name = "tf-apigateway-update-%s"
  description = "test update"
  target_arns = ["${aws_lb.test_a.arn}"]
}
`, rName)
}
