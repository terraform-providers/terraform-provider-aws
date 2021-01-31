package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/mq"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccDataSourceAWSMqBroker_basic(t *testing.T) {
	rString := acctest.RandString(7)
	prefix := "tf-acc-test-d-mq-broker"
	brokerName := fmt.Sprintf("%s-%s", prefix, rString)

	dataSourceByIdName := "data.aws_mq_broker.by_id"
	dataSourceByNameName := "data.aws_mq_broker.by_name"
	resourceName := "aws_mq_broker.acctest"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccPartitionHasServicePreCheck(mq.EndpointsID, t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccDataSourceAWSMqBrokerConfig_byId(brokerName, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "broker_name", resourceName, "broker_name"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "auto_minor_version_upgrade", resourceName, "auto_minor_version_upgrade"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "deployment_mode", resourceName, "deployment_mode"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "configuration.#", resourceName, "configuration.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "encryption_options.#", resourceName, "encryption_options.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "encryption_options.0.use_aws_owned_key", resourceName, "encryption_options.0.use_aws_owned_key"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "engine_type", resourceName, "engine_type"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "host_instance_type", resourceName, "host_instance_type"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "instances.#", resourceName, "instances.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "logs.#", resourceName, "logs.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "maintenance_window_start_time.#", resourceName, "maintenance_window_start_time.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "publicly_accessible", resourceName, "publicly_accessible"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "security_groups.#", resourceName, "security_groups.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceByIdName, "user.#", resourceName, "user.#"),
				),
			},
			{
				Config: testAccDataSourceAWSMqBrokerConfig_byName(brokerName, prefix),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceByNameName, "broker_id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceByNameName, "broker_name", resourceName, "broker_name"),
				),
			},
		},
	})
}

func testAccDataSourceAWSMqBrokerConfig_base(brokerName, prefix string) string {
	return fmt.Sprintf(`
variable "prefix" {
  default = "%s"
}

data "aws_availability_zones" "available" {
  state = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

resource "aws_vpc" "acctest" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = var.prefix
  }
}

resource "aws_internet_gateway" "acctest" {
  vpc_id = aws_vpc.acctest.id
}

resource "aws_route_table" "acctest" {
  vpc_id = aws_vpc.acctest.id

  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.acctest.id
  }
}

resource "aws_subnet" "acctest" {
  count             = 2
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]
  vpc_id            = aws_vpc.acctest.id

  tags = {
    Name = var.prefix
  }
}

resource "aws_route_table_association" "acctest" {
  count          = 2
  subnet_id      = aws_subnet.acctest.*.id[count.index]
  route_table_id = aws_route_table.acctest.id
}

resource "aws_security_group" "acctest" {
  count  = 2
  name   = "${var.prefix}-${count.index}"
  vpc_id = aws_vpc.acctest.id
}

resource "aws_mq_configuration" "acctest" {
  name           = var.prefix
  engine_type    = "ActiveMQ"
  engine_version = "5.15.0"

  data = <<DATA
<?xml version="1.0" encoding="UTF-8" standalone="yes"?>
<broker xmlns="http://activemq.apache.org/schema/core">
</broker>
DATA
}

resource "aws_mq_broker" "acctest" {
  auto_minor_version_upgrade = true
  apply_immediately          = true
  broker_name                = "%s"

  configuration {
    id       = aws_mq_configuration.acctest.id
    revision = aws_mq_configuration.acctest.latest_revision
  }

  deployment_mode    = "ACTIVE_STANDBY_MULTI_AZ"
  engine_type        = "ActiveMQ"
  engine_version     = "5.15.0"
  host_instance_type = "mq.t2.micro"

  maintenance_window_start_time {
    day_of_week = "TUESDAY"
    time_of_day = "02:00"
    time_zone   = "CET"
  }

  publicly_accessible = true
  security_groups     = aws_security_group.acctest[*].id
  subnet_ids          = aws_subnet.acctest[*].id

  user {
    username = "Ender"
    password = "AndrewWiggin"
  }

  user {
    username       = "Petra"
    password       = "PetraArkanian"
    console_access = true
    groups         = ["dragon", "salamander", "leopard"]
  }

  depends_on = [aws_internet_gateway.acctest]
}
`, prefix, brokerName)
}

func testAccDataSourceAWSMqBrokerConfig_byId(brokerName, prefix string) string {
	return testAccDataSourceAWSMqBrokerConfig_base(brokerName, prefix) + `
data "aws_mq_broker" "by_id" {
  broker_id = aws_mq_broker.acctest.id
}
`
}

func testAccDataSourceAWSMqBrokerConfig_byName(brokerName, prefix string) string {
	return testAccDataSourceAWSMqBrokerConfig_base(brokerName, prefix) + `
data "aws_mq_broker" "by_name" {
  broker_name = aws_mq_broker.acctest.broker_name
}
`
}
