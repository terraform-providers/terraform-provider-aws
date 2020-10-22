package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSLB_basic(t *testing.T) {
	lbName := fmt.Sprintf("testaccawslb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	dataSourceName := "data.aws_lb.alb_test_with_arn"
	dataSourceName2 := "data.aws_lb.alb_test_with_name"
	resourceName := "aws_lb.alb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBConfigBasic(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.TestName", resourceName, "tags.TestName"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.TestName", resourceName, "tags.TestName"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLB_outpost(t *testing.T) {
	lbName := fmt.Sprintf("testaccawslb-outpost-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	dataSourceName := "data.aws_lb.alb_test_with_arn"
	resourceName := "aws_lb.alb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPreCheckAWSOutpostsOutposts(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBConfigOutpost(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.TestName", resourceName, "tags.TestName"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_mapping.0.outpost_id", resourceName, "subnet_mapping.0.outpost_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAWSLB_BackwardsCompatibility(t *testing.T) {
	lbName := fmt.Sprintf("testaccawsalb-basic-%s", acctest.RandStringFromCharSet(10, acctest.CharSetAlphaNum))
	dataSourceName1 := "data.aws_alb.alb_test_with_arn"
	dataSourceName2 := "data.aws_alb.alb_test_with_name"
	resourceName := "aws_alb.alb_test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSLBConfigBackardsCompatibility(lbName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName1, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "tags.TestName", resourceName, "tags.TestName"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "access_logs.#", resourceName, "access_logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "internal", resourceName, "internal"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnets.#", resourceName, "subnets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "tags.TestName", resourceName, "tags.TestName"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enable_deletion_protection", resourceName, "enable_deletion_protection"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "idle_timeout", resourceName, "idle_timeout"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "zone_id", resourceName, "zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "subnet_mapping.#", resourceName, "subnet_mapping.#"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "drop_invalid_header_fields", resourceName, "drop_invalid_header_fields"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "enable_http2", resourceName, "enable_http2"),
					resource.TestCheckResourceAttrPair(dataSourceName1, "access_logs.#", resourceName, "access_logs.#"),
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
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test[*].id

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
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-data-source-basic"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
  arn = aws_lb.alb_test.arn
}

data "aws_lb" "alb_test_with_name" {
  name = aws_lb.alb_test.name
}
`, lbName)
}

func testAccDataSourceAWSLBConfigOutpost(lbName string) string {
	return fmt.Sprintf(`
data "aws_outposts_outposts" "test" {}

data "aws_outposts_outpost" "test" {
  id = tolist(data.aws_outposts_outposts.test.ids)[0]
}

resource "aws_lb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = [aws_subnet.alb_test.id]

  idle_timeout               = 30
  enable_deletion_protection = false

  tags = {
    TestName = "TestAccAWSALB_outpost"
  }
}

resource "aws_vpc" "alb_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-data-source-outpost"
  }
}

resource "aws_subnet" "alb_test" {
  vpc_id            = aws_vpc.alb_test.id
  cidr_block        = "10.0.0.0/24"
  availability_zone = data.aws_outposts_outpost.test.availability_zone
  outpost_arn       = data.aws_outposts_outpost.test.arn

  tags = {
    Name = "tf-acc-lb-data-source-outpost"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
    TestName = "TestAccAWSALB_outpost"
  }
}

data "aws_lb" "alb_test_with_arn" {
  arn = aws_lb.alb_test.arn
}
`, lbName)
}

func testAccDataSourceAWSLBConfigBackardsCompatibility(albName string) string {
	return fmt.Sprintf(`
resource "aws_alb" "alb_test" {
  name            = "%s"
  internal        = true
  security_groups = [aws_security_group.alb_test.id]
  subnets         = aws_subnet.alb_test[*].id

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
  vpc_id                  = aws_vpc.alb_test.id
  cidr_block              = element(var.subnets, count.index)
  map_public_ip_on_launch = true
  availability_zone       = element(data.aws_availability_zones.available.names, count.index)

  tags = {
    Name = "tf-acc-lb-data-source-bc"
  }
}

resource "aws_security_group" "alb_test" {
  name        = "allow_all_alb_test"
  description = "Used for ALB Testing"
  vpc_id      = aws_vpc.alb_test.id

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
  arn = aws_alb.alb_test.arn
}

data "aws_alb" "alb_test_with_name" {
  name = aws_alb.alb_test.name
}
`, albName)
}
