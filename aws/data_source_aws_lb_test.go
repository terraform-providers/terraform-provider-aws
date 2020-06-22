package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
)

func TestAccDataSourceAWSLB_basic(t *testing.T) {
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	dataSourceName := "data.aws_lb.alb_test_with_arn"
	dataSourceName2 := "data.aws_lb.alb_test_with_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBConfigBasic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "name", lbName),
					resource.TestCheckResourceAttr(dataSourceName, "internal", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "subnets.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.TestName", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(dataSourceName, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "idle_timeout", "30"),
					resource.TestCheckResourceAttrSet(dataSourceName, "vpc_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "dns_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "ip_address_type"),
					resource.TestCheckResourceAttr(dataSourceName2, "name", lbName),
					resource.TestCheckResourceAttr(dataSourceName2, "internal", "true"),
					resource.TestCheckResourceAttr(dataSourceName2, "subnets.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName2, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName2, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName2, "tags.TestName", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(dataSourceName2, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(dataSourceName2, "idle_timeout", "30"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "vpc_id"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "dns_name"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "arn"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "ip_address_type"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLBBackwardsCompatibility(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsalb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	dataSourceName1 := "data.aws_alb.alb_test_with_arn"
	dataSourceName2 := "data.aws_alb.alb_test_with_name"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBConfigBackardsCompatibility(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName1, "name", lbName),
					resource.TestCheckResourceAttr(dataSourceName1, "internal", "true"),
					resource.TestCheckResourceAttr(dataSourceName1, "subnets.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName1, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName1, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName1, "tags.TestName", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(dataSourceName1, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(dataSourceName1, "idle_timeout", "30"),
					resource.TestCheckResourceAttrSet(dataSourceName1, "vpc_id"),
					resource.TestCheckResourceAttrSet(dataSourceName1, "zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName1, "dns_name"),
					resource.TestCheckResourceAttrSet(dataSourceName1, "arn"),
					resource.TestCheckResourceAttr(dataSourceName2, "name", lbName),
					resource.TestCheckResourceAttr(dataSourceName2, "internal", "true"),
					resource.TestCheckResourceAttr(dataSourceName2, "subnets.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName2, "security_groups.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName2, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName2, "tags.TestName", "TestAccAWSALB_basic"),
					resource.TestCheckResourceAttr(dataSourceName2, "enable_deletion_protection", "false"),
					resource.TestCheckResourceAttr(dataSourceName2, "idle_timeout", "30"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "vpc_id"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "zone_id"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "dns_name"),
					resource.TestCheckResourceAttrSet(dataSourceName2, "arn"),
				),
			},
		},
	})
}

func testAccDataSourceAWSLBConfigBasic(lbName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.0.id}", "${aws_subnet.alb_test.1.id}"]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-data-source-basic"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-data-source-basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    TestName = "TestAccAWSALB_basic"
  }
}

data "aws_lb" "alb_test_with_arn" {
  arn = "${aws_lb.alb_test.arn}"
}

data "aws_lb" "alb_test_with_name" {
  name = "${aws_lb.alb_test.name}"
}
`, lbName)
}

func testAccDataSourceAWSLBConfigBackardsCompatibility(albName string) string {
	return fmt.Sprintf(`
resource "aws_alb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = ["${aws_security_group.alb_test.id}"]
  subnets         = ["${aws_subnet.alb_test.0.id}", "${aws_subnet.alb_test.1.id}"]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccAWSALB_basic"
  }
}

variable "subnets" {
  default = ["10.0.1.0/24", "10.0.2.0/24"]
  type    = "list"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-data-source-bc"
  }
}

resource "aws_subnet" "alb_test" {
  count                   = 2
  vpc_id                  = "${aws_vpc.alb_test.id}"
  cidr_block              = "${element(var.subnets, count.index)}"
  map_public_ip_on_launch = true
  availability_zone       = "${element(data.aws_availability_zones.available.names, count.index)}"

  tags = {
    Name = "tf-acc-lb-data-source-bc"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = "${aws_vpc.alb_test.id}"

  ingress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    TestName = "TestAccAWSALB_basic"
  }
}

data "aws_alb" "alb_test_with_arn" {
  arn = "${aws_alb.alb_test.arn}"
}

data "aws_alb" "alb_test_with_name" {
  name = "${aws_alb.alb_test.name}"
}
`, albName)
}
