package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/elb"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSELB_basic(t *testing.T) {
	// Must be less than 32 characters for ELB name
	rName := acctest.RandomWithPrefix("tf-acc-test")
	dataSourceName := "data.aws_elb.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, elb.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSELBConfigBasic(rName, t.Name()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "cross_zone_load_balancing", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "internal", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.TestName", t.Name()),
					resource.TestCheckResourceAttrSet(dataSourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", "aws_elb.test", "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAWSELBConfigBasic(rName, testName string) string {
	return composeConfig(testAccAvailableAZsNoOptInConfig(), fmt.Sprintf(`
resource "aws_elb" "test" {
  name            = %[1]q
  internal        = true
  security_groups = [aws_security_group.test.id]
  subnets         = aws_subnet.test[*].id

  idle_timeout = 30

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    Name     = %[1]q
    TestName = %[2]q
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  count = 2

  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
  vpc_id            = aws_vpc.test.id

  map_public_ip_on_launch = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_security_group" "test" {
  name        = "%[1]s"
  description = "%[2]s"
  vpc_id      = aws_vpc.test.id

  ingress {
    from_port = 0
    to_port   = 0
    protocol  = "-1"
    self      = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name     = %[1]q
    TestName = %[2]q
  }
}

data "aws_elb" "test" {
  name = aws_elb.test.name
}
`, rName, testName))
}
