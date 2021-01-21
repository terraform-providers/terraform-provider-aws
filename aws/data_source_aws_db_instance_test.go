package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSDbInstanceDataSource_basic(t *testing.T) {
	rInt := acctest.RandInt()
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceDataSourceConfig(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "address"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "allocated_storage"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "auto_minor_version_upgrade"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "db_instance_class"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "db_name"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "db_subnet_group"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "endpoint"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "engine"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "hosted_zone_id"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "master_username"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "port"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "multi_az"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "enabled_cloudwatch_logs_exports.0"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "enabled_cloudwatch_logs_exports.1"),
					resource.TestCheckResourceAttrPair("data.aws_db_instance.bar", "resource_id", "aws_db_instance.bar", "resource_id"),
					resource.TestCheckResourceAttrPair("data.aws_db_instance.bar", "tags.%", "aws_db_instance.bar", "tags.%"),
					resource.TestCheckResourceAttrPair("data.aws_db_instance.bar", "tags.Environment", "aws_db_instance.bar", "tags.Environment"),
				),
			},
		},
	})
}

func TestAccAWSDbInstanceDataSource_ec2Classic(t *testing.T) {
	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		ProviderFactories: testAccProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSDBInstanceDataSourceConfig_ec2Classic(rInt),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.aws_db_instance.bar", "db_subnet_group", ""),
				),
			},
		},
	})
}

func testAccAWSDBInstanceDataSourceConfig(rInt int) string {
	return fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = "mariadb"
  preferred_instance_classes = ["db.t3.micro", "db.t2.micro", "db.t3.small"]
}

resource "aws_db_instance" "bar" {
  identifier = "datasource-test-terraform-%d"

  allocated_storage = 10
  engine            = data.aws_rds_orderable_db_instance.test.engine
  instance_class    = data.aws_rds_orderable_db_instance.test.instance_class
  name              = "baz"
  password          = "barbarbarbar"
  username          = "foo"

  backup_retention_period = 0
  skip_final_snapshot     = true

  enabled_cloudwatch_logs_exports = [
    "audit",
    "error",
  ]

  tags = {
    Environment = "test"
  }
}

data "aws_db_instance" "bar" {
  db_instance_identifier = aws_db_instance.bar.identifier
}
`, rInt)
}

func testAccAWSDBInstanceDataSourceConfig_ec2Classic(rInt int) string {
	return composeConfig(
		testAccEc2ClassicRegionProviderConfig(),
		fmt.Sprintf(`
data "aws_rds_orderable_db_instance" "test" {
  engine                     = "mysql"
  engine_version             = "5.6.41"
  preferred_instance_classes = ["db.m3.medium", "db.m3.large", "db.r3.large"]
}

resource "aws_db_instance" "bar" {
  identifier           = "foobarbaz-test-terraform-%[1]d"
  allocated_storage    = 10
  engine               = data.aws_rds_orderable_db_instance.test.engine
  engine_version       = data.aws_rds_orderable_db_instance.test.engine_version
  instance_class       = data.aws_rds_orderable_db_instance.test.instance_class
  name                 = "baz"
  password             = "barbarbarbar"
  username             = "foo"
  publicly_accessible  = true
  security_group_names = ["default"]
  parameter_group_name = "default.mysql5.6"
  skip_final_snapshot  = true
}

data "aws_db_instance" "bar" {
  db_instance_identifier = aws_db_instance.bar.identifier
}
`, rInt))
}
