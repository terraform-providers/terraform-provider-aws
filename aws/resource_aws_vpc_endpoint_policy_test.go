package aws

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSVpcEndpointPolicy_basic(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("dynamodb:DescribeTable")),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpcEndpointPolicyConfigUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					resource.TestMatchResourceAttr(resourceName, "policy", regexp.MustCompile("dynamodb:*")),
				),
			},
		},
	})
}

func TestAccAWSVpcEndpointPolicy_disappears(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpcEndpointPolicy(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAWSVpcEndpointPolicy_disappears_vpcEndpoint(t *testing.T) {
	var endpoint ec2.VpcEndpoint
	resourceName := "aws_vpc_endpoint_policy.test"
	endpointResourceName := "aws_vpc_endpoint_policy.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckVpcEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointPolicyConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckVpcEndpointExists(resourceName, &endpoint),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsVpcEndpoint(), endpointResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccVpcEndpointPolicyConfig(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_policy" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "ReadOnly",
        "Principal" : "*",
        "Action" : [
          "dynamodb:DescribeTable",
          "dynamodb:ListTables"
        ],
        "Effect" : "Allow",
        "Resource" : "*"
      }
    ]
  })
}
`, rName)
}

func testAccVpcEndpointPolicyConfigUpdated(rName string) string {
	return fmt.Sprintf(`
data "aws_vpc_endpoint_service" "test" {
  service = "dynamodb"
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  service_name = data.aws_vpc_endpoint_service.test.service_name
  vpc_id       = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_policy" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
  policy = jsonencode({
    "Version" : "2012-10-17",
    "Statement" : [
      {
        "Sid" : "AllowAll",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "*"
        },
        "Action" : [
          "dynamodb:*"
        ],
        "Resource" : "*"
      }
    ]
  })
}
`, rName)
}
