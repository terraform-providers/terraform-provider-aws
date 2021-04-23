package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestAccVpcEndpointServiceAccepter_crossAccount(t *testing.T) {
	var providers []*schema.Provider
	resourceName := "aws_vpc_endpoint_connection_accepter.test"
	rName := acctest.RandomWithPrefix("tf-acc-test")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccAlternateAccountPreCheck(t)
		},
		ErrorCheck:        testAccErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: testAccProviderFactoriesAlternate(&providers),
		Steps: []resource.TestStep{
			{
				Config: testAccVpcEndpointServiceAccepterConfig_crossAccount(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "state", "available"),
				),
			},
			{
				Config:            testAccVpcEndpointServiceAccepterConfig_crossAccount(rName),
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func testAccVpcEndpointServiceAccepterConfig_crossAccount(rName string) string {
	return testAccAlternateAccountProviderConfig() + fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test1" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  availability_zone = data.aws_availability_zones.available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test2" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.2.0/24"
  availability_zone = data.aws_availability_zones.available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test3" {
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.3.0/24"
  availability_zone = data.aws_availability_zones.available.names[2]

  tags = {
    Name = %[1]q
  }
}

data "aws_caller_identity" "alternate" {
  provider = "awsalternate"
}

resource "aws_vpc" "test_alternate" {
  provider = "awsalternate"

  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

data "aws_availability_zones" "alternate_available" {
  provider = "awsalternate"

  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_subnet" "test_alternate1" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test_alternate.id
  cidr_block        = "10.1.1.0/24"
  availability_zone = data.aws_availability_zones.alternate_available.names[0]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_alternate2" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test_alternate.id
  cidr_block        = "10.1.2.0/24"
  availability_zone = data.aws_availability_zones.alternate_available.names[1]

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test_alternate3" {
  provider = "awsalternate"

  vpc_id            = aws_vpc.test_alternate.id
  cidr_block        = "10.1.3.0/24"
  availability_zone = data.aws_availability_zones.alternate_available.names[2]

  tags = {
    Name = %[1]q
  }
}

data "aws_subnet_ids" "alternate_intersect" {
  provider = "awsalternate"

  vpc_id = aws_vpc.test_alternate.id

  filter {
    name   = "availabilityZone"
    values = aws_vpc_endpoint_service.test.availability_zones
  }
}

resource "aws_lb" "test" {
  name = %[1]q

  subnets = [
    aws_subnet.test1.id,
    aws_subnet.test2.id,
    aws_subnet.test3.id,
  ]

  load_balancer_type         = "network"
  internal                   = true
  idle_timeout               = 60
  enable_deletion_protection = false

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_service" "test" {
  acceptance_required = true

  network_load_balancer_arns = [
    aws_lb.test.id,
  ]

  allowed_principals = [
    "arn:aws:iam::${data.aws_caller_identity.alternate.account_id}:root",
  ]
}

resource "aws_security_group" "test" {
  provider = "awsalternate"

  vpc_id = aws_vpc.test_alternate.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint" "test" {
  provider = "awsalternate"

  vpc_id              = aws_vpc.test_alternate.id
  service_name        = aws_vpc_endpoint_service.test.service_name
  subnet_ids          = data.aws_subnet_ids.alternate_intersect.ids
  vpc_endpoint_type   = "Interface"
  private_dns_enabled = false

  security_group_ids = [
    aws_security_group.test.id,
  ]

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_endpoint_connection_accepter" "test" {
  service_id  = aws_vpc_endpoint_service.test.id
  endpoint_id = aws_vpc_endpoint.test.id
}
`, rName)
}
