package aws

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
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
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "db_instance_class"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "db_name"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "db_subnet_group"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "endpoint"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "engine"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "hosted_zone_id"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "master_username"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "port"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "enabled_cloudwatch_logs_exports.0"),
					resource.TestCheckResourceAttrSet("data.aws_db_instance.bar", "enabled_cloudwatch_logs_exports.1"),
				),
			},
		},
	})
}

func TestAccAWSDbInstanceDataSource_ec2Classic(t *testing.T) {
	oldvar := os.Getenv("AWS_DEFAULT_REGION")
	os.Setenv("AWS_DEFAULT_REGION", "us-east-1")
	defer os.Setenv("AWS_DEFAULT_REGION", oldvar)

	rInt := acctest.RandInt()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t); testAccEC2ClassicPreCheck(t) },
		Providers: testAccProviders,
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
resource "aws_db_instance" "bar" {
	identifier = "datasource-test-terraform-%d"

	allocated_storage = 10
	engine = "MySQL"
	instance_class = "db.t2.micro"
	name = "baz"
	password = "barbarbarbar"
	username = "foo"

	backup_retention_period = 0
	skip_final_snapshot = true

	enabled_cloudwatch_logs_exports = [
		"audit",
		"error",
	]
}

data "aws_db_instance" "bar" {
	db_instance_identifier = "${aws_db_instance.bar.identifier}"
}

`, rInt)
}

func testAccAWSDBInstanceDataSourceConfig_ec2Classic(rInt int) string {
	return fmt.Sprintf(`
%s

data "aws_db_instance" "bar" {
	db_instance_identifier = "${aws_db_instance.bar.identifier}"
}

`, testAccAWSDBInstanceConfigEc2Classic(rInt))
}
