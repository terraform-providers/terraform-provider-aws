package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsApiGatewayVpcLink(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", acctest.RandString(8))
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAwsApiGatewayVpcLinkConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair("data.aws_api_gateway_vpc_link.vpc_link", "name", "aws_api_gateway_vpc_link.vpc_link", "name"),
					resource.TestCheckResourceAttrPair("data.aws_api_gateway_vpc_link.vpc_link", "id", "aws_api_gateway_vpc_link.vpc_link", "id"),
				),
			},
		},
	})
}

func testAccDataSourceAwsApiGatewayVpcLinkConfig(r string) string {
	return fmt.Sprintf(`
	resource "aws_vpc" "apigateway_vpclink_test" {
		cidr_block = "10.0.0.0/16"
	  
		tags = {
		  Name = "terraform-testacc-lb-apigateway-vpc-link"
		}
	  }
	  
	  resource "aws_lb" "apigateway_vpclink_test" {
		name = "%s"
	  
		subnets = [
		  "${aws_subnet.apigateway_vpclink_test_subnet1.id}"
		]
	  
		load_balancer_type               = "network"
		internal                         = true
		idle_timeout                     = 60
		enable_deletion_protection       = false
		enable_cross_zone_load_balancing = false
	  
		tags = {
		  Name = "testAccDataSourceAwsApiGatewayVpcLinkConfig_networkLoadbalancer"
		}
	  }

	  resource "aws_lb" "apigateway_vpclink_test2" {
		name = "%s-wrong"
	  
		subnets = [
		  "${aws_subnet.apigateway_vpclink_test_subnet1.id}"
		]
	  
		load_balancer_type               = "network"
		internal                         = true
		idle_timeout                     = 60
		enable_deletion_protection       = false
		enable_cross_zone_load_balancing = false
	  
		tags = {
		  Name = "testAccDataSourceAwsApiGatewayVpcLinkConfig_networkLoadbalancer"
		}
	  }
	  
	  resource "aws_subnet" "apigateway_vpclink_test_subnet1" {
		vpc_id            = "${aws_vpc.apigateway_vpclink_test.id}"
		cidr_block        = "10.0.1.0/24"

		tags = {
		  Name = "tf-acc-lb-apigateway-vpclink"
		}
	  }

	  resource "aws_api_gateway_vpc_link" "vpc_link" {
		name = "%s"
		target_arns = ["${aws_lb.apigateway_vpclink_test.arn}"]
	  
	  }

	  resource "aws_api_gateway_vpc_link" "vpc_link2" {
		name = "%s-wrong"
		target_arns = ["${aws_lb.apigateway_vpclink_test2.arn}"]
	  
	  }

	  data "aws_api_gateway_vpc_link" "vpc_link" {
		  name = "${aws_api_gateway_vpc_link.vpc_link.name}"
	  }
	
`, r, r, r, r)
}
