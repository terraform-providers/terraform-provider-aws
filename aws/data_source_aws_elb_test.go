package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAWSELB_basic(t *testing.T) {
	// Must be less than 32 characters for ELB name
	rName := fmt.Sprintf("TestAccDataSourceAWSELB-%s", acctest.RandStringFromCharSet(5, acctest.CharSetAlphaNum))

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSELBConfigBasic(rName, t.Name()),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "name", rName),
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "cross_zone_load_balancing", "true"),
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "idle_timeout", "30"),
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "internal", "true"),
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "subnets.#", "2"),
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "security_groups.#", "1"),
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "tags.%", "1"),
					resource.TestCheckResourceAttr("data.aws_elb.elb_test", "tags.TestName", t.Name()),
					resource.TestCheckResourceAttrSet("data.aws_elb.elb_test", "dns_name"),
					resource.TestCheckResourceAttrSet("data.aws_elb.elb_test", "zone_id"),
				),
			},
		},
	})
}

func testAccDataSourceAWSELBConfigBasic(rName, testName string) string {
	return fmt.Sprintf(`
resource "aws_elb" "elb_test" {
  name            = "%[1]s"
  internal        = true
  security_groups = ["${aws_security_group.elb_test.id}"]
  subnets         = ["${aws_subnet.elb_test.0.id}", "${aws_subnet.elb_test.1.id}"]

  idle_timeout = 30

  listener {
    instance_port     = 80
    instance_protocol = "http"
    lb_port           = 80
    lb_protocol       = "http"
  }

  tags = {
    TestName = "%[2]s"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {}

resource "aws_vpc" "elb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-elb-data-source"
  }
}

resource "aws_subnet" "elb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.elb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-elb-data-source"
  }
}

resource "aws_security_group" "elb_test" {
  name        = "%[1]s"
  description = "%[2]s"
  vpc_id      = "${aws_vpc.elb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    self        = true
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    TestName = "%[2]s"
  }
}

data "aws_elb" "elb_test" {
	name = "${aws_elb.elb_test.name}"
}`, rName, testName)
}
